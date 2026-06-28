package i18n

import "testing"

func TestInitAndTranslate(t *testing.T) {
	if err := Init("en"); err != nil {
		t.Fatal(err)
	}
	if s := T("app.title", nil); s == "" || s == "app.title" {
		t.Fatalf("en title: %q", s)
	}
	for _, lang := range []string{"ru", "es", "pt", "ja", "de", "fr", "pl", "ko", "zh-Hant", "zh-Hans"} {
		SetLanguage(lang)
		if s := T("app.title", nil); s == "" || s == "app.title" {
			t.Fatalf("%s title: %q", lang, s)
		}
	}
}

func TestLocaleJSONAllLanguages(t *testing.T) {
	for _, lang := range []string{"en", "ru", "es", "pt", "ja", "de", "fr", "pl", "ko", "zh-Hant", "zh-Hans"} {
		b, err := LocaleJSON(lang)
		if err != nil {
			t.Fatalf("%s: %v", lang, err)
		}
		if len(b) == 0 {
			t.Fatalf("%s: empty locale JSON", lang)
		}
	}
}

func TestOnLanguageChange(t *testing.T) {
	called := false
	OnLanguageChange(func() { called = true })
	SetLanguage("en")
	if !called {
		t.Fatal("callback not invoked")
	}
}
