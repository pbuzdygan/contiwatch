package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

type networksResponse struct {
	Scope    string                      `json:"scope"`
	Networks []dockerwatcher.NetworkInfo `json:"networks"`
	Error    string                      `json:"error,omitempty"`
}

type networkDetailsResponse struct {
	Scope   string                       `json:"scope"`
	Network dockerwatcher.NetworkDetails `json:"network"`
	Error   string                       `json:"error,omitempty"`
}

type networkRemoveRequest struct {
	Scope     string `json:"scope"`
	NetworkID string `json:"network_id"`
}

type networkRemoveResponse struct {
	NetworkID string                        `json:"network_id"`
	RemovedAt string                        `json:"removed_at,omitempty"`
	BlockedBy []dockerwatcher.NetworkUsedBy `json:"blocked_by,omitempty"`
	Error     string                        `json:"error,omitempty"`
}

type networkPruneRequest struct {
	Scope string `json:"scope"`
}

type networkPruneResponse struct {
	Deleted []string `json:"deleted"`
}

type networkConnectRequest struct {
	Scope       string `json:"scope"`
	NetworkID   string `json:"network_id"`
	ContainerID string `json:"container_id"`
}

type networkDisconnectRequest struct {
	Scope       string `json:"scope"`
	NetworkID   string `json:"network_id"`
	ContainerID string `json:"container_id"`
}

func (s *Server) handleNetworks(w http.ResponseWriter, r *http.Request) {
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
		result, err := s.listLocalNetworks(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, networksResponse{Scope: "local:" + name, Error: err.Error(), Networks: []dockerwatcher.NetworkInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, networksResponse{Scope: "local:" + name, Networks: result})
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
		result, err := s.listLocalNetworks(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, networksResponse{Scope: scope, Error: err.Error(), Networks: []dockerwatcher.NetworkInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, networksResponse{Scope: scope, Networks: result})
		return
	}
	result, err := s.listRemoteNetworks(cfg, name)
	if err != nil {
		writeJSON(w, http.StatusOK, networksResponse{Scope: scope, Error: err.Error(), Networks: []dockerwatcher.NetworkInfo{}})
		return
	}
	writeJSON(w, http.StatusOK, networksResponse{Scope: scope, Networks: result})
}

func (s *Server) handleNetworkDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	networkID := strings.TrimSpace(r.URL.Query().Get("id"))
	if networkID == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
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
		result, err := s.getLocalNetworkDetails(cfg, name, networkID)
		if err != nil {
			writeJSON(w, http.StatusOK, networkDetailsResponse{Scope: "local:" + name, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, networkDetailsResponse{Scope: "local:" + name, Network: result})
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
		result, err := s.getLocalNetworkDetails(cfg, name, networkID)
		if err != nil {
			writeJSON(w, http.StatusOK, networkDetailsResponse{Scope: scope, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, networkDetailsResponse{Scope: scope, Network: result})
		return
	}
	result, err := s.getRemoteNetworkDetails(cfg, name, networkID)
	if err != nil {
		writeJSON(w, http.StatusOK, networkDetailsResponse{Scope: scope, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, networkDetailsResponse{Scope: scope, Network: result})
}

func (s *Server) handleNetworksRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload networkRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.NetworkID = strings.TrimSpace(payload.NetworkID)
	if payload.NetworkID == "" {
		writeError(w, http.StatusBadRequest, errors.New("network_id is required"))
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
		blockedBy, err := s.removeLocalNetwork(cfg, name, payload.NetworkID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if len(blockedBy) > 0 {
			writeJSON(w, http.StatusConflict, networkRemoveResponse{NetworkID: payload.NetworkID, BlockedBy: blockedBy, Error: "network is in use"})
			return
		}
		writeJSON(w, http.StatusOK, networkRemoveResponse{NetworkID: payload.NetworkID, RemovedAt: time.Now().Format(time.RFC3339)})
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
		blockedBy, err := s.removeLocalNetwork(cfg, name, payload.NetworkID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if len(blockedBy) > 0 {
			writeJSON(w, http.StatusConflict, networkRemoveResponse{NetworkID: payload.NetworkID, BlockedBy: blockedBy, Error: "network is in use"})
			return
		}
		writeJSON(w, http.StatusOK, networkRemoveResponse{NetworkID: payload.NetworkID, RemovedAt: time.Now().Format(time.RFC3339)})
		return
	}
	blockedBy, err := s.removeRemoteNetwork(cfg, name, payload.NetworkID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(blockedBy) > 0 {
		writeJSON(w, http.StatusConflict, networkRemoveResponse{NetworkID: payload.NetworkID, BlockedBy: blockedBy, Error: "network is in use"})
		return
	}
	writeJSON(w, http.StatusOK, networkRemoveResponse{NetworkID: payload.NetworkID, RemovedAt: time.Now().Format(time.RFC3339)})
}

func (s *Server) handleNetworksPrune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload networkPruneRequest
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
		deleted, err := s.pruneLocalNetworks(cfg, name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, networkPruneResponse{Deleted: deleted})
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
		deleted, err := s.pruneLocalNetworks(cfg, name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, networkPruneResponse{Deleted: deleted})
		return
	}
	deleted, err := s.pruneRemoteNetworks(cfg, name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, networkPruneResponse{Deleted: deleted})
}

func (s *Server) handleNetworkConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload networkConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.NetworkID = strings.TrimSpace(payload.NetworkID)
	payload.ContainerID = strings.TrimSpace(payload.ContainerID)
	if payload.NetworkID == "" || payload.ContainerID == "" {
		writeError(w, http.StatusBadRequest, errors.New("network_id and container_id are required"))
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
		if err := s.connectLocalNetwork(cfg, name, payload.NetworkID, payload.ContainerID); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "connected"})
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
		if err := s.connectLocalNetwork(cfg, name, payload.NetworkID, payload.ContainerID); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "connected"})
		return
	}
	if err := s.connectRemoteNetwork(cfg, name, payload.NetworkID, payload.ContainerID); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "connected"})
}

func (s *Server) handleNetworkDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload networkDisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.NetworkID = strings.TrimSpace(payload.NetworkID)
	payload.ContainerID = strings.TrimSpace(payload.ContainerID)
	if payload.NetworkID == "" || payload.ContainerID == "" {
		writeError(w, http.StatusBadRequest, errors.New("network_id and container_id are required"))
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
		if err := s.disconnectLocalNetwork(cfg, name, payload.NetworkID, payload.ContainerID); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
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
		if err := s.disconnectLocalNetwork(cfg, name, payload.NetworkID, payload.ContainerID); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
		return
	}
	if err := s.disconnectRemoteNetwork(cfg, name, payload.NetworkID, payload.ContainerID); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

func (s *Server) listLocalNetworks(cfg config.Config, name string) ([]dockerwatcher.NetworkInfo, error) {
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
	return watcher.ListNetworks(ctx)
}

func (s *Server) listRemoteNetworks(cfg config.Config, name string) ([]dockerwatcher.NetworkInfo, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/networks"
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
	var payload networksResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, errors.New(payload.Error)
	}
	if payload.Networks == nil {
		return []dockerwatcher.NetworkInfo{}, nil
	}
	return payload.Networks, nil
}

func (s *Server) getLocalNetworkDetails(cfg config.Config, name, networkID string) (dockerwatcher.NetworkDetails, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return dockerwatcher.NetworkDetails{}, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return dockerwatcher.NetworkDetails{}, err
	}
	defer watcher.Close()
	return watcher.NetworkDetails(ctx, networkID)
}

func (s *Server) getRemoteNetworkDetails(cfg config.Config, name, networkID string) (dockerwatcher.NetworkDetails, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return dockerwatcher.NetworkDetails{}, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return dockerwatcher.NetworkDetails{}, errors.New("remote url missing")
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/networks/details?id=" + url.QueryEscape(networkID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return dockerwatcher.NetworkDetails{}, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return dockerwatcher.NetworkDetails{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return dockerwatcher.NetworkDetails{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload networkDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return dockerwatcher.NetworkDetails{}, err
	}
	if payload.Error != "" {
		return dockerwatcher.NetworkDetails{}, errors.New(payload.Error)
	}
	return payload.Network, nil
}

func (s *Server) removeLocalNetwork(cfg config.Config, name, networkID string) ([]dockerwatcher.NetworkUsedBy, error) {
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
	usage, err := watcher.NetworkUsage(ctx, networkID)
	if err != nil {
		return nil, err
	}
	if len(usage) > 0 {
		return usage, nil
	}
	if err := watcher.RemoveNetwork(ctx, networkID); err != nil {
		return nil, err
	}
	return []dockerwatcher.NetworkUsedBy{}, nil
}

func (s *Server) removeRemoteNetwork(cfg config.Config, name, networkID string) ([]dockerwatcher.NetworkUsedBy, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	body, err := json.Marshal(networkRemoveRequest{
		NetworkID: networkID,
	})
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/networks/remove"
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
		var payload networkRemoveResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			return payload.BlockedBy, nil
		}
		return nil, errors.New("network is in use")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return []dockerwatcher.NetworkUsedBy{}, nil
}

func (s *Server) pruneLocalNetworks(cfg config.Config, name string) ([]string, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return nil, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	return watcher.PruneNetworks(ctx)
}

func (s *Server) pruneRemoteNetworks(cfg config.Config, name string) ([]string, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	body, err := json.Marshal(networkPruneRequest{})
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/networks/prune"
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload networkPruneResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Deleted, nil
}

func (s *Server) connectLocalNetwork(cfg config.Config, name, networkID, containerID string) error {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watcher.ConnectNetwork(ctx, networkID, containerID)
}

func (s *Server) disconnectLocalNetwork(cfg config.Config, name, networkID, containerID string) error {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watcher.DisconnectNetwork(ctx, networkID, containerID)
}

func (s *Server) connectRemoteNetwork(cfg config.Config, name, networkID, containerID string) error {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return errors.New("remote server not found")
	}
	if remote.URL == "" {
		return errors.New("remote url missing")
	}
	body, err := json.Marshal(networkConnectRequest{NetworkID: networkID, ContainerID: containerID})
	if err != nil {
		return err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/networks/connect"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (s *Server) disconnectRemoteNetwork(cfg config.Config, name, networkID, containerID string) error {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return errors.New("remote server not found")
	}
	if remote.URL == "" {
		return errors.New("remote url missing")
	}
	body, err := json.Marshal(networkDisconnectRequest{NetworkID: networkID, ContainerID: containerID})
	if err != nil {
		return err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/networks/disconnect"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
