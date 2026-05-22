package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestLayoutSizeBucketStableWithinQuantum(t *testing.T) {
	a := layoutSizeBucket(800, 600)
	b := layoutSizeBucket(820, 620)
	if a != b {
		t.Fatalf("expected same bucket, got %x vs %x", a, b)
	}
}

func TestAdaptLayoutIgnoresTinyCanvas(t *testing.T) {
	shell := &layoutShell{mode: "h"}
	adaptLayout(fyne.NewSize(10, 10), shell, 0.4)
	if shell.lastBucket != 0 {
		t.Fatalf("bucket %x", shell.lastBucket)
	}
}
