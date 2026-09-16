package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

type serverHealthResponse struct {
	Scope     string                             `json:"scope"`
	CheckedAt time.Time                          `json:"checked_at"`
	Health    dockerwatcher.DockerHealthSnapshot `json:"health"`
}

func (s *Server) handleServerHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg := s.store.Get()
	if s.agentMode {
		name := strings.TrimSpace(r.URL.Query().Get("server"))
		if name == "" && len(cfg.LocalServers) == 1 {
			name = cfg.LocalServers[0].Name
		}
		if name == "" {
			writeError(w, http.StatusBadRequest, errors.New("server name required"))
			return
		}
		health, err := collectLocalServerHealth(r.Context(), cfg, name)
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeJSON(w, http.StatusOK, serverHealthResponse{Scope: "local:" + name, CheckedAt: time.Now(), Health: health})
		return
	}

	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	serverType, name, err := parseScope(scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var health dockerwatcher.DockerHealthSnapshot
	if serverType == "local" {
		health, err = collectLocalServerHealth(r.Context(), cfg, name)
	} else {
		health, err = collectRemoteServerHealth(r.Context(), cfg, name)
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, serverHealthResponse{Scope: scope, CheckedAt: time.Now(), Health: health})
}

func collectLocalServerHealth(parent context.Context, cfg config.Config, name string) (dockerwatcher.DockerHealthSnapshot, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return dockerwatcher.DockerHealthSnapshot{}, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return dockerwatcher.DockerHealthSnapshot{}, err
	}
	defer watcher.Close()
	return watcher.DockerHealth(ctx)
}

func collectRemoteServerHealth(parent context.Context, cfg config.Config, name string) (dockerwatcher.DockerHealthSnapshot, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return dockerwatcher.DockerHealthSnapshot{}, errors.New("remote server not found")
	}
	if strings.TrimSpace(remote.URL) == "" {
		return dockerwatcher.DockerHealthSnapshot{}, errors.New("remote url missing")
	}
	ctx, cancel := context.WithTimeout(parent, 35*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(remote.URL, "/")+"/api/servers/health", nil)
	if err != nil {
		return dockerwatcher.DockerHealthSnapshot{}, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	resp, err := (&http.Client{Timeout: 35 * time.Second}).Do(req)
	if err != nil {
		return dockerwatcher.DockerHealthSnapshot{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return dockerwatcher.DockerHealthSnapshot{}, errors.New(resp.Status)
	}
	var payload serverHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return dockerwatcher.DockerHealthSnapshot{}, err
	}
	return payload.Health, nil
}
