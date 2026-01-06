package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mini-kanban/db/dbgen"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Long:  `Show detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

var showJSON bool

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "output as JSON")
}

func runShow(cmd *cobra.Command, args []string) error {
	id, err := parseID(args[0])
	if err != nil {
		return err
	}

	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	q := dbgen.New(database)
	projectID, _, err := getProjectID(q)
	if err != nil {
		return err
	}

	ctx := context.Background()
	task, err := q.GetTask(ctx, dbgen.GetTaskParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("task #%d not found", id)
	}

	tags, _ := q.ListTagsForTask(ctx, task.ID)
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	if showJSON {
		out := taskOutput{
			ID:        task.ID,
			Project:   task.ProjectName,
			Status:    task.Status,
			Title:     task.Title,
			Body:      task.Body,
			Tags:      tagNames,
			CreatedAt: formatUnix(task.CreatedAt),
			UpdatedAt: formatUnix(task.UpdatedAt),
		}
		if task.DueAt != nil {
			s := formatUnix(*task.DueAt)
			out.DueAt = &s
		}
		if task.DoneAt != nil {
			s := formatUnix(*task.DoneAt)
			out.DoneAt = &s
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Human-readable output
	fmt.Printf("Task #%d\n", task.ID)
	fmt.Printf("  Project:  %s\n", task.ProjectName)
	fmt.Printf("  Status:   %s\n", task.Status)
	fmt.Printf("  Title:    %s\n", task.Title)
	if task.Body != "" {
		fmt.Printf("  Body:\n%s\n", indentText(task.Body, "    "))
	}
	if len(tagNames) > 0 {
		fmt.Printf("  Tags:     %v\n", tagNames)
	}
	if task.DueAt != nil {
		fmt.Printf("  Due:      %s\n", formatUnix(*task.DueAt))
	}
	fmt.Printf("  Created:  %s\n", formatUnix(task.CreatedAt))
	fmt.Printf("  Updated:  %s\n", formatUnix(task.UpdatedAt))
	if task.DoneAt != nil {
		fmt.Printf("  Done at:  %s\n", formatUnix(*task.DoneAt))
	}

	return nil
}

func indentText(s string, indent string) string {
	lines := splitLines(s)
	for i, line := range lines {
		lines[i] = indent + line
	}
	return joinLines(lines)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
