package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"vadlp/internal/configdir"
	"vadlp/internal/core"
)

const fileVersion = 4

type App struct {
	Version             int         `json:"version"`
	Config              core.Config `json:"config"`
	FFmpegPath          string      `json:"ffmpegPath,omitempty"`
	SessionPath         string      `json:"sessionPath,omitempty"`
	WindowWidth         float32     `json:"windowWidth,omitempty"`
	WindowHeight        float32     `json:"windowHeight,omitempty"`
	ActivityPanelOffset float64     `json:"activityPanelOffset,omitempty"`
	ActivityPanelOpen   bool        `json:"activityPanelOpen,omitempty"`
	QueueParallel       int         `json:"queueParallel,omitempty"`
	Language            string      `json:"language,omitempty"`
	DenoPath            string      `json:"denoPath,omitempty"`
	LastProfile         string      `json:"lastProfile,omitempty"`
	DebugLog            bool        `json:"debugLog,omitempty"`
	UIScale             float32     `json:"uiScale,omitempty"`
}

func Default() App {
	return App{
		Version:             fileVersion,
		Config:              core.DefaultConfig(),
		ActivityPanelOffset: 0.4,
		QueueParallel:       1,
		UIScale:             0,
	}
}

func Dir() (string, error) {
	return configdir.Dir()
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
	migrate(&app)
	return app, nil
}

func migrate(app *App) {
	if app.Version < 1 {
		app.Version = 1
	}
	if app.Version < 2 {
		if app.ActivityPanelOffset <= 0 || app.ActivityPanelOffset >= 1 {
			app.ActivityPanelOffset = 0.4
		}
		app.Version = 2
	}
	if app.Version < 3 {
		if app.UIScale <= 0 {
			app.UIScale = 1.15
		}
		app.Version = 3
	}
	if app.Version < 4 {
		if app.UIScale == 1.15 {
			app.UIScale = 0
		}
		app.Version = 4
	}
	if app.QueueParallel < 1 {
		app.QueueParallel = 1
	}
	app.Config.Normalize()
}

func Validate(app App) error {
	if app.QueueParallel < 1 || app.QueueParallel > 32 {
		return core.ValidationError{Key: "err.config.queue_workers"}
	}
	if app.UIScale != 0 && (app.UIScale < 0.9 || app.UIScale > 1.5) {
		return core.ValidationError{Key: "err.config.ui_scale"}
	}
	return app.Config.Validate()
}

func Save(app App) error {
	if err := Validate(app); err != nil {
		return err
	}
	path, err := Path()
	if err != nil {
		return err
	}
	migrate(&app)
	app.Version = fileVersion
	b, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
