package server

import (
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
