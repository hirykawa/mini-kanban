// Package cli implements the command-line interface for mini-kanban.
package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mini-kanban/ai"
	"mini-kanban/config"
	"mini-kanban/db"
	"mini-kanban/db/dbgen"
	"mini-kanban/i18n"
)

var (
	flagProject string
	flagLang    string
	rootCmd     = &cobra.Command{
		Use: "mini-kanban",
	}
)

func init() {
	// Initialize i18n from environment FIRST (before help is displayed)
	// This ensures subcommand help is translated when --help is used
	i18n.Init()
	updateCommandHelp()

	// Register hook to re-initialize with --lang flag if provided
	cobra.OnInitialize(initI18n)

	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", i18n.T("root.flag.project"))
	rootCmd.PersistentFlags().StringVar(&flagLang, "lang", "", i18n.T("root.flag.lang"))

	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(undoCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(initCmd)
}

// initI18n initializes the i18n package based on --lang flag or environment.
func initI18n() {
	if flagLang != "" {
		i18n.SetLanguage(flagLang)
	} else {
		i18n.Init()
	}

	// Update all command help texts after i18n is initialized
	updateCommandHelp()
}

// updateCommandHelp updates all command help texts with i18n translations.
func updateCommandHelp() {
	// Root command
	rootCmd.Short = i18n.T("root.short")
	rootCmd.Short = i18n.T("root.short")
	rootCmd.Long = i18n.T("root.long")

	// init command
	initCmd.Use = i18n.T("init.use")
	initCmd.Short = i18n.T("init.short")
	initCmd.Long = i18n.T("init.long")

	// add command
	addCmd.Use = i18n.T("add.use")
	addCmd.Short = i18n.T("add.short")
	addCmd.Long = i18n.T("add.long")

	// ls command
	lsCmd.Use = i18n.T("ls.use")
	lsCmd.Short = i18n.T("ls.short")
	lsCmd.Long = i18n.T("ls.long")

	// show command
	showCmd.Use = i18n.T("show.use")
	showCmd.Short = i18n.T("show.short")
	showCmd.Long = i18n.T("show.long")

	// edit command
	editCmd.Use = i18n.T("edit.use")
	editCmd.Short = i18n.T("edit.short")
	editCmd.Long = i18n.T("edit.long")

	// done command
	doneCmd.Use = i18n.T("done.use")
	doneCmd.Short = i18n.T("done.short")
	doneCmd.Long = i18n.T("done.long")

	// undo command
	undoCmd.Use = i18n.T("undo.use")
	undoCmd.Short = i18n.T("undo.short")
	undoCmd.Long = i18n.T("undo.long")

	// rm command
	rmCmd.Use = i18n.T("rm.use")
	rmCmd.Short = i18n.T("rm.short")
	rmCmd.Long = i18n.T("rm.long")

	// project command
	projectCmd.Use = i18n.T("project.use")
	projectCmd.Short = i18n.T("project.short")
	projectCmd.Long = i18n.T("project.long")
	projectLsCmd.Use = i18n.T("project.list.use")
	projectLsCmd.Short = i18n.T("project.list.short")

	// config command
	configCmd.Use = i18n.T("config.use")
	configCmd.Short = i18n.T("config.short")
	configCmd.Long = i18n.T("config.long")
	configInitCmd.Use = i18n.T("config.init.use")
	configInitCmd.Short = i18n.T("config.init.short")
	configShowCmd.Use = i18n.T("config.show.use")
	configShowCmd.Short = i18n.T("config.show.short")

	// db command
	dbCmd.Use = i18n.T("db.use")
	dbCmd.Short = i18n.T("db.short")
	dbCmd.Long = i18n.T("db.long")
	dbPathCmd.Use = i18n.T("db.path.use")
	dbPathCmd.Short = i18n.T("db.path.short")

	// web command
	webCmd.Use = i18n.T("web.use")
	webCmd.Short = i18n.T("web.short")
	webCmd.Long = i18n.T("web.long")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// getDB opens the database and runs migrations.
func getDB() (*sql.DB, error) {
	dbPath := config.DBPath()
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return database, nil
}

// getProjectID resolves the project name and returns its ID, creating if needed.
func getProjectID(q *dbgen.Queries) (int64, string, error) {
	projectName := config.ResolveProject(flagProject)
	ctx := context.Background()

	proj, err := q.GetProjectByName(ctx, projectName)
	if err == sql.ErrNoRows {
		// Create the project
		proj, err = q.CreateProject(ctx, projectName)
		if err != nil {
			return 0, "", fmt.Errorf("create project %q: %w", projectName, err)
		}
	} else if err != nil {
		return 0, "", fmt.Errorf("get project %q: %w", projectName, err)
	}

	return proj.ID, proj.Name, nil
}

// getAIProvider returns the configured AI provider.
func getAIProvider() (ai.Provider, config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cfg, fmt.Errorf("load config: %w", err)
	}

	aiCfg := ai.Config{
		Provider:  cfg.LLM.Provider,
		Model:     cfg.LLM.Model,
		APIKey:    "",
		ProjectID: cfg.Vertex.ProjectID,
		Location:  cfg.Vertex.Location,
	}

	// Set API key based on provider
	switch strings.ToLower(cfg.LLM.Provider) {
	case "openai":
		aiCfg.APIKey = cfg.OpenAI.APIKey
	case "anthropic":
		aiCfg.APIKey = cfg.Anthropic.APIKey
	case "gemini":
		aiCfg.APIKey = cfg.Gemini.APIKey
	}

	provider, err := ai.NewProvider(aiCfg)
	if err != nil {
		return nil, cfg, fmt.Errorf("new ai provider: %w", err)
	}

	return provider, cfg, nil
}

// getProjectAIContext returns the AI context and project config for the current project.
func getProjectAIContext() (string, *config.ProjectConfig, error) {
	cwd, _ := os.Getwd()
	pc, dir, err := config.LoadProjectConfig(cwd)
	if err != nil {
		return "", nil, fmt.Errorf("load project config: %w", err)
	}
	return config.GetAIContext(dir, pc), pc, nil
}
