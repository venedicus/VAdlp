package downloader

import "testing"

func TestProgressRegex(t *testing.T) {
	m := progressRegex.FindStringSubmatch("[download]  42.5% of 10MiB")
	if len(m) < 2 || m[1] != "42.5" {
		t.Fatalf("got %v", m)
	}
}

func TestPlaylistRegex(t *testing.T) {
	m := playlistRegex.FindStringSubmatch("[download] Downloading item 3 of 12")
	if len(m) < 3 || m[1] != "3" || m[2] != "12" {
		t.Fatalf("got %v", m)
	}
}

func TestSpeedEtaRegex(t *testing.T) {
	line := "[download]  10% at 1.23MiB/s ETA 01:42"
	if speedRegex.FindStringSubmatch(line) == nil {
		t.Fatal("speed")
	}
	if etaRegex.FindStringSubmatch(line) == nil {
		t.Fatal("eta")
	}
}
