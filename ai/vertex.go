package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// Vertex implements the Provider interface for Vertex AI.
type Vertex struct {
	projectID string
	location  string
	model     string
	client    *http.Client
}

// NewVertex creates a new Vertex AI provider.
func NewVertex(projectID, location, model string) (*Vertex, error) {
	if projectID == "" {
		return nil, fmt.Errorf("Vertex AI project ID is required")
	}
	if location == "" {
		location = "us-central1"
	}
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &Vertex{
		projectID: projectID,
		location:  location,
		model:     model,
		client:    &http.Client{},
	}, nil
}

func (v *Vertex) Name() string {
	return "vertex"
}

func (v *Vertex) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	// Get access token from ADC
	token, err := v.getAccessToken()
	if err != nil {
		return "", fmt.Errorf("get access token: %w (run 'gcloud auth application-default login')", err)
	}

	fullPrompt := prompt
	if opts.SystemPrompt != "" {
		fullPrompt = opts.SystemPrompt + "\n\n" + prompt
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": fullPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": opts.MaxTokens,
			"temperature":     opts.Temperature,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		v.location, v.projectID, v.location, v.model,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Vertex AI error: %s", string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Vertex AI")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func (v *Vertex) getAccessToken() (string, error) {
	cmd := exec.Command("gcloud", "auth", "application-default", "print-access-token")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
