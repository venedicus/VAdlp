package core

var PresetIDs = []string{
	"youtube_playlist",
	"audio_only",
	"video_best",
	"video_1080",
	"video_4k",
	"podcast",
}

var QualityPresets = []struct {
	Key   string
	Value string
}{
	{Key: "quality.best", Value: "best"},
	{Key: "quality.best_av", Value: "bestvideo+bestaudio/best"},
	{Key: "quality.1080", Value: "bestvideo[height<=1080]+bestaudio/best[height<=1080]"},
	{Key: "quality.720", Value: "bestvideo[height<=720]+bestaudio/best[height<=720]"},
	{Key: "quality.480", Value: "bestvideo[height<=480]+bestaudio/best[height<=480]"},
	{Key: "quality.360", Value: "bestvideo[height<=360]+bestaudio/best[height<=360]"},
	{Key: "quality.4k", Value: "bestvideo[height<=2160]+bestaudio/best[height<=2160]"},
	{Key: "quality.best_audio", Value: "bestaudio/best"},
	{Key: "quality.worst", Value: "worst"},
}

var MergeFormats = []string{"mp4", "webm", "mkv", "mov", "flv", "avi", ""}

func ApplyPreset(cfg *Config, id string) bool {
	switch id {
	case "youtube_playlist":
		ApplyYouTubePlaylistPreset(cfg)
	case "audio_only":
		ApplyAudioOnlyPreset(cfg)
	case "video_best":
		cfg.Quality = "best"
		cfg.Format = "mp4"
		cfg.AudioOnly = false
	case "video_1080":
		cfg.Quality = "bestvideo[height<=1080]+bestaudio/best[height<=1080]"
		cfg.Format = "mp4"
		cfg.AudioOnly = false
	case "video_4k":
		cfg.Quality = "bestvideo[height<=2160]+bestaudio/best[height<=2160]"
		cfg.Format = "mp4"
		cfg.AudioOnly = false
	case "podcast":
		cfg.Quality = "bestaudio/best"
		cfg.AudioOnly = true
		cfg.AudioFormat = "m4a"
		cfg.Format = ""
		cfg.OutputTemplate = "%(title)s.%(ext)s"
		cfg.EmbedMetadata = true
		cfg.EmbedThumbnail = true
		cfg.WriteThumbnail = false
	default:
		return false
	}
	return true
}

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
