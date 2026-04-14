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
		"default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self' ws: wss:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'",
	)
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
	originHost := strings.TrimSpace(originURL.Hostname())
	requestHost := strings.TrimSpace(r.Host)
	if originHost == "" || requestHost == "" {
		return false
	}
	requestHost = strings.Split(requestHost, ":")[0]
	return strings.EqualFold(originHost, requestHost)
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
