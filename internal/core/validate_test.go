package core

import "testing"

func TestConfigValidateRetries(t *testing.T) {
	c := Config{Retries: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if v, ok := AsValidation(err); !ok || v.Key != "err.config.retries" {
		t.Fatalf("got %v", err)
	}
}

func TestConfigValidatePlaylistRange(t *testing.T) {
	c := Config{PlaylistStart: 5, PlaylistEnd: 2}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigValidateForDownloadNoURL(t *testing.T) {
	c := DefaultConfig()
	c.URL = ""
	c.BatchURLs = ""
	if err := c.ValidateForDownload(); err == nil {
		t.Fatal("expected error")
	}
}
