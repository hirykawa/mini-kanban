package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mini-kanban/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default config file",
	RunE:  runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	path := config.GlobalConfigPath()
	dir := filepath.Dir(path)

	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	// Write template
	content := config.GenerateDefaultConfig()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Printf("Created config file: %s\n", path)
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	path := config.GlobalConfigPath()
	fmt.Printf("Config file: %s\n\n", path)

	fmt.Println("[ai]")
	fmt.Printf("  mode = %q\n", cfg.AI.Mode)
	fmt.Printf("  max_questions = %d\n", cfg.AI.MaxQuestions)
	fmt.Println()

	fmt.Println("[llm]")
	fmt.Printf("  provider = %q\n", cfg.LLM.Provider)
	fmt.Printf("  model = %q\n", cfg.LLM.Model)
	fmt.Println()

	// Show provider configs with masked keys
	if cfg.OpenAI.APIKey != "" {
		fmt.Println("[openai]")
		fmt.Printf("  api_key = %q\n", maskKey(cfg.OpenAI.APIKey))
	}

	if cfg.Anthropic.APIKey != "" {
		fmt.Println("[anthropic]")
		fmt.Printf("  api_key = %q\n", maskKey(cfg.Anthropic.APIKey))
	}

	if cfg.Gemini.APIKey != "" {
		fmt.Println("[gemini]")
		fmt.Printf("  api_key = %q\n", maskKey(cfg.Gemini.APIKey))
	}

	if cfg.Vertex.ProjectID != "" {
		fmt.Println("[vertex]")
		fmt.Printf("  project_id = %q\n", cfg.Vertex.ProjectID)
		fmt.Printf("  location = %q\n", cfg.Vertex.Location)
	}

	return nil
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
