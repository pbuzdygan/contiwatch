package dockerwatcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDockerClientStillNegotiatesLegacyAgentHostAPI(t *testing.T) {
	listed := false
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.39")
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path != "/v1.39/containers/json" {
			t.Errorf("unexpected daemon request %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		listed = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer daemon.Close()
	watcher, err := NewWithHost(daemon.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()
	containers, err := watcher.Containers(context.Background())
	if err != nil || !listed || len(containers) != 0 {
		t.Fatalf("legacy host API negotiation failed: %v", err)
	}
}
