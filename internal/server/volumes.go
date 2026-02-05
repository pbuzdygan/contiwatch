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

type volumesResponse struct {
	Scope   string                     `json:"scope"`
	Volumes []dockerwatcher.VolumeInfo `json:"volumes"`
	Error   string                     `json:"error,omitempty"`
}

type volumeRemoveRequest struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
}

type volumeRemoveResponse struct {
	Name      string                       `json:"name"`
	RemovedAt string                       `json:"removed_at,omitempty"`
	BlockedBy []dockerwatcher.VolumeUsedBy `json:"blocked_by,omitempty"`
	Error     string                       `json:"error,omitempty"`
}

type volumePruneRequest struct {
	Scope string `json:"scope"`
}

type volumePruneResponse struct {
	Deleted        []string `json:"deleted"`
	ReclaimedBytes uint64   `json:"reclaimed_bytes"`
}

func (s *Server) handleVolumes(w http.ResponseWriter, r *http.Request) {
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
		result, err := s.listLocalVolumes(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, volumesResponse{Scope: "local:" + name, Error: err.Error(), Volumes: []dockerwatcher.VolumeInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, volumesResponse{Scope: "local:" + name, Volumes: result})
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
		result, err := s.listLocalVolumes(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, volumesResponse{Scope: scope, Error: err.Error(), Volumes: []dockerwatcher.VolumeInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, volumesResponse{Scope: scope, Volumes: result})
		return
	}
	result, err := s.listRemoteVolumes(cfg, name)
	if err != nil {
		writeJSON(w, http.StatusOK, volumesResponse{Scope: scope, Error: err.Error(), Volumes: []dockerwatcher.VolumeInfo{}})
		return
	}
	writeJSON(w, http.StatusOK, volumesResponse{Scope: scope, Volumes: result})
}

func (s *Server) handleVolumesRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload volumeRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
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
		blockedBy, err := s.removeLocalVolume(cfg, name, payload.Name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if len(blockedBy) > 0 {
			writeJSON(w, http.StatusConflict, volumeRemoveResponse{Name: payload.Name, BlockedBy: blockedBy, Error: "volume is in use"})
			return
		}
		writeJSON(w, http.StatusOK, volumeRemoveResponse{Name: payload.Name, RemovedAt: time.Now().Format(time.RFC3339)})
		return
	}
	payload.Scope = strings.TrimSpace(payload.Scope)
	if payload.Scope == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope is required"))
		return
	}
	serverType, name, err := parseScope(payload.Scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "local" {
		blockedBy, err := s.removeLocalVolume(cfg, name, payload.Name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if len(blockedBy) > 0 {
			writeJSON(w, http.StatusConflict, volumeRemoveResponse{Name: payload.Name, BlockedBy: blockedBy, Error: "volume is in use"})
			return
		}
		writeJSON(w, http.StatusOK, volumeRemoveResponse{Name: payload.Name, RemovedAt: time.Now().Format(time.RFC3339)})
		return
	}
	blockedBy, err := s.removeRemoteVolume(cfg, name, payload.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(blockedBy) > 0 {
		writeJSON(w, http.StatusConflict, volumeRemoveResponse{Name: payload.Name, BlockedBy: blockedBy, Error: "volume is in use"})
		return
	}
	writeJSON(w, http.StatusOK, volumeRemoveResponse{Name: payload.Name, RemovedAt: time.Now().Format(time.RFC3339)})
}

func (s *Server) handleVolumesPrune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload volumePruneRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
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
		deleted, reclaimed, err := s.pruneLocalVolumes(cfg, name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, volumePruneResponse{Deleted: deleted, ReclaimedBytes: reclaimed})
		return
	}
	payload.Scope = strings.TrimSpace(payload.Scope)
	if payload.Scope == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope is required"))
		return
	}
	serverType, name, err := parseScope(payload.Scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "local" {
		deleted, reclaimed, err := s.pruneLocalVolumes(cfg, name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, volumePruneResponse{Deleted: deleted, ReclaimedBytes: reclaimed})
		return
	}
	deleted, reclaimed, err := s.pruneRemoteVolumes(cfg, name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, volumePruneResponse{Deleted: deleted, ReclaimedBytes: reclaimed})
}

func (s *Server) listLocalVolumes(cfg config.Config, name string) ([]dockerwatcher.VolumeInfo, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return nil, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	return watcher.ListVolumes(ctx)
}

func (s *Server) listRemoteVolumes(cfg config.Config, name string) ([]dockerwatcher.VolumeInfo, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/volumes"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload volumesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, errors.New(payload.Error)
	}
	if payload.Volumes == nil {
		return []dockerwatcher.VolumeInfo{}, nil
	}
	return payload.Volumes, nil
}

func (s *Server) removeLocalVolume(cfg config.Config, name, volumeName string) ([]dockerwatcher.VolumeUsedBy, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return nil, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	usage, err := watcher.VolumeUsage(ctx, volumeName)
	if err != nil {
		return nil, err
	}
	if len(usage) > 0 {
		return usage, nil
	}
	if err := watcher.RemoveVolume(ctx, volumeName); err != nil {
		return nil, err
	}
	return []dockerwatcher.VolumeUsedBy{}, nil
}

func (s *Server) removeRemoteVolume(cfg config.Config, name, volumeName string) ([]dockerwatcher.VolumeUsedBy, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	body, err := json.Marshal(volumeRemoveRequest{
		Name: volumeName,
	})
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/volumes/remove"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		var payload volumeRemoveResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			return payload.BlockedBy, nil
		}
		return nil, errors.New("volume is in use")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return []dockerwatcher.VolumeUsedBy{}, nil
}

func (s *Server) pruneLocalVolumes(cfg config.Config, name string) ([]string, uint64, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return nil, 0, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return nil, 0, err
	}
	defer watcher.Close()
	return watcher.PruneVolumes(ctx)
}

func (s *Server) pruneRemoteVolumes(cfg config.Config, name string) ([]string, uint64, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, 0, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, 0, errors.New("remote url missing")
	}
	body, err := json.Marshal(volumePruneRequest{})
	if err != nil {
		return nil, 0, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/volumes/prune"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload volumePruneResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, 0, err
	}
	return payload.Deleted, payload.ReclaimedBytes, nil
}
