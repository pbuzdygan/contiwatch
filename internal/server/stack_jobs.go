package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"contiwatch/internal/config"
)

const (
	// stackJobExtension is how much time a single user "extend" request adds.
	stackJobExtension = 10 * time.Minute
	// stackJobRetention keeps finished jobs queryable long enough for slow clients to read the result.
	stackJobRetention     = 15 * time.Minute
	stackJobStartTimeout  = 60 * time.Second
	stackJobStatusTimeout = 20 * time.Second
	stackJobErrorBodySize = 64 * 1024
)

// stackJobPollInterval is a variable so tests can poll faster.
var stackJobPollInterval = 2 * time.Second

var (
	errStackActionTimedOut        = errors.New("stack action timed out before completion; the host or registry may be slow — extend the timeout while it runs or retry")
	errStackJobNotFound           = errors.New("stack job not found")
	errStackJobConflict           = errors.New("another action is already running for this stack")
	errStackJobFinished           = errors.New("stack job already finished")
	errStackJobNotExtendable      = errors.New("this stack job cannot be extended; update the remote agent to enable timeout extension")
	errRemoteStackJobsUnsupported = errors.New("remote agent does not support stack jobs")
)

type stackJobStatus string

const (
	stackJobRunning   stackJobStatus = "running"
	stackJobSucceeded stackJobStatus = "succeeded"
	stackJobFailed    stackJobStatus = "failed"
)

// stackJobView is the wire representation of a stack job, shared by the
// browser API and the controller-to-agent API.
type stackJobView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	Scope       string `json:"scope,omitempty"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at,omitempty"`
	Deadline    string `json:"deadline"`
	RemainingMs int64  `json:"remaining_ms"`
	ExtendMs    int64  `json:"extend_ms"`
	Extendable  bool   `json:"extendable"`
}

type stackJobIDRequest struct {
	ID string `json:"id"`
}

// extendableDeadline cancels its context when the deadline passes, unless the
// deadline is pushed back first. Unlike context.WithTimeout, the deadline can
// be moved while the work is running.
type extendableDeadline struct {
	mu       sync.Mutex
	deadline time.Time
	timer    *time.Timer
	cancel   context.CancelCauseFunc
}

func newExtendableDeadline(parent context.Context, timeout time.Duration) (context.Context, *extendableDeadline) {
	ctx, cancel := context.WithCancelCause(parent)
	d := &extendableDeadline{deadline: time.Now().Add(timeout), cancel: cancel}
	d.timer = time.AfterFunc(timeout, func() { cancel(errStackActionTimedOut) })
	return ctx, d
}

// Extend moves the deadline by the given duration. It reports false when the
// deadline already fired or the work was stopped.
func (d *extendableDeadline) Extend(by time.Duration) (time.Time, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.timer.Stop() {
		return d.deadline, false
	}
	d.deadline = d.deadline.Add(by)
	d.timer.Reset(time.Until(d.deadline))
	return d.deadline, true
}

func (d *extendableDeadline) Deadline() time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.deadline
}

// Stop releases the timer and the context once the work has finished.
func (d *extendableDeadline) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.timer.Stop()
	d.cancel(nil)
}

type stackJob struct {
	id        string
	key       string
	name      string
	action    string
	scope     string
	startedAt time.Time
	deadline  *extendableDeadline

	mu           sync.Mutex
	status       stackJobStatus
	err          string
	finishedAt   time.Time
	extendable   bool
	extendRemote func(context.Context) error
}

func (j *stackJob) view(now time.Time) stackJobView {
	j.mu.Lock()
	defer j.mu.Unlock()
	deadline := j.deadline.Deadline()
	view := stackJobView{
		ID:         j.id,
		Name:       j.name,
		Action:     j.action,
		Scope:      j.scope,
		Status:     string(j.status),
		Error:      j.err,
		StartedAt:  j.startedAt.UTC().Format(time.RFC3339),
		Deadline:   deadline.UTC().Format(time.RFC3339),
		ExtendMs:   stackJobExtension.Milliseconds(),
		Extendable: j.extendable && j.status == stackJobRunning,
	}
	if j.status == stackJobRunning {
		view.RemainingMs = max(0, deadline.Sub(now).Milliseconds())
	} else {
		view.FinishedAt = j.finishedAt.UTC().Format(time.RFC3339)
	}
	return view
}

func (j *stackJob) running() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.status == stackJobRunning
}

func (j *stackJob) finish(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.finishedAt = time.Now()
	if err != nil {
		j.status = stackJobFailed
		j.err = err.Error()
		return
	}
	j.status = stackJobSucceeded
}

// setRemoteExtender makes extension requests go to the agent first, so the
// agent-side deadline moves together with the controller-side one.
func (j *stackJob) setRemoteExtender(fn func(context.Context) error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.extendRemote = fn
}

func (j *stackJob) disableExtension() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.extendable = false
}

func (j *stackJob) extend(ctx context.Context) error {
	j.mu.Lock()
	status, extendable, extendRemote := j.status, j.extendable, j.extendRemote
	j.mu.Unlock()
	if status != stackJobRunning {
		return errStackJobFinished
	}
	if !extendable {
		return errStackJobNotExtendable
	}
	if extendRemote != nil {
		if err := extendRemote(ctx); err != nil {
			return err
		}
	}
	if _, ok := j.deadline.Extend(stackJobExtension); !ok {
		return errStackJobFinished
	}
	return nil
}

type stackJobSpec struct {
	key     string
	name    string
	action  string
	scope   string
	timeout time.Duration
}

// stackJobManager runs stack actions in the background so that slow hosts are
// not bounded by a single HTTP request lifetime.
type stackJobManager struct {
	mu   sync.Mutex
	jobs map[string]*stackJob
}

func newStackJobManager() *stackJobManager {
	return &stackJobManager{jobs: map[string]*stackJob{}}
}

// start launches run in the background. Only one job per stack key may run at
// a time, which also prevents duplicate submissions from the UI.
func (m *stackJobManager) start(spec stackJobSpec, run func(context.Context, *stackJob) error) (*stackJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(time.Now())
	for _, existing := range m.jobs {
		if existing.key == spec.key && existing.running() {
			return nil, errStackJobConflict
		}
	}
	id, err := newStackJobID()
	if err != nil {
		return nil, err
	}
	ctx, deadline := newExtendableDeadline(context.Background(), spec.timeout)
	job := &stackJob{
		id:         id,
		key:        spec.key,
		name:       spec.name,
		action:     spec.action,
		scope:      spec.scope,
		startedAt:  time.Now(),
		deadline:   deadline,
		status:     stackJobRunning,
		extendable: true,
	}
	m.jobs[id] = job
	go func() {
		err := run(ctx, job)
		deadline.Stop()
		job.finish(err)
	}()
	return job, nil
}

func (m *stackJobManager) get(id string) (*stackJob, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	return job, ok
}

func (m *stackJobManager) pruneLocked(now time.Time) {
	for id, job := range m.jobs {
		job.mu.Lock()
		expired := job.status != stackJobRunning && now.Sub(job.finishedAt) > stackJobRetention
		job.mu.Unlock()
		if expired {
			delete(m.jobs, id)
		}
	}
}

func newStackJobID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate stack job id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// stackContextError converts a finished context into an actionable error and
// returns fallback when the context is still active.
func stackContextError(ctx context.Context, fallback error) error {
	if ctx.Err() == nil {
		return fallback
	}
	cause := context.Cause(ctx)
	if errors.Is(cause, errStackActionTimedOut) || errors.Is(cause, context.DeadlineExceeded) {
		return errStackActionTimedOut
	}
	return fmt.Errorf("stack action cancelled: %w", cause)
}

func (s *Server) handleStackJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleStackJobStatus(w, r)
	case http.MethodPost:
		s.handleStackJobStart(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStackJobStart(w http.ResponseWriter, r *http.Request) {
	payload, err := s.decodeStackActionRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg := s.store.Get()
	spec := stackJobSpec{
		key:     payload.Scope + "/" + payload.Name,
		name:    payload.Name,
		action:  payload.Action,
		scope:   payload.Scope,
		timeout: s.stackActionTimeout(payload),
	}
	job, err := s.stackJobs.start(spec, func(ctx context.Context, job *stackJob) error {
		err := s.executeStackAction(ctx, cfg, payload, job)
		if err != nil {
			s.addLog("warn", fmt.Sprintf("stack %s action %s failed: %s", payload.Name, payload.Action, err.Error()))
		}
		return err
	})
	if errors.Is(err, errStackJobConflict) {
		writeError(w, http.StatusConflict, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusAccepted, job.view(time.Now()))
}

func (s *Server) handleStackJobStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}
	job, ok := s.stackJobs.get(id)
	if !ok {
		writeError(w, http.StatusNotFound, errStackJobNotFound)
		return
	}
	writeJSON(w, http.StatusOK, job.view(time.Now()))
}

func (s *Server) handleStackJobExtend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload stackJobIDRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job, ok := s.stackJobs.get(strings.TrimSpace(payload.ID))
	if !ok {
		writeError(w, http.StatusNotFound, errStackJobNotFound)
		return
	}
	if err := job.extend(r.Context()); err != nil {
		switch {
		case errors.Is(err, errStackJobFinished), errors.Is(err, errStackJobNotExtendable):
			writeError(w, http.StatusConflict, err)
		default:
			writeError(w, http.StatusBadGateway, fmt.Errorf("extend on remote agent failed: %w", err))
		}
		return
	}
	writeJSON(w, http.StatusOK, job.view(time.Now()))
}

// stackActionTimeout is the initial deadline for a stack action handled by
// this instance. Controllers waiting on an agent get extra grace so the agent
// can report its own timeout first.
func (s *Server) stackActionTimeout(payload parsedStackActionRequest) time.Duration {
	if !s.agentMode && payload.serverType == "remote" {
		return remoteStackActionTimeout(payload.Action)
	}
	return stackActionBudget(payload.Action)
}

// executeStackAction runs a validated stack action on the target server. When
// job is set, remote agents are driven through their job API so the timeout
// can be extended end-to-end.
func (s *Server) executeStackAction(ctx context.Context, cfg config.Config, payload parsedStackActionRequest, job *stackJob) error {
	if s.agentMode {
		return runComposeFromPayload(ctx, payload.stackActionRequest, agentDockerHost(cfg))
	}
	if payload.serverType == "remote" {
		remote, request, err := buildRemoteStackActionRequest(cfg, payload.serverName, payload.stackActionRequest)
		if err != nil {
			return err
		}
		if job != nil {
			err = s.runRemoteStackJob(ctx, job, remote, request)
		} else {
			err = postRemoteStackAction(ctx, remote, request)
		}
		if err != nil {
			return err
		}
	} else if err := s.runLocalStackAction(ctx, cfg, payload); err != nil {
		return err
	}
	if payload.Action == "rm" {
		if err := deleteStackDir(payload.serverName, payload.Name); err != nil {
			s.addLog("warn", fmt.Sprintf("stack %s remove cleanup failed: %s", payload.Name, err.Error()))
		}
	}
	return nil
}

func (s *Server) runLocalStackAction(ctx context.Context, cfg config.Config, payload parsedStackActionRequest) error {
	envFile, invalid, err := prepareEnvFileFromDisk(payload.serverName, payload.Name)
	if err != nil {
		return err
	}
	if envFile != "" {
		defer os.Remove(envFile)
	}
	if len(invalid) > 0 {
		s.addLog("warn", fmt.Sprintf("stack %s env invalid lines ignored: %s", payload.Name, strings.Join(invalid, ", ")))
	}
	return runComposeFromStorage(ctx, cfg, payload.serverName, payload.Name, payload.Action, envFile)
}

// runRemoteStackJob starts the action as a job on the agent and polls it until
// it finishes. Agents without the job API fall back to the synchronous call.
func (s *Server) runRemoteStackJob(ctx context.Context, job *stackJob, remote config.RemoteServer, request stackActionRequest) error {
	startCtx, cancel := context.WithTimeout(ctx, stackJobStartTimeout)
	started, err := startRemoteStackJob(startCtx, remote, request)
	cancel()
	if errors.Is(err, errRemoteStackJobsUnsupported) {
		job.disableExtension()
		s.addLog("info", fmt.Sprintf("stack %s: agent %s does not support stack jobs, using synchronous action", request.Name, remote.Name))
		return postRemoteStackAction(ctx, remote, request)
	}
	if err != nil {
		return stackContextError(ctx, err)
	}
	job.setRemoteExtender(func(extendCtx context.Context) error {
		return extendRemoteStackJob(extendCtx, remote, started.ID)
	})
	return pollRemoteStackJob(ctx, remote, started.ID)
}

// pollRemoteStackJob tolerates transient status failures because slow agents
// may stall while busy; the controller-side deadline bounds the wait.
func pollRemoteStackJob(ctx context.Context, remote config.RemoteServer, id string) error {
	ticker := time.NewTicker(stackJobPollInterval)
	defer ticker.Stop()
	var lastErr error
	for {
		select {
		case <-ctx.Done():
			err := stackContextError(ctx, ctx.Err())
			if lastErr != nil {
				return fmt.Errorf("%w (last agent error: %v)", err, lastErr)
			}
			return err
		case <-ticker.C:
		}
		statusCtx, cancel := context.WithTimeout(ctx, stackJobStatusTimeout)
		view, err := getRemoteStackJob(statusCtx, remote, id)
		cancel()
		if errors.Is(err, errStackJobNotFound) {
			return errors.New("remote agent no longer tracks the stack job (agent restarted?)")
		}
		if err != nil {
			lastErr = err
			continue
		}
		lastErr = nil
		switch stackJobStatus(view.Status) {
		case stackJobSucceeded:
			return nil
		case stackJobFailed:
			if view.Error == "" {
				return errors.New("remote stack action failed")
			}
			return errors.New(view.Error)
		}
	}
}

func startRemoteStackJob(ctx context.Context, remote config.RemoteServer, request stackActionRequest) (stackJobView, error) {
	resp, err := doRemoteStackRequest(ctx, remote, http.MethodPost, "/api/stacks/jobs", request)
	if err != nil {
		return stackJobView{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		return stackJobView{}, errRemoteStackJobsUnsupported
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return stackJobView{}, remoteStackResponseError(resp)
	}
	return decodeStackJobView(resp)
}

func getRemoteStackJob(ctx context.Context, remote config.RemoteServer, id string) (stackJobView, error) {
	resp, err := doRemoteStackRequest(ctx, remote, http.MethodGet, "/api/stacks/jobs?id="+url.QueryEscape(id), nil)
	if err != nil {
		return stackJobView{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return stackJobView{}, errStackJobNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return stackJobView{}, remoteStackResponseError(resp)
	}
	return decodeStackJobView(resp)
}

func extendRemoteStackJob(ctx context.Context, remote config.RemoteServer, id string) error {
	resp, err := doRemoteStackRequest(ctx, remote, http.MethodPost, "/api/stacks/jobs/extend", stackJobIDRequest{ID: id})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return remoteStackResponseError(resp)
	}
	return nil
}

func decodeStackJobView(resp *http.Response) (stackJobView, error) {
	var view stackJobView
	if err := json.NewDecoder(resp.Body).Decode(&view); err != nil {
		return stackJobView{}, fmt.Errorf("decode agent stack job: %w", err)
	}
	return view, nil
}

// doRemoteStackRequest sends an authenticated request to an agent. The
// context bounds the whole exchange, so no client-level timeout is set.
func doRemoteStackRequest(ctx context.Context, remote config.RemoteServer, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(remote.URL, "/")+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}
	return (&http.Client{}).Do(req)
}

// remoteStackResponseError surfaces the agent's error message instead of a
// bare status code, so compose failures are visible in the UI.
func remoteStackResponseError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, stackJobErrorBodySize))
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err == nil && strings.TrimSpace(payload.Error) != "" {
		return fmt.Errorf("agent returned status %d: %s", resp.StatusCode, payload.Error)
	}
	return fmt.Errorf("agent returned status %d", resp.StatusCode)
}
