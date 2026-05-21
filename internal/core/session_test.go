package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	cfg := DefaultConfig()
	cfg.URL = "https://example.com/v"
	cfg.PlaylistStart = 3

	in := Session{
		Config:          cfg,
		PlaylistCurrent: 3,
		PlaylistTotal:   10,
		QueueDone:       2,
		QueueTotal:      5,
	}
	if err := SaveSession(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := LoadSession(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.PlaylistCurrent != 3 || out.QueueTotal != 5 {
		t.Fatalf("session: %+v", out)
	}
}

func TestSessionApplyResumeHints(t *testing.T) {
	s := Session{
		Config:          DefaultConfig(),
		PlaylistCurrent: 4,
	}
	cfg := s.ApplyResumeHints()
	if cfg.PlaylistStart != 4 || !cfg.Continue {
		t.Fatalf("resume hints: %+v", cfg)
	}
}

func TestLoadSessionUnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"version":99,"savedAt":"2020-01-01T00:00:00Z","config":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSession(path); err == nil {
		t.Fatal("expected version error")
	}
}
