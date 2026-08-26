package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const (
	discordWebhookKeepSentinel  = "__keep__"
	discordWebhookClearSentinel = "__clear__"
	defaultRequestBodyLimit     = int64(128 * 1024)
	configRequestBodyLimit      = int64(512 * 1024)
	stackRequestBodyLimit       = int64(4 * 1024 * 1024)
	pinRequestBodyLimit         = int64(1024)
)

func applySecurityHeaders(w http.ResponseWriter) {
	if w == nil {
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
	w.Header().Set(
		"Content-Security-Policy",
		"default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'",
	)
}

func requestBodyLimit(path string) int64 {
	switch {
	case path == "/api/pin/verify":
		return pinRequestBodyLimit
	case path == "/api/config", path == "/api/servers", path == "/api/locals":
		return configRequestBodyLimit
	case strings.HasPrefix(path, "/api/stacks"):
		return stackRequestBodyLimit
	default:
		return defaultRequestBodyLimit
	}
}

func applyRequestBodyLimit(w http.ResponseWriter, r *http.Request) error {
	if r == nil || r.Body == nil {
		return nil
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return nil
	}
	limit := requestBodyLimit(r.URL.Path)
	if r.ContentLength > limit {
		return &http.MaxBytesError{Limit: limit}
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	return nil
}

func isWebSocketOriginAllowed(r *http.Request) bool {
	if r == nil {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if !strings.EqualFold(originURL.Scheme, "http") && !strings.EqualFold(originURL.Scheme, "https") {
		return false
	}
	originHost := strings.TrimSpace(originURL.Host)
	requestHost := strings.TrimSpace(r.Host)
	if originHost == "" || requestHost == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSuffix(originHost, "."), strings.TrimSuffix(requestHost, "."))
}

func validateDiscordWebhookURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return errors.New("webhook_url is required")
	}
	u, err := url.Parse(value)
	if err != nil {
		return errors.New("webhook_url is invalid")
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return errors.New("webhook_url must use https")
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return errors.New("webhook_url host is required")
	}
	allowedHosts := map[string]struct{}{
		"discord.com":        {},
		"ptb.discord.com":    {},
		"canary.discord.com": {},
		"discordapp.com":     {},
	}
	if _, ok := allowedHosts[host]; !ok {
		return errors.New("webhook_url must point to Discord")
	}
	if !strings.HasPrefix(u.EscapedPath(), "/api/webhooks/") {
		return errors.New("webhook_url path is invalid")
	}
	return nil
}

func validateRemoteServerURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return errors.New("url is required")
	}
	u, err := url.Parse(value)
	if err != nil {
		return errors.New("url is invalid")
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return errors.New("url must use http or https")
	}
	if strings.TrimSpace(u.Hostname()) == "" {
		return errors.New("url host is required")
	}
	return nil
}
