package server

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const agentResponseLimit = 32 * 1024 * 1024

var errAgentResponseTooLarge = errors.New("agent response exceeds the 32 MiB limit")

// newAgentHTTPClient keeps existing request budgets and private-network access.
// The standard client's credential forwarding policy is retained on redirects.
func newAgentHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: boundedAgentTransport{base: http.DefaultTransport},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many agent redirects")
			}
			for _, previous := range via {
				if strings.EqualFold(previous.URL.Scheme, "https") && !strings.EqualFold(req.URL.Scheme, "https") {
					return errors.New("agent redirect would downgrade HTTPS; configure the agent's final URL")
				}
			}
			return nil
		},
	}
}

type boundedAgentTransport struct {
	base http.RoundTripper
}

func (t boundedAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if response.ContentLength > agentResponseLimit {
		_ = response.Body.Close()
		return nil, errAgentResponseTooLarge
	}
	response.Body = &boundedAgentBody{ReadCloser: response.Body, remaining: agentResponseLimit}
	return response, nil
}

type boundedAgentBody struct {
	io.ReadCloser
	remaining int64
}

func (b *boundedAgentBody) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if b.remaining == 0 {
		var probe [1]byte
		n, err := b.ReadCloser.Read(probe[:])
		if n > 0 {
			return 0, errAgentResponseTooLarge
		}
		return 0, err
	}
	if int64(len(data)) > b.remaining {
		data = data[:b.remaining]
	}
	n, err := b.ReadCloser.Read(data)
	b.remaining -= int64(n)
	return n, err
}
