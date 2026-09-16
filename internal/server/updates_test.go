package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

func TestBuildScanSummarySeparatesDetectedUpdatedAndRemaining(t *testing.T) {
	s := &Server{}
	result := dockerwatcher.ScanResult{Containers: []dockerwatcher.ContainerStatus{
		{Name: "updated", Updated: true, Policy: config.PolicyUpdate},
		{Name: "skipped", UpdateAvailable: true, Policy: config.PolicyUpdate, Error: "skipped: policy"},
		{Name: "failed", UpdateAvailable: true, Policy: config.PolicyUpdate, Error: "update failed: pull error"},
	}}

	summary := s.buildScanSummary(result)

	if summary.detected != 3 || summary.updated != 1 || summary.remaining != 2 {
		t.Fatalf("unexpected update counts: %+v", summary)
	}
	if summary.skipped != 1 || summary.failed != 1 || summary.ready != 0 {
		t.Fatalf("unexpected outcome counts: %+v", summary)
	}
}

func TestRemoteAgentUpdateTargetSupportsLegacyAgentWithoutMisclassifyingOtherContainers(t *testing.T) {
	tests := []struct {
		name      string
		container dockerwatcher.ContainerStatus
		expected  bool
	}{
		{name: "explicit self marker", container: dockerwatcher.ContainerStatus{Self: true}, expected: true},
		{name: "legacy agent with tagged image", container: dockerwatcher.ContainerStatus{Name: "contiwatch-agent-dev", Image: "ghcr.io/pbuzdygan/contiwatch:dev_latest"}, expected: true},
		{name: "legacy agent with image ID", container: dockerwatcher.ContainerStatus{Name: "contiwatch-agent-dev", Image: "eccc16ed2d1a"}, expected: true},
		{name: "other contiwatch service", container: dockerwatcher.ContainerStatus{Name: "contiwatch-dashboard", Image: "ghcr.io/pbuzdygan/contiwatch:dev_latest"}, expected: false},
		{name: "unrelated agent", container: dockerwatcher.ContainerStatus{Name: "metrics-agent", Image: "example/metrics:latest"}, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isRemoteAgentUpdateTarget(test.container); got != test.expected {
				t.Fatalf("expected %t, got %t", test.expected, got)
			}
		})
	}
}

func TestAutoUpdateRemoteUpdatesAgentLastAndConfirmsRestart(t *testing.T) {
	requestOrder := []string{}
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/self-update":
			id := r.URL.Query().Get("container")
			requestOrder = append(requestOrder, "self:"+id)
			writeJSON(w, http.StatusOK, map[string]string{"status": "scheduled"})
		case strings.HasPrefix(r.URL.Path, "/api/update/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/update/")
			requestOrder = append(requestOrder, id)
			writeJSON(w, http.StatusOK, dockerwatcher.UpdateResult{ID: id, Name: "app", Updated: true})
		case r.URL.Path == "/api/status":
			writeJSON(w, http.StatusOK, dockerwatcher.ScanResult{
				CheckedAt: time.Now(),
				Containers: []dockerwatcher.ContainerStatus{
					{ID: "app-new-id", Name: "app", Updated: true},
					{ID: "agent-new-id", Name: "contiwatch-agent-dev", Self: true, Updated: true},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer agent.Close()

	s := &Server{scanStates: map[string]scanState{}}
	result := dockerwatcher.ScanResult{
		ServerName: "remote",
		Containers: []dockerwatcher.ContainerStatus{
			{ID: "agent-id", Name: "contiwatch-agent-dev", Image: "eccc16ed2d1a", Policy: config.PolicyUpdate, UpdateAvailable: true},
			{ID: "app-id", Name: "app", Image: "app:latest", Policy: config.PolicyUpdate, UpdateAvailable: true},
		},
	}

	updated, err := s.autoUpdateRemote(context.Background(), config.DefaultConfig(), config.RemoteServer{
		Name: "remote", URL: agent.URL,
	}, &result)
	if err != nil {
		t.Fatalf("auto update remote: %v", err)
	}
	if updated != 2 {
		t.Fatalf("expected two confirmed updates, got %d", updated)
	}
	if len(requestOrder) != 2 || requestOrder[0] != "app-id" || requestOrder[1] != "self:agent-id" {
		t.Fatalf("unexpected update order: %v", requestOrder)
	}
	if len(result.Containers) != 2 || !result.Containers[0].Updated || !result.Containers[1].Updated {
		t.Fatalf("expected refreshed confirmed status, got %+v", result.Containers)
	}
}

func TestHandleSelfUpdateRejectsNonAgentContainer(t *testing.T) {
	s := &Server{agentMode: true}
	req := httptest.NewRequest(http.MethodPost, "/api/self-update?container=not-the-running-agent", nil)
	recorder := httptest.NewRecorder()

	s.handleSelfUpdate(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
