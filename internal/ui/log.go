package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu     sync.Mutex
	silent bool
	debug  bool
}

func NewLogger(silent, debug bool) *Logger {
	return &Logger{silent: silent, debug: debug}
}

func (l *Logger) Infof(format string, a ...any) {
	if l.silent {
		return
	}
	l.print("[INF]", ansiCyan, format, a...)
}

func (l *Logger) Warnf(format string, a ...any) {
	if l.silent {
		return
	}
	l.print("[WRN]", ansiYellow, format, a...)
}

func (l *Logger) Errorf(format string, a ...any) {
	// Errors should still show even if silent? Requirement says silent prints only results.
	// So: respect silent (no extra output).
	if l.silent {
		return
	}
	l.print("[ERR]", ansiRed, format, a...)
}

func (l *Logger) Debugf(format string, a ...any) {
	if l.silent || !l.debug {
		return
	}
	l.print("[DBG]", ansiMagenta, format, a...)
}

func (l *Logger) print(tag string, color string, format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().Format("15:04:05")
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s %s%s%s %s\n", ts, color, tag, ansiReset, msg)
}

// minimal ANSI palette (PD-like vibe)
const (
	ansiReset   = "\x1b[0m"
	ansiRed     = "\x1b[31m"
	ansiYellow  = "\x1b[33m"
	ansiCyan    = "\x1b[36m"
	ansiMagenta = "\x1b[35m"
)
