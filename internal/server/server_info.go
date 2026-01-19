package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

const (
	serverInfoMaxConcurrency  = 10
	serverInfoQueueBufferSize = 128
	serverInfoEventBufferSize = 32
)

type serverInfoSnapshot struct {
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
	Checking  bool      `json:"checking"`
}

type streamEvent struct {
	Event string
	Data  []byte
}

func (s *Server) startServerInfoMonitor() {
	if s.store == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.serverInfoStop = cancel
	s.serverInfoQueue = make(chan config.RemoteServer, serverInfoQueueBufferSize)
	s.serverInfoPending = map[string]bool{}
	s.serverInfoSubs = map[chan streamEvent]struct{}{}
	for i := 0; i < serverInfoMaxConcurrency; i++ {
		go s.serverInfoWorker(ctx)
	}
	cfg := s.store.Get()
	s.refreshServerInfoCache(cfg)
}

func serverInfoKey(serverType, name string) string {
	if strings.TrimSpace(name) == "" {
		name = "unknown"
	}
	return serverType + ":" + name
}

func (s *Server) serverInfoSnapshot(cfg config.Config) []serverInfoSnapshot {
	s.serverInfoMu.RLock()
	defer s.serverInfoMu.RUnlock()
	items := make([]serverInfoSnapshot, 0, len(cfg.LocalServers)+len(cfg.RemoteServers))
	for _, local := range cfg.LocalServers {
		key := serverInfoKey("local", local.Name)
		if entry, ok := s.serverInfo[key]; ok {
			items = append(items, entry)
		} else {
			items = append(items, serverInfoSnapshot{
				Type:    "local",
				Name:    local.Name,
				Address: local.Socket,
				Status:  "unknown",
			})
		}
	}
	for _, remote := range cfg.RemoteServers {
		key := serverInfoKey("remote", remote.Name)
		if entry, ok := s.serverInfo[key]; ok {
			items = append(items, entry)
		} else {
			items = append(items, serverInfoSnapshot{
				Type:    "remote",
				Name:    remote.Name,
				Address: remote.URL,
				Status:  "unknown",
			})
		}
	}
	return items
}

func (s *Server) refreshServerInfoCache(cfg config.Config) {
	next := make(map[string]serverInfoSnapshot, len(cfg.LocalServers)+len(cfg.RemoteServers))
	now := time.Now()
	s.serverInfoMu.RLock()
	prev := make(map[string]serverInfoSnapshot, len(s.serverInfo))
	for key, entry := range s.serverInfo {
		prev[key] = entry
	}
	s.serverInfoMu.RUnlock()

	for _, local := range cfg.LocalServers {
		key := serverInfoKey("local", local.Name)
		entry := prev[key]
		entry.Type = "local"
		entry.Name = local.Name
		entry.Address = local.Socket
		entry.Version = s.version
		entry.Checking = false
		entry.CheckedAt = now
		if local.Maintenance {
			entry.Status = "maintenance"
		} else {
			entry.Status = localSocketStatus(local.Socket)
		}
		next[key] = entry
	}

	for _, remote := range cfg.RemoteServers {
		key := serverInfoKey("remote", remote.Name)
		prevEntry := prev[key]
		entry := prevEntry
		entry.Type = "remote"
		entry.Name = remote.Name
		entry.Address = remote.URL
		if remote.Maintenance {
			entry.Status = "maintenance"
			entry.Checking = false
			if entry.CheckedAt.IsZero() {
				entry.CheckedAt = now
			}
		} else if remote.URL == "" {
			entry.Status = "offline"
			entry.Checking = false
			entry.CheckedAt = now
		} else if entry.Status == "" || entry.Status == "maintenance" || prevEntry.Address != remote.URL {
			entry.Status = "unknown"
			entry.CheckedAt = time.Time{}
			entry.Checking = s.ensureRemoteCheck(remote)
		} else {
			nextEntry := entry
			nextEntry.Address = remote.URL
			entry = nextEntry
		}
		next[key] = entry
	}

	s.serverInfoMu.Lock()
	s.serverInfo = next
	s.serverInfoMu.Unlock()

	for key, entry := range next {
		if prevEntry, ok := prev[key]; !ok || !serverInfoEqual(prevEntry, entry) {
			s.broadcastServerInfo(entry)
		}
	}
}

func serverInfoEqual(a, b serverInfoSnapshot) bool {
	return a.Type == b.Type &&
		a.Name == b.Name &&
		a.Address == b.Address &&
		a.Version == b.Version &&
		a.Status == b.Status &&
		a.Checking == b.Checking &&
		a.CheckedAt.Equal(b.CheckedAt)
}

func (s *Server) ensureRemoteCheck(remote config.RemoteServer) bool {
	if remote.Maintenance {
		return false
	}
	if s.serverInfoQueue == nil {
		return false
	}
	key := serverInfoKey("remote", remote.Name)
	s.serverInfoMu.Lock()
	if s.serverInfoPending == nil {
		s.serverInfoPending = map[string]bool{}
	}
	if s.serverInfoPending[key] {
		s.serverInfoMu.Unlock()
		return true
	}
	s.serverInfoPending[key] = true
	s.serverInfoMu.Unlock()

	select {
	case s.serverInfoQueue <- remote:
		return true
	default:
		s.serverInfoMu.Lock()
		delete(s.serverInfoPending, key)
		s.serverInfoMu.Unlock()
		return false
	}
}

func (s *Server) triggerServerInfoRefresh(cfg config.Config) []serverInfoSnapshot {
	s.refreshServerInfoCache(cfg)
	now := time.Now()
	updates := make([]serverInfoSnapshot, 0, len(cfg.RemoteServers))
	toCheck := make([]config.RemoteServer, 0, len(cfg.RemoteServers))

	s.serverInfoMu.Lock()
	for _, remote := range cfg.RemoteServers {
		key := serverInfoKey("remote", remote.Name)
		entry := s.serverInfo[key]
		entry.Type = "remote"
		entry.Name = remote.Name
		entry.Address = remote.URL
		if remote.Maintenance {
			entry.Status = "maintenance"
			entry.Checking = false
			if entry.CheckedAt.IsZero() {
				entry.CheckedAt = now
			}
			s.serverInfo[key] = entry
			continue
		}
		if remote.URL == "" {
			entry.Status = "offline"
			entry.Checking = false
			entry.CheckedAt = now
			s.serverInfo[key] = entry
			continue
		}
		entry.Checking = false
		s.serverInfo[key] = entry
		toCheck = append(toCheck, remote)
	}
	s.serverInfoMu.Unlock()

	for _, remote := range toCheck {
		queued := s.ensureRemoteCheck(remote)
		if !queued {
			continue
		}
		key := serverInfoKey("remote", remote.Name)
		s.serverInfoMu.Lock()
		entry := s.serverInfo[key]
		entry.Checking = true
		s.serverInfo[key] = entry
		s.serverInfoMu.Unlock()
		updates = append(updates, entry)
	}

	for _, entry := range updates {
		s.broadcastServerInfo(entry)
	}
	return s.serverInfoSnapshot(cfg)
}

func (s *Server) handleServersRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg := s.store.Get()
	s.addLog("info", fmt.Sprintf("reachability refresh requested: servers=%d", len(cfg.RemoteServers)))
	items := s.triggerServerInfoRefresh(cfg)
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) serverInfoWorker(ctx context.Context) {
	client := &http.Client{Timeout: 5 * time.Second}
	for {
		select {
		case <-ctx.Done():
			return
		case remote := <-s.serverInfoQueue:
			key := serverInfoKey("remote", remote.Name)
			s.serverInfoMu.Lock()
			entry := s.serverInfo[key]
			if entry.Name == "" {
				entry = serverInfoSnapshot{
					Type:    "remote",
					Name:    remote.Name,
					Address: remote.URL,
				}
			}
			entry.Checking = true
			s.serverInfo[key] = entry
			s.serverInfoMu.Unlock()
			s.broadcastServerInfo(entry)

			version, online := fetchRemoteReachability(ctx, client, remote)
			checkedAt := time.Now()
			status := "online"
			if remote.Maintenance {
				status = "maintenance"
			} else if !online {
				status = "offline"
			}

			entry.Address = remote.URL
			entry.Version = version
			entry.Status = status
			entry.CheckedAt = checkedAt
			entry.Checking = false
			s.serverInfoMu.Lock()
			s.serverInfo[key] = entry
			delete(s.serverInfoPending, key)
			s.serverInfoMu.Unlock()
			s.broadcastServerInfo(entry)
		}
	}
}

func fetchRemoteReachability(ctx context.Context, client *http.Client, remote config.RemoteServer) (string, bool) {
	if remote.URL == "" {
		return "unknown", false
	}
	versionURL := strings.TrimSuffix(remote.URL, "/") + "/api/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionURL, nil)
	if err != nil {
		return "unknown", false
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "unknown", false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "unknown", false
	}
	var payload struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "unknown", false
	}
	version := strings.TrimSpace(payload.Version)
	if version == "" {
		version = "unknown"
	}
	return version, true
}

func (s *Server) broadcastServerInfo(entry serverInfoSnapshot) {
	payload, err := json.Marshal(entry)
	if err != nil {
		return
	}
	s.broadcastStreamEvent(streamEvent{
		Event: "server_info",
		Data:  payload,
	})
}

func (s *Server) broadcastScanResult(result dockerwatcher.ScanResult) {
	payload, err := json.Marshal(result)
	if err != nil {
		return
	}
	s.broadcastStreamEvent(streamEvent{
		Event: "scan_result",
		Data:  payload,
	})
}

func (s *Server) broadcastScanSnapshot(results []dockerwatcher.ScanResult) {
	payload, err := json.Marshal(results)
	if err != nil {
		return
	}
	s.broadcastStreamEvent(streamEvent{
		Event: "scan_snapshot",
		Data:  payload,
	})
}

func (s *Server) broadcastStreamEvent(event streamEvent) {
	s.serverInfoMu.RLock()
	subs := make([]chan streamEvent, 0, len(s.serverInfoSubs))
	for ch := range s.serverInfoSubs {
		subs = append(subs, ch)
	}
	s.serverInfoMu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *Server) handleServersStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := make(chan streamEvent, serverInfoEventBufferSize)
	s.serverInfoMu.Lock()
	if s.serverInfoSubs == nil {
		s.serverInfoSubs = map[chan streamEvent]struct{}{}
	}
	s.serverInfoSubs[ch] = struct{}{}
	s.serverInfoMu.Unlock()

	defer func() {
		s.serverInfoMu.Lock()
		delete(s.serverInfoSubs, ch)
		close(ch)
		s.serverInfoMu.Unlock()
	}()

	cfg := s.store.Get()
	s.refreshServerInfoCache(cfg)
	s.sendStreamEvent(w, "server_info_snapshot", s.serverInfoSnapshot(cfg))
	s.sendStreamEvent(w, "scan_snapshot", s.buildAggregateResults(cfg))
	flusher.Flush()

	keepAlive := time.NewTicker(20 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			if len(event.Data) == 0 {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\n", event.Event)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", event.Data)
			flusher.Flush()
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) sendStreamEvent(w http.ResponseWriter, event string, payload any) {
  data, err := json.Marshal(payload)
  if err != nil {
    return
  }
  _, _ = fmt.Fprintf(w, "event: %s\n", event)
  _, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}

func (s *Server) scanInProgress() bool {
  s.scanMutex.Lock()
  active := s.scanRunning || s.updateRunning
  s.scanMutex.Unlock()
  return active || s.remoteScanRunning.Load() > 0
}

func (s *Server) getServerInfoStatus(serverType, name string) string {
  key := serverInfoKey(serverType, name)
  s.serverInfoMu.RLock()
  entry := s.serverInfo[key]
  s.serverInfoMu.RUnlock()
  return strings.ToLower(entry.Status)
}

func (s *Server) buildAggregateResults(cfg config.Config) []dockerwatcher.ScanResult {
	results := []dockerwatcher.ScanResult{}
	s.lastScanMutex.RLock()
	scans := append([]dockerwatcher.ScanResult(nil), s.lastScans...)
	s.lastScanMutex.RUnlock()
	byKey := make(map[string]dockerwatcher.ScanResult, len(scans))
	for _, scan := range scans {
		key := scanKey(scan.Local, scan.ServerName)
		s.applyScanState(&scan, scan.Local, scan.ServerName)
		byKey[key] = scan
	}

	for _, local := range cfg.LocalServers {
		key := scanKey(true, local.Name)
		if existing, ok := byKey[key]; ok {
			results = append(results, existing)
			continue
		}
		placeholder := dockerwatcher.ScanResult{
			ServerName: local.Name,
			ServerURL:  local.Socket,
			Local:      true,
		}
		s.applyScanState(&placeholder, true, local.Name)
		results = append(results, placeholder)
	}

	for _, remote := range cfg.RemoteServers {
		key := scanKey(false, remote.Name)
		if existing, ok := byKey[key]; ok {
			results = append(results, existing)
			continue
		}
		placeholder := dockerwatcher.ScanResult{
			ServerName: remote.Name,
			ServerURL:  remote.URL,
			Local:      false,
		}
		s.applyScanState(&placeholder, false, remote.Name)
		results = append(results, placeholder)
	}
	return results
}
