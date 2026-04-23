package core

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const SessionFileVersion = 1

// Session captures recoverable download state (playlist position, batch progress, config).
type Session struct {
	Version         int       `json:"version"`
	SavedAt         time.Time `json:"savedAt"`
	Config          Config    `json:"config"`
	PlaylistCurrent int       `json:"playlistCurrent"`
	PlaylistTotal   int       `json:"playlistTotal"`
	QueueDone       int       `json:"queueDone"`
	QueueTotal      int       `json:"queueTotal"`
}

func (s *Session) ApplyResumeHints() Config {
	out := s.Config
	if s.PlaylistCurrent > 0 {
		out.PlaylistStart = s.PlaylistCurrent
	}
	out.Continue = true
	return out
}

func SaveSession(path string, s Session) error {
	s.Version = SessionFileVersion
	s.SavedAt = time.Now().UTC()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func LoadSession(path string) (Session, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return Session{}, err
	}
	if s.Version < 1 || s.Version > SessionFileVersion {
		return Session{}, fmt.Errorf("unsupported session version %d", s.Version)
	}
	return s, nil
}
