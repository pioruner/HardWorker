package logger

import (
	"log"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

var base = log.New(os.Stderr, "", log.LstdFlags)

func Infof(format string, args ...any) {
	logf("INFO", colorCyan, format, args...)
}

func Warnf(format string, args ...any) {
	logf("WARN", colorYellow, format, args...)
}

func Errorf(format string, args ...any) {
	logf("ERROR", colorRed, format, args...)
}

func logf(level string, color string, format string, args ...any) {
	base.Printf(colorize(level, color)+" "+format, args...)
}

func colorize(level string, color string) string {
	if os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return "[" + level + "]"
	}
	return color + "[" + level + "]" + colorReset
}
