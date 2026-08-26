package server

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPinSessionAuthorizedRefreshesLastSeen(t *testing.T) {
	server := &Server{
		pinGuardEnabled: true,
		pinSessions: map[string]pinSessionEntry{
			"valid-token": {LastSeen: time.Now().Add(-2 * time.Hour)},
		},
	}

	if !server.pinSessionAuthorized("valid-token") {
		t.Fatalf("expected token to be authorized")
	}

	entry, ok := server.pinSessions["valid-token"]
	if !ok {
		t.Fatalf("expected token to remain stored")
	}
	if time.Since(entry.LastSeen) > time.Second {
		t.Fatalf("expected LastSeen to be refreshed, got %v", entry.LastSeen)
	}
}

func newPinGuardTestServer(t *testing.T, pin string) *Server {
	t.Helper()
	salt, hash, err := initializePinGuardMaterial(pin)
	if err != nil {
		t.Fatalf("initialize PIN guard: %v", err)
	}
	mux := http.NewServeMux()
	server := &Server{
		mux:                 mux,
		pinGuardEnabled:     true,
		pinSalt:             salt,
		pinHash:             hash,
		pinSessions:         map[string]pinSessionEntry{},
		pinAttempts:         map[string]pinAttemptEntry{},
		pinWebSocketTickets: map[string]pinWebSocketTicketEntry{},
		pinGlobalTokens:     pinGlobalBurst,
		pinGlobalLastRefill: time.Now(),
	}
	mux.HandleFunc("/api/pin/verify", server.handlePinVerify)
	return server
}

func performPinVerify(server *Server, remoteAddr, forwardedFor, pin string) *httptest.ResponseRecorder {
	body := `{"pin":"` + pin + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/pin/verify", strings.NewReader(body))
	request.RemoteAddr = remoteAddr
	request.Header.Set("Content-Type", "application/json")
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func TestPinRateLimitIgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	server := newPinGuardTestServer(t, "2468")
	for attempt := 1; attempt <= 4; attempt++ {
		response := performPinVerify(server, "203.0.113.10:44000", "198.51.100."+strconv.Itoa(attempt), "0000")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", attempt, response.Code)
		}
	}
	response := performPinVerify(server, "203.0.113.10:44000", "198.51.100.99", "0000")
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("expected lockout despite changing X-Forwarded-For, got %d", response.Code)
	}
	response = performPinVerify(server, "203.0.113.10:44000", "198.51.100.100", "2468")
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("expected locked peer to remain locked, got %d", response.Code)
	}
}

func TestClientIPUsesForwardedForOnlyForTrustedProxy(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/pin/verify", nil)
	request.RemoteAddr = "10.0.0.2:51000"
	request.Header.Set("X-Forwarded-For", "198.51.100.8, 10.0.0.3")

	if got := clientIPFromRequest(request, nil); got != "10.0.0.2" {
		t.Fatalf("expected direct peer without trusted proxies, got %q", got)
	}
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	if got := clientIPFromRequest(request, trusted); got != "198.51.100.8" {
		t.Fatalf("expected forwarded client through trusted proxies, got %q", got)
	}
}

func TestPinVerifyRejectsOversizedBody(t *testing.T) {
	server := newPinGuardTestServer(t, "2468")
	payload := `{"pin":"` + strings.Repeat("1", int(pinRequestBodyLimit)) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/pin/verify", strings.NewReader(payload))
	request.RemoteAddr = "203.0.113.10:44000"
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", response.Code)
	}
}

func TestPinWebSocketTicketIsSingleUse(t *testing.T) {
	server := &Server{
		pinGuardEnabled:     true,
		pinSessions:         map[string]pinSessionEntry{"session": {LastSeen: time.Now()}},
		pinWebSocketTickets: map[string]pinWebSocketTicketEntry{},
	}
	ticket, err := server.createPinWebSocketTicket("session")
	if err != nil {
		t.Fatalf("create WebSocket ticket: %v", err)
	}
	if !server.consumePinWebSocketTicket(ticket) {
		t.Fatalf("expected ticket to authorize first use")
	}
	if server.consumePinWebSocketTicket(ticket) {
		t.Fatalf("expected ticket to be rejected after first use")
	}
}

func TestPinSessionAuthorizedRejectsExpiredSession(t *testing.T) {
	server := &Server{
		pinGuardEnabled: true,
		pinSessions: map[string]pinSessionEntry{
			"expired-token": {LastSeen: time.Now().Add(-(pinSessionIdleMaxAge + time.Minute))},
		},
	}

	if server.pinSessionAuthorized("expired-token") {
		t.Fatalf("expected expired token to be rejected")
	}
	if _, ok := server.pinSessions["expired-token"]; ok {
		t.Fatalf("expected expired token to be removed from storage")
	}
}
