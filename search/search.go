// Package search implements query parsing and search logic.
package search

import (
	"strings"
	"unicode"
)

// ParsedQuery contains parsed search tokens.
type ParsedQuery struct {
	FreeText string   // Full-text search query
	Tags     []string // tag: filters
	Status   string   // status: filter (open/done)
	HasDue   bool     // has:due filter
	Overdue  bool     // is:overdue filter
	DueBefore string  // due:<= filter
	DueAfter  string  // due:>= filter
}

// ParseQuery tokenizes and parses a search query string.
// Supports: tag:value, status:open|done, has:due, is:overdue, due:<=YYYY-MM-DD
func ParseQuery(q string) ParsedQuery {
	var result ParsedQuery
	var freeTextParts []string

	tokens := tokenize(q)

	for _, token := range tokens {
		lower := strings.ToLower(token)

		switch {
		case strings.HasPrefix(lower, "tag:"):
			value := token[4:]
			if value != "" {
				result.Tags = append(result.Tags, value)
			}

		case strings.HasPrefix(lower, "status:"):
			value := strings.ToLower(token[7:])
			if value == "open" || value == "done" {
				result.Status = value
			}

		case lower == "has:due":
			result.HasDue = true

		case lower == "is:overdue":
			result.Overdue = true

		case strings.HasPrefix(lower, "due:<="):
			result.DueBefore = token[6:]

		case strings.HasPrefix(lower, "due:>="):
			result.DueAfter = token[6:]

		case strings.HasPrefix(lower, "due:<"):
			result.DueBefore = token[5:]

		case strings.HasPrefix(lower, "due:>"):
			result.DueAfter = token[5:]

		default:
			freeTextParts = append(freeTextParts, token)
		}
	}

	result.FreeText = strings.Join(freeTextParts, " ")
	return result
}

// tokenize splits query respecting quoted strings.
func tokenize(q string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false

	for _, r := range q {
		switch {
		case r == '"':
			inQuote = !inQuote
		case unicode.IsSpace(r) && !inQuote:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// ShouldUseLike returns true if the query contains CJK characters
// that require LIKE search instead of FTS.
func ShouldUseLike(q string) bool {
	for _, r := range q {
		if isCJK(r) {
			return true
		}
	}
	return false
}

// isCJK returns true if the rune is a CJK character (Chinese, Japanese, Korean).
func isCJK(r rune) bool {
	return unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hangul, r)
}
