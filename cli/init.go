package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"mini-kanban/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration and database", // overwritten by i18n in updateCommandHelp
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// 1. Initialize Config
	path := config.GlobalConfigPath()
	dir := filepath.Dir(path)

	fmt.Println("Initializing mini-kanban...")

	// Create config directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Global config file already exists at: %s\n", path)
	} else {
		// Write template
		content := config.GenerateDefaultConfig()
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return fmt.Errorf("write config file: %w", err)
		}
		fmt.Printf("Created global config file at: %s\n", path)
	}

	// 2. Initialize Project Config (Local)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %w", err)
	}
	localConfigPath := filepath.Join(cwd, ".mini-kanban.toml")

	if _, err := os.Stat(localConfigPath); err == nil {
		fmt.Printf("Local config file already exists at: %s\n", localConfigPath)
	} else {
		projectName := filepath.Base(cwd)
		content := config.GenerateDefaultProjectConfig(projectName)
		if err := os.WriteFile(localConfigPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write local config file: %w", err)
		}
		fmt.Printf("Created local config file at: %s\n", localConfigPath)
	}

	// 2. Database Info
	dbPath := config.DBPath()
	fmt.Printf("Database path: %s\n", dbPath)
	
	// Ensure DB directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	fmt.Println("\nSetup complete! You can now use mini-kanban.")
	fmt.Println("Try running: mini-kanban ls")

	return nil
}
