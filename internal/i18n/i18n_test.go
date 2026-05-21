package i18n

import "testing"

func TestInitAndTranslate(t *testing.T) {
	if err := Init("en"); err != nil {
		t.Fatal(err)
	}
	if s := T("app.title", nil); s == "" || s == "app.title" {
		t.Fatalf("en title: %q", s)
	}
	SetLanguage("ru")
	if s := T("app.title", nil); s == "" || s == "app.title" {
		t.Fatalf("ru title: %q", s)
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
