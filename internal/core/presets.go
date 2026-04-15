package core

func ApplyYouTubePlaylistPreset(cfg *Config) {
	cfg.Quality = "best"
	cfg.Format = "mp4"
	cfg.AudioOnly = false
	cfg.PlaylistReverse = true
	cfg.UseCookiesBrowser = true
	cfg.CookiesBrowser = "chrome"
}

func ApplyAudioOnlyPreset(cfg *Config) {
	cfg.AudioOnly = true
	cfg.Format = ""
	cfg.OutputTemplate = "%(title)s.%(ext)s"
}
