// Package i18n provides internationalization support for mini-kanban CLI.
package i18n

import (
	"net/http"
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

// T returns the translated string for the given key using the global current language (for CLI).
func T(key string) string {
	return Translate(GetLanguage(), key)
}

// Translate returns the translated string for the given language and key.
func Translate(lang, key string) string {
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

// DetectLanguageFromRequest detects the preferred language from the HTTP request.
func DetectLanguageFromRequest(r *http.Request) string {
	// 1. Check query param ?lang=xx
	if lang := r.URL.Query().Get("lang"); lang != "" {
		return normalizeLang(lang)
	}

	// 2. Check cookie if we had one (skipping for now as per plan, keeping it simple)

	// 3. Check Accept-Language header
	accept := r.Header.Get("Accept-Language")
	if accept != "" {
		// Simple parser: take the first one or split by comma
		// "en-US,en;q=0.9,ja;q=0.8"
		parts := strings.Split(accept, ",")
		for _, part := range parts {
			// clean up q values "en;q=0.9" -> "en"
			if idx := strings.Index(part, ";"); idx != -1 {
				part = part[:idx]
			}
			part = strings.TrimSpace(part)
			part = strings.TrimSpace(part)
			// check if it's a supported language (normalizeLang returns DefaultLang if not supported, 
			// but we want to know if it *matched* a supported one)
			
			// Re-use logic: if normalizeLang returns a supported lang that isn't just a fallback
			// simpler: just check against known supported languages directly or trust normalizeLang's logic?
			// normalizeLang returns DefaultLang if it doesn't match.
			// So if we have "fr", normalizeLang("fr") -> "en". We shouldn't necessarily stop there if "ja" is next.
			// But normalizeLang implementation:
			// case LangJapanese: return LangJapanese
			// case LangEnglish: return LangEnglish
			// default: return DefaultLang
			
			// So we need to check if the input maps to a supported language distinctively.
			// Let's copy logic from normalizeLang but return empty if not supported
			langCode := strings.ToLower(part)
			// Remove region
			if idx := strings.Index(langCode, "-"); idx != -1 {
				langCode = langCode[:idx]
			}
			
			if langCode == LangJapanese || langCode == "jp" {
				return LangJapanese
			}
			if langCode == LangEnglish {
				return LangEnglish
			}
		}
	}

	return DefaultLang 
}
