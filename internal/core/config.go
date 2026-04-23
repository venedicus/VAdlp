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
	AudioFormat       string
	OutputPath        string
	OutputTemplate    string
	UseCookiesFile    bool
	CookiesFile       string
	UseCookiesBrowser bool
	CookiesBrowser    string
	Proxy             string
	RateLimit         string
	PlaylistReverse   bool

	Continue            bool
	NoPart              bool
	PlaylistStart       int
	PlaylistEnd         int
	MaxDownloads        int
	DownloadArchive     string
	NoPlaylist          bool
	FlatPlaylist        bool
	WriteSubs           bool
	WriteAutoSub        bool
	EmbedSubs           bool
	SubLangs            string
	WriteThumbnail      bool
	EmbedThumbnail      bool
	EmbedMetadata       bool
	EmbedChapters       bool
	Retries             int
	FragmentRetries     int
	ConcurrentFragments int
	SocketTimeout       int
	NoWarnings          bool
	Verbose             bool
	Quiet               bool
	WriteInfoJSON       bool
	LoadInfoJSON        string
	WindowsFilenames    bool
	NoMtime             bool
	AbortOnError        bool
	IgnoreErrors        bool
	ExtraArgs           string
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
		Quality:             "best",
		Format:              "mp4",
		OutputPath:          outputPath,
		OutputTemplate:      "%(upload_date)s - %(title)s.%(ext)s",
		UseCookiesBrowser:   true,
		CookiesBrowser:      "chrome",
		PlaylistReverse:     true,
		Continue:            true,
		SubLangs:            "en.*,ru.*",
		Retries:             10,
		FragmentRetries:     10,
		ConcurrentFragments: 1,
	}
}
