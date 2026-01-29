package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

const maxContainersResources = 50

type containersResourcesRequest struct {
	Scope        string   `json:"scope"`
	ContainerIDs []string `json:"container_ids"`
}

type containersResourcesPayload struct {
	ContainerIDs []string `json:"container_ids"`
}

type containersResourcesResponse struct {
	Scope     string                           `json:"scope"`
	Resources []dockerwatcher.ContainerResource `json:"resources"`
	Error     string                           `json:"error,omitempty"`
}

func (s *Server) handleContainersResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload containersResourcesRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	containerIDs := normalizeContainerIDs(payload.ContainerIDs)
	if len(containerIDs) > maxContainersResources {
		writeError(w, http.StatusBadRequest, errors.New("too many containers requested"))
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
		if len(containerIDs) == 0 {
			writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: "local:" + name, Resources: []dockerwatcher.ContainerResource{}})
			return
		}
		result, err := s.listLocalContainersResources(cfg, name, containerIDs)
		if err != nil {
			writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: "local:" + name, Error: err.Error(), Resources: []dockerwatcher.ContainerResource{}})
			return
		}
		writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: "local:" + name, Resources: result})
		return
	}
	scope := strings.TrimSpace(payload.Scope)
	if scope == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope is required"))
		return
	}
	if len(containerIDs) == 0 {
		writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: scope, Resources: []dockerwatcher.ContainerResource{}})
		return
	}
	serverType, name, err := parseScope(scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "local" {
		result, err := s.listLocalContainersResources(cfg, name, containerIDs)
		if err != nil {
			writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: scope, Error: err.Error(), Resources: []dockerwatcher.ContainerResource{}})
			return
		}
		writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: scope, Resources: result})
		return
	}
	result, err := s.listRemoteContainersResources(cfg, name, containerIDs)
	if err != nil {
		writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: scope, Error: err.Error(), Resources: []dockerwatcher.ContainerResource{}})
		return
	}
	writeJSON(w, http.StatusOK, containersResourcesResponse{Scope: scope, Resources: result})
}

func normalizeContainerIDs(ids []string) []string {
	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

func (s *Server) listLocalContainersResources(cfg config.Config, name string, ids []string) ([]dockerwatcher.ContainerResource, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return nil, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	return watcher.ContainerResources(ctx, ids)
}

func (s *Server) listRemoteContainersResources(cfg config.Config, name string, ids []string) ([]dockerwatcher.ContainerResource, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	body, err := json.Marshal(containersResourcesPayload{ContainerIDs: ids})
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/containers/resources"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	var payload containersResourcesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, errors.New(payload.Error)
	}
	if payload.Resources == nil {
		return []dockerwatcher.ContainerResource{}, nil
	}
	return payload.Resources, nil
}
