package core

import "strings"

// AppendNetworkArgs adds cookies, proxy, and deno runtime flags shared by download and probe commands.
func AppendNetworkArgs(args []string, cfg Config) []string {
	if cfg.UseCookiesFile && strings.TrimSpace(cfg.CookiesFile) != "" {
		args = append(args, "--cookies", strings.TrimSpace(cfg.CookiesFile))
	}
	if cfg.UseCookiesBrowser && strings.TrimSpace(cfg.CookiesBrowser) != "" {
		args = append(args, "--cookies-from-browser", cfg.CookiesBrowser)
	}
	if strings.TrimSpace(cfg.Proxy) != "" {
		args = append(args, "--proxy", strings.TrimSpace(cfg.Proxy))
	}
	args = append(args, DenoRuntimeArgs(cfg.DenoPath)...)
	return args
}

// DenoRuntimeArgs returns yt-dlp --js-runtimes flags when deno path is set.
func DenoRuntimeArgs(denoPath string) []string {
	p := strings.TrimSpace(denoPath)
	if p == "" {
		return nil
	}
	return []string{"--js-runtimes", "deno:" + p}
}
