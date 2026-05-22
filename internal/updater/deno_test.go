package updater

import (
	"runtime"
	"strings"
	"testing"
)

func TestDenoDownloadURLMatchesPlatform(t *testing.T) {
	url, err := DenoDownloadURL()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://github.com/denoland/deno/releases/latest/download/deno-") {
		t.Fatalf("url %q", url)
	}
	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(url, "windows") {
			t.Fatalf("url %q", url)
		}
	case "darwin":
		if !strings.Contains(url, "apple-darwin") {
			t.Fatalf("url %q", url)
		}
	default:
		if !strings.Contains(url, "linux-gnu") {
			t.Fatalf("url %q", url)
		}
	}
}
