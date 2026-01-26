package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

const containerActionTimeoutSec = 10

type containersResponse struct {
	Scope      string                        `json:"scope"`
	Containers []dockerwatcher.ContainerInfo `json:"containers"`
	Error      string                        `json:"error,omitempty"`
}

type containerActionRequest struct {
	Scope       string `json:"scope"`
	ContainerID string `json:"container_id"`
	Action      string `json:"action"`
}

type containerActionPayload struct {
	ContainerID string `json:"container_id"`
	Action      string `json:"action"`
}

type containerActionResponse struct {
	Container dockerwatcher.ContainerInfo `json:"container"`
}

func (s *Server) handleContainers(w http.ResponseWriter, r *http.Request) {
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
		result, err := s.listLocalContainers(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, containersResponse{Scope: "local:" + name, Error: err.Error(), Containers: []dockerwatcher.ContainerInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, containersResponse{Scope: "local:" + name, Containers: result})
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
		result, err := s.listLocalContainers(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, containersResponse{Scope: scope, Error: err.Error(), Containers: []dockerwatcher.ContainerInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, containersResponse{Scope: scope, Containers: result})
		return
	}
	result, err := s.listRemoteContainers(cfg, name)
	if err != nil {
		writeJSON(w, http.StatusOK, containersResponse{Scope: scope, Error: err.Error(), Containers: []dockerwatcher.ContainerInfo{}})
		return
	}
	writeJSON(w, http.StatusOK, containersResponse{Scope: scope, Containers: result})
}

func (s *Server) handleContainerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload containerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Action = strings.TrimSpace(strings.ToLower(payload.Action))
	payload.ContainerID = strings.TrimSpace(payload.ContainerID)
	if payload.ContainerID == "" {
		writeError(w, http.StatusBadRequest, errors.New("container_id is required"))
		return
	}
	if payload.Action == "" {
		writeError(w, http.StatusBadRequest, errors.New("action is required"))
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
		updated, err := s.applyContainerAction(cfg, name, payload.ContainerID, payload.Action)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, containerActionResponse{Container: updated})
		return
	}
	scope := strings.TrimSpace(payload.Scope)
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
		updated, err := s.applyContainerAction(cfg, name, payload.ContainerID, payload.Action)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, containerActionResponse{Container: updated})
		return
	}
	updated, err := s.applyRemoteContainerAction(cfg, name, payload.ContainerID, payload.Action)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, containerActionResponse{Container: updated})
}

func (s *Server) listLocalContainers(cfg config.Config, name string) ([]dockerwatcher.ContainerInfo, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return nil, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	return watcher.Containers(ctx)
}

func (s *Server) listRemoteContainers(cfg config.Config, name string) ([]dockerwatcher.ContainerInfo, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/containers"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload containersResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, errors.New(payload.Error)
	}
	if payload.Containers == nil {
		return []dockerwatcher.ContainerInfo{}, nil
	}
	return payload.Containers, nil
}

func (s *Server) applyContainerAction(cfg config.Config, name, containerID, action string) (dockerwatcher.ContainerInfo, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return dockerwatcher.ContainerInfo{}, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	defer watcher.Close()
	if err := applyContainerActionWithWatcher(ctx, watcher, containerID, action); err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	info, err := watcher.Containers(ctx)
	if err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	for _, item := range info {
		if item.ID == containerID {
			return item, nil
		}
	}
	return dockerwatcher.ContainerInfo{}, errors.New("container not found after update")
}

func (s *Server) applyRemoteContainerAction(cfg config.Config, name, containerID, action string) (dockerwatcher.ContainerInfo, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return dockerwatcher.ContainerInfo{}, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return dockerwatcher.ContainerInfo{}, errors.New("remote url missing")
	}
	body, err := json.Marshal(containerActionPayload{
		ContainerID: containerID,
		Action:      action,
	})
	if err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/containers/action"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return dockerwatcher.ContainerInfo{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload containerActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return dockerwatcher.ContainerInfo{}, err
	}
	return payload.Container, nil
}

func applyContainerActionWithWatcher(ctx context.Context, watcher *dockerwatcher.Watcher, containerID, action string) error {
	switch action {
	case "start":
		return watcher.StartContainer(ctx, containerID)
	case "stop":
		return watcher.StopContainer(ctx, containerID, containerActionTimeoutSec)
	case "restart":
		return watcher.RestartContainer(ctx, containerID, containerActionTimeoutSec)
	case "pause":
		return watcher.PauseContainer(ctx, containerID)
	case "unpause":
		return watcher.UnpauseContainer(ctx, containerID)
	case "kill":
		return watcher.KillContainer(ctx, containerID)
	default:
		return errors.New("unknown action")
	}
}

func parseScope(scope string) (string, string, error) {
	parts := strings.SplitN(scope, ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid scope")
	}
	serverType := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])
	if name == "" {
		return "", "", errors.New("scope name required")
	}
	if serverType != "local" && serverType != "remote" {
		return "", "", errors.New("scope must be local or remote")
	}
	return serverType, name, nil
}
