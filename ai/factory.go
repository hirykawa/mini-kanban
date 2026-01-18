package ai

import (
	"fmt"
	"strings"

	"mini-kanban/config"
)

// NewProviderFromConfig creates an AI provider from the global config file.
// It loads the config, determines the appropriate API key based on provider,
// and returns the configured provider along with the full config.
func NewProviderFromConfig() (Provider, config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cfg, fmt.Errorf("load config: %w", err)
	}

	aiCfg := Config{
		Provider:  cfg.LLM.Provider,
		Model:     cfg.LLM.Model,
		APIKey:    "",
		ProjectID: cfg.Vertex.ProjectID,
		Location:  cfg.Vertex.Location,
	}

	// Set API key based on provider
	switch strings.ToLower(cfg.LLM.Provider) {
	case "openai":
		aiCfg.APIKey = cfg.OpenAI.APIKey
	case "anthropic":
		aiCfg.APIKey = cfg.Anthropic.APIKey
	case "gemini":
		aiCfg.APIKey = cfg.Gemini.APIKey
	}

	provider, err := NewProvider(aiCfg)
	if err != nil {
		return nil, cfg, fmt.Errorf("create ai provider: %w", err)
	}

	return provider, cfg, nil
}
