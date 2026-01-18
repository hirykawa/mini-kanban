package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI implements the Provider interface for OpenAI API.
type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(apiKey, model string) (*OpenAI, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAI{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (o *OpenAI) Name() string {
	return "openai"
}

func (o *OpenAI) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	messages := []map[string]string{}
	
	if opts.SystemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": opts.SystemPrompt,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	reqBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  opts.MaxTokens,
		"temperature": opts.Temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}
