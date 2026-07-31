package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxChecksumManifestSize = 4 << 20 // 4 MiB

func ytDlpChecksumURL() string {
	return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/SHA2-256SUMS"
}

func ffmpegChecksumURL() string {
	return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/SHA2-256SUMS"
}

func denoChecksumURL() string {
	return "https://github.com/denoland/deno/releases/latest/download/SHASUMS256.txt"
}

// fetchChecksum downloads a release checksum manifest and returns the SHA-256
// hex digest for the given artifact file name. A missing entry is an error.
func fetchChecksum(ctx context.Context, manifestURL, artifactName string) (string, error) {
	resp, err := httpGet(ctx, manifestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums %s: HTTP %d", manifestURL, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxChecksumManifestSize))
	if err != nil {
		return "", err
	}
	return parseChecksumLine(data, artifactName)
}

// parseChecksumLine finds the entry for artifactName in a sha256sum-style
// manifest ("<hash>  <name>" or "<hash> *<name>") and returns its hex digest.
func parseChecksumLine(data []byte, artifactName string) (string, error) {
	want := filepath.Base(artifactName)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		if filepath.Base(name) != want {
			continue
		}
		hash := strings.ToLower(fields[0])
		if len(hash) != sha256.Size*2 {
			return "", fmt.Errorf("malformed hash for %s", want)
		}
		return hash, nil
	}
	return "", fmt.Errorf("no entry for %s", want)
}

// verifyChecksum computes the SHA-256 of the file at path and compares it to
// the expected hex digest.
func verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("checksum mismatch: got %s, want %s", got, expected)
	}
	return nil
}

// verifyDownload downloads the checksum manifest for the artifact at
// artifactURL and verifies the already-downloaded file at destPath against it.
func verifyDownload(ctx context.Context, manifestURL, artifactURL, destPath string) error {
	expected, err := fetchChecksum(ctx, manifestURL, artifactURL)
	if err != nil {
		return err
	}
	return verifyChecksum(destPath, expected)
}
