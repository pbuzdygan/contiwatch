package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
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
}

func New(store *config.Store, watcher *dockerwatcher.Watcher, agentMode bool, agentToken string) *Server {
	s := &Server{
		store:      store,
		watcher:    watcher,
		mux:        http.NewServeMux(),
		agentMode:  agentMode,
		agentToken: agentToken,
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
	s.mux.HandleFunc("/api/locals", s.handleLocals)
	s.mux.HandleFunc("/api/locals/", s.handleLocalByName)
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/scan", s.handleScan)
	s.mux.HandleFunc("/api/update/", s.handleUpdateContainer)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/aggregate", s.handleAggregate)

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
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
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

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.lastScanMutex.RLock()
	defer s.lastScanMutex.RUnlock()
	writeJSON(w, http.StatusOK, s.lastScan)
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
	s.addLog("info", "manual scan started")
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
	s.addLog("info", fmt.Sprintf("manual scan finished: containers=%d", len(result.Containers)))
	writeJSON(w, http.StatusOK, result)
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
	serverName := strings.TrimSpace(r.URL.Query().Get("server"))
	if serverName == "" {
		if len(cfg.LocalServers) == 1 {
			serverName = cfg.LocalServers[0].Name
		} else {
			writeError(w, http.StatusBadRequest, errors.New("server name required"))
			return
		}
	}
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

	result, err := watcher.UpdateContainer(r.Context(), containerID, cfg)
	if err != nil {
		s.addLog("error", fmt.Sprintf("update failed: %s: %v", containerID, err))
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	log.Printf("manual update: container=%s updated=%t previous=%s current=%s msg=%s", result.Name, result.Updated, result.PreviousState, result.CurrentState, result.Message)
	if result.Updated {
		s.addLog("info", fmt.Sprintf("updated %s (%s → %s)", result.Name, result.PreviousState, result.CurrentState))
	} else {
		s.addLog("info", fmt.Sprintf("update skipped %s: %s", result.Name, result.Message))
	}

	if s.discord != nil && result.Updated && discordNotificationsEnabled(cfg) {
		const discordBlue = 0x3498DB
		description := fmt.Sprintf("Updated: 1\n\nContainers updated:\n- %s\n\nState: %s → %s", result.Name, result.PreviousState, result.CurrentState)
		_ = s.discord.SendEmbed("Contiwatch updates applied", description, discordBlue)
	}

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
	localScans := s.lastScans
	s.lastScanMutex.RUnlock()
	for _, local := range localScans {
		if !local.CheckedAt.IsZero() {
			results = append(results, local)
		}
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
	scannedCount := 0
	updatesAvailableCount := 0
	readyToUpdateCount := 0
	skippedCount := 0
	updatedCount := 0
	containersWithUpdates := []string{}
	updatedContainers := []string{}

	if len(cfg.LocalServers) == 0 {
		s.addLog("info", "scan skipped: no local servers configured")
	} else {
		for _, local := range cfg.LocalServers {
			watcher, err := dockerwatcher.NewWithHost(dockerHostFromSocket(local.Socket))
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
			result, err := watcher.Scan(ctx, cfg)
			_ = watcher.Close()
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
			result.ServerName = local.Name
			result.ServerURL = local.Socket
			result.Local = true
			results = append(results, result)

			scannedCount += len(result.Containers)
			for _, container := range result.Containers {
				if container.Updated {
					updatedCount++
					updatedContainers = append(updatedContainers, container.Name)
				}
				if !container.UpdateAvailable {
					continue
				}
				updatesAvailableCount++
				containersWithUpdates = append(containersWithUpdates, container.Name)
				if strings.HasPrefix(container.Error, "skipped:") {
					skippedCount++
					continue
				}
				if container.Policy == config.PolicyUpdate && !container.Updated {
					readyToUpdateCount++
				}
			}
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

	if s.discord != nil && discordNotificationsEnabled(cfg) {
		if len(containersWithUpdates) > 0 || updatedCount > 0 {
			description := fmt.Sprintf(
				"Scanned images: %d\nUpdates available: %d\nReady to update: %d\nSkipped: %d\nUpdated: %d\n\nContainers with updates:\n- %s",
				scannedCount,
				updatesAvailableCount,
				readyToUpdateCount,
				skippedCount,
				updatedCount,
				strings.Join(containersWithUpdates, "\n- "),
			)
			if updatedCount > 0 {
				description += "\n\nContainers updated:\n- " + strings.Join(updatedContainers, "\n- ")
			}
			const discordBlue = 0x3498DB
			_ = s.discord.SendEmbed("Contiwatch updates:", description, discordBlue)
		}
	}

	if len(results) == 0 {
		return dockerwatcher.ScanResult{}, nil
	}
	return results[0], nil
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

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				_, _ = s.runScan(runCtx)
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
