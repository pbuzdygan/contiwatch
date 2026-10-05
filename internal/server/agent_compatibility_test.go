package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"contiwatch/internal/config"
	"github.com/gorilla/websocket"
)

func TestCompatibilityRedactionPreservesAgentCommunication(t *testing.T) {
	var requests atomic.Int32
	agent := &Server{agentMode: true, agentToken: "compatibility-fixture-token", mux: http.NewServeMux()}
	agent.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/api/scan":
			writeJSON(w, http.StatusOK, map[string]any{"containers": []any{}})
		case "/api/update/container-fixture":
			writeJSON(w, http.StatusOK, map[string]any{"id": "container-fixture", "updated": true})
		case "/api/self-update":
			writeJSON(w, http.StatusOK, map[string]string{"status": "scheduled"})
		case "/api/config":
			if r.Method == http.MethodPut {
				var cfg config.Config
				if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
					t.Error(err)
				}
				if cfg.GlobalPolicy != config.PolicyUpdate {
					t.Error("Policy update lost")
				}
			}
			writeJSON(w, http.StatusOK, sanitizeConfigForResponse(config.DefaultConfig()))
		case "/api/stacks/action":
			var payload stackActionRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			writeJSON(w, http.StatusOK, stackActionResponse{Name: payload.Name, Action: payload.Action})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	endpoint := httptest.NewServer(agent)
	defer endpoint.Close()
	store, err := config.NewStore(t.TempDir() + "/config.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Update(func(cfg *config.Config) {
		cfg.RemoteServers = []config.RemoteServer{{Name: "remote-fixture", URL: endpoint.URL, Token: agent.agentToken}}
	})
	if err != nil {
		t.Fatal(err)
	}
	controller := &Server{store: store}
	cfg := store.Get()
	encoded, err := json.Marshal(sanitizeConfigForResponse(cfg))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), agent.agentToken) {
		t.Fatal("Credential exposed")
	}
	var response struct {
		RemoteServers []remoteServerResponse `json:"remote_servers"`
	}
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.RemoteServers) != 1 || !response.RemoteServers[0].TokenConfigured {
		t.Fatal("Configured state lost")
	}
	remote, ok := findRemoteServer(store.Get().RemoteServers, "remote-fixture")
	if !ok || remote.Token != agent.agentToken {
		t.Fatal("Stored credential changed")
	}
	ctx := context.Background()
	if _, err := controller.scanRemoteServer(ctx, remote); err != nil {
		t.Fatal(err)
	}
	result, err := controller.updateRemoteContainer(ctx, remote, "container-fixture", false)
	if err != nil || !result.Updated {
		t.Fatalf("Remote update failed: %v", err)
	}
	if _, err := controller.updateRemoteSelfUpdate(ctx, remote, "agent-fixture"); err != nil {
		t.Fatal(err)
	}
	cfg.GlobalPolicy = config.PolicyUpdate
	changed, err := controller.syncRemotePolicy(ctx, cfg, remote)
	if err != nil || !changed {
		t.Fatalf("Policy sync failed: %v", err)
	}
	for _, action := range []string{"up", "down", "redeploy", "restart"} {
		if err := postRemoteStackAction(ctx, remote, stackActionRequest{Name: "stack-fixture", Action: action}); err != nil {
			t.Fatal(err)
		}
	}
	if requests.Load() != 9 {
		t.Fatalf("Expected 9 authorized agent requests, got %d", requests.Load())
	}
	if store.Get().RemoteServers[0].Token != agent.agentToken {
		t.Fatal("Credential not retained")
	}
}

func TestCompatibilityRemoteWebSocketKeepsBearerHeader(t *testing.T) {
	agent := &Server{agentMode: true, agentToken: "compatibility-fixture-token", mux: http.NewServeMux()}
	var upgraded atomic.Bool
	agent.mux.HandleFunc("/api/containers/shell", func(w http.ResponseWriter, r *http.Request) {
		conn, err := shellUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		upgraded.Store(true)
		typ, message, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
			return
		}
		if err := conn.WriteMessage(typ, message); err != nil {
			t.Error(err)
		}
		conn.ReadMessage()
	})
	backend := httptest.NewServer(agent)
	defer backend.Close()
	cfg := config.DefaultConfig()
	cfg.RemoteServers = []config.RemoteServer{{Name: "remote-fixture", URL: backend.URL, Token: agent.agentToken}}
	controller := &Server{}
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller.handleRemoteContainerShell(w, r, cfg, "remote-fixture", "container-fixture")
	}))
	defer frontend.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(frontend.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte("synthetic-shell-message")); err != nil {
		t.Fatal(err)
	}
	typ, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !upgraded.Load() || typ != websocket.BinaryMessage || string(message) != "synthetic-shell-message" {
		t.Fatal("Authenticated remote WebSocket did not preserve the message")
	}
}

func TestCompatibilityRemoteLogsStreamsLargeLegacyMessage(t *testing.T) {
	// A pre-hardening agent may send a whole large Docker log frame as one
	// message. The proxy must not apply the smaller browser-input limit to it.
	message := strings.Repeat("synthetic-log-line\n", 32768)
	agent := &Server{agentMode: true, agentToken: "legacy-fixture", mux: http.NewServeMux()}
	agent.mux.HandleFunc("/api/containers/logs", func(w http.ResponseWriter, r *http.Request) {
		conn, err := shellUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			t.Error(err)
			return
		}
		_, _, _ = conn.ReadMessage()
	})
	backend := httptest.NewServer(agent)
	defer backend.Close()
	cfg := config.DefaultConfig()
	cfg.RemoteServers = []config.RemoteServer{{Name: "remote-fixture", URL: backend.URL, Token: agent.agentToken}}
	controller := &Server{}
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller.handleRemoteContainerLogs(w, r, cfg, "remote-fixture", "container-fixture")
	}))
	defer frontend.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(frontend.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	typ, data, err := conn.ReadMessage()
	if err != nil || typ != websocket.TextMessage || string(data) != message {
		t.Fatalf("legacy large log frame not preserved: %v", err)
	}
}
