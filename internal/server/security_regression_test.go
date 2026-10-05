package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"contiwatch/internal/config"

	"github.com/gorilla/websocket"
)

func TestConfigResponsesRedactTokensAndKeepSavedCredentials(t *testing.T) {
	store, err := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Update(func(cfg *config.Config) {
		cfg.RemoteServers = []config.RemoteServer{{Name: "fixture", Token: "synthetic-agent-credential", Maintenance: true}}
	})
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{store: store}
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		body, _ := json.Marshal(store.Get())
		response := httptest.NewRecorder()
		s.handleConfig(response, httptest.NewRequest(method, "/api/config", bytes.NewReader(body)))
		if response.Code != http.StatusOK {
			t.Fatalf("%s failed: %s", method, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "synthetic-agent-credential") {
			t.Fatalf("%s exposed agent token", method)
		}
		var payload struct {
			RemoteServers []remoteServerResponse `json:"remote_servers"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.RemoteServers) != 1 || !payload.RemoteServers[0].TokenConfigured || payload.RemoteServers[0].Token != "" {
			t.Fatal("configured-state contract lost")
		}
		if store.Get().RemoteServers[0].Token != "synthetic-agent-credential" {
			t.Fatal("redaction changed saved credential")
		}
	}
}

func TestLegacySSESessionRevokedOnLogout(t *testing.T) {
	store, err := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{store: store, pinGuardEnabled: true, mux: http.NewServeMux()}
	s.mux.HandleFunc("/api/servers/stream", s.handleServersStream)
	token, err := s.createPinSession()
	if err != nil {
		t.Fatal(err)
	}
	endpoint := httptest.NewServer(s)
	defer endpoint.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.URL+"/api/servers/stream?pin_session="+token, nil)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("legacy SSE authentication failed")
	}
	s.revokePinSession(token)
	if _, err := io.ReadAll(response.Body); err != nil {
		t.Fatalf("SSE did not finish after logout: %v", err)
	}
}

func TestWebSocketTicketRetainsSessionForRevocation(t *testing.T) {
	s := &Server{pinGuardEnabled: true}
	token, err := s.createPinSession()
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := s.createPinWebSocketTicket(token)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/containers/shell?ws_ticket="+ticket, nil)
	if !s.pinRequestAuthorized(request) {
		t.Fatal("ticket rejected")
	}
	done := s.pinSessionDone(request)
	if done == nil {
		t.Fatal("ticket lost owning session")
	}
	s.revokePinSession(token)
	select {
	case <-done:
	default:
		t.Fatal("ticket-authenticated connection missed logout")
	}
	if s.pinRequestAuthorized(httptest.NewRequest(http.MethodGet, "/api/containers/shell?ws_ticket="+ticket, nil)) {
		t.Fatal("ticket reused after logout")
	}
}

func TestAgentBodyBoundedWithUnknownLength(t *testing.T) {
	for _, test := range []struct {
		input string
		fail  bool
	}{{"1234", false}, {"12345", true}} {
		body := &boundedAgentBody{ReadCloser: io.NopCloser(strings.NewReader(test.input)), remaining: 4}
		data, err := io.ReadAll(body)
		if string(data) != "1234" || errors.Is(err, errAgentResponseTooLarge) != test.fail {
			t.Fatalf("input %q: data=%q error=%v", test.input, data, err)
		}
	}
}

type fixtureAgentTransport struct {
	response *http.Response
}

func (t fixtureAgentTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return t.response, nil
}

type fixtureAgentBody struct {
	io.Reader
	closed bool
}

func (b *fixtureAgentBody) Close() error { b.closed = true; return nil }

func TestAgentTransportRejectsDeclaredOversizedBodyAndClosesIt(t *testing.T) {
	body := &fixtureAgentBody{Reader: strings.NewReader("fixture")}
	transport := boundedAgentTransport{base: fixtureAgentTransport{response: &http.Response{StatusCode: http.StatusOK, ContentLength: agentResponseLimit + 1, Body: body}}}
	request, _ := http.NewRequest(http.MethodGet, "http://agent.test/", nil)
	response, err := transport.RoundTrip(request)
	if response != nil || !errors.Is(err, errAgentResponseTooLarge) || !body.closed {
		t.Fatal("oversized body accepted or leaked")
	}
}

func TestAgentHTTPRedirectCompatibility(t *testing.T) {
	var authorized bool
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusTemporaryRedirect)
			return
		}
		authorized = r.Header.Get("Authorization") == "Bearer fixture"
		_, _ = w.Write([]byte("ok"))
	}))
	defer agent.Close()
	client := newAgentHTTPClient(10 * time.Minute)
	if client.Timeout != 10*time.Minute {
		t.Fatal("operation budget changed")
	}
	req, _ := http.NewRequest(http.MethodGet, agent.URL+"/redirect", nil)
	req.Header.Set("Authorization", "Bearer fixture")
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if !authorized {
		t.Fatal("private HTTP agent lost bearer authentication during redirect")
	}
	https, _ := http.NewRequest(http.MethodGet, "https://example.test/", nil)
	httpURL, _ := http.NewRequest(http.MethodGet, "http://example.test/", nil)
	if client.CheckRedirect(httpURL, []*http.Request{https}) == nil {
		t.Fatal("HTTPS downgrade accepted")
	}
	if client.CheckRedirect(https, []*http.Request{httpURL}) != nil {
		t.Fatal("HTTPS upgrade rejected")
	}
}

func TestComposeEnvironmentFiltersOnlyApplicationSecrets(t *testing.T) {
	secretKeys := []string{"APP_PIN", "CONTIWATCH_APP_PIN", "CONTIWATCH_AGENT_TOKEN", "CONTIWATCH_GITHUB_TOKEN", "GITHUB_TOKEN"}
	for _, key := range secretKeys {
		t.Setenv(key, "fixture")
	}
	for _, key := range []string{"DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_CONFIG", "HTTPS_PROXY", "PATH", "CUSTOM_STACK_VALUE"} {
		t.Setenv(key, "required-fixture")
	}
	values := map[string]string{}
	for _, item := range mergeComposeEnv(nil, "") {
		key, value, _ := strings.Cut(item, "=")
		values[key] = value
	}
	for _, key := range secretKeys {
		if _, ok := values[key]; ok {
			t.Errorf("application secret %s inherited", key)
		}
	}
	for _, key := range []string{"DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_CONFIG", "HTTPS_PROXY", "PATH", "CUSTOM_STACK_VALUE"} {
		if values[key] != "required-fixture" {
			t.Errorf("required environment %s changed", key)
		}
	}
}

func TestComposeDiagnosticsRemainBoundedWithoutStoppingCommand(t *testing.T) {
	// This process prints synthetic output only; it never invokes Docker.
	cmd := exec.Command("sh", "-c", "i=0; while [ $i -lt 10000 ]; do printf 'synthetic-progress-line\\n'; i=$((i+1)); done; printf 'completed-fixture'; exit 7")
	output, err := boundedComposeOutput(cmd)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("command outcome changed: %v", err)
	}
	if !strings.HasPrefix(output, "[output truncated;") || !strings.HasSuffix(output, "completed-fixture") || len(output) > composeOutputLimit+100 {
		t.Fatalf("unbounded or incomplete diagnostics: %d bytes", len(output))
	}
}

func TestStackStorageRejectsEscapeAndReplacesLinkSafely(t *testing.T) {
	directory := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("original-fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "docker-compose.yml")
	if err := os.Symlink(outside, path); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := readStackFile(root, "docker-compose.yml"); err == nil {
		t.Fatal("external symlink read accepted")
	}
	if err := writePrivateFileInRoot(root, "docker-compose.yml", []byte("services: {}\n")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != "original-fixture" {
		t.Fatal("file outside stack changed")
	}
	info, _ := os.Stat(outside)
	if info.Mode().Perm() != 0o644 {
		t.Fatal("permissions outside stack changed")
	}
	if err := writePrivateFileInRoot(root, ".env", bytes.Repeat([]byte("x"), stackFileLimit+1)); err == nil {
		t.Fatal("oversized stack file accepted")
	}
	detail, err := loadStackDetailFromRoot(root, "fixture")
	if err != nil || detail.ComposeYml != "services: {}\n" || detail.HasEnv {
		t.Fatalf("normal stack no longer loads: %+v %v", detail, err)
	}
}

func TestContainerStreamSlotsAreReleased(t *testing.T) {
	s := &Server{}
	var releases []func()
	for i := 0; i < maxContainerStreams; i++ {
		release, ok := s.reserveContainerStream(httptest.NewRecorder())
		if !ok {
			t.Fatal("slot rejected too early")
		}
		releases = append(releases, release)
	}
	w := httptest.NewRecorder()
	if _, ok := s.reserveContainerStream(w); ok || w.Code != http.StatusTooManyRequests {
		t.Fatal("unbounded stream concurrency")
	}
	releases[0]()
	release, ok := s.reserveContainerStream(httptest.NewRecorder())
	if !ok {
		t.Fatal("closed connection did not release slot")
	}
	release()
	for _, release := range releases[1:] {
		release()
	}
}

func TestContainerStreamLimitAndSessionRevocation(t *testing.T) {
	for _, revoke := range []bool{false, true} {
		t.Run(map[bool]string{false: "message_limit", true: "logout"}[revoke], func(t *testing.T) {
			s := &Server{pinGuardEnabled: true}
			token, err := s.createPinSession()
			if err != nil {
				t.Fatal(err)
			}
			ready := make(chan struct{})
			finished := make(chan error, 1)
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r = r.WithContext(context.WithValue(r.Context(), pinSessionContextKey{}, token))
				conn, err := shellUpgrader.Upgrade(w, r, nil)
				if err != nil {
					finished <- err
					return
				}
				defer conn.Close()
				defer s.protectContainerStream(r, conn, 8)()
				close(ready)
				_, _, err = conn.ReadMessage()
				finished <- err
			}))
			defer endpoint.Close()
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(endpoint.URL, "http"), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			<-ready
			if revoke {
				s.revokePinSession(token)
			} else if err := conn.WriteMessage(websocket.BinaryMessage, []byte("123456789")); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-finished:
				if err == nil || (!revoke && !errors.Is(err, websocket.ErrReadLimit)) {
					t.Fatalf("stream remained usable: %v", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("stream not closed")
			}
		})
	}
}

func TestLogChunkingPreservesUnicodeAcrossChunkBoundary(t *testing.T) {
	input := strings.Repeat("a", 32767) + strings.Repeat("żółw", 10000)
	finished := make(chan error, 1)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := shellUpgrader.Upgrade(w, r, nil)
		if err != nil {
			finished <- err
			return
		}
		defer conn.Close()
		_, err = (&wsTextWriter{conn: conn}).Write([]byte(input))
		finished <- err
	}))
	defer endpoint.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(endpoint.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var output strings.Builder
	for output.Len() < len(input) {
		kind, chunk, err := conn.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if kind != websocket.TextMessage || len(chunk) > 32768 || !utf8.Valid(chunk) {
			t.Fatal("invalid text chunk")
		}
		output.Write(chunk)
	}
	if output.String() != input {
		t.Fatal("log output corrupted")
	}
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
}
