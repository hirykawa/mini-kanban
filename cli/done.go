package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"mini-kanban/db/dbgen"
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
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
	task, err := q.MarkTaskDone(ctx, dbgen.MarkTaskDoneParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("mark task done: %w", err)
	}

	fmt.Printf("✓ Task #%d marked as done: %s\n", task.ID, task.Title)
	return nil
}

var undoCmd = &cobra.Command{
	Use:   "undo <id>",
	Short: "Mark a task as open (undo done)",
	Args:  cobra.ExactArgs(1),
	RunE:  runUndo,
}

func runUndo(cmd *cobra.Command, args []string) error {
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
	task, err := q.MarkTaskOpen(ctx, dbgen.MarkTaskOpenParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("mark task open: %w", err)
	}

	fmt.Printf("○ Task #%d reopened: %s\n", task.ID, task.Title)
	return nil
}

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runRm,
}

func runRm(cmd *cobra.Command, args []string) error {
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

	// Get task first to show title
	task, err := q.GetTask(ctx, dbgen.GetTaskParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("task #%d not found", id)
	}

	if err := q.DeleteTask(ctx, dbgen.DeleteTaskParams{
		ID:        id,
		ProjectID: projectID,
	}); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	fmt.Printf("Deleted task #%d: %s\n", task.ID, task.Title)
	return nil
}
