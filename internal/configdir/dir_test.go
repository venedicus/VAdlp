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
