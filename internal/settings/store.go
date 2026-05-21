package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"vadlp/internal/core"
)

const fileVersion = 1

type App struct {
	Version      int         `json:"version"`
	Config       core.Config `json:"config"`
	YtDlpPath    string      `json:"ytDlpPath,omitempty"`
	FFmpegPath   string      `json:"ffmpegPath,omitempty"`
	SessionPath  string      `json:"sessionPath,omitempty"`
	WindowWidth     float32 `json:"windowWidth,omitempty"`
	WindowHeight    float32 `json:"windowHeight,omitempty"`
	QueueParallel   int     `json:"queueParallel,omitempty"`
	Language        string  `json:"language,omitempty"`
	DenoPath        string  `json:"denoPath,omitempty"`
	LastProfile     string  `json:"lastProfile,omitempty"`
}

func Default() App {
	return App{
		Version: fileVersion,
		Config:  core.DefaultConfig(),
	}
}

func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, err2 := os.UserHomeDir()
		if err2 != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	dir = filepath.Join(dir, "vadlp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

func Load() (App, error) {
	path, err := Path()
	if err != nil {
		return Default(), err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), err
	}
	var app App
	if err := json.Unmarshal(b, &app); err != nil {
		return Default(), err
	}
	if app.Version < 1 {
		app.Version = fileVersion
	}
	return app, nil
}

func Save(app App) error {
	path, err := Path()
	if err != nil {
		return err
	}
	app.Version = fileVersion
	b, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
