package main

import (
	"context"
	"embed"
	"io"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vadlp/internal/app"
	"vadlp/internal/settings"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	log.SetOutput(io.Discard)
	appSettings, err := settings.Load()
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("vadlp: load settings: %v", err)
		log.SetOutput(io.Discard)
	}
	width, height := float32(1280), float32(820)
	if appSettings.WindowWidth >= 960 {
		width = appSettings.WindowWidth
	}
	if appSettings.WindowHeight >= 640 {
		height = appSettings.WindowHeight
	}
	application := app.New()
	err = wails.Run(&options.App{
		Title:     "VAdlp",
		Width:     int(width),
		Height:    int(height),
		MinWidth:  960,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 15, A: 255},
		OnStartup:        application.Startup,
		OnShutdown: func(ctx context.Context) {
			application.Shutdown(ctx)
		},
		OnBeforeClose: func(ctx context.Context) bool {
			if application.ShouldQuit() {
				return false
			}
			runtime.WindowHide(ctx)
			return true
		},
		Bind: []interface{}{
			application,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
