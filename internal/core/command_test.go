package core

import "testing"

func TestURLsFromConfigBatch(t *testing.T) {
	cfg := Config{
		URL:       "https://single.example/v",
		BatchURLs: "https://a.example/1\n# comment\n\nhttps://b.example/2",
	}
	urls := URLsFromConfig(cfg)
	if len(urls) != 2 || urls[0] != "https://a.example/1" || urls[1] != "https://b.example/2" {
		t.Fatalf("got %v", urls)
	}
}

func TestBuildCommandSponsorBlock(t *testing.T) {
	cfg := Config{SponsorBlockRemove: true, URL: "https://x.example/v"}
	args := BuildCommand(cfg)
	found := false
	for i, a := range args {
		if a == "--sponsorblock-remove" && i+1 < len(args) && args[i+1] == "all" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing sponsorblock flags: %v", args)
	}
}
