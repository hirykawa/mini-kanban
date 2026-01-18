package srv

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"mini-kanban/ai"
	"mini-kanban/config"
	"mini-kanban/db/dbgen"
)

// JSON utility functions for parsing AI responses.
// These are used by both the srv and ai packages.

// AIAssistRequest is the request body for AI assist API.
type AIAssistRequest struct {
	Title        string   `json:"title"`
	Answers      []string `json:"answers"`
	ExistingTags []string `json:"existingTags"`
	Project      string   `json:"project"`
}

// AIAssistResponse is the response for AI assist API.
type AIAssistResponse struct {
	Questions []string `json:"questions,omitempty"`
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ShouldAsk bool     `json:"shouldAsk"`
	Phase     string   `json:"phase"` // "questions" or "result"
}

// getAIProvider creates an AI provider from config.
func getAIProvider() (ai.Provider, config.Config, error) {
	return ai.NewProviderFromConfig()
}

// handleAPIAIAssist handles AI assist requests for web.
func (s *Server) handleAPIAIAssist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AIAssistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	// Check if AI is enabled
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		http.Error(w, "config error", http.StatusInternalServerError)
		return
	}

	// If mode is "off", return early
	if strings.ToLower(cfg.AI.Mode) == "off" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIAssistResponse{
			ShouldAsk: false,
			Phase:     "skip",
		})
		return
	}

	provider, _, err := getAIProvider()
	if err != nil {
		slog.Error("get ai provider", "error", err)
		// If AI provider is not configured, skip assist
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIAssistResponse{
			ShouldAsk: false,
			Phase:     "skip",
		})
		return
	}

	// Get existing tags for context
	existingTags := req.ExistingTags
	if len(existingTags) == 0 {
		q := dbgen.New(s.DB)
		projectID, _ := s.getProjectID(ctx, req.Project)
		topTags, _ := q.TopTagsForProject(ctx, dbgen.TopTagsForProjectParams{
			ProjectID: projectID,
			Limit:     20,
		})
		for _, t := range topTags {
			existingTags = append(existingTags, t.Name)
		}
	}

	maxQuestions := cfg.AI.MaxQuestions
	if maxQuestions <= 0 {
		maxQuestions = 3
	}

	// Load project context from .mini-kanban.toml
	cwd, _ := os.Getwd()
	pc, dir, _ := config.LoadProjectConfig(cwd)
	projectContext := config.GetAIContext(dir, pc)

	// Phase 1: Generate questions (no answers yet)
	if len(req.Answers) == 0 {
		questions, err := generateQuestions(ctx, provider, req.Title, existingTags, projectContext)
		if err != nil {
			slog.Error("generate questions", "error", err)
			http.Error(w, "ai error", http.StatusInternalServerError)
			return
		}

		// Limit questions
		if len(questions) > maxQuestions {
			questions = questions[:maxQuestions]
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AIAssistResponse{
			Questions: questions,
			ShouldAsk: true,
			Phase:     "questions",
		})
		return
	}

	// Phase 2: Generate final result with answers
	result, err := generateFinalResult(ctx, provider, req.Title, req.Answers, existingTags, projectContext)
	if err != nil {
		slog.Error("generate result", "error", err)
		http.Error(w, "ai error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AIAssistResponse{
		Title:     result.Title,
		Body:      result.Body,
		Tags:      result.Tags,
		ShouldAsk: true,
		Phase:     "result",
	})
}

func generateQuestions(ctx context.Context, provider ai.Provider, title string, existingTags []string, projectContext string) ([]string, error) {
	systemPrompt := `You are a task clarification assistant for a task management app called mini-kanban.
Your role is to help users refine vague or ambiguous task descriptions into clear, actionable items.

Guidelines:
- Ask clarifying questions to understand scope, acceptance criteria, and edge cases
- Keep questions concise and focused
- Ask 2-3 questions maximum
- Output structured JSON`

	tagsStr := "none"
	if len(existingTags) > 0 {
		tagsStr = strings.Join(existingTags, ", ")
	}

	contextSection := ""
	if projectContext != "" {
		contextSection = "\nProject context:\n" + projectContext + "\n"
	}

	prompt := `Analyze this task and generate clarifying questions.

Task: "` + title + `"
` + contextSection + `
Existing tags in project: ` + tagsStr + `

Respond with JSON only:
{
  "questions": ["question 1", "question 2", "question 3"]
}

Focus on:
- Scope clarity (what exactly needs to be done?)
- Acceptance criteria (how do we know it's done?)
- Edge cases or exceptions`

	opts := ai.GenerateOptions{
		MaxTokens:    0, // unlimited
		Temperature:  0.7,
		SystemPrompt: systemPrompt,
	}

	response, err := provider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	type questionsResponse struct {
		Questions []string `json:"questions"`
	}

	jsonStr := extractJSON(response)
	var parsed questionsResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// Fallback: extract lines ending with ?
		var questions []string
		for _, line := range strings.Split(response, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasSuffix(line, "?") {
				questions = append(questions, line)
			}
		}
		return questions, nil
	}

	return parsed.Questions, nil
}

type aiResult struct {
	Title string
	Body  string
	Tags  []string
}

func generateFinalResult(ctx context.Context, provider ai.Provider, title string, answers []string, existingTags []string, projectContext string) (*aiResult, error) {
	systemPrompt := `You are a task clarification assistant for mini-kanban.
Generate a refined task based on the original title and user's answers to clarifying questions.
Output structured JSON only.`

	tagsStr := "none"
	if len(existingTags) > 0 {
		tagsStr = strings.Join(existingTags, ", ")
	}

	// Build conversation
	var conversation strings.Builder
	conversation.WriteString("Task: " + title + "\n\n")
	for i, answer := range answers {
		conversation.WriteString("Q" + string(rune('1'+i)) + " Answer: " + answer + "\n")
	}

	contextSection := ""
	if projectContext != "" {
		contextSection = "\nProject context:\n" + projectContext + "\n"
	}

	prompt := `Based on the following conversation, generate a refined task.

` + conversation.String() + contextSection + `
Existing tags: ` + tagsStr + `

Respond with JSON only:
{
  "title": "refined, clear title (short)",
  "tags": ["tag1", "tag2"],
  "body": "## メモ\n- point1\n- point2\n\n## 決めること\n- decision1\n\n## 受け入れ条件\n- [ ] criteria1\n- [ ] criteria2"
}

Rules:
- Title should be concise and actionable
- Use existing tags when applicable, minimize new tags
- Body uses lightweight template format
- Keep メモ to 3-7 bullet points
- Keep 受け入れ条件 to 3-6 items
- IMPORTANT: Escape all newlines in JSON strings with \n`

	opts := ai.GenerateOptions{
		MaxTokens:    0, // unlimited
		Temperature:  0.7,
		SystemPrompt: systemPrompt,
	}

	response, err := provider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	type finalResponse struct {
		Title string   `json:"title"`
		Tags  []string `json:"tags"`
		Body  string   `json:"body"`
	}

	jsonStr := extractJSON(response)
	var parsed finalResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// Try to sanitize JSON (fix unescaped newlines)
		sanitized := sanitizeJSON(jsonStr)
		if err2 := json.Unmarshal([]byte(sanitized), &parsed); err2 != nil {
			return &aiResult{
				Title: title,
				Body:  response,
			}, nil
		}
	}

	resultTitle := parsed.Title
	if resultTitle == "" {
		resultTitle = title
	}
	resultTitle = strings.Trim(resultTitle, "\"'")

	return &aiResult{
		Title: resultTitle,
		Body:  parsed.Body,
		Tags:  parsed.Tags,
	}, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func sanitizeJSON(s string) string {
	var result strings.Builder
	inString := false
	escaped := false

	for _, r := range s {
		if escaped {
			result.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			result.WriteRune(r)
			escaped = true
			continue
		}

		if r == '"' {
			inString = !inString
		}

		if r == '\n' && inString {
			result.WriteString("\\n")
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
