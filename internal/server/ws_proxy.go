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
		buffer := make([]byte, 32*1024)
		for {
			msgType, reader, readErr := src.NextReader()
			if readErr != nil {
				code, text := websocketCloseCode(readErr)
				closeBoth(code, text)
				return
			}
			if err := dst.SetWriteDeadline(time.Now().Add(streamWriteTimeout)); err != nil {
				closeBoth(websocket.CloseGoingAway, "")
				return
			}
			writer, writeErr := dst.NextWriter(msgType)
			if writeErr == nil {
				_, writeErr = io.CopyBuffer(streamDeadlineWriter{conn: dst, writer: writer}, reader, buffer)
				closeErr := writer.Close()
				if writeErr == nil {
					writeErr = closeErr
				}
			}
			if writeErr != nil {
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

type streamDeadlineWriter struct {
	conn   *websocket.Conn
	writer io.Writer
}

func (w streamDeadlineWriter) Write(data []byte) (int, error) {
	if err := w.conn.SetWriteDeadline(time.Now().Add(streamWriteTimeout)); err != nil {
		return 0, err
	}
	return w.writer.Write(data)
}

func websocketCloseCode(err error) (int, string) {
	if err == nil {
		return websocket.CloseNormalClosure, ""
	}
	if errors.Is(err, websocket.ErrReadLimit) {
		return websocket.CloseMessageTooBig, "message exceeds stream limit"
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
