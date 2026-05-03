package core

import (
	"strconv"
	"strings"
)

func BuildCommand(cfg Config) []string {
	args := []string{}
	if cfg.Continue { args = append(args, "--continue") }
	if cfg.NoPart { args = append(args, "--no-part") }
	if cfg.PlaylistStart > 0 { args = append(args, "--playlist-start", strconv.Itoa(cfg.PlaylistStart)) }
	if cfg.PlaylistEnd > 0 { args = append(args, "--playlist-end", strconv.Itoa(cfg.PlaylistEnd)) }
	if cfg.MaxDownloads > 0 { args = append(args, "--max-downloads", strconv.Itoa(cfg.MaxDownloads)) }
	if strings.TrimSpace(cfg.DownloadArchive) != "" { args = append(args, "--download-archive", strings.TrimSpace(cfg.DownloadArchive)) }
	if cfg.NoPlaylist { args = append(args, "--no-playlist") }
	if cfg.FlatPlaylist { args = append(args, "--flat-playlist") }
	if cfg.Quality != "" { args = append(args, "-f", cfg.Quality) }
	if cfg.Format != "" { args = append(args, "--merge-output-format", cfg.Format) }
	if cfg.AudioOnly {
		args = append(args, "-x")
		if strings.TrimSpace(cfg.AudioFormat) != "" { args = append(args, "--audio-format", strings.TrimSpace(cfg.AudioFormat)) }
	}
	if cfg.OutputPath != "" { args = append(args, "-P", cfg.OutputPath) }
	if cfg.UseCookiesFile && cfg.CookiesFile != "" { args = append(args, "--cookies", cfg.CookiesFile) }
	if cfg.UseCookiesBrowser && cfg.CookiesBrowser != "" { args = append(args, "--cookies-from-browser", cfg.CookiesBrowser) }
	if cfg.Proxy != "" { args = append(args, "--proxy", cfg.Proxy) }
	if cfg.RateLimit != "" { args = append(args, "--limit-rate", cfg.RateLimit) }
	if cfg.PlaylistReverse { args = append(args, "--playlist-reverse") }
	if cfg.OutputTemplate != "" { args = append(args, "-o", cfg.OutputTemplate) }
	if cfg.WriteSubs { args = append(args, "--write-subs") }
	if cfg.WriteAutoSub { args = append(args, "--write-auto-sub") }
	if cfg.EmbedSubs { args = append(args, "--embed-subs") }
	if strings.TrimSpace(cfg.SubLangs) != "" { args = append(args, "--sub-langs", strings.TrimSpace(cfg.SubLangs)) }
	if cfg.WriteThumbnail { args = append(args, "--write-thumbnail") }
	if cfg.EmbedThumbnail { args = append(args, "--embed-thumbnail") }
	if cfg.EmbedMetadata { args = append(args, "--embed-metadata") }
	if cfg.EmbedChapters { args = append(args, "--embed-chapters") }
	if cfg.WriteInfoJSON { args = append(args, "--write-info-json") }
	if strings.TrimSpace(cfg.LoadInfoJSON) != "" { args = append(args, "--load-info-json", strings.TrimSpace(cfg.LoadInfoJSON)) }
	if cfg.WindowsFilenames { args = append(args, "--windows-filenames") }
	if cfg.NoMtime { args = append(args, "--no-mtime") }
	if cfg.Retries > 0 { args = append(args, "--retries", strconv.Itoa(cfg.Retries)) }
	if cfg.FragmentRetries >= 0 { args = append(args, "--fragment-retries", strconv.Itoa(cfg.FragmentRetries)) }
	if cfg.ConcurrentFragments > 0 { args = append(args, "--concurrent-fragments", strconv.Itoa(cfg.ConcurrentFragments)) }
	if cfg.SocketTimeout > 0 { args = append(args, "--socket-timeout", strconv.Itoa(cfg.SocketTimeout)) }
	if cfg.NoWarnings { args = append(args, "--no-warnings") }
	if cfg.Verbose { args = append(args, "-v") }
	if cfg.Quiet { args = append(args, "-q") }
	if cfg.AbortOnError { args = append(args, "--abort-on-error") }
	if cfg.IgnoreErrors { args = append(args, "--ignore-errors") }
	args = appendParsedExtra(args, cfg.ExtraArgs)
	if strings.TrimSpace(cfg.LoadInfoJSON) == "" && cfg.URL != "" {
		args = append(args, cfg.URL)
	}
	return args
}

func appendParsedExtra(args []string, raw string) []string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		parts := strings.Fields(line)
		if len(parts) > 0 { args = append(args, parts...) }
	}
	return args
}

func quoteArg(arg string) string {
	if arg == "" { return `""` }
	if strings.ContainsAny(arg, " \t\"") { return strconv.Quote(arg) }
	return arg
}

func PreviewCommand(cfg Config, program string) string {
	if strings.TrimSpace(program) == "" { program = "yt-dlp" }
	args := BuildCommand(cfg)
	quoted := make([]string, 0, len(args))
	for _, arg := range args { quoted = append(quoted, quoteArg(arg)) }
	return quoteArg(program) + " " + strings.Join(quoted, " ")
}
