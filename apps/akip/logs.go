package main

import (
	"strings"
	"sync"
	"time"
)

type LogLevel string

const (
	LogInfo  = "INFO"
	LogWarn  = "WARN"
	LogError = "ERROR"
	LogDebug = "DEBUG"
)

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type LogBuffer struct {
	mu    sync.RWMutex
	limit int
	lines []LogEntry
}

func NewLogBuffer(limit int) *LogBuffer {
	if limit < 1 {
		limit = 200
	}
	return &LogBuffer{
		limit: limit,
		lines: make([]LogEntry, 0, limit),
	}
}

func (b *LogBuffer) Add(level string, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	entry := LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   level,
		Message: message,
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.lines = append(b.lines, entry)
	if len(b.lines) > b.limit {
		b.lines = b.lines[len(b.lines)-b.limit:]
	}
}

func (b *LogBuffer) Snapshot() []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]LogEntry, len(b.lines))
	copy(out, b.lines)
	return out
}
