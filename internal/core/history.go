package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"vadlp/internal/configdir"
)

const maxHistory = 200

type HistoryItem struct {
	At          time.Time `json:"at"`
	URL         string    `json:"url"`
	Status      string    `json:"status"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	DurationSec int       `json:"durationSec,omitempty"`
}

type History struct {
	Items []HistoryItem `json:"items"`
}

func historyPath() (string, error) {
	dir, err := configdir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

func LoadHistory() (History, error) {
	path, err := historyPath()
	if err != nil {
		return History{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return History{}, nil
		}
		return History{}, err
	}
	var h History
	return h, json.Unmarshal(b, &h)
}

func AppendHistory(item HistoryItem) error {
	h, err := LoadHistory()
	if err != nil {
		return err
	}
	if item.At.IsZero() {
		item.At = time.Now().UTC()
	}
	h.Items = append([]HistoryItem{item}, h.Items...)
	if len(h.Items) > maxHistory {
		h.Items = h.Items[:maxHistory]
	}
	path, err := historyPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
