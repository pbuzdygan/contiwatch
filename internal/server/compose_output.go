package server

import (
	"bytes"
	"os/exec"
	"sync"
	"time"
)

const composeOutputLimit = 64 * 1024

// composeOutput retains a bounded diagnostic tail without stopping slow pulls
// when Docker prints a large amount of progress output.
type composeOutput struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
}

func (o *composeOutput) Write(data []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	n := len(data)
	if len(o.data)+n > composeOutputLimit {
		o.truncated = true
		if n >= composeOutputLimit {
			o.data = append(o.data[:0], data[n-composeOutputLimit:]...)
			return n, nil
		}
		o.data = o.data[len(o.data)+n-composeOutputLimit:]
	}
	o.data = append(o.data, data...)
	return n, nil
}

func boundedComposeOutput(cmd *exec.Cmd) (string, error) {
	output := &composeOutput{}
	cmd.Stdout = output
	cmd.Stderr = output
	// Bound pipe cleanup even if a child process inherits Docker's output pipes.
	cmd.WaitDelay = 5 * time.Second
	err := cmd.Run()
	text := string(bytes.ToValidUTF8(output.data, []byte("?")))
	if output.truncated {
		text = "[output truncated; showing final 64 KiB]\n" + text
	}
	return text, err
}
