package applog

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"vadlp/internal/configdir"
)

var (
	mu     sync.Mutex
	logger *slog.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
)

func Init(debug bool) error {
	dir, err := configdir.Dir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "vadlp.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	mu.Lock()
	logger = slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: level}))
	mu.Unlock()
	return nil
}

func Info(msg string, args ...any) {
	mu.Lock()
	l := logger
	mu.Unlock()
	l.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	mu.Lock()
	l := logger
	mu.Unlock()
	l.Debug(msg, args...)
}
