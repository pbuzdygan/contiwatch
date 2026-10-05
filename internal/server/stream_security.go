package server

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxContainerStreams      = 64
	clientStreamMessageLimit = 256 * 1024
	agentStreamMessageLimit  = 16 * 1024 * 1024
	streamWriteTimeout       = 30 * time.Second
	streamPongTimeout        = 90 * time.Second
	streamPingInterval       = 30 * time.Second
)

func (s *Server) reserveContainerStream(w http.ResponseWriter) (func(), bool) {
	s.streamMu.Lock()
	defer s.streamMu.Unlock()
	if s.activeStreams >= maxContainerStreams {
		writeError(w, http.StatusTooManyRequests, errors.New("too many active container streams"))
		return nil, false
	}
	s.activeStreams++
	return func() {
		s.streamMu.Lock()
		s.activeStreams--
		s.streamMu.Unlock()
	}, true
}

// protectContainerStream uses standard WebSocket control frames supported by
// older agents and browsers; it does not add a new application-level protocol.
func (s *Server) protectContainerStream(r *http.Request, conn *websocket.Conn, limit int64) func() {
	conn.SetReadLimit(limit)
	_ = conn.SetReadDeadline(time.Now().Add(streamPongTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(streamPongTimeout))
	})
	sessionDone := s.pinSessionDone(r)
	stop := make(chan struct{})
	var once sync.Once
	go func() {
		ticker := time.NewTicker(streamPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-sessionDone:
				_ = conn.Close()
				return
			case <-r.Context().Done():
				_ = conn.Close()
				return
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(streamWriteTimeout)); err != nil {
					_ = conn.Close()
					return
				}
			}
		}
	}()
	return func() { once.Do(func() { close(stop) }) }
}

func writeStreamMessage(conn *websocket.Conn, kind int, data []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(streamWriteTimeout)); err != nil {
		return err
	}
	return conn.WriteMessage(kind, data)
}
