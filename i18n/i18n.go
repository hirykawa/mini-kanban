// Package i18n provides internationalization support for mini-kanban CLI.
package i18n

import (
	"os"
	"strings"
	"sync"
)

// Supported languages
const (
	LangEnglish  = "en"
	LangJapanese = "ja"
	DefaultLang  = LangEnglish
)

var (
	currentLang string
	once        sync.Once
)

// DetectLanguage detects the preferred language from environment variables.
// Priority: MINI_KANBAN_LANG > LC_ALL > LC_MESSAGES > LANG
func DetectLanguage() string {
	// Check app-specific env var first
	if lang := os.Getenv("MINI_KANBAN_LANG"); lang != "" {
		return normalizeLang(lang)
	}

	// Check standard locale env vars
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := os.Getenv(key); val != "" {
			return normalizeLang(val)
		}
	}

	return DefaultLang
}

// normalizeLang extracts the language code from a locale string.
// Examples: "ja_JP.UTF-8" -> "ja", "en_US" -> "en", "ja" -> "ja"
func normalizeLang(locale string) string {
	// Remove encoding suffix (e.g., ".UTF-8")
	if idx := strings.Index(locale, "."); idx != -1 {
		locale = locale[:idx]
	}

	// Get language part (before "_" or "-")
	for _, sep := range []string{"_", "-"} {
		if idx := strings.Index(locale, sep); idx != -1 {
			locale = locale[:idx]
		}
	}

	locale = strings.ToLower(locale)

	// Validate supported language
	switch locale {
	case LangJapanese, "jp":
		return LangJapanese
	case LangEnglish:
		return LangEnglish
	default:
		return DefaultLang
	}
}

// Init initializes the i18n package with auto-detection.
func Init() {
	once.Do(func() {
		currentLang = DetectLanguage()
	})
}

// SetLanguage sets the current language explicitly.
func SetLanguage(lang string) {
	currentLang = normalizeLang(lang)
}

// GetLanguage returns the current language.
func GetLanguage() string {
	if currentLang == "" {
		Init()
	}
	return currentLang
}

// T returns the translated string for the given key.
func T(key string) string {
	lang := GetLanguage()

	var msgs map[string]string
	switch lang {
	case LangJapanese:
		msgs = messagesJA
	default:
		msgs = messagesEN
	}

	if msg, ok := msgs[key]; ok {
		return msg
	}

	// Fallback to English
	if msg, ok := messagesEN[key]; ok {
		return msg
	}

	// Return key itself as last resort
	return key
}
