package downloader

import "testing"

func TestDetectStage(t *testing.T) {
	cases := []struct {
		line string
		want Stage
	}{
		{"[download]  50.0% of ~10MiB", StageDownloading},
		{"[download] Destination: out.mp4", StageDownloading},
		{"[ffmpeg] Merging formats into mkv", StagePostProcess},
		{"[info] Downloading 1 format(s)", StageExtracting},
		{"random log line", StageUnknown},
	}
	for _, tc := range cases {
		if got := detectStage(tc.line); got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.line, got, tc.want)
		}
	}
}
