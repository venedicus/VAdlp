package core

import "testing"

func TestConfigForProfileStripsEphemeral(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com/watch"
	cfg.BatchURLs = "https://a.example/1"
	cfg.LoadInfoJSON = "/tmp/info.json"
	cfg.Quality = "1080"

	out := ConfigForProfile(cfg)
	if out.URL != "" || out.BatchURLs != "" || out.LoadInfoJSON != "" {
		t.Fatalf("ephemeral fields not cleared: %+v", out)
	}
	if out.Quality != "1080" {
		t.Fatalf("quality=%q", out.Quality)
	}
}

func TestApplyProfileConfigKeepsURL(t *testing.T) {
	live := DefaultConfig()
	live.URL = "https://keep.me/v"
	live.BatchURLs = "https://batch/1"

	profile := DefaultConfig()
	profile.Quality = "worst"
	profile.OutputTemplate = "custom.%(ext)s"

	ApplyProfileConfig(&live, profile)
	if live.URL != "https://keep.me/v" || live.BatchURLs != "https://batch/1" {
		t.Fatalf("url/batch changed: %+v", live)
	}
	if live.Quality != "worst" || live.OutputTemplate != "custom.%(ext)s" {
		t.Fatalf("profile not applied: %+v", live)
	}
}

func TestValidateProfileName(t *testing.T) {
	if err := ValidateProfileName(""); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateProfileName("bad/name"); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateProfileName("My 1080p"); err != nil {
		t.Fatal(err)
	}
}

func TestApplyPreset(t *testing.T) {
	cfg := DefaultConfig()
	if !ApplyPreset(&cfg, "video_1080") {
		t.Fatal("preset not applied")
	}
	if cfg.Quality == "" || cfg.Format != "mp4" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if ApplyPreset(&cfg, "no_such") {
		t.Fatal("unknown preset should fail")
	}
}
