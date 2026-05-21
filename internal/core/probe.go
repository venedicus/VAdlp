package core

import "strings"

func ProbeFlags(cfg Config) []string {
	var args []string
	if cfg.UseCookiesFile && strings.TrimSpace(cfg.CookiesFile) != "" {
		args = append(args, "--cookies", strings.TrimSpace(cfg.CookiesFile))
	}
	if cfg.UseCookiesBrowser && strings.TrimSpace(cfg.CookiesBrowser) != "" {
		args = append(args, "--cookies-from-browser", cfg.CookiesBrowser)
	}
	if strings.TrimSpace(cfg.Proxy) != "" {
		args = append(args, "--proxy", strings.TrimSpace(cfg.Proxy))
	}
	if cfg.NoPlaylist {
		args = append(args, "--no-playlist")
	}
	if strings.TrimSpace(cfg.DenoPath) != "" {
		args = append(args, "--js-runtimes", "deno:"+strings.TrimSpace(cfg.DenoPath))
	}
	return args
}
