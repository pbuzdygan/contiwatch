package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
)

func TestCollectRemoteServerHealthUsesAgentToken(t *testing.T) {
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/servers/health" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		writeJSON(w, http.StatusOK, serverHealthResponse{
			Scope: "local:agent",
			Health: dockerwatcher.DockerHealthSnapshot{
				EngineVersion: "29.0.0",
			},
		})
	}))
	defer agent.Close()

	health, err := collectRemoteServerHealth(context.Background(), config.Config{
		RemoteServers: []config.RemoteServer{{Name: "remote", URL: agent.URL, Token: "test-token"}},
	}, "remote")
	if err != nil {
		t.Fatalf("collect remote health: %v", err)
	}
	if health.EngineVersion != "29.0.0" {
		encoded, _ := json.Marshal(health)
		t.Fatalf("unexpected health payload %s", encoded)
	}
}
