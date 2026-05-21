package i18n

import (
	"embed"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	onChange  func()
)

func Init(lang string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := localeFS.ReadFile("locales/" + e.Name())
		if err != nil {
			return err
		}
		if _, err := bundle.ParseMessageFileBytes(b, e.Name()); err != nil {
			return err
		}
	}
	SetLanguage(lang)
	return nil
}

func SetLanguage(lang string) {
	tag := language.English
	switch lang {
	case "ru", "ru-RU":
		tag = language.Russian
	}
	if bundle == nil {
		_ = Init("en")
	}
	localizer = i18n.NewLocalizer(bundle, tag.String(), language.English.String())
	if onChange != nil {
		onChange()
	}
}

func OnLanguageChange(fn func()) {
	onChange = fn
}

func T(id string, data map[string]interface{}) string {
	if localizer == nil {
		_ = Init("en")
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	s, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: id, TemplateData: data})
	if err != nil {
		return id
	}
	return s
}
