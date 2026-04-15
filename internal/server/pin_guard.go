package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	pinSessionHeaderName = "X-Contiwatch-Pin-Session"
	pinVerifyMinDelay    = 250 * time.Millisecond
	pinSessionIdleMaxAge = 24 * time.Hour
)

type pinSessionEntry struct {
	LastSeen time.Time
}

type pinAttemptEntry struct {
	Failures  int
	LockedTo  time.Time
	UpdatedAt time.Time
}

func parsePinGuardFromEnv(agentMode bool) (bool, string, time.Duration, error) {
	if agentMode {
		return false, "", 0, nil
	}
	pin := strings.TrimSpace(os.Getenv("APP_PIN"))
	if pin == "" {
		pin = strings.TrimSpace(os.Getenv("CONTIWATCH_APP_PIN"))
	}
	if pin == "" {
		return false, "", 0, errors.New("controller mode requires APP_PIN (4-8 digits)")
	}
	return true, pin, pinVerifyMinDelay, nil
}

func isValidPinValue(value string) bool {
	if len(value) < 4 || len(value) > 8 {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func hashPinValue(value string, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + value))
	return hex.EncodeToString(sum[:])
}

func generateToken(size int) (string, error) {
	if size < 16 {
		size = 16
	}
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func pinPublicAPIPath(path string) bool {
	switch path {
	case "/api/health":
		return true
	case "/api/version":
		return true
	case "/api/meta":
		return true
	case "/api/release":
		return true
	case "/api/pin/status":
		return true
	case "/api/pin/verify":
		return true
	case "/api/pin/logout":
		return true
	default:
		return false
	}
}

func (s *Server) pinSessionTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if value := strings.TrimSpace(r.Header.Get(pinSessionHeaderName)); value != "" {
		return value
	}
	if value := strings.TrimSpace(r.URL.Query().Get("pin_session")); value != "" {
		return value
	}
	return ""
}

func (s *Server) pinSessionAuthorized(token string) bool {
	if !s.pinGuardEnabled || token == "" {
		return false
	}
	now := time.Now()
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	s.pruneExpiredPinSessionsLocked(now)
	entry, ok := s.pinSessions[token]
	if !ok {
		return false
	}
	entry.LastSeen = now
	s.pinSessions[token] = entry
	return true
}

func (s *Server) createPinSession() (string, error) {
	token, err := generateToken(32)
	if err != nil {
		return "", err
	}
	now := time.Now()
	s.pinMu.Lock()
	if s.pinSessions == nil {
		s.pinSessions = map[string]pinSessionEntry{}
	}
	s.pruneExpiredPinSessionsLocked(now)
	s.pinSessions[token] = pinSessionEntry{LastSeen: now}
	s.pinMu.Unlock()
	return token, nil
}

func (s *Server) revokePinSession(token string) {
	if token == "" {
		return
	}
	s.pinMu.Lock()
	delete(s.pinSessions, token)
	s.pinMu.Unlock()
}

func (s *Server) pruneExpiredPinSessionsLocked(now time.Time) {
	if len(s.pinSessions) == 0 {
		return
	}
	for token, entry := range s.pinSessions {
		if entry.LastSeen.IsZero() || now.Sub(entry.LastSeen) > pinSessionIdleMaxAge {
			delete(s.pinSessions, token)
		}
	}
}

func clientIPFromRequest(r *http.Request) string {
	if r == nil {
		return "unknown"
	}
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		part := strings.TrimSpace(strings.Split(xff, ",")[0])
		if part != "" {
			return part
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	addr := strings.TrimSpace(r.RemoteAddr)
	if addr == "" {
		return "unknown"
	}
	return addr
}

func (s *Server) pinBeforeVerify(ip string) (bool, int) {
	now := time.Now()
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	if s.pinAttempts == nil {
		s.pinAttempts = map[string]pinAttemptEntry{}
	}
	entry := s.pinAttempts[ip]
	if !entry.LockedTo.IsZero() && entry.LockedTo.After(now) {
		seconds := int(entry.LockedTo.Sub(now).Seconds())
		if seconds < 1 {
			seconds = 1
		}
		return false, seconds
	}
	return true, 0
}

func (s *Server) pinOnVerifySuccess(ip string) {
	s.pinMu.Lock()
	delete(s.pinAttempts, ip)
	s.pinMu.Unlock()
}

func (s *Server) pinOnVerifyFailure(ip string) (bool, int) {
	now := time.Now()
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	if s.pinAttempts == nil {
		s.pinAttempts = map[string]pinAttemptEntry{}
	}
	entry := s.pinAttempts[ip]
	entry.Failures++
	entry.UpdatedAt = now
	lockApplied := false
	retryAfterSeconds := 0
	if entry.Failures >= 5 {
		lockApplied = true
		steps := entry.Failures - 5
		if steps > 5 {
			steps = 5
		}
		lockFor := time.Duration(1<<steps) * time.Minute
		entry.LockedTo = now.Add(lockFor)
		retryAfterSeconds = int(lockFor.Seconds())
	}
	s.pinAttempts[ip] = entry
	return lockApplied, retryAfterSeconds
}

func (s *Server) verifyPinCandidate(candidate string) bool {
	if !s.pinGuardEnabled {
		return false
	}
	if !isValidPinValue(candidate) {
		return false
	}
	hashed := hashPinValue(candidate, s.pinSalt)
	return subtleStringCompare(hashed, s.pinHash)
}

func subtleStringCompare(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	result := 0
	for i := 0; i < len(left); i++ {
		result |= int(left[i] ^ right[i])
	}
	return result == 0
}

func withMinPinDelay(startedAt time.Time, minDelay time.Duration, fn func()) {
	if fn == nil {
		return
	}
	elapsed := time.Since(startedAt)
	if elapsed >= minDelay {
		fn()
		return
	}
	time.Sleep(minDelay - elapsed)
	fn()
}

func writePinRequired(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "pin_required"})
}

func (s *Server) handlePinStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !s.pinGuardEnabled {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false, "unlocked": true})
		return
	}
	token := s.pinSessionTokenFromRequest(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":  true,
		"unlocked": s.pinSessionAuthorized(token),
	})
}

func (s *Server) handlePinVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !s.pinGuardEnabled {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "pin_guard_disabled"})
		return
	}
	startedAt := time.Now()
	ip := clientIPFromRequest(r)
	allowed, retryAfterSeconds := s.pinBeforeVerify(ip)
	if !allowed {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		withMinPinDelay(startedAt, s.pinMinResponseDelay, func() {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"ok":                  false,
				"error":               "LOCKOUT",
				"retry_after_seconds": retryAfterSeconds,
			})
		})
		return
	}
	var payload struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		withMinPinDelay(startedAt, s.pinMinResponseDelay, func() {
			writeError(w, http.StatusBadRequest, err)
		})
		return
	}
	pin := strings.TrimSpace(payload.Pin)
	if s.verifyPinCandidate(pin) {
		s.pinOnVerifySuccess(ip)
		token, err := s.createPinSession()
		if err != nil {
			withMinPinDelay(startedAt, s.pinMinResponseDelay, func() {
				writeError(w, http.StatusInternalServerError, err)
			})
			return
		}
		withMinPinDelay(startedAt, s.pinMinResponseDelay, func() {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "session_token": token})
		})
		return
	}
	lockApplied, lockRetry := s.pinOnVerifyFailure(ip)
	if lockApplied {
		w.Header().Set("Retry-After", strconv.Itoa(lockRetry))
		withMinPinDelay(startedAt, s.pinMinResponseDelay, func() {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"ok":                  false,
				"error":               "LOCKOUT",
				"retry_after_seconds": lockRetry,
			})
		})
		return
	}
	withMinPinDelay(startedAt, s.pinMinResponseDelay, func() {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "Wrong PIN"})
	})
}

func (s *Server) handlePinLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	token := s.pinSessionTokenFromRequest(r)
	s.revokePinSession(token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func validateStartupPin(pin string) error {
	value := strings.TrimSpace(pin)
	if value == "" {
		return errors.New("pin is required")
	}
	if !isValidPinValue(value) {
		return errors.New("pin must contain 4-8 digits")
	}
	return nil
}

func initializePinGuardMaterial(pin string) (string, string, error) {
	value := strings.TrimSpace(pin)
	if err := validateStartupPin(value); err != nil {
		return "", "", err
	}
	saltToken, err := generateToken(16)
	if err != nil {
		return "", "", err
	}
	return saltToken, hashPinValue(value, saltToken), nil
}
