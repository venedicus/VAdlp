package configdir

import (
	"strings"
	"testing"
)

func TestDirEndsWithVadlp(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, "vadlp") {
		t.Fatalf("expected vadlp suffix, got %q", dir)
	}
}

func TestToolsDirUnderConfig(t *testing.T) {
	dir, err := ToolsDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, "tools") {
		t.Fatalf("expected tools suffix, got %q", dir)
	}
	cfg, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dir, cfg) {
		t.Fatalf("tools dir %q should be under config %q", dir, cfg)
	}
}
