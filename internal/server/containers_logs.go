package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/websocket"
)

const (
	defaultLogsTail = "100"
	maxLogsTail     = 5000
)

type logsErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (s *Server) handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	release, ok := s.reserveContainerStream(w)
	if !ok {
		return
	}
	defer release()
	cfg := s.store.Get()
	if !s.agentMode && (!cfg.ExperimentalFeatures.Containers || !cfg.ExperimentalFeatures.ContainerLogs) {
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
		s.handleLocalContainerLogs(w, r, cfg, name, containerID)
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
		s.handleLocalContainerLogs(w, r, cfg, name, containerID)
		return
	}
	s.handleRemoteContainerLogs(w, r, cfg, name, containerID)
}

func (s *Server) handleLocalContainerLogs(w http.ResponseWriter, r *http.Request, cfg config.Config, name, containerID string) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		writeError(w, http.StatusBadRequest, errors.New("local server not found"))
		return
	}
	conn, err := shellUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("logs: websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	defer s.protectContainerStream(r, conn, clientStreamMessageLimit)()

	follow := parseBoolQuery(r, "follow", true)
	timestamps := parseBoolQuery(r, "timestamps", false)
	tail := parseTailQuery(r, "tail")
	since := strings.TrimSpace(r.URL.Query().Get("since"))

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		s.addLog("error", "logs: docker client init failed: "+err.Error())
		log.Printf("logs: docker client init failed: %v", err)
		sendLogsError(conn, err.Error())
		return
	}
	defer watcher.Close()

	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: timestamps,
		Tail:       tail,
	}
	if since != "" {
		opts.Since = since
	}

	reader, tty, err := watcher.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		s.addLog("error", "logs: container logs failed: "+err.Error())
		log.Printf("logs: container logs failed: %v", err)
		sendLogsError(conn, err.Error())
		return
	}
	defer reader.Close()

	log.Printf("logs: connected local server=%s container=%s", name, containerID)

	go func() {
		for {
			if _, _, readErr := conn.ReadMessage(); readErr != nil {
				cancel()
				break
			}
		}
	}()

	writer := &wsTextWriter{conn: conn}
	if tty {
		_, _ = io.Copy(writer, reader)
	} else {
		_, _ = stdcopy.StdCopy(writer, writer, reader)
	}
}

func (s *Server) handleRemoteContainerLogs(w http.ResponseWriter, r *http.Request, cfg config.Config, name, containerID string) {
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
		log.Printf("logs: websocket upgrade failed (remote): %v", err)
		return
	}
	defer clientConn.Close()
	defer s.protectContainerStream(r, clientConn, clientStreamMessageLimit)()

	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/containers/logs"
	wsURL, err := url.Parse(endpoint)
	if err != nil {
		sendLogsError(clientConn, err.Error())
		return
	}
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	query := wsURL.Query()
	query.Set("container_id", containerID)
	if follow := strings.TrimSpace(r.URL.Query().Get("follow")); follow != "" {
		query.Set("follow", follow)
	}
	if tail := strings.TrimSpace(r.URL.Query().Get("tail")); tail != "" {
		query.Set("tail", tail)
	}
	if since := strings.TrimSpace(r.URL.Query().Get("since")); since != "" {
		query.Set("since", since)
	}
	if timestamps := strings.TrimSpace(r.URL.Query().Get("timestamps")); timestamps != "" {
		query.Set("timestamps", timestamps)
	}
	wsURL.RawQuery = query.Encode()

	headers := http.Header{}
	if remote.Token != "" {
		headers.Set("Authorization", "Bearer "+remote.Token)
	}
	dialer := websocket.Dialer{HandshakeTimeout: 30 * time.Second}
	remoteConn, _, err := dialer.Dial(wsURL.String(), headers)
	if err != nil {
		log.Printf("logs: remote websocket dial failed: %v", err)
		sendLogsError(clientConn, err.Error())
		return
	}
	defer remoteConn.Close()
	defer s.protectContainerStream(r, remoteConn, agentStreamMessageLimit)()

	proxyWebSockets(clientConn, remoteConn)
}

type wsTextWriter struct {
	conn *websocket.Conn
}

func (w *wsTextWriter) Write(p []byte) (int, error) {
	if len(p) == 0 || w.conn == nil {
		return len(p), nil
	}
	// Chunk output independently of Docker's log frame size. Old controllers can
	// consume these ordinary text messages without a coordinated agent upgrade.
	written := 0
	for len(p) > 0 {
		size := min(len(p), 32*1024)
		if size < len(p) {
			for size > 0 && !utf8.RuneStart(p[size]) {
				size--
			}
			if size == 0 {
				return written, errors.New("invalid UTF-8 log output")
			}
		}
		if err := writeStreamMessage(w.conn, websocket.TextMessage, p[:size]); err != nil {
			return written, err
		}
		written += size
		p = p[size:]
	}
	return written, nil
}

func parseBoolQuery(r *http.Request, key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(r.URL.Query().Get(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseTailQuery(r *http.Request, key string) string {
	value := strings.TrimSpace(strings.ToLower(r.URL.Query().Get(key)))
	if value == "" {
		return defaultLogsTail
	}
	if value == "all" {
		return "all"
	}
	num, err := strconv.Atoi(value)
	if err != nil || num < 0 {
		return defaultLogsTail
	}
	if num > maxLogsTail {
		num = maxLogsTail
	}
	return fmt.Sprintf("%d", num)
}

func sendLogsError(conn *websocket.Conn, message string) {
	if conn == nil {
		return
	}
	payload, _ := json.Marshal(logsErrorMessage{Type: "error", Message: message})
	_ = writeStreamMessage(conn, websocket.TextMessage, payload)
}
