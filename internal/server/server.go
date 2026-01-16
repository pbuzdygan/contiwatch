package server

import (
	"bytes"
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
	statePath     string
	lastScan      dockerwatcher.ScanResult
	lastScanMutex sync.RWMutex
	lastScans     []dockerwatcher.ScanResult
	scanStateMu   sync.RWMutex
	scanStates    map[string]scanState
	scanMutex     sync.Mutex
	scanRunning   bool
	updateRunning bool
	scanCancel    context.CancelFunc

	schedulerMu       sync.Mutex
	schedulerCancel   context.CancelFunc
	schedulerInterval time.Duration
	schedulerEnabled  bool
	schedulerParent   context.Context

	logMu  sync.RWMutex
	logs   []logEntry
	logSeq int64

	agentMode         bool
	agentToken        string
	version           string
	remoteScanRunning atomic.Int64
}

type scanState struct {
	State      string
	StartedAt  time.Time
	FinishedAt time.Time
}

const (
	scanStateIdle      = "idle"
	scanStatePending   = "pending"
	scanStateScanning  = "scanning"
	scanStateUpdating  = "updating"
	scanStateDone      = "done"
	scanStateCancelled = "cancelled"
	scanStateError     = "error"
)

func New(store *config.Store, watcher *dockerwatcher.Watcher, agentMode bool, agentToken string, version string) *Server {
	statePath := ""
	if store != nil {
		configPath := strings.TrimSpace(store.Path())
		if configPath != "" {
			statePath = path.Join(path.Dir(configPath), "scan_state.json")
		}
	}
	s := &Server{
		store:      store,
		watcher:    watcher,
		mux:        http.NewServeMux(),
		agentMode:  agentMode,
		agentToken: agentToken,
		version:    version,
		scanStates: map[string]scanState{},
		statePath:  statePath,
	}
	s.routes()
	s.loadScanState()
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

func scanKey(local bool, name string) string {
	if strings.TrimSpace(name) == "" {
		name = "unknown"
	}
	if local {
		return "local:" + name
	}
	return "remote:" + name
}

func (s *Server) setScanStateKey(key string, state string) {
	now := time.Now()
	s.scanStateMu.Lock()
	defer s.scanStateMu.Unlock()
	current := s.scanStates[key]
	current.State = state
	switch state {
	case scanStatePending, scanStateScanning, scanStateUpdating:
		current.StartedAt = now
		current.FinishedAt = time.Time{}
	case scanStateDone, scanStateCancelled, scanStateError:
		if current.StartedAt.IsZero() {
			current.StartedAt = now
		}
		current.FinishedAt = now
	case scanStateIdle:
		current.StartedAt = time.Time{}
		current.FinishedAt = time.Time{}
	}
	s.scanStates[key] = current
}

func (s *Server) setScanState(local bool, name string, state string) {
	s.setScanStateKey(scanKey(local, name), state)
}

func (s *Server) resetScanStates(cfg config.Config) {
	s.scanStateMu.Lock()
	defer s.scanStateMu.Unlock()
	if s.scanStates == nil {
		s.scanStates = map[string]scanState{}
	}
	for _, local := range cfg.LocalServers {
		s.scanStates[scanKey(true, local.Name)] = scanState{State: scanStateIdle}
	}
	for _, remote := range cfg.RemoteServers {
		s.scanStates[scanKey(false, remote.Name)] = scanState{State: scanStateIdle}
	}
}

func (s *Server) setAllScanStates(cfg config.Config, state string) {
	if state == scanStateIdle {
		s.resetScanStates(cfg)
		return
	}
	for _, local := range cfg.LocalServers {
		s.setScanState(true, local.Name, state)
	}
	for _, remote := range cfg.RemoteServers {
		s.setScanState(false, remote.Name, state)
	}
}

func (s *Server) cancelActiveScanStates() {
	s.scanStateMu.Lock()
	defer s.scanStateMu.Unlock()
	now := time.Now()
	for key, current := range s.scanStates {
		if current.State != scanStatePending && current.State != scanStateScanning && current.State != scanStateUpdating {
			continue
		}
		current.State = scanStateCancelled
		if current.StartedAt.IsZero() {
			current.StartedAt = now
		}
		current.FinishedAt = now
		s.scanStates[key] = current
	}
}

func (s *Server) applyScanState(result *dockerwatcher.ScanResult, local bool, name string) {
	key := scanKey(local, name)
	s.scanStateMu.RLock()
	state, ok := s.scanStates[key]
	s.scanStateMu.RUnlock()
	if !ok {
		return
	}
	if state.State == "" || state.State == scanStateIdle || state.State == scanStateDone {
		return
	}
	result.ScanState = state.State
	if !state.StartedAt.IsZero() {
		started := state.StartedAt
		result.ScanStartedAt = &started
	}
	if !state.FinishedAt.IsZero() {
		finished := state.FinishedAt
		result.ScanFinishedAt = &finished
	}
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
	s.mux.HandleFunc("/api/scan/stop", s.handleScanStop)
	s.mux.HandleFunc("/api/update/", s.handleUpdateContainer)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/notifications/test", s.handleNotificationsTest)
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
	case path == "/api/config":
		return true
	case path == "/api/scan":
		return true
	case path == "/api/scan/stop":
		return true
	case path == "/api/logs":
		return true
	case path == "/api/notifications/test":
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
		before := s.store.Get()
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
			if payload.DiscordNotifyOnStart != nil {
				cfg.DiscordNotifyOnStart = payload.DiscordNotifyOnStart
			}
			if payload.DiscordNotifyOnUpdateDetected != nil {
				cfg.DiscordNotifyOnUpdateDetected = payload.DiscordNotifyOnUpdateDetected
			}
			if payload.DiscordNotifyOnContainerUpdated != nil {
				cfg.DiscordNotifyOnContainerUpdated = payload.DiscordNotifyOnContainerUpdated
			}
			cfg.UpdateStoppedContainers = payload.UpdateStoppedContainers
			cfg.PruneDanglingImages = payload.PruneDanglingImages
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		changes := []string{}
		if before.SchedulerEnabled != updated.SchedulerEnabled {
			changes = append(changes, fmt.Sprintf("scheduler_enabled=%t", updated.SchedulerEnabled))
		}
		if before.ScanIntervalSec != updated.ScanIntervalSec {
			changes = append(changes, fmt.Sprintf("scan_interval_sec=%d", updated.ScanIntervalSec))
		}
		if before.GlobalPolicy != updated.GlobalPolicy {
			changes = append(changes, fmt.Sprintf("global_policy=%s→%s", before.GlobalPolicy, updated.GlobalPolicy))
		}
		if before.DiscordNotificationsEnabled == nil ||
			updated.DiscordNotificationsEnabled == nil ||
			*before.DiscordNotificationsEnabled != *updated.DiscordNotificationsEnabled {
			enabled := updated.DiscordNotificationsEnabled != nil && *updated.DiscordNotificationsEnabled
			changes = append(changes, fmt.Sprintf("discord_notifications_enabled=%t", enabled))
		}
		if before.DiscordNotifyOnStart == nil ||
			updated.DiscordNotifyOnStart == nil ||
			*before.DiscordNotifyOnStart != *updated.DiscordNotifyOnStart {
			enabled := updated.DiscordNotifyOnStart != nil && *updated.DiscordNotifyOnStart
			changes = append(changes, fmt.Sprintf("discord_notify_on_start=%t", enabled))
		}
		if before.DiscordNotifyOnUpdateDetected == nil ||
			updated.DiscordNotifyOnUpdateDetected == nil ||
			*before.DiscordNotifyOnUpdateDetected != *updated.DiscordNotifyOnUpdateDetected {
			enabled := updated.DiscordNotifyOnUpdateDetected != nil && *updated.DiscordNotifyOnUpdateDetected
			changes = append(changes, fmt.Sprintf("discord_notify_on_update_detected=%t", enabled))
		}
		if before.DiscordNotifyOnContainerUpdated == nil ||
			updated.DiscordNotifyOnContainerUpdated == nil ||
			*before.DiscordNotifyOnContainerUpdated != *updated.DiscordNotifyOnContainerUpdated {
			enabled := updated.DiscordNotifyOnContainerUpdated != nil && *updated.DiscordNotifyOnContainerUpdated
			changes = append(changes, fmt.Sprintf("discord_notify_on_container_updated=%t", enabled))
		}
		if before.UpdateStoppedContainers != updated.UpdateStoppedContainers {
			changes = append(changes, fmt.Sprintf("update_stopped_containers=%t", updated.UpdateStoppedContainers))
		}
		if before.PruneDanglingImages != updated.PruneDanglingImages {
			changes = append(changes, fmt.Sprintf("prune_dangling_images=%t", updated.PruneDanglingImages))
		}
		if before.DiscordWebhookURL != updated.DiscordWebhookURL {
			state := "set"
			if updated.DiscordWebhookURL == "" {
				state = "cleared"
			}
			changes = append(changes, fmt.Sprintf("discord_webhook_url=%s", state))
		}
		if len(changes) > 0 {
			s.addLog("info", fmt.Sprintf("settings updated: %s", strings.Join(changes, ", ")))
		}
		s.UpdateDiscord(updated.DiscordWebhookURL)
		s.UpdateScheduler(updated)
		if before.GlobalPolicy != updated.GlobalPolicy && len(updated.RemoteServers) > 0 {
			go s.syncRemotePolicies(updated, updated.RemoteServers)
		}
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
		cfg := s.store.Get()
		if localNameExists(cfg.LocalServers, payload.Name) {
			writeError(w, http.StatusBadRequest, errors.New("name already used by local server"))
			return
		}
		updated, err := s.store.Update(func(cfg *config.Config) {
			cfg.RemoteServers = upsertServer(cfg.RemoteServers, payload)
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if remote, ok := findRemoteServer(updated.RemoteServers, payload.Name); ok {
			go s.syncRemotePolicies(updated, []config.RemoteServer{remote})
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
		cfg := s.store.Get()
		if remoteNameExists(cfg.RemoteServers, payload.Name) {
			writeError(w, http.StatusBadRequest, errors.New("name already used by remote server"))
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
		status := localSocketStatus(local.Socket)
		items = append(items, serverInfo{
			Type:    "local",
			Name:    local.Name,
			Address: local.Socket,
			Version: s.version,
			Status:  status,
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

func (s *Server) handleScanStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.scanMutex.Lock()
	cancel := s.scanCancel
	active := s.scanRunning || s.remoteScanRunning.Load() > 0
	s.scanMutex.Unlock()
	if cancel == nil || !active {
		writeJSON(w, http.StatusOK, map[string]string{"status": "idle"})
		return
	}
	cancel()
	s.cancelActiveScanStates()
	s.addLog("warn", "scan cancelled manually")
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopping"})
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
	case http.MethodPost:
		var payload struct {
			Level   string `json:"level"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		level := strings.TrimSpace(payload.Level)
		if level == "" {
			level = "info"
		}
		message := strings.TrimSpace(payload.Message)
		if message == "" {
			writeError(w, http.StatusBadRequest, errors.New("message is required"))
			return
		}
		s.addLog(level, message)
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case http.MethodDelete:
		s.clearLogs()
		writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNotificationsTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	webhookURL := strings.TrimSpace(payload.WebhookURL)
	if webhookURL == "" {
		writeError(w, http.StatusBadRequest, errors.New("webhook_url is required"))
		return
	}
	client := notify.NewDiscordClient(webhookURL)
	const discordBlue = 0x3498DB
	if err := client.SendEmbed("Contiwatch test", "Webhook verified.", discordBlue); err != nil {
		s.addLog("warn", fmt.Sprintf("Webhook test - Failed: %v", err))
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.addLog("info", "Webhook test - Success")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	scope := strings.TrimSpace(r.URL.Query().Get("server"))
	cfg := s.store.Get()
	s.resetScanStates(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	s.scanMutex.Lock()
	s.scanCancel = cancel
	s.scanMutex.Unlock()
	cleanup := true
	defer func() {
		if cleanup {
			cancel()
			s.scanMutex.Lock()
			s.scanCancel = nil
			s.scanMutex.Unlock()
		}
	}()

	s.addLog("info", "manual scan started")
	switch {
	case scope == "" || scope == "all":
		s.setAllScanStates(cfg, scanStatePending)
		result, err := s.runScan(ctx)
		if err != nil {
			if errors.Is(err, errScanInProgress) {
				writeError(w, http.StatusConflict, err)
				return
			}
			s.addLog("error", fmt.Sprintf("manual scan failed: %v", err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		cleanup = false
		remoteCtx, remoteCancel := context.WithTimeout(ctx, 2*time.Minute)
		go func() {
			defer remoteCancel()
			s.triggerRemoteScans(remoteCtx, cfg.RemoteServers)
			s.scanMutex.Lock()
			s.scanCancel = nil
			s.scanMutex.Unlock()
			cancel()
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
		s.setScanState(true, local.Name, scanStateScanning)
		s.scanMutex.Lock()
		if s.scanRunning || s.updateRunning {
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
		s.addLog("info", fmt.Sprintf("local scan started: %s", local.Name))
		result, err := s.scanLocalServer(ctx, cfg, local, true)
		if err != nil {
			if ctx.Err() != nil {
				s.setScanState(true, local.Name, scanStateCancelled)
				result.Error = "scan cancelled manually"
				s.updateLastScans(result)
				writeJSON(w, http.StatusOK, result)
				return
			}
			s.setScanState(true, local.Name, scanStateError)
			s.addLog("error", fmt.Sprintf("manual scan failed: %s: %v", local.Name, err))
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if result.Error != "" {
			s.setScanState(true, local.Name, scanStateError)
			s.updateLastScans(result)
			writeJSON(w, http.StatusOK, result)
			return
		}
		s.addLog("info", fmt.Sprintf("local scan finished: %s %s", local.Name, s.formatScanSummary(result)))
		if !s.agentMode {
			updatedResult, _, err := s.autoUpdateLocal(ctx, cfg, local, result)
			if err != nil {
				if ctx.Err() != nil {
					s.setScanState(true, local.Name, scanStateCancelled)
				} else {
					s.setScanState(true, local.Name, scanStateError)
				}
				s.addLog("error", fmt.Sprintf("local update failed: %s: %v", local.Name, err))
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			result = updatedResult
		}
		s.sendScanNotification(cfg, result)
		s.setScanState(true, local.Name, scanStateDone)
		s.updateLastScans(result)
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
		if _, err := s.syncRemotePolicy(ctx, cfg, remote); err != nil {
			s.addLog("warn", fmt.Sprintf("remote policy sync failed: %s: %v", remote.Name, err))
		}
		s.addLog("info", fmt.Sprintf("remote scan started: %s", remote.Name))
		s.setScanState(false, remote.Name, scanStateScanning)
		s.remoteScanRunning.Add(1)
		result, err := s.scanRemoteServer(ctx, remote)
		s.remoteScanRunning.Add(-1)
		if err != nil {
			if ctx.Err() != nil {
				s.setScanState(false, remote.Name, scanStateCancelled)
				result = dockerwatcher.ScanResult{
					ServerName: remote.Name,
					ServerURL:  remote.URL,
					Local:      false,
					CheckedAt:  time.Now(),
					Error:      "scan cancelled manually",
				}
				writeJSON(w, http.StatusOK, result)
				return
			}
			s.setScanState(false, remote.Name, scanStateError)
			s.addLog("error", fmt.Sprintf("manual scan failed: %s: %v", remote.Name, err))
			writeError(w, http.StatusBadGateway, err)
			return
		}
		s.addLog("info", fmt.Sprintf("remote scan finished: %s %s", remote.Name, s.formatScanSummary(result)))
		s.sendScanNotification(cfg, result)
		if _, err := s.autoUpdateRemote(ctx, cfg, remote, result); err != nil {
			if ctx.Err() != nil {
				s.setScanState(false, remote.Name, scanStateCancelled)
			} else {
				s.setScanState(false, remote.Name, scanStateError)
			}
			s.addLog("error", fmt.Sprintf("remote update failed: %s: %v", remote.Name, err))
			writeError(w, http.StatusBadGateway, err)
			return
		}
		s.setScanState(false, remote.Name, scanStateDone)
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
	if s.scanRunning || s.updateRunning {
		s.scanMutex.Unlock()
		writeError(w, http.StatusConflict, errors.New("another operation is in progress"))
		return
	}
	s.updateRunning = true
	s.scanMutex.Unlock()
	defer func() {
		s.scanMutex.Lock()
		s.updateRunning = false
		s.scanMutex.Unlock()
	}()

	cfg := s.store.Get()
	if pruneOverride := strings.TrimSpace(r.URL.Query().Get("prune")); pruneOverride != "" {
		switch strings.ToLower(pruneOverride) {
		case "1", "true", "yes", "on":
			cfg.PruneDanglingImages = true
		default:
			cfg.PruneDanglingImages = false
		}
	}
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
		if err := s.watcher.TriggerSelfUpdateWithLogs(updateCtx, containerID, func(level, message string) {
			s.addLog(level, fmt.Sprintf("self-update helper: %s", message))
		}); err != nil {
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
		result, err = s.updateRemoteContainer(updateCtx, remote, containerID, cfg.PruneDanglingImages)
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
	s.logUpdateResult(cfg, serverLabel, serverScope, result, isRemote)

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
		s.applyScanState(&local, true, local.ServerName)
		localByName[local.ServerName] = local
	}
	for _, local := range cfg.LocalServers {
		if existing, ok := localByName[local.Name]; ok {
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

	client := &http.Client{Timeout: 5 * time.Second}
	for _, remote := range cfg.RemoteServers {
		remoteResult := dockerwatcher.ScanResult{ServerName: remote.Name, ServerURL: remote.URL, Local: false}
		if remote.URL == "" {
			remoteResult.Error = "missing url"
			s.applyScanState(&remoteResult, false, remote.Name)
			results = append(results, remoteResult)
			continue
		}
		url := strings.TrimSuffix(remote.URL, "/") + "/api/status"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			remoteResult.Error = "invalid url"
			remoteResult.CheckedAt = time.Now()
			s.applyScanState(&remoteResult, false, remote.Name)
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
			s.applyScanState(&remoteResult, false, remote.Name)
			results = append(results, remoteResult)
			continue
		}
		if resp.Body == nil {
			remoteResult.Error = "empty response"
			remoteResult.CheckedAt = time.Now()
			s.applyScanState(&remoteResult, false, remote.Name)
			results = append(results, remoteResult)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			remoteResult.Error = fmt.Sprintf("status %d", resp.StatusCode)
			_ = resp.Body.Close()
			remoteResult.CheckedAt = time.Now()
			s.applyScanState(&remoteResult, false, remote.Name)
			results = append(results, remoteResult)
			continue
		}
		var payload dockerwatcher.ScanResult
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			remoteResult.Error = "invalid payload"
			_ = resp.Body.Close()
			remoteResult.CheckedAt = time.Now()
			s.applyScanState(&remoteResult, false, remote.Name)
			results = append(results, remoteResult)
			continue
		}
		_ = resp.Body.Close()
		payload.ServerName = remote.Name
		payload.ServerURL = remote.URL
		payload.Local = false
		s.applyScanState(&payload, false, remote.Name)
		results = append(results, payload)
	}

	writeJSON(w, http.StatusOK, results)
}

var errScanInProgress = errors.New("scan already in progress")

func (s *Server) runScan(ctx context.Context) (dockerwatcher.ScanResult, error) {
	s.scanMutex.Lock()
	if s.scanRunning || s.updateRunning {
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
		for i, local := range cfg.LocalServers {
			if ctx.Err() != nil {
				for _, remaining := range cfg.LocalServers[i:] {
					s.setScanState(true, remaining.Name, scanStateCancelled)
					results = append(results, dockerwatcher.ScanResult{
						ServerName: remaining.Name,
						ServerURL:  remaining.Socket,
						Local:      true,
						CheckedAt:  time.Now(),
						Error:      "scan cancelled manually (not started)",
					})
				}
				break
			}
			s.addLog("info", fmt.Sprintf("local scan started: %s", local.Name))
			s.setScanState(true, local.Name, scanStateScanning)
			result, err := s.scanLocalServer(ctx, cfg, local, true)
			if err != nil {
				if ctx.Err() != nil {
					s.setScanState(true, local.Name, scanStateCancelled)
				} else {
					s.setScanState(true, local.Name, scanStateError)
				}
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
			if result.Error != "" {
				s.setScanState(true, local.Name, scanStateError)
				results = append(results, result)
				continue
			}
			s.addLog("info", fmt.Sprintf("local scan finished: %s %s", local.Name, s.formatScanSummary(result)))
			if !s.agentMode {
				updatedResult, _, err := s.autoUpdateLocal(ctx, cfg, local, result)
				if err != nil {
					if ctx.Err() != nil {
						s.setScanState(true, local.Name, scanStateCancelled)
					} else {
						s.setScanState(true, local.Name, scanStateError)
					}
					s.addLog("error", fmt.Sprintf("local update failed: %s: %v", local.Name, err))
					results = append(results, dockerwatcher.ScanResult{
						ServerName: local.Name,
						ServerURL:  local.Socket,
						Local:      true,
						CheckedAt:  time.Now(),
						Error:      err.Error(),
					})
					continue
				}
				result = updatedResult
			}
			s.setScanState(true, local.Name, scanStateDone)
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
		go s.saveScanState()
	} else {
		s.lastScanMutex.Lock()
		s.lastScan = dockerwatcher.ScanResult{}
		s.lastScans = nil
		s.lastScanMutex.Unlock()
		go s.saveScanState()
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
	go s.saveScanState()
}

func (s *Server) scanLocalServer(ctx context.Context, cfg config.Config, local config.LocalServer, scanOnly bool) (dockerwatcher.ScanResult, error) {
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	defer watcher.Close()
	var result dockerwatcher.ScanResult
	if scanOnly {
		result, err = watcher.ScanNoUpdate(ctx, cfg)
	} else {
		result, err = watcher.Scan(ctx, cfg)
	}
	if err != nil {
		return dockerwatcher.ScanResult{}, err
	}
	result.ServerName = local.Name
	result.ServerURL = local.Socket
	result.Local = true
	return result, nil
}

type scanStateFile struct {
	LastScans []dockerwatcher.ScanResult `json:"last_scans"`
}

func (s *Server) loadScanState() {
	if s.statePath == "" {
		return
	}
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("scan state load failed: %v", err)
		}
		return
	}
	var payload scanStateFile
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("scan state decode failed: %v", err)
		return
	}
	if len(payload.LastScans) == 0 {
		return
	}
	localScans := make([]dockerwatcher.ScanResult, 0, len(payload.LastScans))
	for _, scan := range payload.LastScans {
		if scan.Local {
			localScans = append(localScans, scan)
		}
	}
	if len(localScans) == 0 {
		return
	}
	s.lastScanMutex.Lock()
	s.lastScans = localScans
	s.lastScan = latestScan(localScans)
	s.lastScanMutex.Unlock()
}

func (s *Server) saveScanState() {
	if s.statePath == "" {
		return
	}
	s.lastScanMutex.RLock()
	scans := append([]dockerwatcher.ScanResult(nil), s.lastScans...)
	s.lastScanMutex.RUnlock()
	localScans := make([]dockerwatcher.ScanResult, 0, len(scans))
	for _, scan := range scans {
		if scan.Local {
			localScans = append(localScans, slimScanResult(scan))
		}
	}
	if len(localScans) == 0 {
		_ = os.Remove(s.statePath)
		return
	}
	payload := scanStateFile{LastScans: localScans}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("scan state encode failed: %v", err)
		return
	}
	tmp := s.statePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		log.Printf("scan state write failed: %v", err)
		return
	}
	if err := os.Rename(tmp, s.statePath); err != nil {
		log.Printf("scan state rename failed: %v", err)
	}
}

func latestScan(scans []dockerwatcher.ScanResult) dockerwatcher.ScanResult {
	if len(scans) == 0 {
		return dockerwatcher.ScanResult{}
	}
	latest := scans[0]
	for _, scan := range scans[1:] {
		if scan.CheckedAt.After(latest.CheckedAt) {
			latest = scan
		}
	}
	return latest
}

func slimScanResult(scan dockerwatcher.ScanResult) dockerwatcher.ScanResult {
	scan.Containers = nil
	return scan
}

func localSocketStatus(socket string) string {
	socket = strings.TrimSpace(socket)
	if socket == "" {
		return "offline"
	}
	info, err := os.Stat(socket)
	if err != nil {
		return "offline"
	}
	if info.Mode()&os.ModeSocket == 0 {
		return "offline"
	}
	return "online"
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

func (s *Server) updateRemoteContainer(ctx context.Context, remote config.RemoteServer, containerID string, prune bool) (dockerwatcher.UpdateResult, error) {
	if remote.URL == "" {
		return dockerwatcher.UpdateResult{}, errors.New("missing url")
	}
	updateURL := strings.TrimSuffix(remote.URL, "/") + "/api/update/" + url.PathEscape(containerID)
	if prune {
		updateURL += "?prune=1"
	}
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
	for i, remote := range remotes {
		if ctx.Err() != nil {
			for _, remaining := range remotes[i:] {
				s.setScanState(false, remaining.Name, scanStateCancelled)
			}
			break
		}
		cfg := s.store.Get()
		if _, err := s.syncRemotePolicy(ctx, cfg, remote); err != nil {
			s.addLog("warn", fmt.Sprintf("remote policy sync failed: %s: %v", remote.Name, err))
		}
		s.setScanState(false, remote.Name, scanStateScanning)
		s.addLog("info", fmt.Sprintf("remote scan started: %s", remote.Name))
		result, err := s.scanRemoteServer(ctx, remote)
		if err != nil {
			if ctx.Err() != nil {
				s.setScanState(false, remote.Name, scanStateCancelled)
			} else {
				s.setScanState(false, remote.Name, scanStateError)
			}
			s.addLog("error", fmt.Sprintf("remote scan failed: %s: %v", remote.Name, err))
			failCount++
			continue
		}
		s.addLog("info", fmt.Sprintf("remote scan finished: %s %s", remote.Name, s.formatScanSummary(result)))
		s.sendScanNotification(cfg, result)
		if _, err := s.autoUpdateRemote(ctx, cfg, remote, result); err != nil {
			if ctx.Err() != nil {
				s.setScanState(false, remote.Name, scanStateCancelled)
			} else {
				s.setScanState(false, remote.Name, scanStateError)
			}
			s.addLog("error", fmt.Sprintf("remote update failed: %s: %v", remote.Name, err))
			failCount++
			continue
		}
		s.setScanState(false, remote.Name, scanStateDone)
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
	total        int
	updates      int
	ready        int
	skipped      int
	updated      int
	updateNames  []string
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
	notifyUpdates := cfg.DiscordNotifyOnUpdateDetected != nil && *cfg.DiscordNotifyOnUpdateDetected
	notifyUpdated := cfg.DiscordNotifyOnContainerUpdated != nil && *cfg.DiscordNotifyOnContainerUpdated
	summary := s.buildScanSummary(result)
	if (!notifyUpdates || summary.updates == 0) && (!notifyUpdated || summary.updated == 0) {
		return
	}
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

	lines := []string{
		fmt.Sprintf("Server: %s (%s)", serverLabel, scope),
		fmt.Sprintf("Scanned images: %d", summary.total),
		fmt.Sprintf("Skipped: %d", summary.skipped),
	}
	if notifyUpdates {
		lines = append(lines,
			fmt.Sprintf("Updates available: %d", summary.updates),
			fmt.Sprintf("Ready to update: %d", summary.ready),
		)
	}
	if notifyUpdated {
		lines = append(lines, fmt.Sprintf("Updated: %d", summary.updated))
	}
	description := strings.Join(lines, "\n")
	if notifyUpdates && summary.updates > 0 {
		description += "\n\nContainers with updates:\n- " + strings.Join(summary.updateNames, "\n- ")
	}
	if notifyUpdated && summary.updated > 0 {
		description += "\n\nContainers updated:\n- " + strings.Join(summary.updatedNames, "\n- ")
	}
	const discordBlue = 0x3498DB
	_ = s.discord.SendEmbed("Contiwatch updates:", description, discordBlue)
}

func (s *Server) sendUpdateNotification(cfg config.Config, serverName string, result dockerwatcher.UpdateResult) {
	if s.discord == nil || !discordNotificationsEnabled(cfg) {
		return
	}
	if cfg.DiscordNotifyOnContainerUpdated == nil || !*cfg.DiscordNotifyOnContainerUpdated {
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

func (s *Server) logUpdateResult(cfg config.Config, serverLabel, serverScope string, result dockerwatcher.UpdateResult, isRemote bool) {
	if result.Updated {
		s.addLog("info", fmt.Sprintf("updated %s (%s → %s) on %s (%s)", result.Name, result.PreviousState, result.CurrentState, serverLabel, serverScope))
		if strings.Contains(result.Message, "prune") {
			if strings.Contains(result.Message, "prune failed") {
				s.addLog("error", fmt.Sprintf("prune dangling images failed after update: %s on %s (%s)", result.Name, serverLabel, serverScope))
			} else {
				s.addLog("info", fmt.Sprintf("pruned dangling images after update: %s on %s (%s) %s", result.Name, serverLabel, serverScope, result.Message))
			}
		} else if cfg.PruneDanglingImages && isRemote {
			s.addLog("warn", fmt.Sprintf("prune not reported by agent: %s on %s (%s)", result.Name, serverLabel, serverScope))
		}
	} else {
		s.addLog("info", fmt.Sprintf("update skipped %s on %s (%s): %s", result.Name, serverLabel, serverScope, result.Message))
	}
	s.sendUpdateNotification(cfg, serverLabel, result)
}

func (s *Server) autoUpdateRemote(ctx context.Context, cfg config.Config, remote config.RemoteServer, result dockerwatcher.ScanResult) (int, error) {
	targets := []dockerwatcher.ContainerStatus{}
	for _, container := range result.Containers {
		if !container.UpdateAvailable {
			continue
		}
		if container.Policy != config.PolicyUpdate {
			continue
		}
		if container.Error != "" && !strings.HasPrefix(container.Error, "skipped:") {
			continue
		}
		targets = append(targets, container)
	}
	if len(targets) == 0 {
		return 0, nil
	}
	s.setScanState(false, remote.Name, scanStateUpdating)
	s.addLog("info", fmt.Sprintf("remote update started: %s containers=%d", remote.Name, len(targets)))
	updatedCount := 0
	for _, container := range targets {
		if ctx.Err() != nil {
			return updatedCount, ctx.Err()
		}
		updateResult, err := s.updateRemoteContainer(ctx, remote, container.ID, cfg.PruneDanglingImages)
		if err != nil {
			if isRemoteUpdateDisconnect(err) {
				updateResult = dockerwatcher.UpdateResult{
					ID:      container.ID,
					Name:    shortID(container.ID),
					Updated: false,
					Message: "update triggered; agent restarting",
				}
				s.addLog("warn", fmt.Sprintf("update connection closed: %s (agent restarting)", container.ID))
			} else {
				s.addLog("error", fmt.Sprintf("update failed: %s: %v", container.ID, err))
				continue
			}
		}
		s.logUpdateResult(cfg, remote.Name, "remote", updateResult, true)
		if updateResult.Updated {
			updatedCount++
		}
	}
	s.addLog("info", fmt.Sprintf("remote update finished: %s updated=%d", remote.Name, updatedCount))
	return updatedCount, nil
}

func (s *Server) syncRemotePolicies(cfg config.Config, remotes []config.RemoteServer) {
	for _, remote := range remotes {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		changed, err := s.syncRemotePolicy(ctx, cfg, remote)
		cancel()
		if err != nil {
			s.addLog("warn", fmt.Sprintf("remote policy sync failed: %s: %v", remote.Name, err))
			continue
		}
		if changed {
			s.addLog("info", fmt.Sprintf("remote policy synced: %s policy=%s", remote.Name, cfg.GlobalPolicy))
		}
	}
}

func (s *Server) syncRemotePolicy(ctx context.Context, cfg config.Config, remote config.RemoteServer) (bool, error) {
	if remote.URL == "" {
		return false, errors.New("missing url")
	}
	configURL := strings.TrimSuffix(remote.URL, "/") + "/api/config"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configURL, nil)
	if err != nil {
		return false, err
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload config.Config
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false, err
	}
	if payload.GlobalPolicy == cfg.GlobalPolicy {
		return false, nil
	}
	payload.GlobalPolicy = cfg.GlobalPolicy
	body, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}
	updateReq, err := http.NewRequestWithContext(ctx, http.MethodPut, configURL, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	updateReq.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		updateReq.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	updateResp, err := client.Do(updateReq)
	if err != nil {
		return false, err
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode < 200 || updateResp.StatusCode >= 300 {
		return false, fmt.Errorf("status %d", updateResp.StatusCode)
	}
	return true, nil
}

func (s *Server) autoUpdateLocal(ctx context.Context, cfg config.Config, local config.LocalServer, result dockerwatcher.ScanResult) (dockerwatcher.ScanResult, int, error) {
	targets := []dockerwatcher.ContainerStatus{}
	for _, container := range result.Containers {
		if !container.UpdateAvailable {
			continue
		}
		if container.Policy != config.PolicyUpdate {
			continue
		}
		if container.Error != "" && !strings.HasPrefix(container.Error, "skipped:") {
			continue
		}
		targets = append(targets, container)
	}
	if len(targets) == 0 {
		return result, 0, nil
	}
	watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
	if err != nil {
		return result, 0, err
	}
	defer watcher.Close()
	s.setScanState(true, local.Name, scanStateUpdating)
	s.addLog("info", fmt.Sprintf("local update started: %s containers=%d", local.Name, len(targets)))
	updatedCount := 0
	for _, container := range targets {
		if ctx.Err() != nil {
			return result, updatedCount, ctx.Err()
		}
		updateResult, err := watcher.UpdateContainer(ctx, container.ID, cfg)
		if err != nil {
			s.addLog("error", fmt.Sprintf("update failed: %s: %v", container.ID, err))
			continue
		}
		s.logUpdateResult(cfg, local.Name, "local", updateResult, false)
		if updateResult.Updated {
			updatedCount++
		}
		updateScanResultContainer(&result, container.ID, updateResult)
	}
	s.addLog("info", fmt.Sprintf("local update finished: %s updated=%d", local.Name, updatedCount))
	return result, updatedCount, nil
}

func updateScanResultContainer(result *dockerwatcher.ScanResult, containerID string, updateResult dockerwatcher.UpdateResult) {
	if result == nil {
		return
	}
	for i := range result.Containers {
		if result.Containers[i].ID != containerID {
			continue
		}
		result.Containers[i].LastChecked = time.Now()
		result.Containers[i].Error = ""
		result.Containers[i].Updated = updateResult.Updated
		if updateResult.Updated {
			result.Containers[i].UpdateAvailable = false
		}
		return
	}
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
				currentCfg := s.store.Get()
				s.setAllScanStates(currentCfg, scanStatePending)
				s.addLog("info", "scheduled scan started")
				result, err := s.runScan(runCtx)
				if err != nil && !errors.Is(err, errScanInProgress) {
					s.addLog("error", fmt.Sprintf("scheduled scan failed: %v", err))
				}
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

func remoteNameExists(servers []config.RemoteServer, name string) bool {
	for _, srv := range servers {
		if strings.EqualFold(srv.Name, name) {
			return true
		}
	}
	return false
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

func localNameExists(servers []config.LocalServer, name string) bool {
	for _, srv := range servers {
		if strings.EqualFold(srv.Name, name) {
			return true
		}
	}
	return false
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
