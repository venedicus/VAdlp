package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchChecksumManifestFormats(t *testing.T) {
	const (
		hashA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		hashB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		hashF = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	)
	tests := []struct {
		name     string
		manifest string
		artifact string
		wantHash string
		wantErr  bool
	}{
		{
			name:     "sha256sum style with space",
			manifest: hashA + "  yt-dlp.exe\n" + hashB + "  SHA2-256SUMS\n",
			artifact: "yt-dlp.exe",
			wantHash: hashA,
		},
		{
			name:     "sha256sum style with asterisk",
			manifest: hashA + " *yt-dlp_macos\n",
			artifact: "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos",
			wantHash: hashA,
		},
		{
			name:     "nested path in manifest",
			manifest: hashF + "  bin/ffmpeg-master-latest-win64-gpl.zip\n",
			artifact: "ffmpeg-master-latest-win64-gpl.zip",
			wantHash: hashF,
		},
		{
			name:     "missing entry",
			manifest: hashA + "  other.zip\n",
			artifact: "missing.zip",
			wantErr:  true,
		},
		{
			name:     "malformed hash",
			manifest: "zz  yt-dlp.exe\n",
			artifact: "yt-dlp.exe",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := parseChecksumLine([]byte(tt.manifest), tt.artifact)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got hash %q", hash)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if hash != tt.wantHash {
				t.Fatalf("hash %q, want %q", hash, tt.wantHash)
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tool.bin")
	content := []byte("hello tool")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	want := hex.EncodeToString(sum[:])
	if err := verifyChecksum(path, want); err != nil {
		t.Fatal(err)
	}
	if err := verifyChecksum(path, hex.EncodeToString(make([]byte, 32))); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestVerifyChecksumRejectsMissingFile(t *testing.T) {
	if err := verifyChecksum(filepath.Join(t.TempDir(), "nope"), "00"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
