package core

import (
	"strings"
	"testing"
)

func TestPreviewCommandQuotesArgs(t *testing.T) {
	cfg := Config{
		URL:            "https://example.com/watch?v=1",
		Quality:        "best",
		OutputTemplate: "my videos/%(title)s.%(ext)s",
	}
	got := PreviewCommand(cfg, "yt-dlp")
	if got == "" {
		t.Fatal("empty preview")
	}
	if !strings.Contains(got, `"my videos/%(title)s.%(ext)s"`) {
		t.Fatalf("expected quoted template in %q", got)
	}
}

func TestPreviewCommandEmptyProgram(t *testing.T) {
	cfg := Config{URL: "https://example.com/v"}
	got := PreviewCommand(cfg, "")
	if got == "" || !strings.Contains(got, "yt-dlp") {
		t.Fatalf("preview: %q", got)
	}
}

func TestProbeFlags(t *testing.T) {
	cfg := Config{
		UseCookiesFile:    true,
		CookiesFile:       "/tmp/cookies.txt",
		UseCookiesBrowser: true,
		CookiesBrowser:    "firefox",
		Proxy:             "http://127.0.0.1:8080",
		NoPlaylist:        true,
		DenoPath:          "/usr/bin/deno",
	}
	args := ProbeFlags(cfg)
	want := []string{
		"--cookies", "/tmp/cookies.txt",
		"--cookies-from-browser", "firefox",
		"--proxy", "http://127.0.0.1:8080",
		"--no-playlist",
		"--js-runtimes", "deno:/usr/bin/deno",
	}
	if len(args) != len(want) {
		t.Fatalf("got %v want %v", args, want)
	}
	for i, w := range want {
		if args[i] != w {
			t.Fatalf("args[%d]=%q want %q", i, args[i], w)
		}
	}
}

func TestBuildCommandSkipsURLWithLoadInfoJSON(t *testing.T) {
	cfg := Config{LoadInfoJSON: "/tmp/info.json", URL: "https://example.com/v"}
	args := BuildCommand(cfg)
	for _, a := range args {
		if a == "https://example.com/v" {
			t.Fatalf("url should be omitted: %v", args)
		}
	}
	if !containsArg(args, "--load-info-json") {
		t.Fatalf("missing load-info-json: %v", args)
	}
}
