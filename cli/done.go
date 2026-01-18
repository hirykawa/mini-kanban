package cli

import (
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
	defer notifyWeb()
	id, err := parseID(args[0])
	if err != nil {
		return err
	}

	cc, err := NewCmdContext()
	if err != nil {
		return err
	}
	defer cc.Close()

	task, err := cc.Queries.MarkTaskDone(cc.Context(), dbgen.MarkTaskDoneParams{
		ID:        id,
		ProjectID: cc.ProjectID,
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
	defer notifyWeb()
	id, err := parseID(args[0])
	if err != nil {
		return err
	}

	cc, err := NewCmdContext()
	if err != nil {
		return err
	}
	defer cc.Close()

	task, err := cc.Queries.MarkTaskOpen(cc.Context(), dbgen.MarkTaskOpenParams{
		ID:        id,
		ProjectID: cc.ProjectID,
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
	defer notifyWeb()
	id, err := parseID(args[0])
	if err != nil {
		return err
	}

	cc, err := NewCmdContext()
	if err != nil {
		return err
	}
	defer cc.Close()

	ctx := cc.Context()

	// Get task first to show title
	task, err := cc.Queries.GetTask(ctx, dbgen.GetTaskParams{
		ID:        id,
		ProjectID: cc.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("task #%d not found", id)
	}

	if err := cc.Queries.DeleteTask(ctx, dbgen.DeleteTaskParams{
		ID:        id,
		ProjectID: cc.ProjectID,
	}); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	fmt.Printf("Deleted task #%d: %s\n", task.ID, task.Title)
	return nil
}
