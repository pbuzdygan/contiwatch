package server

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func proxyWebSockets(a, b *websocket.Conn) {
	if a == nil || b == nil {
		return
	}

	done := make(chan struct{})
	var once sync.Once

	closeBoth := func(code int, text string) {
		once.Do(func() {
			deadline := time.Now().Add(800 * time.Millisecond)
			_ = a.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), deadline)
			_ = b.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), deadline)
			_ = a.Close()
			_ = b.Close()
			close(done)
		})
	}

	pump := func(dst, src *websocket.Conn) {
		for {
			msgType, data, readErr := src.ReadMessage()
			if readErr != nil {
				code, text := websocketCloseCode(readErr)
				closeBoth(code, text)
				return
			}
			if writeErr := dst.WriteMessage(msgType, data); writeErr != nil {
				code, text := websocketCloseCode(writeErr)
				closeBoth(code, text)
				return
			}
		}
	}

	go pump(b, a)
	go pump(a, b)
	<-done
}

func websocketCloseCode(err error) (int, string) {
	if err == nil {
		return websocket.CloseNormalClosure, ""
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		code := closeErr.Code
		if code == 0 {
			code = websocket.CloseNormalClosure
		}
		return code, closeErr.Text
	}
	if errors.Is(err, io.EOF) {
		return websocket.CloseGoingAway, ""
	}
	return websocket.CloseGoingAway, ""
}

