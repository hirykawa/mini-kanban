package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"mini-kanban/config"
	"mini-kanban/db/dbgen"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

var projectLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all projects",
	RunE:  runProjectLs,
}

var projectAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectAdd,
}

var projectRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectRm,
}

var projectRmForce bool

func init() {
	projectCmd.AddCommand(projectLsCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectRmCmd)

	projectRmCmd.Flags().BoolVar(&projectRmForce, "force", false, "delete even if project has tasks")
}

func runProjectLs(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	q := dbgen.New(database)
	ctx := context.Background()

	projects, err := q.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	currentProject := config.ResolveProject(flagProject)

	for _, p := range projects {
		count, _ := q.CountTasksByProject(ctx, p.ID)
		marker := " "
		if p.Name == currentProject {
			marker = "*"
		}
		fmt.Printf("%s %s (%d tasks)\n", marker, p.Name, count)
	}
	return nil
}

func runProjectAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	q := dbgen.New(database)
	ctx := context.Background()

	// Check if exists
	if _, err := q.GetProjectByName(ctx, name); err == nil {
		return fmt.Errorf("project %q already exists", name)
	}

	proj, err := q.CreateProject(ctx, name)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	fmt.Printf("Created project: %s\n", proj.Name)
	return nil
}

func runProjectRm(cmd *cobra.Command, args []string) error {
	name := args[0]

	if name == "default" {
		return fmt.Errorf("cannot delete the default project")
	}

	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	q := dbgen.New(database)
	ctx := context.Background()

	proj, err := q.GetProjectByName(ctx, name)
	if err != nil {
		return fmt.Errorf("project %q not found", name)
	}

	count, err := q.CountTasksByProject(ctx, proj.ID)
	if err != nil {
		return fmt.Errorf("count tasks: %w", err)
	}

	if count > 0 && !projectRmForce {
		return fmt.Errorf("project %q has %d tasks; use --force to delete", name, count)
	}

	if err := q.DeleteProject(ctx, proj.ID); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	fmt.Printf("Deleted project: %s\n", name)
	return nil
}
