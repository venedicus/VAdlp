package core

import (
	"strconv"
	"strings"
)

// BuildCommand converts config to yt-dlp args.
func BuildCommand(cfg Config) []string {
	args := []string{}

	if cfg.Quality != "" {
		args = append(args, "-f", cfg.Quality)
	}
	if cfg.Format != "" {
		args = append(args, "--merge-output-format", cfg.Format)
	}
	if cfg.AudioOnly {
		args = append(args, "-x")
	}
	if cfg.OutputPath != "" {
		args = append(args, "-P", cfg.OutputPath)
	}
	if cfg.UseCookiesFile && cfg.CookiesFile != "" {
		args = append(args, "--cookies", cfg.CookiesFile)
	}
	if cfg.UseCookiesBrowser && cfg.CookiesBrowser != "" {
		args = append(args, "--cookies-from-browser", cfg.CookiesBrowser)
	}
	if cfg.Proxy != "" {
		args = append(args, "--proxy", cfg.Proxy)
	}
	if cfg.RateLimit != "" {
		args = append(args, "--limit-rate", cfg.RateLimit)
	}
	if cfg.PlaylistReverse {
		args = append(args, "--playlist-reverse")
	}
	if cfg.OutputTemplate != "" {
		args = append(args, "-o", cfg.OutputTemplate)
	}
	if cfg.URL != "" {
		args = append(args, cfg.URL)
	}
	return args
}

func quoteArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if strings.ContainsAny(arg, " \t\"") {
		return strconv.Quote(arg)
	}
	return arg
}

func PreviewCommand(cfg Config) string {
	args := BuildCommand(cfg)
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, quoteArg(arg))
	}
	return "yt-dlp " + strings.Join(quoted, " ")
}
