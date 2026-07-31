package downloader

import "testing"

func TestProgressRegex(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"[download]  42.5% of 10MiB", "42.5"},
		{"[download]   7% of 1.2GiB at 3.4MiB/s", "7"},
		{"[download] 100% of 10MiB", "100"},
		{"[download] Destination: 100% Barbie (2023).mp4", ""},
		{"[youtube] 25%: some title with 50% in it", ""},
		{"Merging formats into " + `"video 100% better.webm"`, ""},
		{"Extracting URL: https://example.com/50%20off", ""},
	}
	for _, tc := range cases {
		m := progressRegex.FindStringSubmatch(tc.line)
		if tc.want == "" {
			if m != nil {
				t.Errorf("%q: unexpected match %v", tc.line, m)
			}
			continue
		}
		if len(m) < 2 || m[1] != tc.want {
			t.Errorf("%q: got %v, want %q", tc.line, m, tc.want)
		}
	}
}

func TestIsProgressLine(t *testing.T) {
	if !IsProgressLine("[download]  42.5% of 10MiB") {
		t.Fatal("progress line not detected")
	}
	if IsProgressLine("[download] Destination: 100% Barbie (2023).mp4") {
		t.Fatal("destination line should not be treated as progress")
	}
	if IsProgressLine("[youtube] extracting 50% of video") {
		t.Fatal("non-download percentage should not be treated as progress")
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

func TestSpeedRegexApproximate(t *testing.T) {
	m := speedRegex.FindStringSubmatch("[download]  10% at ~1.23MiB/s ETA 01:42")
	if len(m) < 2 || m[1] != "1.23MiB/s" {
		t.Fatalf("got %v", m)
	}
}
