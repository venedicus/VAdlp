package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCandidatesCustomFile(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "ffmpeg-test-not-in-path.exe")
	if err := writeEmpty(bin); err != nil {
		t.Fatal(err)
	}
	cands := buildCandidates("ffmpeg-test-not-in-path.exe", bin)
	if len(cands) == 0 {
		t.Fatal("no candidates")
	}
	if filepath.Clean(cands[0].path) != filepath.Clean(bin) {
		t.Fatalf("expected custom file first, got %+v", cands[0])
	}
	if cands[0].source != SourceCustom {
		t.Fatalf("source %q", cands[0].source)
	}
}

func TestBuildCandidatesCustomDir(t *testing.T) {
	dir := t.TempDir()
	name := "ffmpeg-test-not-in-path.exe"
	bin := filepath.Join(dir, name)
	if err := writeEmpty(bin); err != nil {
		t.Fatal(err)
	}
	cands := buildCandidates(name, dir)
	found := false
	for _, c := range cands {
		if filepath.Clean(c.path) == filepath.Clean(bin) && c.source == SourceCustom {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("custom dir candidate missing: %+v", cands)
	}
}

func TestResolveDependencyMissing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-tool-vadlp-test.bin")
	info := ResolveDependency(DepFFmpeg, DependencyPaths{FFmpeg: missing})
	if info.Status != DepMissing {
		t.Fatalf("status %q path %q", info.Status, info.Path)
	}
}

func writeEmpty(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}
