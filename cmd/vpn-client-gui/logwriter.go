package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogEntry is emitted to the frontend for each log line.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// EventLogWriter implements io.Writer and emits log lines as Wails events.
type EventLogWriter struct {
	ctx context.Context
}

func (w *EventLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	entry := LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Message:   msg,
	}

	runtime.EventsEmit(w.ctx, "vpn:log", entry)

	// Also write to stderr for debugging
	fmt.Fprintf(defaultStderr, "%s %s\n", entry.Timestamp, entry.Message)

	return len(p), nil
}
