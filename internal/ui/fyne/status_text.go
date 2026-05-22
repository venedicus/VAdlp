package fyneui

import (
	"strings"

	"vadlp/internal/i18n"
)

// LocalizedStatus returns an upper-case UI label for a queue/history status code.
func LocalizedStatus(status string, compact bool) string {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "complete" {
		s = "completed"
	}
	if s == "" {
		s = "ready"
	}
	key := "status." + s
	text := strings.ToUpper(i18n.T(key, nil))
	if text == strings.ToUpper(key) {
		return strings.ToUpper(strings.TrimSpace(status))
	}
	if compact {
		r := []rune(text)
		if len(r) > 4 {
			return string(r[:4])
		}
	}
	return text
}
