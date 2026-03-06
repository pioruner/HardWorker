package loger

import (
	"strings"
	"sync"
)

const maxLines = 100

type logLevel int

const (
	levelInfo logLevel = iota
	levelWarn
	levelError
	levelDebug
)

type logLine struct {
	text  string
	level logLevel
}

type LogUI struct {
	mu         sync.Mutex
	lines      []logLine
	autoScroll bool
}

func New() *LogUI {
	return &LogUI{
		autoScroll: true,
	}
}

type UiWriter struct {
	Ui *LogUI
}

func (w *UiWriter) Write(p []byte) (n int, err error) {
	text := strings.TrimRight(string(p), "\n")

	level := detectLevel(text)

	w.Ui.add(text, level)

	return len(p), nil
}

func detectLevel(text string) logLevel {

	switch {
	case strings.Contains(text, "ERROR"):
		return levelError
	case strings.Contains(text, "WARN"):
		return levelWarn
	case strings.Contains(text, "DEBUG"):
		return levelDebug
	default:
		return levelInfo
	}
}

func (ui *LogUI) add(text string, level logLevel) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ui.lines = append(ui.lines, logLine{text, level})

	if len(ui.lines) > maxLines {
		ui.lines = ui.lines[1:]
	}
}
