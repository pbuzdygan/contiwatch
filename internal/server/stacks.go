package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

const (
	stacksBaseDir         = "/data/stacks"
	stackActionTimeoutSec = 180
)

var stackNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type stackSummary struct {
	Name              string `json:"name"`
	Status            string `json:"status"`
	ContainersTotal   int    `json:"containers_total"`
	ContainersRunning int    `json:"containers_running"`
	ContainersStopped int    `json:"containers_stopped"`
	UpdatedAt         string `json:"updated_at"`
}

type stacksResponse struct {
	Scope  string         `json:"scope"`
	Stacks []stackSummary `json:"stacks"`
	Error  string         `json:"error,omitempty"`
}

type stackDetailResponse struct {
	Scope      string `json:"scope"`
	Name       string `json:"name"`
	ComposeYml string `json:"compose_yaml"`
	Env        string `json:"env"`
	HasEnv     bool   `json:"has_env"`
	UpdatedAt  string `json:"updated_at"`
}

type stackSaveRequest struct {
	Scope      string `json:"scope"`
	Name       string `json:"name"`
	ComposeYml string `json:"compose_yaml"`
	Env        string `json:"env"`
	UseEnv     bool   `json:"use_env"`
}

type stackActionRequest struct {
	Scope      string `json:"scope"`
	Name       string `json:"name"`
	Action     string `json:"action"`
	ComposeYml string `json:"compose_yaml"`
	Env        string `json:"env"`
	UseEnv     bool   `json:"use_env"`
}

type stackActionResponse struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

type stackValidateRequest struct {
	ComposeYml string `json:"compose_yaml"`
	Env        string `json:"env"`
	UseEnv     bool   `json:"use_env"`
}

type stackValidateResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

func (s *Server) handleStacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.agentMode {
		writeError(w, http.StatusForbidden, errors.New("stacks are disabled in agent mode"))
		return
	}
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	if scope == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope is required"))
		return
	}
	serverType, serverName, err := parseScope(scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateStackSegment(serverName, "server"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	stacks, err := listStacksOnDisk(serverName)
	if err != nil {
		writeJSON(w, http.StatusOK, stacksResponse{Scope: scope, Error: err.Error(), Stacks: []stackSummary{}})
		return
	}
	cfg := s.store.Get()
	containers, contErr := s.listContainersByScope(cfg, serverType, serverName)
	list := buildStackSummaries(stacks, containers)
	if contErr != nil {
		writeJSON(w, http.StatusOK, stacksResponse{Scope: scope, Error: contErr.Error(), Stacks: list})
		return
	}
	writeJSON(w, http.StatusOK, stacksResponse{Scope: scope, Stacks: list})
}

func (s *Server) handleStackGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.agentMode {
		writeError(w, http.StatusForbidden, errors.New("stacks are disabled in agent mode"))
		return
	}
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if scope == "" || name == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope and name are required"))
		return
	}
	serverType, serverName, err := parseScope(scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateStackSegment(serverName, "server"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateStackSegment(name, "stack"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "" {
		writeError(w, http.StatusBadRequest, errors.New("invalid scope"))
		return
	}

	detail, err := loadStackDetail(serverName, name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	detail.Scope = scope
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleStackSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.agentMode {
		writeError(w, http.StatusForbidden, errors.New("stacks are disabled in agent mode"))
		return
	}
	var payload stackSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Scope = strings.TrimSpace(payload.Scope)
	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Scope == "" || payload.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope and name are required"))
		return
	}
	if strings.TrimSpace(payload.ComposeYml) == "" {
		writeError(w, http.StatusBadRequest, errors.New("compose_yaml is required"))
		return
	}
	if payload.UseEnv {
		if err := validateEnvContent(payload.Env); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}
	serverType, serverName, err := parseScope(payload.Scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "" {
		writeError(w, http.StatusBadRequest, errors.New("invalid scope"))
		return
	}
	if err := validateStackSegment(serverName, "server"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateStackSegment(payload.Name, "stack"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := saveStackFiles(serverName, payload.Name, payload.ComposeYml, payload.Env, payload.UseEnv); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"name":     payload.Name,
		"saved_at": time.Now().Format(time.RFC3339),
		"server":   serverName,
		"scope":    payload.Scope,
	})
}

func (s *Server) handleStackValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.agentMode {
		writeError(w, http.StatusForbidden, errors.New("stacks are disabled in agent mode"))
		return
	}
	var payload stackValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(payload.ComposeYml) == "" {
		writeJSON(w, http.StatusOK, stackValidateResponse{Valid: false, Error: "compose_yaml is required"})
		return
	}
	if payload.UseEnv {
		if err := validateEnvContent(payload.Env); err != nil {
			writeJSON(w, http.StatusOK, stackValidateResponse{Valid: false, Error: err.Error()})
			return
		}
	}
	if err := runComposeConfigFromPayload(payload); err != nil {
		writeJSON(w, http.StatusOK, stackValidateResponse{Valid: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stackValidateResponse{Valid: true})
}

func (s *Server) handleStackAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload stackActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Action = strings.TrimSpace(strings.ToLower(payload.Action))
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Scope = strings.TrimSpace(payload.Scope)
	if payload.Name == "" || payload.Action == "" {
		writeError(w, http.StatusBadRequest, errors.New("name and action are required"))
		return
	}
	if err := validateStackSegment(payload.Name, "stack"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	cfg := s.store.Get()
	if s.agentMode {
		if strings.TrimSpace(payload.ComposeYml) == "" {
			writeError(w, http.StatusBadRequest, errors.New("compose_yaml is required"))
			return
		}
		sanitized, invalid := sanitizeEnvContent(payload.Env)
		payload.Env = sanitized
		if payload.UseEnv && len(invalid) > 0 {
			s.addLog("warn", fmt.Sprintf("stack %s env invalid lines ignored: %s", payload.Name, strings.Join(invalid, ", ")))
		}
		dockerHost := ""
		if len(cfg.LocalServers) == 1 {
			dockerHost = dockerHostFromSocket(cfg.LocalServers[0].Socket)
		}
		if err := runComposeFromPayload(payload, dockerHost); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, stackActionResponse{Name: payload.Name, Action: payload.Action})
		return
	}

	if payload.Scope == "" {
		writeError(w, http.StatusBadRequest, errors.New("scope is required"))
		return
	}
	serverType, serverName, err := parseScope(payload.Scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateStackSegment(serverName, "server"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if serverType == "remote" {
		if err := s.runRemoteStackAction(cfg, serverName, payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if payload.Action == "rm" {
			if err := deleteStackDir(serverName, payload.Name); err != nil {
				s.addLog("warn", fmt.Sprintf("stack %s remove cleanup failed: %s", payload.Name, err.Error()))
			}
		}
		writeJSON(w, http.StatusOK, stackActionResponse{Name: payload.Name, Action: payload.Action})
		return
	}
	envFile, invalid, err := prepareEnvFileFromDisk(serverName, payload.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if envFile != "" {
		defer os.Remove(envFile)
	}
	if len(invalid) > 0 {
		s.addLog("warn", fmt.Sprintf("stack %s env invalid lines ignored: %s", payload.Name, strings.Join(invalid, ", ")))
	}
	if err := runComposeFromStorage(cfg, serverName, payload.Name, payload.Action, envFile); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if payload.Action == "rm" {
		if err := deleteStackDir(serverName, payload.Name); err != nil {
			s.addLog("warn", fmt.Sprintf("stack %s remove cleanup failed: %s", payload.Name, err.Error()))
		}
	}
	writeJSON(w, http.StatusOK, stackActionResponse{Name: payload.Name, Action: payload.Action})
}

func buildStackSummaries(stacks []stackSummary, containers []dockerwatcher.ContainerInfo) []stackSummary {
	grouped := map[string][]dockerwatcher.ContainerInfo{}
	for _, item := range containers {
		stack := strings.TrimSpace(item.Stack)
		if stack == "" || stack == "-" {
			continue
		}
		grouped[stack] = append(grouped[stack], item)
	}
	result := make([]stackSummary, 0, len(stacks))
	for _, item := range stacks {
		list := grouped[item.Name]
		running, stopped := countContainerStates(list)
		status := deriveStackStatus(list, running, stopped)
		summary := stackSummary{
			Name:              item.Name,
			Status:            status,
			ContainersTotal:   running + stopped,
			ContainersRunning: running,
			ContainersStopped: stopped,
			UpdatedAt:         item.UpdatedAt,
		}
		result = append(result, summary)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result
}

func (s *Server) listContainersByScope(cfg config.Config, serverType, serverName string) ([]dockerwatcher.ContainerInfo, error) {
	if serverType == "local" {
		return s.listLocalContainers(cfg, serverName)
	}
	return s.listRemoteContainers(cfg, serverName)
}

func listStacksOnDisk(serverName string) ([]stackSummary, error) {
	base := stacksServerDir(serverName)
	entries, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []stackSummary{}, nil
		}
		return nil, err
	}
	var result []stackSummary
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if err := validateStackSegment(name, "stack"); err != nil {
			continue
		}
		composePath := filepath.Join(base, name, "docker-compose.yml")
		stat, err := os.Stat(composePath)
		if err != nil {
			continue
		}
		updatedAt := stat.ModTime()
		if envStat, err := os.Stat(filepath.Join(base, name, ".env")); err == nil {
			if envStat.ModTime().After(updatedAt) {
				updatedAt = envStat.ModTime()
			}
		}
		result = append(result, stackSummary{
			Name:      name,
			UpdatedAt: updatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func loadStackDetail(serverName, stackName string) (stackDetailResponse, error) {
	dir := stackDir(serverName, stackName)
	composePath := filepath.Join(dir, "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	if err != nil {
		return stackDetailResponse{}, err
	}
	detail := stackDetailResponse{
		Name:       stackName,
		ComposeYml: string(content),
	}
	if envBytes, err := os.ReadFile(filepath.Join(dir, ".env")); err == nil {
		detail.Env = string(envBytes)
		detail.HasEnv = true
	}
	if stat, err := os.Stat(composePath); err == nil {
		detail.UpdatedAt = stat.ModTime().Format(time.RFC3339)
	}
	return detail, nil
}

func saveStackFiles(serverName, stackName, composeYml, env string, useEnv bool) error {
	dir := stackDir(serverName, stackName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	composePath := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(composeYml), 0o644); err != nil {
		return err
	}
	envPath := filepath.Join(dir, ".env")
	if useEnv {
		if err := os.WriteFile(envPath, []byte(env), 0o644); err != nil {
			return err
		}
	} else {
		if err := os.Remove(envPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

func (s *Server) runRemoteStackAction(cfg config.Config, serverName string, payload stackActionRequest) error {
	remote, ok := findRemoteServer(cfg.RemoteServers, serverName)
	if !ok {
		return errors.New("remote server not found")
	}
	if remote.URL == "" {
		return errors.New("remote url missing")
	}
	detail, err := loadStackDetail(serverName, payload.Name)
	if err != nil {
		return err
	}
	body, err := json.Marshal(stackActionRequest{
		Name:       payload.Name,
		Action:     payload.Action,
		ComposeYml: detail.ComposeYml,
		Env:        detail.Env,
		UseEnv:     detail.HasEnv,
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimSuffix(remote.URL, "/") + "/api/stacks/action"
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

func runComposeFromStorage(cfg config.Config, serverName, stackName, action, envFile string) error {
	local, ok := findLocalServer(cfg.LocalServers, serverName)
	if !ok {
		return errors.New("local server not found")
	}
	dir := stackDir(serverName, stackName)
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err != nil {
		return err
	}
	env := map[string]string{}
	if host := dockerHostFromSocket(local.Socket); host != "" {
		env["DOCKER_HOST"] = host
	}
	return runComposeAction(dir, stackName, action, env, envFile)
}

func runComposeFromPayload(payload stackActionRequest, dockerHost string) error {
	tempDir, err := os.MkdirTemp("", "contiwatch-stack-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(payload.ComposeYml), 0o644); err != nil {
		return err
	}
	envFile := ""
	if payload.UseEnv {
		sanitized, _ := sanitizeEnvContent(payload.Env)
		var err error
		envFile, err = writeTempEnvFile(sanitized)
		if err != nil {
			return err
		}
		defer os.Remove(envFile)
	}
	env := map[string]string{}
	if dockerHost != "" {
		env["DOCKER_HOST"] = dockerHost
	}
	return runComposeAction(tempDir, payload.Name, payload.Action, env, envFile)
}

func runComposeAction(dir, projectName, action string, env map[string]string, envFile string) error {
	projectName = composeProjectName(projectName)
	args, err := buildComposeArgs(projectName, action, envFile)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), stackActionTimeoutSec*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	cmd.Env = mergeEnv(env)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("compose action timed out")
	}
	if err != nil {
		return fmt.Errorf("compose failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func composeProjectName(name string) string {
	value := strings.ToLower(strings.TrimSpace(name))
	if value == "" {
		return "stack"
	}
	first := value[0]
	isAlphaNum := (first >= 'a' && first <= 'z') || (first >= '0' && first <= '9')
	if !isAlphaNum {
		return "s" + value
	}
	return value
}

func runComposeConfigFromPayload(payload stackValidateRequest) error {
	tempDir, err := os.MkdirTemp("", "contiwatch-stack-validate-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(payload.ComposeYml), 0o644); err != nil {
		return err
	}
	envFile := ""
	if payload.UseEnv {
		sanitized, _ := sanitizeEnvContent(payload.Env)
		var err error
		envFile, err = writeTempEnvFile(sanitized)
		if err != nil {
			return err
		}
		defer os.Remove(envFile)
	}
	return runComposeConfig(tempDir, envFile)
}

func runComposeConfig(dir, envFile string) error {
	args := []string{"compose"}
	if strings.TrimSpace(envFile) != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args, "config", "--no-interpolate")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("compose config timed out")
	}
	if err != nil {
		return fmt.Errorf("compose config failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func buildComposeArgs(projectName, action, envFile string) ([]string, error) {
	args := []string{"compose"}
	if strings.TrimSpace(envFile) != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args, "--project-name", projectName)
	switch action {
	case "up":
		return append(args, "up", "-d"), nil
	case "down":
		return append(args, "down"), nil
	case "restart":
		return append(args, "restart"), nil
	case "start":
		return append(args, "start"), nil
	case "stop":
		return append(args, "stop"), nil
	case "kill":
		return append(args, "kill"), nil
	case "rm":
		return append(args, "rm", "-f", "-s"), nil
	default:
		return nil, errors.New("unknown action")
	}
}

func mergeEnv(overrides map[string]string) []string {
	env := os.Environ()
	if len(overrides) == 0 {
		return env
	}
	for key, value := range overrides {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func stacksServerDir(serverName string) string {
	return filepath.Join(stacksBaseDir, serverName)
}

func stackDir(serverName, stackName string) string {
	return filepath.Join(stacksBaseDir, serverName, stackName)
}

func validateStackSegment(value, label string) error {
	if value == "" {
		return fmt.Errorf("%s name is required", label)
	}
	if !stackNamePattern.MatchString(value) {
		return fmt.Errorf("%s name must match [A-Za-z0-9_-]+", label)
	}
	return nil
}

func deleteStackDir(serverName, stackName string) error {
	base := stacksServerDir(serverName)
	target := stackDir(serverName, stackName)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return err
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return errors.New("invalid stack path")
	}
	return os.RemoveAll(target)
}

func validateEnvContent(content string) error {
	_, invalid := sanitizeEnvContent(content)
	if len(invalid) == 0 {
		return nil
	}
	return fmt.Errorf("invalid .env entries: %s", strings.Join(invalid, ", "))
}

func sanitizeEnvContent(content string) (string, []string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var cleaned []string
	var invalid []string
	for idx, line := range lines {
		raw := line
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			cleaned = append(cleaned, raw)
			continue
		}
		if strings.ContainsRune(raw, 0) {
			invalid = append(invalid, fmt.Sprintf("line %d", idx+1))
			continue
		}
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			invalid = append(invalid, fmt.Sprintf("line %d", idx+1))
			continue
		}
		key := strings.TrimSpace(parts[0])
		if !envKeyPattern.MatchString(key) {
			invalid = append(invalid, fmt.Sprintf("line %d (%s)", idx+1, key))
			continue
		}
		value := parts[1]
		cleaned = append(cleaned, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(cleaned, "\n"), invalid
}

func prepareEnvFileFromDisk(serverName, stackName string) (string, []string, error) {
	envPath := filepath.Join(stackDir(serverName, stackName), ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil, nil
		}
		return "", nil, err
	}
	cleaned, invalid := sanitizeEnvContent(string(data))
	path, err := writeTempEnvFile(cleaned)
	if err != nil {
		return "", nil, err
	}
	return path, invalid, nil
}

func writeTempEnvFile(content string) (string, error) {
	file, err := os.CreateTemp("", "contiwatch-env-*")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func countContainerStates(list []dockerwatcher.ContainerInfo) (int, int) {
	running := 0
	stopped := 0
	for _, item := range list {
		state := strings.ToLower(strings.TrimSpace(item.State))
		switch state {
		case "running", "paused", "restarting":
			running++
		case "exited", "dead", "created", "oomkilled":
			stopped++
		default:
			stopped++
		}
	}
	return running, stopped
}

func deriveStackStatus(list []dockerwatcher.ContainerInfo, running, stopped int) string {
	total := running + stopped
	if total == 0 {
		return "not_deployed"
	}
	if running == total {
		return "running"
	}
	if running == 0 {
		return "stopped"
	}
	down := false
	for _, item := range list {
		state := strings.ToLower(strings.TrimSpace(item.State))
		if state == "exited" || state == "dead" || state == "oomkilled" {
			down = true
			break
		}
	}
	if down {
		return "degraded"
	}
	return "partial"
}
