// Package ai provides LLM integration for task creation assistance.
package ai

import (
	"context"
	"fmt"
	"strings"
)

// Provider is the interface for LLM providers.
type Provider interface {
	// Generate sends a prompt and returns the response.
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
	// Name returns the provider name.
	Name() string
}

// GenerateOptions contains options for generation.
type GenerateOptions struct {
	MaxTokens   int
	Temperature float64
	SystemPrompt string
}

// DefaultOptions returns sensible defaults for generation.
func DefaultOptions() GenerateOptions {
	return GenerateOptions{
		MaxTokens:   1024,
		Temperature: 0.7,
	}
}

// Config holds provider configuration.
type Config struct {
	Provider  string
	Model     string
	APIKey    string
	ProjectID string // For Vertex AI
	Location  string // For Vertex AI
}

// NewProvider creates a provider based on config.
func NewProvider(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		return NewOpenAI(cfg.APIKey, cfg.Model)
	case "anthropic":
		return NewAnthropic(cfg.APIKey, cfg.Model)
	case "gemini":
		return NewGemini(cfg.APIKey, cfg.Model)
	case "vertex":
		return NewVertex(cfg.ProjectID, cfg.Location, cfg.Model)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}
