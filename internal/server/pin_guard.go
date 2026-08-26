package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	pinSessionHeaderName  = "X-Contiwatch-Pin-Session"
	pinVerifyMinDelay     = 250 * time.Millisecond
	pinSessionIdleMaxAge  = 24 * time.Hour
	pinWebSocketTicketAge = 30 * time.Second
	pinAttemptRetention   = 24 * time.Hour
	pinAttemptPruneEvery  = 10 * time.Minute
	pinAttemptMaxEntries  = 4096
	pinGlobalBurst        = 20.0
	pinGlobalRefillRate   = 1.0
)

type pinSessionEntry struct {
	LastSeen time.Time
}

type pinAttemptEntry struct {
	Failures  int
	LockedTo  time.Time
	UpdatedAt time.Time
}

type pinWebSocketTicketEntry struct {
	ExpiresAt    time.Time
	SessionToken string
}

func parseTrustedProxyPrefixes(raw string) ([]netip.Prefix, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	prefixes := make([]netip.Prefix, 0)
	for _, item := range strings.Split(value, ",") {
		candidate := strings.TrimSpace(item)
		if candidate == "" {
			continue
		}
		if prefix, err := netip.ParsePrefix(candidate); err == nil {
			prefixes = append(prefixes, prefix.Masked())
			continue
		}
		address, err := netip.ParseAddr(candidate)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy %q", candidate)
		}
		address = address.Unmap()
		prefixes = append(prefixes, netip.PrefixFrom(address, address.BitLen()))
	}
	return prefixes, nil
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

func pinWebSocketAPIPath(path string) bool {
	return path == "/api/containers/shell" || path == "/api/containers/logs"
}

func (s *Server) pinRequestAuthorized(r *http.Request) bool {
	if r == nil {
		return false
	}
	if s.pinSessionAuthorized(s.pinSessionTokenFromRequest(r)) {
		return true
	}
	if !pinWebSocketAPIPath(r.URL.Path) {
		return false
	}
	ticket := strings.TrimSpace(r.URL.Query().Get("ws_ticket"))
	return s.consumePinWebSocketTicket(ticket)
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
	for ticket, entry := range s.pinWebSocketTickets {
		if entry.SessionToken == token {
			delete(s.pinWebSocketTickets, ticket)
		}
	}
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
	for ticket, entry := range s.pinWebSocketTickets {
		if entry.ExpiresAt.IsZero() || !entry.ExpiresAt.After(now) {
			delete(s.pinWebSocketTickets, ticket)
		}
	}
}

func (s *Server) createPinWebSocketTicket(sessionToken string) (string, error) {
	if sessionToken == "" {
		return "", errors.New("pin session is required")
	}
	ticket, err := generateToken(24)
	if err != nil {
		return "", err
	}
	now := time.Now()
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	if s.pinWebSocketTickets == nil {
		s.pinWebSocketTickets = map[string]pinWebSocketTicketEntry{}
	}
	s.pruneExpiredPinSessionsLocked(now)
	s.pinWebSocketTickets[ticket] = pinWebSocketTicketEntry{
		ExpiresAt:    now.Add(pinWebSocketTicketAge),
		SessionToken: sessionToken,
	}
	return ticket, nil
}

func (s *Server) consumePinWebSocketTicket(ticket string) bool {
	if ticket == "" {
		return false
	}
	now := time.Now()
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	s.pruneExpiredPinSessionsLocked(now)
	entry, ok := s.pinWebSocketTickets[ticket]
	if !ok || !entry.ExpiresAt.After(now) {
		delete(s.pinWebSocketTickets, ticket)
		return false
	}
	delete(s.pinWebSocketTickets, ticket)
	_, sessionExists := s.pinSessions[entry.SessionToken]
	return sessionExists
}

func requestRemoteIP(r *http.Request) netip.Addr {
	if r == nil {
		return netip.Addr{}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	address, err := netip.ParseAddr(strings.TrimSpace(host))
	if err != nil {
		return netip.Addr{}
	}
	return address.Unmap()
}

func isTrustedProxy(address netip.Addr, trusted []netip.Prefix) bool {
	if !address.IsValid() {
		return false
	}
	address = address.Unmap()
	for _, prefix := range trusted {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func clientIPFromRequest(r *http.Request, trusted []netip.Prefix) string {
	peer := requestRemoteIP(r)
	if !peer.IsValid() {
		return "unknown"
	}
	if len(trusted) == 0 || !isTrustedProxy(peer, trusted) {
		return peer.String()
	}
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff == "" {
		return peer.String()
	}
	parts := strings.Split(xff, ",")
	forwarded := make([]netip.Addr, 0, len(parts))
	for _, part := range parts {
		address, err := netip.ParseAddr(strings.TrimSpace(part))
		if err != nil {
			return peer.String()
		}
		forwarded = append(forwarded, address.Unmap())
	}
	candidate := peer
	for i := len(forwarded) - 1; i >= 0; i-- {
		if !isTrustedProxy(candidate, trusted) {
			break
		}
		candidate = forwarded[i]
	}
	return candidate.String()
}

func (s *Server) prunePinAttemptsLocked(now time.Time) {
	if len(s.pinAttempts) == 0 {
		s.pinAttemptsLastPruned = now
		return
	}
	if len(s.pinAttempts) < pinAttemptMaxEntries &&
		!s.pinAttemptsLastPruned.IsZero() &&
		now.Sub(s.pinAttemptsLastPruned) < pinAttemptPruneEvery {
		return
	}
	for ip, entry := range s.pinAttempts {
		if entry.UpdatedAt.IsZero() || now.Sub(entry.UpdatedAt) > pinAttemptRetention {
			delete(s.pinAttempts, ip)
		}
	}
	s.pinAttemptsLastPruned = now
}

func (s *Server) ensurePinAttemptCapacityLocked() {
	if len(s.pinAttempts) < pinAttemptMaxEntries {
		return
	}
	oldestIP := ""
	oldestAt := time.Time{}
	for ip, entry := range s.pinAttempts {
		if oldestIP == "" || entry.UpdatedAt.Before(oldestAt) {
			oldestIP = ip
			oldestAt = entry.UpdatedAt
		}
	}
	if oldestIP != "" {
		delete(s.pinAttempts, oldestIP)
	}
}

func (s *Server) consumeGlobalPinAttemptLocked(now time.Time) (bool, int) {
	if s.pinGlobalLastRefill.IsZero() {
		s.pinGlobalTokens = pinGlobalBurst
		s.pinGlobalLastRefill = now
	}
	elapsed := now.Sub(s.pinGlobalLastRefill).Seconds()
	if elapsed > 0 {
		s.pinGlobalTokens = math.Min(pinGlobalBurst, s.pinGlobalTokens+(elapsed*pinGlobalRefillRate))
		s.pinGlobalLastRefill = now
	}
	if s.pinGlobalTokens < 1 {
		waitSeconds := int(math.Ceil((1 - s.pinGlobalTokens) / pinGlobalRefillRate))
		if waitSeconds < 1 {
			waitSeconds = 1
		}
		return false, waitSeconds
	}
	s.pinGlobalTokens--
	return true, 0
}

func (s *Server) pinBeforeVerify(ip string) (bool, int) {
	now := time.Now()
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	if s.pinAttempts == nil {
		s.pinAttempts = map[string]pinAttemptEntry{}
	}
	s.prunePinAttemptsLocked(now)
	entry := s.pinAttempts[ip]
	if !entry.LockedTo.IsZero() && entry.LockedTo.After(now) {
		seconds := int(entry.LockedTo.Sub(now).Seconds())
		if seconds < 1 {
			seconds = 1
		}
		return false, seconds
	}
	return s.consumeGlobalPinAttemptLocked(now)
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
	s.prunePinAttemptsLocked(now)
	if _, exists := s.pinAttempts[ip]; !exists {
		s.ensurePinAttemptCapacityLocked()
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
	ip := clientIPFromRequest(r, s.trustedProxies)
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

func (s *Server) handlePinWebSocketTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sessionToken := s.pinSessionTokenFromRequest(r)
	if !s.pinSessionAuthorized(sessionToken) {
		writePinRequired(w)
		return
	}
	ticket, err := s.createPinWebSocketTicket(sessionToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ticket":     ticket,
		"expires_in": int(pinWebSocketTicketAge.Seconds()),
	})
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
