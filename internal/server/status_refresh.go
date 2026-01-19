package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

func (s *Server) handleStatusRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg := s.store.Get()
	s.addLog("info", "remote scan snapshot refresh requested")
	go s.refreshRemoteStatusCache(cfg)
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) refreshRemoteStatusCache(cfg config.Config) {
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for _, remote := range cfg.RemoteServers {
		if remote.Maintenance {
			continue
		}
		if s.getServerInfoStatus("remote", remote.Name) != "online" {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(remote config.RemoteServer) {
			defer wg.Done()
			defer func() { <-sem }()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result, err := fetchRemoteScanStatus(ctx, remote)
			if err != nil {
				return
			}
			s.updateLastScans(result)
		}(remote)
	}
	wg.Wait()
}

func fetchRemoteScanStatus(ctx context.Context, remote config.RemoteServer) (dockerwatcher.ScanResult, error) {
	result := dockerwatcher.ScanResult{
		ServerName: remote.Name,
		ServerURL:  remote.URL,
		Local:      false,
		CheckedAt:  time.Now(),
	}
	if remote.URL == "" {
		return result, context.DeadlineExceeded
	}
	statusURL := strings.TrimSuffix(remote.URL, "/") + "/api/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return result, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, context.DeadlineExceeded
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	result.ServerName = remote.Name
	result.ServerURL = remote.URL
	result.Local = false
	if result.CheckedAt.IsZero() {
		result.CheckedAt = time.Now()
	}
	return result, nil
}
