package core

import (
	"runtime"
	"strings"
)

func (c *Config) Validate() error {
	c.Normalize()
	if err := validateRetries(c.Retries); err != nil {
		return err
	}
	if err := validateRetries(c.FragmentRetries); err != nil {
		return err
	}
	if c.ConcurrentFragments < 0 || c.ConcurrentFragments > 128 {
		return ValidationError{Key: "err.config.concurrent_frags"}
	}
	if c.SocketTimeout < 0 || c.SocketTimeout > 86400 {
		return ValidationError{Key: "err.config.socket_timeout"}
	}
	if c.PlaylistStart < 0 || c.PlaylistEnd < 0 {
		return ValidationError{Key: "err.config.playlist_range"}
	}
	if c.PlaylistEnd > 0 && c.PlaylistStart > 0 && c.PlaylistStart > c.PlaylistEnd {
		return ValidationError{Key: "err.config.playlist_range"}
	}
	if c.MaxDownloads < 0 || c.MaxDownloads > 10000 {
		return ValidationError{Key: "err.config.max_downloads"}
	}
	if p := strings.TrimSpace(c.OutputPath); p != "" {
		if err := ValidatePath(p); err != nil {
			return err
		}
	}
	if p := strings.TrimSpace(c.CookiesFile); p != "" && c.UseCookiesFile {
		if err := ValidatePath(p); err != nil {
			return ValidationError{Key: "err.config.cookies_path"}
		}
	}
	return nil
}

func (c *Config) ValidateForDownload() error {
	if err := c.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.LoadInfoJSON) == "" && len(URLsFromConfig(*c)) == 0 {
		return ValidationError{Key: "err.queue_no_url"}
	}
	return nil
}

func validateRetries(n int) error {
	if n < 0 || n > 100 {
		return ValidationError{Key: "err.config.retries"}
	}
	return nil
}

func ValidatePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if strings.Contains(path, "\x00") {
		return ValidationError{Key: "err.config.invalid_path"}
	}
	// Do not forbid ':' — required for Windows drive letters (C:\...).
	bad := `<>|?*"`
	if runtime.GOOS == "windows" {
		bad = `<>|?*"`
	}
	if strings.ContainsAny(path, bad) {
		return ValidationError{Key: "err.config.invalid_path"}
	}
	return nil
}
