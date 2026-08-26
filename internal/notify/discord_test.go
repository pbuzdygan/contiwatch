package notify

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

type failingRoundTripper struct {
	err error
}

func (f failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, f.err
}

func TestDiscordRequestErrorDoesNotExposeWebhookURL(t *testing.T) {
	const secret = "secret-webhook-token"
	client := NewDiscordClient("https://discord.com/api/webhooks/123/" + secret)
	client.Client = &http.Client{
		Transport: failingRoundTripper{err: errors.New("request failed for https://discord.com/api/webhooks/123/" + secret)},
	}

	err := client.Send("test")
	if err == nil {
		t.Fatalf("expected request error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("webhook secret leaked in error: %v", err)
	}
}
