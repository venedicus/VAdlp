package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryAppendAndClear(t *testing.T) {
	dir := t.TempDir()
	prev := historyConfigDir
	historyConfigDir = func() (string, error) { return dir, nil }
	defer func() { historyConfigDir = prev }()

	if err := AppendHistory(HistoryItem{URL: "https://example.com/a", Status: "ok"}); err != nil {
		t.Fatal(err)
	}
	h, err := LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Items) != 1 || h.Items[0].URL != "https://example.com/a" {
		t.Fatalf("history: %+v", h)
	}
	if h.Items[0].At.IsZero() {
		t.Fatal("expected timestamp")
	}

	if err := ClearHistory(); err != nil {
		t.Fatal(err)
	}
	h, err = LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Items) != 0 {
		t.Fatalf("expected empty history, got %+v", h)
	}
	path := filepath.Join(dir, "history.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("history file should be gone: %v", err)
	}
}

func TestAppendHistoryPreservesExplicitTime(t *testing.T) {
	dir := t.TempDir()
	prev := historyConfigDir
	historyConfigDir = func() (string, error) { return dir, nil }
	defer func() { historyConfigDir = prev }()

	at := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := AppendHistory(HistoryItem{At: at, URL: "u", Status: "ok"}); err != nil {
		t.Fatal(err)
	}
	h, err := LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	if !h.Items[0].At.Equal(at) {
		t.Fatalf("at=%v", h.Items[0].At)
	}
}
