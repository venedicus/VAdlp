package core

import "testing"

func TestBuildCommandTable(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		want   []string
		absent []string
	}{
		{
			name: "continue and quality",
			cfg:  Config{URL: "https://x.test/v", Quality: "best", Continue: true},
			want: []string{"--continue", "-f", "best"},
		},
		{
			name: "verbose not quiet",
			cfg:  Config{URL: "https://x.test/v", Verbose: true, Quiet: true},
			want: []string{"-v"},
			absent: []string{"-q"},
		},
		{
			name: "no playlist",
			cfg:  Config{URL: "https://x.test/v", NoPlaylist: true},
			want: []string{"--no-playlist"},
		},
		{
			name: "audio only",
			cfg:  Config{URL: "https://x.test/v", AudioOnly: true, AudioFormat: "mp3"},
			want: []string{"-x", "--audio-format", "mp3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := BuildCommand(tt.cfg)
			joined := " " + joinArgs(args) + " "
			for _, w := range tt.want {
				if !containsArg(args, w) {
					t.Fatalf("missing %q in %v", w, args)
				}
				_ = joined
			}
			for _, a := range tt.absent {
				if containsArg(args, a) {
					t.Fatalf("unexpected %q in %v", a, args)
				}
			}
		})
	}
}

func containsArg(args []string, s string) bool {
	for _, a := range args {
		if a == s {
			return true
		}
	}
	return false
}

func joinArgs(args []string) string {
	out := ""
	for _, a := range args {
		out += a + " "
	}
	return out
}
