package fyneui

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"vadlp/internal/core"
	"vadlp/internal/downloader"
)

func journalFromErr(tr func(string) string, addJournal func(string, error), fallbackKey string, err error) {
	if err == nil {
		return
	}
	if verr, ok := core.AsValidation(err); ok {
		addJournal(tr(verr.Key), nil)
		return
	}
	if downloader.IsCancelled(err) || errors.Is(err, context.Canceled) {
		return
	}
	addJournal(tr(fallbackKey), err)
}

func atoiOrZero(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func itoaOrEmpty(v int) string {
	if v == 0 {
		return ""
	}
	return strconv.Itoa(v)
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
