package cli

import (
	"context"
	"fmt"
	"strings"
	"time"
	"mini-kanban/ai"
	"github.com/spf13/cobra"

	"mini-kanban/db/dbgen"
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new task",
	Long:  "Add a new task to the current project.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

var (
	addBody  string
	addTags  []string
	addDue   string
	addAI    bool
	addNoAI  bool
)

func init() {
	addCmd.Flags().StringVar(&addBody, "body", "", "task body/description")
	addCmd.Flags().StringArrayVar(&addTags, "tag", nil, "tag (can be specified multiple times)")
	addCmd.Flags().StringVar(&addDue, "due", "", "due date (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	addCmd.Flags().BoolVar(&addAI, "ai", false, "force AI assist on")
	addCmd.Flags().BoolVar(&addNoAI, "no-ai", false, "force AI assist off")
}

func runAdd(cmd *cobra.Command, args []string) error {
	defer notifyWeb()
	title := strings.Join(args, " ")
	if title == "" {
		return fmt.Errorf("title is required")
	}

	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	q := dbgen.New(database)
	projectID, projectName, err := getProjectID(q)
	if err != nil {
		return err
	}

	// Parse due date
	var dueAt *int64
	if addDue != "" {
		t, err := parseDateTime(addDue)
		if err != nil {
			return fmt.Errorf("invalid due date: %w", err)
		}
		unix := t.Unix()
		dueAt = &unix
	}

	// Handle AI assist
	shouldAI := addAI
	if !addNoAI && !shouldAI {
		_, cfg, err := getAIProvider()
		if err == nil && strings.ToLower(cfg.AI.Mode) == "auto" {
			if ai.ShouldAssist(title) {
				shouldAI = true
			}
		}
	}

	if shouldAI && !addNoAI {
		provider, cfg, err := getAIProvider()
		if err != nil {
			return err
		}
		projectContext, _, _ := getProjectAIContext()
		
		fmt.Printf("🤖 AI assist is starting for: %s\n", title)
		result, err := ai.Assist(context.Background(), provider, title, nil, projectContext, cfg.AI.MaxQuestions)
		if err != nil {
			return fmt.Errorf("ai assist: %w", err)
		}

		if result.Canceled {
			fmt.Println("AI assist canceled.")
			return nil
		}

		if !result.Skipped {
			title = result.Title
			addBody = result.Body
			addTags = append(addTags, result.Tags...)
		}
	}

	// Parse tags (support comma-separated within a single --tag)
	allTags := parseTags(addTags)

	// Create task
	task, err := q.CreateTask(context.Background(), dbgen.CreateTaskParams{
		ProjectID: projectID,
		Title:     title,
		Body:      addBody,
		Status:    "open",
		DueAt:     dueAt,
	})
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	// Add tags
	for _, tagName := range allTags {
		tag, err := q.GetOrCreateTag(context.Background(), tagName)
		if err != nil {
			return fmt.Errorf("create tag %q: %w", tagName, err)
		}
		if err := q.AddTagToTask(context.Background(), dbgen.AddTagToTaskParams{
			TaskID: task.ID,
			TagID:  tag.ID,
		}); err != nil {
			return fmt.Errorf("add tag to task: %w", err)
		}
	}

	fmt.Printf("Created task #%d in project %q: %s\n", task.ID, projectName, task.Title)
	return nil
}

// parseDateTime parses a date/datetime string in local timezone.
func parseDateTime(s string) (time.Time, error) {
	loc := time.Local
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse %q (use YYYY-MM-DD or YYYY-MM-DD HH:MM)", s)
}

// parseTags normalizes tag input (supports comma-separated and trims whitespace).
func parseTags(tags []string) []string {
	var result []string
	for _, t := range tags {
		// Split by comma
		parts := strings.Split(t, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// Collapse multiple spaces
			p = strings.Join(strings.Fields(p), " ")
			if p != "" {
				result = append(result, p)
			}
		}
	}
	return result
}
