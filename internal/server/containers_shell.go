package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"

	"github.com/docker/docker/api/types"
	"github.com/gorilla/websocket"
)

const shellExecTimeout = 15 * time.Second

var shellUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type shellControlMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type shellErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (s *Server) handleContainerShell(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg := s.store.Get()
	if !s.agentMode && (!cfg.ExperimentalFeatures.Containers || !cfg.ExperimentalFeatures.ContainerShell) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	containerID := strings.TrimSpace(r.URL.Query().Get("container_id"))
	if containerID == "" {
		writeError(w, http.StatusBadRequest, errors.New("container_id is required"))
		return
	}
	if s.agentMode {
		name := strings.TrimSpace(r.URL.Query().Get("server"))
		if name == "" && len(cfg.LocalServers) == 1 {
			name = cfg.LocalServers[0].Name
		}
		if name == "" {
			writeError(w, http.StatusBadRequest, errors.New("server name required"))
			return
		}
		s.handleLocalContainerShell(w, r, cfg, name, containerID)
		return
	}
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	if scope == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope is required"))
		return
	}
	serverType, name, err := parseScope(scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "local" {
		s.handleLocalContainerShell(w, r, cfg, name, containerID)
		return
	}
	s.handleRemoteContainerShell(w, r, cfg, name, containerID)
}

func (s *Server) handleLocalContainerShell(w http.ResponseWriter, r *http.Request, cfg config.Config, name, containerID string) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		writeError(w, http.StatusBadRequest, errors.New("local server not found"))
		return
	}
	conn, err := shellUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("shell: websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		s.addLog("error", "shell: docker client init failed: "+err.Error())
		log.Printf("shell: docker client init failed: %v", err)
		sendShellError(conn, err.Error())
		return
	}
	defer watcher.Close()

	log.Printf("shell: connected local server=%s container=%s", name, containerID)
	execID, hijacked, err := openShellExec(ctx, watcher, containerID)
	if err != nil {
		s.addLog("error", "shell: exec failed: "+err.Error())
		log.Printf("shell: exec failed: %v", err)
		sendShellError(conn, err.Error())
		return
	}
	defer hijacked.Close()
	log.Printf("shell: exec started id=%s", execID)

	done := make(chan struct{})
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := hijacked.Reader.Read(buf)
			if n > 0 {
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					log.Printf("shell: ws write failed: %v", writeErr)
					break
				}
			}
			if readErr != nil {
				log.Printf("shell: exec stream closed: %v", readErr)
				break
			}
		}
		inspectCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		inspect, inspectErr := watcher.InspectExec(inspectCtx, execID)
		cancel()
		if inspectErr == nil && !inspect.Running && inspect.ExitCode != 0 {
			s.addLog("warn", "shell: exec exited with code "+fmt.Sprint(inspect.ExitCode))
			log.Printf("shell: exec exited with code %d", inspect.ExitCode)
		}
		close(done)
	}()

	for {
		msgType, payload, readErr := conn.ReadMessage()
		if readErr != nil {
			log.Printf("shell: ws read closed: %v", readErr)
			break
		}
		if msgType == websocket.TextMessage {
			if handled := handleShellControlMessage(watcher, execID, payload); handled {
				continue
			}
		}
		if msgType == websocket.BinaryMessage || msgType == websocket.TextMessage {
			if _, err := hijacked.Conn.Write(payload); err != nil {
				log.Printf("shell: exec stdin write failed: %v", err)
				break
			}
		}
	}
	hijacked.Close()
	<-done
}

func (s *Server) handleRemoteContainerShell(w http.ResponseWriter, r *http.Request, cfg config.Config, name, containerID string) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		writeError(w, http.StatusBadRequest, errors.New("remote server not found"))
		return
	}
	if remote.URL == "" {
		writeError(w, http.StatusBadRequest, errors.New("remote url missing"))
		return
	}
	clientConn, err := shellUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("shell: websocket upgrade failed (remote): %v", err)
		return
	}
	defer clientConn.Close()

	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/containers/shell"
	wsURL, err := url.Parse(endpoint)
	if err != nil {
		sendShellError(clientConn, err.Error())
		return
	}
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	query := wsURL.Query()
	query.Set("container_id", containerID)
	wsURL.RawQuery = query.Encode()

	headers := http.Header{}
	if remote.Token != "" {
		headers.Set("Authorization", "Bearer "+remote.Token)
	}
	dialer := websocket.Dialer{}
	remoteConn, _, err := dialer.Dial(wsURL.String(), headers)
	if err != nil {
		log.Printf("shell: remote websocket dial failed: %v", err)
		sendShellError(clientConn, err.Error())
		return
	}
	defer remoteConn.Close()

	proxyWebSockets(clientConn, remoteConn)
}

func openShellExec(ctx context.Context, watcher *dockerwatcher.Watcher, containerID string) (string, types.HijackedResponse, error) {
	commands := [][]string{{"sh"}, {"/bin/sh"}, {"/bin/bash"}, {"/bin/ash"}}
	var lastErr error
	for _, cmd := range commands {
		execID, hijacked, err := watcher.ExecShell(ctx, containerID, cmd)
		if err == nil {
			inspectCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			inspect, inspectErr := watcher.InspectExec(inspectCtx, execID)
			cancel()
			if inspectErr == nil && !inspect.Running && inspect.ExitCode != 0 {
				hijacked.Close()
				lastErr = errors.New("shell not available")
				continue
			}
			return execID, hijacked, nil
		}
		lastErr = err
		if !isShellNotFound(err) {
			break
		}
	}
	if lastErr == nil {
		lastErr = errors.New("shell unavailable")
	}
	return "", types.HijackedResponse{}, lastErr
}

func isShellNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "executable file not found") || strings.Contains(msg, "no such file") || strings.Contains(msg, "file not found")
}

func handleShellControlMessage(watcher *dockerwatcher.Watcher, execID string, payload []byte) bool {
	var msg shellControlMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return false
	}
	if msg.Type != "resize" && msg.Type != "init" {
		return false
	}
	if msg.Cols <= 0 || msg.Rows <= 0 {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = watcher.ResizeExec(ctx, execID, uint(msg.Cols), uint(msg.Rows))
	return true
}

func sendShellError(conn *websocket.Conn, message string) {
	if conn == nil {
		return
	}
	payload, _ := json.Marshal(shellErrorMessage{Type: "error", Message: message})
	_ = conn.WriteMessage(websocket.TextMessage, payload)
}
