package server

import (
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
	stacksBaseDir                  = "/data/stacks"
	stackActionTimeout             = 3 * time.Minute
	stackImageActionTimeout        = 10 * time.Minute
	remoteStackActionGraceDuration = 30 * time.Second
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
	payload, err := s.decodeStackActionRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Synchronous endpoint kept for controllers older than 1.3.4 that call
	// agents directly; the UI uses the extendable /api/stacks/jobs flow.
	ctx, cancel := context.WithTimeout(context.Background(), s.stackActionTimeout(payload))
	defer cancel()
	if err := s.executeStackAction(ctx, s.store.Get(), payload, nil); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, stackActionResponse{Name: payload.Name, Action: payload.Action})
}

// parsedStackActionRequest is a validated stack action together with the
// resolved target server (controller mode only).
type parsedStackActionRequest struct {
	stackActionRequest
	serverType string
	serverName string
}

// decodeStackActionRequest reads and validates a stack action request body.
// Agents require the compose payload; controllers require a valid scope.
func (s *Server) decodeStackActionRequest(r *http.Request) (parsedStackActionRequest, error) {
	var payload stackActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return parsedStackActionRequest{}, err
	}
	payload.Action = strings.TrimSpace(strings.ToLower(payload.Action))
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Scope = strings.TrimSpace(payload.Scope)
	if payload.Name == "" || payload.Action == "" {
		return parsedStackActionRequest{}, errors.New("name and action are required")
	}
	if !isValidStackAction(payload.Action) {
		return parsedStackActionRequest{}, fmt.Errorf("unknown stack action: %s", payload.Action)
	}
	if err := validateStackSegment(payload.Name, "stack"); err != nil {
		return parsedStackActionRequest{}, err
	}
	if s.agentMode {
		if strings.TrimSpace(payload.ComposeYml) == "" {
			return parsedStackActionRequest{}, errors.New("compose_yaml is required")
		}
		sanitized, invalid := sanitizeEnvContent(payload.Env)
		payload.Env = sanitized
		if payload.UseEnv && len(invalid) > 0 {
			s.addLog("warn", fmt.Sprintf("stack %s env invalid lines ignored: %s", payload.Name, strings.Join(invalid, ", ")))
		}
		return parsedStackActionRequest{stackActionRequest: payload}, nil
	}
	if payload.Scope == "" {
		return parsedStackActionRequest{}, errors.New("scope is required")
	}
	serverType, serverName, err := parseScope(payload.Scope)
	if err != nil {
		return parsedStackActionRequest{}, err
	}
	if err := validateStackSegment(serverName, "server"); err != nil {
		return parsedStackActionRequest{}, err
	}
	return parsedStackActionRequest{stackActionRequest: payload, serverType: serverType, serverName: serverName}, nil
}

func agentDockerHost(cfg config.Config) string {
	if len(cfg.LocalServers) == 1 {
		return dockerHostFromSocket(cfg.LocalServers[0].Socket)
	}
	return ""
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
	if err := validateStackSegment(serverName, "server"); err != nil {
		return nil, err
	}
	base, err := os.OpenRoot(stacksBaseDir)
	if errors.Is(err, fs.ErrNotExist) {
		return []stackSummary{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer base.Close()
	root, err := base.OpenRoot(serverName)
	if errors.Is(err, fs.ErrNotExist) {
		return []stackSummary{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer root.Close()
	directory, err := root.Open(".")
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	entries, err := directory.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	result := []stackSummary{}
	for _, entry := range entries {
		if !entry.IsDir() || validateStackSegment(entry.Name(), "stack") != nil {
			continue
		}
		stat, err := root.Stat(filepath.Join(entry.Name(), "docker-compose.yml"))
		if err != nil || !stat.Mode().IsRegular() {
			continue
		}
		updatedAt := stat.ModTime()
		if envStat, err := root.Stat(filepath.Join(entry.Name(), ".env")); err == nil && envStat.Mode().IsRegular() && envStat.ModTime().After(updatedAt) {
			updatedAt = envStat.ModTime()
		}
		result = append(result, stackSummary{Name: entry.Name(), UpdatedAt: updatedAt.Format(time.RFC3339)})
	}
	return result, nil
}

// buildRemoteStackActionRequest resolves the remote agent and embeds the stored
// stack files, because agents do not keep stack definitions on their own disk.
func buildRemoteStackActionRequest(cfg config.Config, serverName string, payload stackActionRequest) (config.RemoteServer, stackActionRequest, error) {
	remote, ok := findRemoteServer(cfg.RemoteServers, serverName)
	if !ok {
		return config.RemoteServer{}, stackActionRequest{}, errors.New("remote server not found")
	}
	if remote.URL == "" {
		return config.RemoteServer{}, stackActionRequest{}, errors.New("remote url missing")
	}
	detail, err := loadStackDetail(serverName, payload.Name)
	if err != nil {
		return config.RemoteServer{}, stackActionRequest{}, err
	}
	return remote, stackActionRequest{
		Name:       payload.Name,
		Action:     payload.Action,
		ComposeYml: detail.ComposeYml,
		Env:        detail.Env,
		UseEnv:     detail.HasEnv,
	}, nil
}

// postRemoteStackAction runs a stack action through the synchronous agent
// endpoint. The caller's context bounds the whole request.
func postRemoteStackAction(ctx context.Context, remote config.RemoteServer, request stackActionRequest) error {
	resp, err := doRemoteStackRequest(ctx, remote, http.MethodPost, "/api/stacks/action", request)
	if err != nil {
		return stackContextError(ctx, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return remoteStackResponseError(resp)
	}
	return nil
}

func runComposeFromStorage(ctx context.Context, cfg config.Config, serverName, stackName, action, envFile string) error {
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
	return runComposeAction(ctx, dir, stackName, action, env, envFile)
}

func runComposeFromPayload(ctx context.Context, payload stackActionRequest, dockerHost string) error {
	tempDir, err := os.MkdirTemp("", "contiwatch-stack-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(payload.ComposeYml), 0o600); err != nil {
		return err
	}
	envFile := ""
	if payload.UseEnv {
		sanitized, _ := sanitizeEnvContent(payload.Env)
		envFile = filepath.Join(tempDir, ".env")
		if err := os.WriteFile(envFile, []byte(sanitized), 0o600); err != nil {
			return err
		}
	}
	env := map[string]string{}
	if dockerHost != "" {
		env["DOCKER_HOST"] = dockerHost
	}
	return runComposeAction(ctx, tempDir, payload.Name, payload.Action, env, envFile)
}

// runComposeAction runs a stack action until it completes or ctx ends. The
// context carries the whole action budget, so redeploy shares one deadline
// across its pull and up steps.
func runComposeAction(ctx context.Context, dir, projectName, action string, env map[string]string, envFile string) error {
	projectName = composeProjectName(projectName)

	if action == "redeploy" {
		if err := runComposeSingleAction(ctx, dir, projectName, "pull", env, envFile); err != nil {
			return fmt.Errorf("compose pull failed: %w", err)
		}
		if err := runComposeSingleAction(ctx, dir, projectName, "up", env, envFile); err != nil {
			return fmt.Errorf("compose up failed: %w", err)
		}
		return nil
	}
	return runComposeSingleAction(ctx, dir, projectName, action, env, envFile)
}

func runComposeSingleAction(ctx context.Context, dir, projectName, action string, env map[string]string, envFile string) error {
	args, err := buildComposeArgs(projectName, action, envFile)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	cmd.Env = mergeComposeEnv(env, envFile)
	output, err := boundedComposeOutput(cmd)
	if ctx.Err() != nil {
		return stackContextError(ctx, ctx.Err())
	}
	if err != nil {
		return fmt.Errorf("compose failed: %s", strings.TrimSpace(output))
	}
	return nil
}

func composeActionTimeout(action string) time.Duration {
	switch action {
	case "up", "pull":
		return stackImageActionTimeout
	default:
		return stackActionTimeout
	}
}

// stackActionBudget is the initial time allowed for a whole stack action.
func stackActionBudget(action string) time.Duration {
	if action == "redeploy" {
		return composeActionTimeout("pull") + composeActionTimeout("up")
	}
	return composeActionTimeout(action)
}

// remoteStackActionTimeout outlives the agent-side budget so the agent can
// report its own result before the controller gives up.
func remoteStackActionTimeout(action string) time.Duration {
	return stackActionBudget(action) + remoteStackActionGraceDuration
}

func isValidStackAction(action string) bool {
	if action == "redeploy" {
		return true
	}
	_, err := buildComposeArgs("stack", action, "")
	return err == nil
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
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(payload.ComposeYml), 0o600); err != nil {
		return err
	}
	envFile := ""
	if payload.UseEnv {
		sanitized, _ := sanitizeEnvContent(payload.Env)
		envFile = filepath.Join(tempDir, ".env")
		if err := os.WriteFile(envFile, []byte(sanitized), 0o600); err != nil {
			return err
		}
	}
	return runComposeConfig(tempDir, envFile)
}

func runComposeConfig(dir, envFile string) error {
	args := []string{"compose"}
	if strings.TrimSpace(envFile) != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args, "config", "--quiet")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = dir
	cmd.Env = mergeComposeEnv(nil, envFile)
	output, err := boundedComposeOutput(cmd)
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("compose config timed out")
	}
	if err != nil {
		return fmt.Errorf("compose config failed: %s", strings.TrimSpace(output))
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
	case "pull":
		return append(args, "pull"), nil
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

func mergeComposeEnv(overrides map[string]string, envFile string) []string {
	blocked := map[string]struct{}{}
	if strings.TrimSpace(envFile) != "" {
		if content, err := os.ReadFile(envFile); err == nil {
			cleaned, _ := sanitizeEnvContent(string(content))
			for _, line := range strings.Split(cleaned, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				if parts := strings.SplitN(trimmed, "=", 2); len(parts) == 2 {
					blocked[strings.TrimSpace(parts[0])] = struct{}{}
				}
			}
		}
	}
	for key := range overrides {
		blocked[key] = struct{}{}
	}

	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, item := range os.Environ() {
		key, _, found := strings.Cut(item, "=")
		if found {
			// Strip application credentials, while preserving Docker, TLS, proxy,
			// credential-helper and custom interpolation settings used by deployments.
			switch key {
			case "APP_PIN", "CONTIWATCH_APP_PIN", "CONTIWATCH_AGENT_TOKEN", "CONTIWATCH_GITHUB_TOKEN", "GITHUB_TOKEN":
				continue
			}
			if _, shouldUseEnvFile := blocked[key]; shouldUseEnvFile {
				continue
			}
		}
		env = append(env, item)
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
	if err := validateStackSegment(serverName, "server"); err != nil {
		return err
	}
	if err := validateStackSegment(stackName, "stack"); err != nil {
		return err
	}
	base, err := os.OpenRoot(stacksBaseDir)
	if err != nil {
		return err
	}
	defer base.Close()
	return base.RemoveAll(filepath.Join(serverName, stackName))
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
	root, err := openStackRoot(serverName, stackName, false)
	if err != nil {
		return "", nil, err
	}
	defer root.Close()
	data, err := readStackFile(root, ".env")
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
