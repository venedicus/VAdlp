package updater

import "testing"

func TestVersionTokenRe(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"2025.06.25", "2025.06.25"},
		{"ffmpeg version 6.1.1-essentials_build 20231121-ge1e3f4c4f7", "6.1.1-essentials_build"},
		{"deno 2.9.0 (stable, release, x86_64-pc-windows-msvc)", "2.9.0"},
	}
	for _, c := range cases {
		if got := versionTokenRe.FindString(c.line); got != c.want {
			t.Errorf("FindString(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}
