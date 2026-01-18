package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// AssistResult contains the AI-enhanced task data.
type AssistResult struct {
	Title    string   // Refined title
	Body     string   // Generated body with template
	Tags     []string // Suggested tags
	Skipped  bool     // True if AI assist was skipped
	Canceled bool     // True if user canceled
}

// Assist runs the AI assist flow for task creation.
func Assist(ctx context.Context, provider Provider, title string, existingTags []string, projectContext string, maxQuestions int) (*AssistResult, error) {
	if maxQuestions <= 0 {
		maxQuestions = 3
	}

	// Build system prompt
	systemPrompt := buildSystemPrompt(projectContext)

	// Initial analysis
	analysisPrompt := buildAnalysisPrompt(title, existingTags)

	opts := GenerateOptions{
		MaxTokens:    0, // unlimited
		Temperature:  0.7,
		SystemPrompt: systemPrompt,
	}

	// Get initial questions
	response, err := provider.Generate(ctx, analysisPrompt, opts)
	if err != nil {
		return nil, fmt.Errorf("generate questions: %w", err)
	}

	parsed := parseAIResponse(response)

	// Interactive Q&A
	reader := bufio.NewReader(os.Stdin)
	var conversation []string
	conversation = append(conversation, fmt.Sprintf("Task: %s", title))

	questionsAsked := 0
	for _, q := range parsed.Questions {
		if questionsAsked >= maxQuestions {
			break
		}

		fmt.Printf("\n🤖 %s\n> ", q)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			continue
		}
		conversation = append(conversation, fmt.Sprintf("Q: %s\nA: %s", q, answer))
		questionsAsked++
	}

	// Generate final result
	finalPrompt := buildFinalPrompt(title, conversation, existingTags)
	finalResponse, err := provider.Generate(ctx, finalPrompt, opts)
	if err != nil {
		return nil, fmt.Errorf("generate result: %w", err)
	}

	result := parseFinalResponse(finalResponse)
	result.Title = cleanTitle(result.Title, title)

	// Show preview
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("📋 Preview:")
	fmt.Printf("   Title: %s\n", result.Title)
	if len(result.Tags) > 0 {
		fmt.Printf("   Tags:  %v\n", result.Tags)
	}
	if result.Body != "" {
		fmt.Println("\n   Body:")
		for _, line := range strings.Split(result.Body, "\n") {
			fmt.Printf("   %s\n", line)
		}
	}
	fmt.Println(strings.Repeat("─", 50))

	// Confirm
	fmt.Print("\nSave this task? [y/n]: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		result.Canceled = true
	}

	return result, nil
}

func buildSystemPrompt(projectContext string) string {
	prompt := `You are a task clarification assistant for a task management app called mini-kanban.
Your role is to help users refine vague or ambiguous task descriptions into clear, actionable items.

Guidelines:
- Ask clarifying questions to understand scope, acceptance criteria, and edge cases
- Keep questions concise and focused
- Suggest improvements to make tasks more actionable
- Recommend relevant tags from the existing tag list when applicable
- Output structured JSON when generating the final task`

	if projectContext != "" {
		prompt += fmt.Sprintf("\n\nProject context:\n%s", projectContext)
	}
	return prompt
}

func buildAnalysisPrompt(title string, existingTags []string) string {
	tagsStr := "none"
	if len(existingTags) > 0 {
		tagsStr = strings.Join(existingTags, ", ")
	}

	return fmt.Sprintf(`Analyze this task and identify what clarification is needed.

Task: "%s"

Existing tags in project: %s

Respond with JSON:
{
  "questions": ["question 1", "question 2", "question 3"],
  "initial_thoughts": "brief analysis"
}

Focus on:
- Scope clarity (what exactly needs to be done?)
- Acceptance criteria (how do we know it's done?)
- Edge cases or exceptions
- Dependencies or prerequisites`, title, tagsStr)
}

func buildFinalPrompt(title string, conversation []string, existingTags []string) string {
	tagsStr := "none"
	if len(existingTags) > 0 {
		tagsStr = strings.Join(existingTags, ", ")
	}

	return fmt.Sprintf(`Based on the following conversation, generate a refined task.

%s

Existing tags: %s

Respond with JSON:
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
- IMPORTANT: Escape all newlines in JSON strings with \n`, strings.Join(conversation, "\n\n"), tagsStr)
}

type analysisResponse struct {
	Questions       []string `json:"questions"`
	InitialThoughts string   `json:"initial_thoughts"`
}

func parseAIResponse(response string) analysisResponse {
	// Try to extract JSON from response
	response = extractJSON(response)

	var result analysisResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback: treat each line as a question
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasSuffix(line, "?") {
				result.Questions = append(result.Questions, line)
			}
		}
	}
	return result
}

type finalResponse struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
	Body  string   `json:"body"`
}

func parseFinalResponse(response string) *AssistResult {
	response = extractJSON(response)

	var parsed finalResponse
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Try to sanitize JSON
		sanitized := sanitizeJSON(response)
		if err2 := json.Unmarshal([]byte(sanitized), &parsed); err2 != nil {
			return &AssistResult{
				Title: "",
				Body:  response,
				Tags:  nil,
			}
		}
	}

	return &AssistResult{

		Title: parsed.Title,
		Body:  parsed.Body,
		Tags:  parsed.Tags,
	}
}

func extractJSON(s string) string {
	// Find JSON block in response
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

func cleanTitle(aiTitle, originalTitle string) string {
	if aiTitle == "" {
		return originalTitle
	}
	// Remove quotes if present
	aiTitle = strings.Trim(aiTitle, "\"'")
	return aiTitle
}

// IsTTY checks if stdin is a terminal.
func IsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
