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
	onChange  []func()
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
	if bundle == nil {
		_ = Init("en")
	}
	tag := tagFor(lang)
	localizer = i18n.NewLocalizer(bundle, tag.String(), language.English.String())
	for _, fn := range onChange {
		fn()
	}
}

// OnLanguageChange registers a callback invoked every time the active
// language changes. Multiple callbacks may be registered; each is called
// in registration order.
func OnLanguageChange(fn func()) {
	onChange = append(onChange, fn)
}

// LocaleJSON returns raw locale entries for the given language code.
func LocaleJSON(lang string) ([]byte, error) {
	b, err := localeFS.ReadFile("locales/" + localeFile(lang))
	if err != nil {
		return localeFS.ReadFile("locales/en.json")
	}
	return b, nil
}

// tagFor maps a settings language code to its golang.org/x/text/language tag.
func tagFor(lang string) language.Tag {
	switch lang {
	case "ru", "ru-RU":
		return language.Russian
	case "es", "es-ES":
		return language.Spanish
	case "pt", "pt-BR", "pt-PT":
		return language.Portuguese
	case "ja", "ja-JP":
		return language.Japanese
	case "de", "de-DE":
		return language.German
	case "fr", "fr-FR":
		return language.French
	case "pl", "pl-PL":
		return language.Polish
	case "ko", "ko-KR":
		return language.Korean
	case "zh-Hant", "zh-TW", "zh-HK":
		return language.TraditionalChinese
	case "zh-Hans", "zh-CN", "zh":
		return language.SimplifiedChinese
	default:
		return language.English
	}
}

// localeFile maps a settings language code to its embedded JSON file name.
func localeFile(lang string) string {
	switch lang {
	case "ru", "ru-RU":
		return "ru.json"
	case "es", "es-ES":
		return "es.json"
	case "pt", "pt-BR", "pt-PT":
		return "pt.json"
	case "ja", "ja-JP":
		return "ja.json"
	case "de", "de-DE":
		return "de.json"
	case "fr", "fr-FR":
		return "fr.json"
	case "pl", "pl-PL":
		return "pl.json"
	case "ko", "ko-KR":
		return "ko.json"
	case "zh-Hant", "zh-TW", "zh-HK":
		return "zh-Hant.json"
	case "zh-Hans", "zh-CN", "zh":
		return "zh-Hans.json"
	default:
		return "en.json"
	}
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
