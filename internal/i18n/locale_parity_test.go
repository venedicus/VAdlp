package i18n

import (
	"encoding/json"
	"testing"
)

type localeEntry struct {
	ID string `json:"id"`
}

func TestLocaleKeyParity(t *testing.T) {
	en := loadLocaleIDs(t, "locales/en.json")
	ru := loadLocaleIDs(t, "locales/ru.json")
	if len(en) == 0 || len(ru) == 0 {
		t.Fatal("empty locale files")
	}
	missingInRu := diff(en, ru)
	missingInEn := diff(ru, en)
	if len(missingInRu) > 0 || len(missingInEn) > 0 {
		t.Fatalf("en-only: %v\nru-only: %v", missingInRu, missingInEn)
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
