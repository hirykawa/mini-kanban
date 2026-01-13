package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"mini-kanban/ai"
	"mini-kanban/db/dbgen"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a task",
	Long:  `Edit an existing task's title, body, tags, or due date.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

var (
	editTitle string
	editBody  string
	editTags  []string
	editDue   string
	editStatus string
	editAI    bool
	editNoAI  bool
)

func init() {
	editCmd.Flags().StringVar(&editTitle, "title", "", "new title")
	editCmd.Flags().StringVar(&editBody, "body", "", "new body")
	editCmd.Flags().StringArrayVar(&editTags, "tag", nil, "replace tags (can be specified multiple times)")
	editCmd.Flags().StringVar(&editDue, "due", "", "new due date (YYYY-MM-DD or YYYY-MM-DD HH:MM, use 'none' to clear)")
	editCmd.Flags().StringVarP(&editStatus, "status", "s", "", "new status (todo, doing, review, done)")
	editCmd.Flags().BoolVar(&editAI, "ai", false, "force AI assist on")
	editCmd.Flags().BoolVar(&editNoAI, "no-ai", false, "force AI assist off")
}

func runEdit(cmd *cobra.Command, args []string) error {
	defer notifyWeb()
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

	// Verify task exists
	_, err = q.GetTask(ctx, dbgen.GetTaskParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("task #%d not found", id)
	}

	// Build update params
	params := dbgen.UpdateTaskParams{
		ID:        id,
		ProjectID: projectID,
	}

	if editTitle != "" {
		params.Title = &editTitle
	}
	if editBody != "" {
		params.Body = &editBody
	}
	if editDue == "none" {
		clearDue := int64(1)
		params.ClearDue = &clearDue
	} else if editDue != "" {
		t, err := parseDateTime(editDue)
		if err != nil {
			return fmt.Errorf("invalid due date: %w", err)
		}
		unix := t.Unix()
		params.DueAt = &unix
	}

	// Handle AI assist
	if editAI && !editNoAI {
		provider, cfg, err := getAIProvider()
		if err != nil {
			return err
		}
		projectContext, _, _ := getProjectAIContext()

		// Get current task for context
		currentTask, err := q.GetTask(ctx, dbgen.GetTaskParams{
			ID:        id,
			ProjectID: projectID,
		})
		if err != nil {
			return fmt.Errorf("get task for context: %w", err)
		}

		title := editTitle
		if title == "" {
			title = currentTask.Title
		}
		
		fmt.Printf("🤖 AI assist is starting for task #%d: %s\n", id, title)
		result, err := ai.Assist(ctx, provider, title, nil, projectContext, cfg.AI.MaxQuestions)
		if err != nil {
			return fmt.Errorf("ai assist: %w", err)
		}

		if result.Canceled {
			fmt.Println("AI assist canceled.")
			return nil
		}

		if !result.Skipped {
			editTitle = result.Title
			editBody = result.Body
			editTags = append(editTags, result.Tags...)
			
			// Force update params
			params.Title = &editTitle
			params.Body = &editBody
		}
	}

	task, err := q.UpdateTask(ctx, params)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	if editStatus != "" {
		task, err = q.UpdateTaskStatus(ctx, dbgen.UpdateTaskStatusParams{
			ID:        id,
			ProjectID: projectID,
			Status:    editStatus,
		})
		if err != nil {
			return fmt.Errorf("update task status: %w", err)
		}
	}

	// Handle tags if specified
	if cmd.Flags().Changed("tag") {
		// Remove all existing tags
		if err := q.RemoveAllTagsFromTask(ctx, id); err != nil {
			return fmt.Errorf("remove tags: %w", err)
		}

		// Add new tags
		allTags := parseTags(editTags)
		for _, tagName := range allTags {
			tag, err := q.GetOrCreateTag(ctx, tagName)
			if err != nil {
				return fmt.Errorf("create tag %q: %w", tagName, err)
			}
			if err := q.AddTagToTask(ctx, dbgen.AddTagToTaskParams{
				TaskID: id,
				TagID:  tag.ID,
			}); err != nil {
				return fmt.Errorf("add tag to task: %w", err)
			}
		}
	}

	fmt.Printf("Updated task #%d: %s\n", task.ID, task.Title)
	return nil
}

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid task ID: %s", s)
	}
	return id, nil
}
