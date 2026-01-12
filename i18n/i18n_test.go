package i18n

import (
	"net/http/httptest"
	"testing"
)

func TestDetectLanguageFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		header string
		query  string
		want   string
	}{
		{"No header", "", "", "en"},
		{"Standard Japanese", "ja", "", "ja"},
		{"Japanese with region", "ja-JP", "", "ja"},
		{"Japanese first", "ja,en;q=0.9", "", "ja"},
		{"English first", "en,ja;q=0.9", "", "en"},
		{"Query param override", "en", "ja", "ja"},
		{"Unsupported language", "fr-FR", "", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				r.Header.Set("Accept-Language", tt.header)
			}
			if tt.query != "" {
				q := r.URL.Query()
				q.Set("lang", tt.query)
				r.URL.RawQuery = q.Encode()
			}

			if got := DetectLanguageFromRequest(r); got != tt.want {
				t.Errorf("DetectLanguageFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
