package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// Gemini implements the Provider interface for Google Gemini API.
type Gemini struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGemini creates a new Gemini provider.
func NewGemini(apiKey, model string) (*Gemini, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &Gemini{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (g *Gemini) Name() string {
	return "gemini"
}

func (g *Gemini) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	fullPrompt := prompt
	if opts.SystemPrompt != "" {
		fullPrompt = opts.SystemPrompt + "\n\n" + prompt
	}

	genConfig := map[string]interface{}{
		"temperature": opts.Temperature,
	}
	if opts.MaxTokens > 0 {
		genConfig["maxOutputTokens"] = opts.MaxTokens
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": fullPrompt},
				},
			},
		},
		"generationConfig": genConfig,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	slog.Debug("gemini raw response", "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: %s", string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason  string `json:"finishReason"`
			SafetyRatings []struct {
				Category    string `json:"category"`
				Probability string `json:"probability"`
			} `json:"safetyRatings"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason   string `json:"blockReason"`
			SafetyRatings []struct {
				Category    string `json:"category"`
				Probability string `json:"probability"`
			} `json:"safetyRatings"`
		} `json:"promptFeedback"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if result.PromptFeedback.BlockReason != "" {
		slog.Warn("gemini prompt blocked", "reason", result.PromptFeedback.BlockReason, "safetyRatings", result.PromptFeedback.SafetyRatings)
		return "", fmt.Errorf("prompt blocked by Gemini: %s", result.PromptFeedback.BlockReason)
	}

	if len(result.Candidates) == 0 {
		slog.Warn("gemini no candidates returned")
		return "", fmt.Errorf("no response from Gemini: no candidates")
	}

	candidate := result.Candidates[0]
	if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
		slog.Warn("gemini finish reason", "reason", candidate.FinishReason, "safetyRatings", candidate.SafetyRatings)
	}

	if len(candidate.Content.Parts) == 0 {
		slog.Warn("gemini no content parts", "finishReason", candidate.FinishReason)
		return "", fmt.Errorf("no response from Gemini: empty content (finishReason=%s)", candidate.FinishReason)
	}

	return candidate.Content.Parts[0].Text, nil
}
