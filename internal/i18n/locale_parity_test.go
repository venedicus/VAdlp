package i18n

import (
	"encoding/json"
	"sort"
	"testing"
)

type localeEntry struct {
	ID string `json:"id"`
}

// TestLocaleKeyParity ensures every locale file defines exactly the same set
// of message ids as en.json (the canonical baseline), so a UI string never
// silently falls back to a raw id key when the user picks another language.
func TestLocaleKeyParity(t *testing.T) {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		t.Fatal(err)
	}
	baseline := loadLocaleIDs(t, "locales/en.json")
	if len(baseline) == 0 {
		t.Fatal("empty en.json")
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "en.json" {
			continue
		}
		ids := loadLocaleIDs(t, "locales/"+e.Name())
		if len(ids) == 0 {
			t.Fatalf("%s: empty locale file", e.Name())
		}
		missing := diff(baseline, ids)
		extra := diff(ids, baseline)
		if len(missing) > 0 || len(extra) > 0 {
			sort.Strings(missing)
			sort.Strings(extra)
			t.Errorf("%s: missing %v, extra %v", e.Name(), missing, extra)
		}
	}
}

func loadLocaleIDs(t *testing.T, path string) map[string]struct{} {
	t.Helper()
	b, err := localeFS.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var entries []localeEntry
	if err := json.Unmarshal(b, &entries); err != nil {
		t.Fatal(err)
	}
	out := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.ID == "" {
			t.Fatalf("empty id in %s", path)
		}
		out[e.ID] = struct{}{}
	}
	return out
}

func diff(a, b map[string]struct{}) []string {
	var missing []string
	for k := range a {
		if _, ok := b[k]; !ok {
			missing = append(missing, k)
		}
	}
	return missing
}
