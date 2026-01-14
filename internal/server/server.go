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
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
	"contiwatch/internal/notify"
)

type Server struct {
	store         *config.Store
	watcher       *dockerwatcher.Watcher
	discord       *notify.DiscordClient
	mux           *http.ServeMux
	lastScan      dockerwatcher.ScanResult
	lastScanMutex sync.RWMutex
	lastScans     []dockerwatcher.ScanResult
	scanMutex     sync.Mutex
	scanRunning   bool

	schedulerMu       sync.Mutex
	schedulerCancel   context.CancelFunc
	schedulerInterval time.Duration
	schedulerEnabled  bool
	schedulerParent   context.Context

	logMu sync.RWMutex
	logs  []logEntry
	logSeq int64

	agentMode  bool
	agentToken string
	version    string
	remoteScanRunning atomic.Int64
}

func New(store *config.Store, watcher *dockerwatcher.Watcher, agentMode bool, agentToken string, version string) *Server {
	s := &Server{
		store:      store,
		watcher:    watcher,
		mux:        http.NewServeMux(),
		agentMode:  agentMode,
		agentToken: agentToken,
		version:    version,
	}
	s.routes()
	return s
}

func (s *Server) UpdateDiscord(webhookURL string) {
	if webhookURL == "" {
		s.discord = nil
		return
	}
	s.discord = notify.NewDiscordClient(webhookURL)
}

func discordNotificationsEnabled(cfg config.Config) bool {
	if cfg.DiscordNotificationsEnabled == nil {
		return true
	}
	return *cfg.DiscordNotificationsEnabled
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.agentMode {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Path != "/api/health" {
			if !s.authorize(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		if !s.agentAllowed(r.URL.Path) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/servers", s.handleServers)
	s.mux.HandleFunc("/api/servers/", s.handleServerByName)
	s.mux.HandleFunc("/api/servers/info", s.handleServersInfo)
	s.mux.HandleFunc("/api/locals", s.handleLocals)
	s.mux.HandleFunc("/api/locals/", s.handleLocalByName)
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/scan/state", s.handleScanState)
	s.mux.HandleFunc("/api/scan", s.handleScan)
	s.mux.HandleFunc("/api/update/", s.handleUpdateContainer)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/aggregate", s.handleAggregate)
	s.mux.HandleFunc("/api/version", s.handleVersion)

	if !s.agentMode {
		staticDir := "/app/web/static"
		if _, err := os.Stat(staticDir); err != nil {
			staticDir = "./web/static"
		}
		static := http.FileServer(http.Dir(staticDir))
		s.mux.Handle("/", static)
	}
}

func (s *Server) authorize(r *http.Request) bool {
	if s.agentToken == "" {
		return false
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	return token == s.agentToken
}

func (s *Server) agentAllowed(path string) bool {
	switch {
	case path == "/api/health":
		return true
	case path == "/api/version":
		return true
	case path == "/api/status":
		return true
	case path == "/api/scan":
		return true
	case path == "/api/logs":
		return true
	case strings.HasPrefix(path, "/api/update/"):
		return true
	default:
		return false
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.Get())
	case http.MethodPut:
		var payload config.Config
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		updated, err := s.store.Update(func(cfg *config.Config) {
			cfg.ScanIntervalSec = payload.ScanIntervalSec
			cfg.SchedulerEnabled = payload.SchedulerEnabled
			cfg.GlobalPolicy = payload.GlobalPolicy
			cfg.DiscordWebhookURL = payload.DiscordWebhookURL
			if payload.DiscordNotificationsEnabled != nil {
				cfg.DiscordNotificationsEnabled = payload.DiscordNotificationsEnabled
			}
			cfg.UpdateStoppedContainers = payload.UpdateStoppedContainers
			cfg.PruneDanglingImages = payload.PruneDanglingImages
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		s.addLog("info", fmt.Sprintf("settings updated: prune_dangling_images=%t", updated.PruneDanglingImages))
		s.UpdateDiscord(updated.DiscordWebhookURL)
		s.UpdateScheduler(updated)
		writeJSON(w, http.StatusOK, updated)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.Get()
		writeJSON(w, http.StatusOK, cfg.RemoteServers)
	case http.MethodPost:
		var payload config.RemoteServer
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if payload.Name == "" || payload.URL == "" {
			writeError(w, http.StatusBadRequest, errors.New("name and url required"))
			return
		}
		updated, err := s.store.Update(func(cfg *config.Config) {
			cfg.RemoteServers = upsertServer(cfg.RemoteServers, payload)
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, updated.RemoteServers)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLocals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.Get()
		writeJSON(w, http.StatusOK, cfg.LocalServers)
	case http.MethodPost:
		var payload config.LocalServer
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if payload.Name == "" || payload.Socket == "" {
			writeError(w, http.StatusBadRequest, errors.New("name and socket required"))
			return
		}
		updated, err := s.store.Update(func(cfg *config.Config) {
			cfg.LocalServers = upsertLocal(cfg.LocalServers, payload)
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, updated.LocalServers)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLocalByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/locals/")
	if name == "" {
		writeError(w, http.StatusBadRequest, errors.New("server name required"))
		return
	}
	updated, err := s.store.Update(func(cfg *config.Config) {
		cfg.LocalServers = removeLocal(cfg.LocalServers, name)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, updated.LocalServers)
}

func (s *Server) handleServerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/servers/")
	if name == "" {
		writeError(w, http.StatusBadRequest, errors.New("server name required"))
		return
	}
	updated, err := s.store.Update(func(cfg *config.Config) {
		cfg.RemoteServers = removeServer(cfg.RemoteServers, name)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, updated.RemoteServers)
}

type serverInfo struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

func (s *Server) handleServersInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg := s.store.Get()
	items := []serverInfo{}
	for _, local := range cfg.LocalServers {
		items = append(items, serverInfo{
			Type:    "local",
			Name:    local.Name,
			Address: local.Socket,
			Version: s.version,
			Status:  "online",
		})
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, remote := range cfg.RemoteServers {
		version, status := fetchRemoteVersion(r.Context(), client, remote)
		items = append(items, serverInfo{
			Type:    "remote",
			Name:    remote.Name,
			Address: remote.URL,
			Version: version,
			Status:  status,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.lastScanMutex.RLock()
	defer s.lastScanMutex.RUnlock()
	writeJSON(w, http.StatusOK, s.lastScan)
}

func (s *Server) handleScanState(w http.ResponseWriter, r *http.Request) {
	s.scanMutex.Lock()
	scanRunning := s.scanRunning
	s.scanMutex.Unlock()
	remoteRunning := s.remoteScanRunning.Load() > 0
	writeJSON(w, http.StatusOK, map[string]bool{"running": scanRunning || remoteRunning})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": s.version})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.logMu.RLock()
		logs := append([]logEntry(nil), s.logs...)
		s.logMu.RUnlock()
		if logs == nil {
			logs = []logEntry{}
		}
		writeJSON(w, http.StatusOK, logs)
	case http.MethodDelete:
		s.clearLogs()
		writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	scope := strings.TrimSpace(r.URL.Query().Get("server"))
	cfg := s.store.Get()

	s.addLog("info", "manual scan started")
	switch {
	case scope == "" || scope == "all":
		result, err := s.runScan(r.Context())
		if err != nil {
			if errors.Is(err, errScanInProgress) {
				writeError(w, http.StatusConflict, err)
				return
			}
			s.addLog("error", fmt.Sprintf("manual scan failed: %v", err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		remoteCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		go func() {
			defer cancel()
			s.triggerRemoteScans(remoteCtx, cfg.RemoteServers)
		}()
		s.addLog("info", fmt.Sprintf("manual scan finished (local batch): servers=%d", len(cfg.LocalServers)))
		writeJSON(w, http.StatusOK, result)
		return
	case strings.HasPrefix(scope, "local:"):
		name := strings.TrimPrefix(scope, "local:")
		local, ok := findLocalServer(cfg.LocalServers, name)
		if !ok {
			writeError(w, http.StatusNotFound, errors.New("local server not found"))
			return
		}
		s.scanMutex.Lock()
		if s.scanRunning {
			s.scanMutex.Unlock()
			writeError(w, http.StatusConflict, errors.New("scan already in progress"))
			return
		}
		s.scanRunning = true
		s.scanMutex.Unlock()
		defer func() {
			s.scanMutex.Lock()
			s.scanRunning = false
			s.scanMutex.Unlock()
		}()
		s.addLog("info", fmt.Sprintf("local scan started: %s", local.Name))
		result, err := s.scanLocalServer(r.Context(), cfg, local)
		if err != nil {
			s.addLog("error", fmt.Sprintf("manual scan failed: %s: %v", local.Name, err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		s.updateLastScans(result)
		s.addLog("info", fmt.Sprintf("local scan finished: %s %s", local.Name, s.formatScanSummary(result)))
		s.sendScanNotification(cfg, result)
		s.addLog("info", fmt.Sprintf("manual scan finished: containers=%d", len(result.Containers)))
		writeJSON(w, http.StatusOK, result)
		return
	case strings.HasPrefix(scope, "remote:"):
		name := strings.TrimPrefix(scope, "remote:")
		remote, ok := findRemoteServer(cfg.RemoteServers, name)
		if !ok {
			writeError(w, http.StatusNotFound, errors.New("remote server not found"))
			return
		}
		s.addLog("info", fmt.Sprintf("remote scan started: %s", remote.Name))
		s.remoteScanRunning.Add(1)
		result, err := s.scanRemoteServer(r.Context(), remote)
		s.remoteScanRunning.Add(-1)
		if err != nil {
			s.addLog("error", fmt.Sprintf("manual scan failed: %s: %v", remote.Name, err))
			writeError(w, http.StatusBadGateway, err)
			return
		}
		s.addLog("info", fmt.Sprintf("remote scan finished: %s %s", remote.Name, s.formatScanSummary(result)))
		s.sendScanNotification(cfg, result)
		s.addLog("info", fmt.Sprintf("manual scan finished: containers=%d", len(result.Containers)))
		writeJSON(w, http.StatusOK, result)
		return
	default:
		writeError(w, http.StatusBadRequest, errors.New("invalid scan scope"))
		return
	}
}

func (s *Server) handleUpdateContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	containerID := strings.TrimPrefix(r.URL.Path, "/api/update/")
	if containerID == "" {
		writeError(w, http.StatusBadRequest, errors.New("container id required"))
		return
	}

	s.scanMutex.Lock()
	if s.scanRunning {
		s.scanMutex.Unlock()
		writeError(w, http.StatusConflict, errors.New("another operation is in progress"))
		return
	}
	s.scanRunning = true
	s.scanMutex.Unlock()
	defer func() {
		s.scanMutex.Lock()
		s.scanRunning = false
		s.scanMutex.Unlock()
	}()

	cfg := s.store.Get()
	serverParam := strings.TrimSpace(r.URL.Query().Get("server"))
	isRemote := false
	serverName := serverParam
	switch {
	case strings.HasPrefix(serverParam, "remote:"):
		isRemote = true
		serverName = strings.TrimPrefix(serverParam, "remote:")
	case strings.HasPrefix(serverParam, "local:"):
		serverName = strings.TrimPrefix(serverParam, "local:")
	}
	if serverName == "" {
		if len(cfg.LocalServers) == 1 {
			serverName = cfg.LocalServers[0].Name
		} else {
			writeError(w, http.StatusBadRequest, errors.New("server name required"))
			return
		}
	}

	updateCtx := r.Context()
	var cancel context.CancelFunc
	if s.agentMode {
		updateCtx, cancel = context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
	}

	var result dockerwatcher.UpdateResult
	var err error
	if dockerwatcher.IsSelfContainer(containerID) {
		if !s.agentMode {
			writeError(w, http.StatusBadRequest, errors.New("self update is only supported in agent mode"))
			return
		}
		if err := s.watcher.TriggerSelfUpdate(updateCtx, containerID); err != nil {
			s.addLog("error", fmt.Sprintf("self update failed: %s: %v", containerID, err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		result = dockerwatcher.UpdateResult{
			ID:            containerID,
			Name:          shortID(containerID),
			Updated:       false,
			PreviousState: "online",
			CurrentState:  "online",
			Message:       "self update scheduled; recheck in 30s",
		}
		s.addLog("info", fmt.Sprintf("self update scheduled: %s", containerID))
		writeJSON(w, http.StatusOK, result)
		return
	}
	if isRemote {
		remote, ok := findRemoteServer(cfg.RemoteServers, serverName)
		if !ok {
			writeError(w, http.StatusNotFound, errors.New("remote server not found"))
			return
		}
		result, err = s.updateRemoteContainer(updateCtx, remote, containerID)
		if err != nil {
			if isRemoteUpdateDisconnect(err) {
				result = dockerwatcher.UpdateResult{
					ID:      containerID,
					Name:    shortID(containerID),
					Updated: false,
					Message: "update triggered; agent restarting",
				}
				s.addLog("warn", fmt.Sprintf("update connection closed: %s (agent restarting)", containerID))
			} else {
				s.addLog("error", fmt.Sprintf("update failed: %s: %v", containerID, err))
				writeError(w, http.StatusBadGateway, err)
				return
			}
		}
	} else {
		localServer, ok := findLocalServer(cfg.LocalServers, serverName)
		if !ok {
			writeError(w, http.StatusNotFound, errors.New("local server not found"))
			return
		}
		watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(localServer.Socket))
		if err != nil {
			s.addLog("error", fmt.Sprintf("update failed: %s: %v", containerID, err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		defer watcher.Close()
		result, err = watcher.UpdateContainer(updateCtx, containerID, cfg)
		if err != nil {
			s.addLog("error", fmt.Sprintf("update failed: %s: %v", containerID, err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	serverLabel := serverName
	if serverLabel == "" {
		serverLabel = "local"
	}
	serverScope := "local"
	if isRemote {
		serverScope = "remote"
	}
	log.Printf("manual update: server=%s (%s) container=%s updated=%t previous=%s current=%s msg=%s", serverLabel, serverScope, result.Name, result.Updated, result.PreviousState, result.CurrentState, result.Message)
	if result.Updated {
		s.addLog("info", fmt.Sprintf("updated %s (%s → %s) on %s (%s)", result.Name, result.PreviousState, result.CurrentState, serverLabel, serverScope))
		if cfg.PruneDanglingImages {
			if strings.Contains(result.Message, "prune failed") {
				s.addLog("error", fmt.Sprintf("prune dangling images failed after update: %s on %s (%s)", result.Name, serverLabel, serverScope))
			} else {
				s.addLog("info", fmt.Sprintf("pruned dangling images after update: %s on %s (%s)", result.Name, serverLabel, serverScope))
			}
		}
	} else {
		s.addLog("info", fmt.Sprintf("update skipped %s on %s (%s): %s", result.Name, serverLabel, serverScope, result.Message))
	}
	s.sendUpdateNotification(cfg, serverLabel, result)

	s.lastScanMutex.Lock()
	for i := range s.lastScan.Containers {
		if s.lastScan.Containers[i].ID == containerID {
			s.lastScan.Containers[i].LastChecked = time.Now()
			s.lastScan.Containers[i].Error = ""
			s.lastScan.Containers[i].Updated = result.Updated
			if result.Updated {
				s.lastScan.Containers[i].UpdateAvailable = false
			}
			break
		}
	}
	s.lastScanMutex.Unlock()

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := s.store.Get()

	results := []dockerwatcher.ScanResult{}
	s.lastScanMutex.RLock()
	localScans := append([]dockerwatcher.ScanResult(nil), s.lastScans...)
	s.lastScanMutex.RUnlock()
	localByName := make(map[string]dockerwatcher.ScanResult, len(localScans))
	for _, local := range localScans {
		localByName[local.ServerName] = local
	}
	for _, local := range cfg.LocalServers {
		if existing, ok := localByName[local.Name]; ok {
			results = append(results, existing)
			continue
		}
		results = append(results, dockerwatcher.ScanResult{
			ServerName: local.Name,
			ServerURL:  local.Socket,
			Local:      true,
		})
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, remote := range cfg.RemoteServers {
		remoteResult := dockerwatcher.ScanResult{ServerName: remote.Name, ServerURL: remote.URL, Local: false}
		if remote.URL == "" {
			remoteResult.Error = "missing url"
			results = append(results, remoteResult)
			continue
		}
		url := strings.TrimSuffix(remote.URL, "/") + "/api/status"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			remoteResult.Error = "invalid url"
			remoteResult.CheckedAt = time.Now()
			results = append(results, remoteResult)
			continue
		}
		if remote.Token != "" {
			req.Header.Set("Authorization", "Bearer "+remote.Token)
		}
		resp, err := client.Do(req)
		if err != nil {
			remoteResult.Error = "connection failed"
			remoteResult.CheckedAt = time.Now()
			results = append(results, remoteResult)
			continue
		}
		if resp.Body == nil {
			remoteResult.Error = "empty response"
			remoteResult.CheckedAt = time.Now()
			results = append(results, remoteResult)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			remoteResult.Error = fmt.Sprintf("status %d", resp.StatusCode)
			_ = resp.Body.Close()
			remoteResult.CheckedAt = time.Now()
			results = append(results, remoteResult)
			continue
		}
		var payload dockerwatcher.ScanResult
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			remoteResult.Error = "invalid payload"
			_ = resp.Body.Close()
			remoteResult.CheckedAt = time.Now()
			results = append(results, remoteResult)
			continue
		}
		_ = resp.Body.Close()
		payload.ServerName = remote.Name
		payload.ServerURL = remote.URL
		payload.Local = false
		results = append(results, payload)
	}

	writeJSON(w, http.StatusOK, results)
}

var errScanInProgress = errors.New("scan already in progress")

func (s *Server) runScan(ctx context.Context) (dockerwatcher.ScanResult, error) {
	s.scanMutex.Lock()
	if s.scanRunning {
		s.scanMutex.Unlock()
		return dockerwatcher.ScanResult{}, errScanInProgress
	}
	s.scanRunning = true
	s.scanMutex.Unlock()
	defer func() {
		s.scanMutex.Lock()
		s.scanRunning = false
		s.scanMutex.Unlock()
	}()

	cfg := s.store.Get()
	results := []dockerwatcher.ScanResult{}

	if len(cfg.LocalServers) == 0 {
		s.addLog("info", "scan skipped: no local servers configured")
	} else {
		for _, local := range cfg.LocalServers {
			s.addLog("info", fmt.Sprintf("local scan started: %s", local.Name))
			result, err := s.scanLocalServer(ctx, cfg, local)
			if err != nil {
				s.addLog("error", fmt.Sprintf("scan failed: %s: %v", local.Name, err))
				results = append(results, dockerwatcher.ScanResult{
					ServerName: local.Name,
					ServerURL:  local.Socket,
					Local:      true,
					CheckedAt:  time.Now(),
					Error:      err.Error(),
				})
				continue
			}
			s.addLog("info", fmt.Sprintf("local scan finished: %s %s", local.Name, s.formatScanSummary(result)))
			results = append(results, result)
			s.sendScanNotification(cfg, result)
		}
	}

	// Preserve last local scan for /api/status and aggregate.
	if len(results) > 0 {
		result := results[0]
		s.lastScanMutex.Lock()
		s.lastScan = result
		s.lastScans = results
		s.lastScanMutex.Unlock()
	} else {
		s.lastScanMutex.Lock()
		s.lastScan = dockerwatcher.ScanResult{}
		s.lastScans = nil
		s.lastScanMutex.Unlock()
	}

	if len(results) == 0 {
		return dockerwatcher.ScanResult{}, nil
	}
	return results[0], nil
}

func (s *Server) updateLastScans(result dockerwatcher.ScanResult) {
	s.lastScanMutex.Lock()
	defer s.lastScanMutex.Unlock()
	updated := false
	for i := range s.lastScans {
		if s.lastScans[i].ServerName == result.ServerName && s.lastScans[i].Local == result.Local {
			s.lastScans[i] = result
			updated = true
			break
		}
	}
	if !updated {
		s.lastScans = append(s.lastScans, result)
	}
	if s.lastScan.ServerName == "" || s.lastScan.ServerName == result.ServerName {
		s.lastScan = result
	}
}

func (s *Server) scanLocalServer(ctx context.Context, cfg config.Config, local config.LocalServer) (dockerwatcher.ScanResult, error) {
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	defer watcher.Close()
	result, err := watcher.Scan(ctx, cfg)
	if err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	result.ServerName = local.Name
	result.ServerURL = local.Socket
	result.Local = true
	return result, nil
}

func (s *Server) scanRemoteServer(ctx context.Context, remote config.RemoteServer) (dockerwatcher.ScanResult, error) {
	if remote.URL == "" {
		return dockerwatcher.ScanResult{}, errors.New("missing url")
	}
	url := strings.TrimSuffix(remote.URL, "/") + "/api/scan"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return dockerwatcher.ScanResult{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload dockerwatcher.ScanResult
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	payload.ServerName = remote.Name
	payload.ServerURL = remote.URL
	payload.Local = false
	return payload, nil
}

func (s *Server) updateRemoteContainer(ctx context.Context, remote config.RemoteServer, containerID string) (dockerwatcher.UpdateResult, error) {
	if remote.URL == "" {
		return dockerwatcher.UpdateResult{}, errors.New("missing url")
	}
	updateURL := strings.TrimSuffix(remote.URL, "/") + "/api/update/" + url.PathEscape(containerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, updateURL, nil)
	if err != nil {
		return dockerwatcher.UpdateResult{}, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return dockerwatcher.UpdateResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var payload struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil && payload.Error != "" {
			return dockerwatcher.UpdateResult{}, errors.New(payload.Error)
		}
		return dockerwatcher.UpdateResult{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	var result dockerwatcher.UpdateResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return dockerwatcher.UpdateResult{}, err
	}
	return result, nil
}

func (s *Server) triggerRemoteScans(ctx context.Context, remotes []config.RemoteServer) {
	s.remoteScanRunning.Add(1)
	defer s.remoteScanRunning.Add(-1)
	successCount := 0
	failCount := 0
	for _, remote := range remotes {
		if ctx.Err() != nil {
			break
		}
		s.addLog("info", fmt.Sprintf("remote scan started: %s", remote.Name))
		result, err := s.scanRemoteServer(ctx, remote)
		if err != nil {
			s.addLog("error", fmt.Sprintf("remote scan failed: %s: %v", remote.Name, err))
			failCount++
			continue
		}
		cfg := s.store.Get()
		s.addLog("info", fmt.Sprintf("remote scan finished: %s %s", remote.Name, s.formatScanSummary(result)))
		s.sendScanNotification(cfg, result)
		successCount++
	}
	if len(remotes) > 0 {
		s.addLog("info", fmt.Sprintf("remote scans completed: ok=%d failed=%d", successCount, failCount))
		if ctx.Err() != nil {
			s.addLog("warn", fmt.Sprintf("remote scans aborted: %v", ctx.Err()))
		}
	}
}

func (s *Server) formatScanSummary(result dockerwatcher.ScanResult) string {
	summary := s.buildScanSummary(result)
	return fmt.Sprintf(
		"containers=%d updates=%d ready=%d skipped=%d updated=%d",
		summary.total,
		summary.updates,
		summary.ready,
		summary.skipped,
		summary.updated,
	)
}

type scanSummary struct {
	total       int
	updates     int
	ready       int
	skipped     int
	updated     int
	updateNames []string
	updatedNames []string
}

func (s *Server) buildScanSummary(result dockerwatcher.ScanResult) scanSummary {
	summary := scanSummary{
		total: len(result.Containers),
	}
	for _, container := range result.Containers {
		if container.Updated {
			summary.updated++
			summary.updatedNames = append(summary.updatedNames, container.Name)
		}
		if !container.UpdateAvailable {
			continue
		}
		summary.updates++
		summary.updateNames = append(summary.updateNames, container.Name)
		if strings.HasPrefix(container.Error, "skipped:") {
			summary.skipped++
			continue
		}
		if container.Policy == config.PolicyUpdate && !container.Updated {
			summary.ready++
		}
	}
	return summary
}

func (s *Server) sendScanNotification(cfg config.Config, result dockerwatcher.ScanResult) {
	if s.discord == nil || !discordNotificationsEnabled(cfg) {
		return
	}
	summary := s.buildScanSummary(result)
	serverLabel := result.ServerName
	if serverLabel == "" {
		if result.Local {
			serverLabel = "local"
		} else {
			serverLabel = "remote"
		}
	}
	scope := "remote"
	if result.Local {
		scope = "local"
	}

	description := fmt.Sprintf(
		"Server: %s (%s)\nScanned images: %d\nUpdates available: %d\nReady to update: %d\nSkipped: %d\nUpdated: %d",
		serverLabel,
		scope,
		summary.total,
		summary.updates,
		summary.ready,
		summary.skipped,
		summary.updated,
	)
	if summary.updates > 0 {
		description += "\n\nContainers with updates:\n- " + strings.Join(summary.updateNames, "\n- ")
	}
	if summary.updated > 0 {
		description += "\n\nContainers updated:\n- " + strings.Join(summary.updatedNames, "\n- ")
	}
	const discordBlue = 0x3498DB
	_ = s.discord.SendEmbed("Contiwatch updates:", description, discordBlue)
}

func (s *Server) sendUpdateNotification(cfg config.Config, serverName string, result dockerwatcher.UpdateResult) {
	if s.discord == nil || !discordNotificationsEnabled(cfg) {
		return
	}
	status := "not updated"
	if result.Updated {
		status = "updated"
	} else if result.Message != "" {
		status = result.Message
	}
	description := fmt.Sprintf(
		"Server: %s\nContainer: %s\nResult: %s\nPrevious state: %s\nCurrent state: %s",
		serverName,
		result.Name,
		status,
		result.PreviousState,
		result.CurrentState,
	)
	const discordBlue = 0x3498DB
	_ = s.discord.SendEmbed("Contiwatch updates:", description, discordBlue)
}

type logEntry struct {
	Seq       int64     `json:"seq"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

func (s *Server) addLog(level, message string) {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	s.logSeq++
	entry := logEntry{Seq: s.logSeq, Timestamp: time.Now(), Level: level, Message: message}
	s.logs = append(s.logs, entry)
	const maxLogs = 200
	if len(s.logs) > maxLogs {
		s.logs = s.logs[len(s.logs)-maxLogs:]
	}
}

func (s *Server) clearLogs() {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	s.logs = nil
}

func (s *Server) StartScheduler(ctx context.Context) {
	s.schedulerMu.Lock()
	defer s.schedulerMu.Unlock()
	s.schedulerParent = ctx
	s.UpdateSchedulerLocked(s.store.Get())
}

func (s *Server) UpdateScheduler(cfg config.Config) {
	s.schedulerMu.Lock()
	defer s.schedulerMu.Unlock()
	s.UpdateSchedulerLocked(cfg)
}

func (s *Server) UpdateSchedulerLocked(cfg config.Config) {
	if s.schedulerParent == nil {
		return
	}

	enabled := cfg.SchedulerEnabled
	interval := time.Duration(cfg.ScanIntervalSec) * time.Second
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	// Stop scheduler if disabled.
	if !enabled {
		if s.schedulerCancel != nil {
			s.schedulerCancel()
			s.schedulerCancel = nil
		}
		if s.schedulerEnabled {
			s.addLog("info", "scheduler disabled")
		}
		s.schedulerEnabled = false
		s.schedulerInterval = 0
		return
	}

	// No change.
	if s.schedulerCancel != nil && s.schedulerEnabled && s.schedulerInterval == interval {
		return
	}

	// Restart if running with different interval.
	if s.schedulerCancel != nil {
		s.schedulerCancel()
		s.schedulerCancel = nil
	}

	runCtx, cancel := context.WithCancel(s.schedulerParent)
	s.schedulerCancel = cancel
	s.schedulerEnabled = true
	s.schedulerInterval = interval
	s.addLog("info", fmt.Sprintf("scheduler enabled (interval=%s)", interval))

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				s.addLog("info", "scheduled scan started")
				result, err := s.runScan(runCtx)
				if err != nil && !errors.Is(err, errScanInProgress) {
					s.addLog("error", fmt.Sprintf("scheduled scan failed: %v", err))
				}
				currentCfg := s.store.Get()
				if len(currentCfg.RemoteServers) > 0 {
					s.addLog("info", fmt.Sprintf("scheduled remote scans started: servers=%d", len(currentCfg.RemoteServers)))
					remoteCtx, cancel := context.WithTimeout(runCtx, 2*time.Minute)
					s.triggerRemoteScans(remoteCtx, currentCfg.RemoteServers)
					cancel()
					s.addLog("info", "scheduled remote scans finished")
				}
				if err == nil {
					s.addLog("info", fmt.Sprintf("scheduled scan finished: containers=%d", len(result.Containers)))
				}
			}
		}
	}()
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func upsertServer(servers []config.RemoteServer, input config.RemoteServer) []config.RemoteServer {
	for i, srv := range servers {
		if strings.EqualFold(srv.Name, input.Name) {
			servers[i] = input
			return servers
		}
	}
	return append(servers, input)
}

func removeServer(servers []config.RemoteServer, name string) []config.RemoteServer {
	filtered := make([]config.RemoteServer, 0, len(servers))
	for _, srv := range servers {
		if !strings.EqualFold(srv.Name, name) {
			filtered = append(filtered, srv)
		}
	}
	return filtered
}

func upsertLocal(servers []config.LocalServer, input config.LocalServer) []config.LocalServer {
	for i, srv := range servers {
		if strings.EqualFold(srv.Name, input.Name) {
			servers[i] = input
			return servers
		}
	}
	return append(servers, input)
}

func removeLocal(servers []config.LocalServer, name string) []config.LocalServer {
	filtered := make([]config.LocalServer, 0, len(servers))
	for _, srv := range servers {
		if !strings.EqualFold(srv.Name, name) {
			filtered = append(filtered, srv)
		}
	}
	return filtered
}

func findRemoteServer(servers []config.RemoteServer, name string) (config.RemoteServer, bool) {
	for _, srv := range servers {
		if strings.EqualFold(srv.Name, name) {
			return srv, true
		}
	}
	return config.RemoteServer{}, false
}

func isRemoteUpdateDisconnect(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "eof") || strings.Contains(msg, "connection reset") || strings.Contains(msg, "broken pipe")
}

func shortID(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func fetchRemoteVersion(ctx context.Context, client *http.Client, remote config.RemoteServer) (string, string) {
	if remote.URL == "" {
		return "unknown", "offline"
	}
	url := strings.TrimSuffix(remote.URL, "/") + "/api/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "unknown", "offline"
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "unknown", "offline"
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "unknown", "offline"
	}
	var payload struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "unknown", "offline"
	}
	version := strings.TrimSpace(payload.Version)
	if version == "" {
		version = "unknown"
	}
	return version, "online"
}

func findLocalServer(servers []config.LocalServer, name string) (config.LocalServer, bool) {
	for _, srv := range servers {
		if strings.EqualFold(srv.Name, name) {
			return srv, true
		}
	}
	return config.LocalServer{}, false
}

func dockerHostFromSocket(socket string) string {
	socket = strings.TrimSpace(socket)
	if socket == "" {
		return ""
	}
	if strings.HasPrefix(socket, "unix://") {
		return socket
	}
	return "unix://" + socket
}

func ResolveConfigPath() string {
	pathFromEnv := strings.TrimSpace(os.Getenv("CONTIWATCH_CONFIG"))
	if pathFromEnv != "" {
		return pathFromEnv
	}
	return path.Join("/data", "config.json")
}
