package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"contiwatch/internal/config"
)

func waitForStackJob(t *testing.T, job *stackJob) stackJobView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !job.running() {
			return job.view(time.Now())
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("stack job did not finish in time")
	return stackJobView{}
}

func useFastStackJobPolling(t *testing.T) {
	t.Helper()
	previous := stackJobPollInterval
	stackJobPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { stackJobPollInterval = previous })
}

func TestExtendableDeadlineCancelsWithTimeoutCause(t *testing.T) {
	ctx, deadline := newExtendableDeadline(context.Background(), 20*time.Millisecond)
	defer deadline.Stop()

	<-ctx.Done()

	if err := stackContextError(ctx, nil); !errors.Is(err, errStackActionTimedOut) {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestExtendableDeadlineExtendPostponesCancellation(t *testing.T) {
	ctx, deadline := newExtendableDeadline(context.Background(), 30*time.Millisecond)
	defer deadline.Stop()
	before := deadline.Deadline()

	next, ok := deadline.Extend(time.Minute)

	if !ok {
		t.Fatal("expected extension to succeed while running")
	}
	if !next.Equal(before.Add(time.Minute)) {
		t.Fatalf("expected deadline %s, got %s", before.Add(time.Minute), next)
	}
	select {
	case <-ctx.Done():
		t.Fatal("context cancelled despite extension")
	case <-time.After(80 * time.Millisecond):
	}
}

func TestExtendableDeadlineRejectsExtensionAfterExpiry(t *testing.T) {
	ctx, deadline := newExtendableDeadline(context.Background(), 5*time.Millisecond)
	defer deadline.Stop()
	<-ctx.Done()

	if _, ok := deadline.Extend(time.Minute); ok {
		t.Fatal("expected extension to fail after the deadline fired")
	}
}

func TestStackJobManagerRejectsConcurrentJobForSameStack(t *testing.T) {
	manager := newStackJobManager()
	release := make(chan struct{})
	spec := stackJobSpec{key: "local:host/web", name: "web", action: "up", timeout: time.Minute}
	first, err := manager.start(spec, func(ctx context.Context, job *stackJob) error {
		<-release
		return nil
	})
	if err != nil {
		t.Fatalf("start first job: %v", err)
	}

	_, err = manager.start(spec, func(ctx context.Context, job *stackJob) error { return nil })
	close(release)

	if !errors.Is(err, errStackJobConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if view := waitForStackJob(t, first); view.Status != string(stackJobSucceeded) {
		t.Fatalf("expected first job to succeed, got %+v", view)
	}
}

func TestStackJobFailsWithTimeoutWhenNotExtended(t *testing.T) {
	manager := newStackJobManager()
	spec := stackJobSpec{key: "k", name: "web", action: "up", timeout: 20 * time.Millisecond}
	job, err := manager.start(spec, func(ctx context.Context, job *stackJob) error {
		<-ctx.Done()
		return stackContextError(ctx, ctx.Err())
	})
	if err != nil {
		t.Fatalf("start job: %v", err)
	}

	view := waitForStackJob(t, job)

	if view.Status != string(stackJobFailed) || view.Error != errStackActionTimedOut.Error() {
		t.Fatalf("expected timeout failure, got %+v", view)
	}
	if view.Extendable {
		t.Fatal("finished job must not be extendable")
	}
}

func TestStackJobExtendKeepsSlowActionAlive(t *testing.T) {
	manager := newStackJobManager()
	spec := stackJobSpec{key: "k", name: "web", action: "up", timeout: 40 * time.Millisecond}
	job, err := manager.start(spec, func(ctx context.Context, job *stackJob) error {
		select {
		case <-ctx.Done():
			return stackContextError(ctx, ctx.Err())
		case <-time.After(120 * time.Millisecond):
			return nil
		}
	})
	if err != nil {
		t.Fatalf("start job: %v", err)
	}

	if err := job.extend(context.Background()); err != nil {
		t.Fatalf("extend job: %v", err)
	}

	if view := waitForStackJob(t, job); view.Status != string(stackJobSucceeded) {
		t.Fatalf("expected extended job to succeed, got %+v", view)
	}
	if err := job.extend(context.Background()); !errors.Is(err, errStackJobFinished) {
		t.Fatalf("expected finished error, got %v", err)
	}
}

func TestRunRemoteStackJobPollsAgentAndForwardsExtension(t *testing.T) {
	useFastStackJobPolling(t)
	var polls, extensions atomic.Int32
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/stacks/jobs":
			writeJSON(w, http.StatusAccepted, stackJobView{ID: "agent-job", Status: string(stackJobRunning)})
		case r.Method == http.MethodPost && r.URL.Path == "/api/stacks/jobs/extend":
			extensions.Add(1)
			writeJSON(w, http.StatusOK, stackJobView{ID: "agent-job", Status: string(stackJobRunning)})
		case r.Method == http.MethodGet && r.URL.Path == "/api/stacks/jobs":
			status := stackJobRunning
			if polls.Add(1) >= 3 && extensions.Load() > 0 {
				status = stackJobSucceeded
			}
			writeJSON(w, http.StatusOK, stackJobView{ID: r.URL.Query().Get("id"), Status: string(status)})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer agent.Close()
	s := &Server{stackJobs: newStackJobManager()}
	remote := config.RemoteServer{Name: "pi", URL: agent.URL, Token: "secret"}

	job, err := s.stackJobs.start(stackJobSpec{key: "k", name: "web", action: "up", timeout: time.Minute}, func(ctx context.Context, job *stackJob) error {
		return s.runRemoteStackJob(ctx, job, remote, stackActionRequest{Name: "web", Action: "up"})
	})
	if err != nil {
		t.Fatalf("start job: %v", err)
	}
	waitUntil(t, func() bool { return polls.Load() > 0 })
	if err := job.extend(context.Background()); err != nil {
		t.Fatalf("extend job: %v", err)
	}

	if view := waitForStackJob(t, job); view.Status != string(stackJobSucceeded) {
		t.Fatalf("expected success, got %+v", view)
	}
	if extensions.Load() != 1 {
		t.Fatalf("expected extension forwarded once, got %d", extensions.Load())
	}
}

func TestRunRemoteStackJobReportsAgentFailure(t *testing.T) {
	useFastStackJobPolling(t)
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusAccepted, stackJobView{ID: "agent-job", Status: string(stackJobRunning)})
			return
		}
		writeJSON(w, http.StatusOK, stackJobView{ID: "agent-job", Status: string(stackJobFailed), Error: "compose failed: pull denied"})
	}))
	defer agent.Close()
	s := &Server{stackJobs: newStackJobManager()}
	job := &stackJob{extendable: true}

	err := s.runRemoteStackJob(context.Background(), job, config.RemoteServer{URL: agent.URL}, stackActionRequest{Name: "web", Action: "up"})

	if err == nil || err.Error() != "compose failed: pull denied" {
		t.Fatalf("expected agent error, got %v", err)
	}
}

func TestRunRemoteStackJobFallsBackToSynchronousActionForOldAgents(t *testing.T) {
	var legacyCalls atomic.Int32
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/stacks/action" {
			legacyCalls.Add(1)
			var payload stackActionRequest
			_ = json.NewDecoder(r.Body).Decode(&payload)
			writeJSON(w, http.StatusOK, stackActionResponse{Name: payload.Name, Action: payload.Action})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer agent.Close()
	s := &Server{stackJobs: newStackJobManager()}
	job := &stackJob{extendable: true, status: stackJobRunning}

	err := s.runRemoteStackJob(context.Background(), job, config.RemoteServer{URL: agent.URL}, stackActionRequest{Name: "web", Action: "down"})

	if err != nil {
		t.Fatalf("expected legacy fallback to succeed, got %v", err)
	}
	if legacyCalls.Load() != 1 {
		t.Fatalf("expected one legacy call, got %d", legacyCalls.Load())
	}
	if err := job.extend(context.Background()); !errors.Is(err, errStackJobNotExtendable) {
		t.Fatalf("expected legacy job to be non-extendable, got %v", err)
	}
}

func TestPostRemoteStackActionSurfacesAgentErrorMessage(t *testing.T) {
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusBadRequest, errors.New("compose failed: no such image"))
	}))
	defer agent.Close()

	err := postRemoteStackAction(context.Background(), config.RemoteServer{URL: agent.URL}, stackActionRequest{Name: "web", Action: "up"})

	if err == nil || err.Error() != "agent returned status 400: compose failed: no such image" {
		t.Fatalf("expected agent message, got %v", err)
	}
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}
