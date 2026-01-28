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

type imagesResponse struct {
	Scope  string                     `json:"scope"`
	Images []dockerwatcher.ImageInfo  `json:"images"`
	Error  string                     `json:"error,omitempty"`
}

type imagePullRequest struct {
	Scope      string `json:"scope"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type imagePullResponse struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	PulledAt   string `json:"pulled_at"`
}

type imagePruneRequest struct {
	Scope string `json:"scope"`
	Mode  string `json:"mode"`
}

type imagePruneResponse struct {
	Deleted        int    `json:"deleted"`
	ReclaimedBytes uint64 `json:"reclaimed_bytes"`
}

type imageRemoveRequest struct {
	Scope   string `json:"scope"`
	ImageID string `json:"image_id"`
}

type imageRemoveResponse struct {
	ImageID   string `json:"image_id"`
	RemovedAt string `json:"removed_at"`
}

func (s *Server) handleImages(w http.ResponseWriter, r *http.Request) {
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
		result, err := s.listLocalImages(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, imagesResponse{Scope: "local:" + name, Error: err.Error(), Images: []dockerwatcher.ImageInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, imagesResponse{Scope: "local:" + name, Images: result})
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
		result, err := s.listLocalImages(cfg, name)
		if err != nil {
			writeJSON(w, http.StatusOK, imagesResponse{Scope: scope, Error: err.Error(), Images: []dockerwatcher.ImageInfo{}})
			return
		}
		writeJSON(w, http.StatusOK, imagesResponse{Scope: scope, Images: result})
		return
	}
	result, err := s.listRemoteImages(cfg, name)
	if err != nil {
		writeJSON(w, http.StatusOK, imagesResponse{Scope: scope, Error: err.Error(), Images: []dockerwatcher.ImageInfo{}})
		return
	}
	writeJSON(w, http.StatusOK, imagesResponse{Scope: scope, Images: result})
}

func (s *Server) handleImagesPull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload imagePullRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Repository = strings.TrimSpace(payload.Repository)
	payload.Tag = strings.TrimSpace(payload.Tag)
	if payload.Repository == "" {
		writeError(w, http.StatusBadRequest, errors.New("repository is required"))
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
		if err := s.pullLocalImage(cfg, name, payload.Repository, payload.Tag); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, imagePullResponse{
			Repository: payload.Repository,
			Tag:        normalizeTag(payload.Tag),
			PulledAt:   time.Now().Format(time.RFC3339),
		})
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
		if err := s.pullLocalImage(cfg, name, payload.Repository, payload.Tag); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, imagePullResponse{
			Repository: payload.Repository,
			Tag:        normalizeTag(payload.Tag),
			PulledAt:   time.Now().Format(time.RFC3339),
		})
		return
	}
	if err := s.pullRemoteImage(cfg, name, payload.Repository, payload.Tag); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, imagePullResponse{
		Repository: payload.Repository,
		Tag:        normalizeTag(payload.Tag),
		PulledAt:   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleImagesPrune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload imagePruneRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Mode = strings.TrimSpace(payload.Mode)
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
		deleted, reclaimed, err := s.pruneLocalImages(cfg, name, payload.Mode)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, imagePruneResponse{Deleted: deleted, ReclaimedBytes: reclaimed})
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
		deleted, reclaimed, err := s.pruneLocalImages(cfg, name, payload.Mode)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, imagePruneResponse{Deleted: deleted, ReclaimedBytes: reclaimed})
		return
	}
	deleted, reclaimed, err := s.pruneRemoteImages(cfg, name, payload.Mode)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, imagePruneResponse{Deleted: deleted, ReclaimedBytes: reclaimed})
}

func (s *Server) handleImagesRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload imageRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.ImageID = strings.TrimSpace(payload.ImageID)
	if payload.ImageID == "" {
		writeError(w, http.StatusBadRequest, errors.New("image_id is required"))
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
		if err := s.removeLocalImage(cfg, name, payload.ImageID); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, imageRemoveResponse{ImageID: payload.ImageID, RemovedAt: time.Now().Format(time.RFC3339)})
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
		if err := s.removeLocalImage(cfg, name, payload.ImageID); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, imageRemoveResponse{ImageID: payload.ImageID, RemovedAt: time.Now().Format(time.RFC3339)})
		return
	}
	if err := s.removeRemoteImage(cfg, name, payload.ImageID); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, imageRemoveResponse{ImageID: payload.ImageID, RemovedAt: time.Now().Format(time.RFC3339)})
}

func (s *Server) listLocalImages(cfg config.Config, name string) ([]dockerwatcher.ImageInfo, error) {
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
	return watcher.ListImages(ctx)
}

func (s *Server) listRemoteImages(cfg config.Config, name string) ([]dockerwatcher.ImageInfo, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return nil, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return nil, errors.New("remote url missing")
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/images"
	if name != "" {
		endpoint += "?server=" + url.QueryEscape(name)
	}
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
	var payload imagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, errors.New(payload.Error)
	}
	if payload.Images == nil {
		return []dockerwatcher.ImageInfo{}, nil
	}
	return payload.Images, nil
}

func (s *Server) pullLocalImage(cfg config.Config, name, repository, tag string) error {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watcher.PullImage(ctx, repository, tag)
}

func (s *Server) pullRemoteImage(cfg config.Config, name, repository, tag string) error {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return errors.New("remote server not found")
	}
	if remote.URL == "" {
		return errors.New("remote url missing")
	}
	body, err := json.Marshal(imagePullRequest{
		Repository: repository,
		Tag:        tag,
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/images/pull"
	if name != "" {
		endpoint += "?server=" + url.QueryEscape(name)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 2 * time.Minute}
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

func (s *Server) pruneLocalImages(cfg config.Config, name, mode string) (int, uint64, error) {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return 0, 0, errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return 0, 0, err
	}
	defer watcher.Close()
	return watcher.PruneImages(ctx, mode)
}

func (s *Server) pruneRemoteImages(cfg config.Config, name, mode string) (int, uint64, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return 0, 0, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return 0, 0, errors.New("remote url missing")
	}
	body, err := json.Marshal(imagePruneRequest{
		Mode: mode,
	})
	if err != nil {
		return 0, 0, err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/images/prune"
	if name != "" {
		endpoint += "?server=" + url.QueryEscape(name)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload imagePruneResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, 0, err
	}
	return payload.Deleted, payload.ReclaimedBytes, nil
}

func (s *Server) removeLocalImage(cfg config.Config, name, imageID string) error {
	local, ok := findLocalServer(cfg.LocalServers, name)
	if !ok {
		return errors.New("local server not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watcher.RemoveImage(ctx, imageID)
}

func (s *Server) removeRemoteImage(cfg config.Config, name, imageID string) error {
	remote, ok := findRemoteServer(cfg.RemoteServers, name)
	if !ok {
		return errors.New("remote server not found")
	}
	if remote.URL == "" {
		return errors.New("remote url missing")
	}
	body, err := json.Marshal(imageRemoveRequest{
		ImageID: imageID,
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/images/remove"
	if name != "" {
		endpoint += "?server=" + url.QueryEscape(name)
	}
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

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return "latest"
	}
	return tag
}
