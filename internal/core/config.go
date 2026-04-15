package core

import (
	"os"
	"path/filepath"
)

// Config stores all yt-dlp options controlled by the UI.
type Config struct {
	URL               string
	Quality           string
	Format            string
	AudioOnly         bool
	OutputPath        string
	OutputTemplate    string
	UseCookiesFile    bool
	CookiesFile       string
	UseCookiesBrowser bool
	CookiesBrowser    string
	Proxy             string
	RateLimit         string
	PlaylistReverse   bool
}

func DefaultConfig() Config {
	outputPath := ""
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		outputPath = filepath.Join(home, "Videos")
	}
	if outputPath == "" {
		outputPath = "."
	}

	return Config{
		Quality:           "best",
		Format:            "mp4",
		OutputPath:        outputPath,
		OutputTemplate:    "%(upload_date)s - %(title)s.%(ext)s",
		UseCookiesBrowser: true,
		CookiesBrowser:    "vivaldi",
		PlaylistReverse:   true,
	}
}
