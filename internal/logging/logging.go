package logging

import (
	"io"
	"log"
	"os"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

type Logger struct {
	level Level
	l     *log.Logger
}

func New(level string) Logger {
	return NewWithWriter(level, os.Stderr)
}

func NewWithWriter(level string, w io.Writer) Logger {
	if w == nil {
		w = os.Stderr
	}
	return Logger{level: parseLevel(level), l: log.New(w, "", log.LstdFlags)}
}

func parseLevel(s string) Level {
	switch s {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	default:
		return Info
	}
}

func (lg Logger) Debugf(format string, args ...any) {
	if lg.level <= Debug {
		lg.l.Printf("[DEBUG] "+format, args...)
	}
}

func (lg Logger) Infof(format string, args ...any) {
	if lg.level <= Info {
		lg.l.Printf("[INFO] "+format, args...)
	}
}

func (lg Logger) Warnf(format string, args ...any) {
	if lg.level <= Warn {
		lg.l.Printf("[WARN] "+format, args...)
	}
}

func (lg Logger) Errorf(format string, args ...any) {
	if lg.level <= Error {
		lg.l.Printf("[ERROR] "+format, args...)
	}
}
