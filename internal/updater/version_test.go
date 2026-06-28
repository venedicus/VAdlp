package updater

import (
	"context"
	"testing"
	"time"
)

func TestNormalizeVersion(t *testing.T) {
	if got := normalizeVersion("v2025.06.25"); got != "2025.06.25" {
		t.Fatalf("got %q", got)
	}
}

func TestVersionOlder(t *testing.T) {
	cases := []struct {
		cur, latest string
		want        bool
	}{
		{"2025.06.24", "2025.06.25", true},
		{"2025.06.25", "2025.06.25", false},
		{"2025.07.01", "2025.06.25", false},
		{"1.44.0", "1.45.0", true},
	}
	for _, c := range cases {
		if got := VersionOlder(c.cur, c.latest); got != c.want {
			t.Fatalf("%q vs %q: got %v want %v", c.cur, c.latest, got, c.want)
		}
	}
}

func TestApplyLatestVersions(t *testing.T) {
	versionCacheMu.Lock()
	versionCache[DepYtDlp] = cachedTag{tag: "2025.06.25", fetched: time.Now()}
	versionCacheMu.Unlock()

	infos := []DependencyInfo{{
		ID: DepYtDlp, Status: DepFound, Version: "2025.06.24",
	}}
	ApplyLatestVersions(infos)
	if !infos[0].UpdateAvail {
		t.Fatal("expected update available")
	}
	if infos[0].Status != DepOutdated {
		t.Fatalf("status %q", infos[0].Status)
	}
}

func TestNeedsAttention(t *testing.T) {
	n := NeedsAttention([]DependencyInfo{
		{ID: DepYtDlp, Status: DepMissing, Level: DepRequired},
		{ID: DepDeno, Status: DepMissing, Level: DepOptional},
	})
	if n != 1 {
		t.Fatalf("got %d", n)
	}
}

func TestRefreshLatestVersionsOffline(t *testing.T) {
	infos, err := RefreshLatestVersions(context.Background(), DependencyPaths{})
	if len(infos) != 3 {
		t.Fatalf("expected 3 infos, got %d", len(infos))
	}
	_ = err // network errors are reported but local resolution still returns
}
