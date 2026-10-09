// Package i18n arayüz dilini seçer.
package i18n

import (
	"os"
	"strings"
)

type Lang string

const (
	TR Lang = "tr"
	EN Lang = "en"
)

func Parse(s string) (Lang, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "tr", "turkish", "türkçe", "turkce":
		return TR, true
	case "en", "english", "ingilizce":
		return EN, true
	}
	return "", false
}

// Sistem dili Türkçe ise TR, değilse EN.
func Detect() Lang {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return fromLocale(v)
		}
	}
	return fromLocale(systemLocale())
}

// "tr_TR.UTF-8", "tr-TR" -> TR
func fromLocale(s string) Lang {
	if strings.HasPrefix(strings.ToLower(s), "tr") {
		return TR
	}
	return EN
}
