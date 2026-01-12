// Package config handles configuration loading for mini-kanban.
// Priority: CLI flags > env vars > .mini-kanban.toml > ~/.config/mini-kanban/config.toml
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds all configuration for mini-kanban.
type Config struct {
	AI  AIConfig  `toml:"ai"`
	LLM LLMConfig `toml:"llm"`

	OpenAI    OpenAIConfig    `toml:"openai"`
	Anthropic AnthropicConfig `toml:"anthropic"`
	Gemini    GeminiConfig    `toml:"gemini"`
	Vertex    VertexConfig    `toml:"vertex"`
}

// AIConfig controls AI assist behavior.
type AIConfig struct {
	Mode         string `toml:"mode"`          // auto, on, off
	MaxQuestions int    `toml:"max_questions"` // default 3
}

// LLMConfig specifies which provider/model to use.
type LLMConfig struct {
	Provider string `toml:"provider"` // openai, anthropic, gemini, vertex
	Model    string `toml:"model"`
}

// OpenAIConfig holds OpenAI API settings.
type OpenAIConfig struct {
	APIKey string `toml:"api_key"`
}

// AnthropicConfig holds Anthropic API settings.
type AnthropicConfig struct {
	APIKey string `toml:"api_key"`
}

// GeminiConfig holds Gemini API settings.
type GeminiConfig struct {
	APIKey string `toml:"api_key"`
}

// VertexConfig holds Vertex AI settings.
type VertexConfig struct {
	ProjectID string `toml:"project_id"`
	Location  string `toml:"location"`
}

// ProjectConfig is per-project config from .mini-kanban.toml
type ProjectConfig struct {
	Project ProjectSettings `toml:"project"`
	AI      ProjectAI       `toml:"ai"`
}

// ProjectSettings holds project-level settings.
type ProjectSettings struct {
	Name string `toml:"name"`
}

// ProjectAI holds AI context for a project.
type ProjectAI struct {
	Context     string `toml:"context"`
	ContextFile string `toml:"context_file"`
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		AI: AIConfig{
			Mode:         "auto",
			MaxQuestions: 3,
		},
	}
}

// GlobalConfigPath returns the path to the global config file.
func GlobalConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mini-kanban", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mini-kanban", "config.toml")
}

// GenerateDefaultProjectConfig returns a template .mini-kanban.toml content.
func GenerateDefaultProjectConfig(projectName string) string {
	if projectName == "" {
		projectName = "my-project"
	}
	return fmt.Sprintf(`# mini-kanban project configuration

[project]
name = "%s"

[ai]
context = ""
context_file = ""
`, projectName)
}

// DBPath returns the path to the database file.
func DBPath() string {
	// Check explicit env var first
	if env := os.Getenv("MINI_KANBAN_DB"); env != "" {
		return env
	}

	// Check XDG_DATA_HOME
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "mini-kanban", "tasks.db")
	}

	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "mini-kanban", "tasks.db")
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "mini-kanban", "tasks.db")
		}
		return filepath.Join(home, "AppData", "Roaming", "mini-kanban", "tasks.db")
	default:
		return filepath.Join(home, ".local", "share", "mini-kanban", "tasks.db")
	}
}

// Load loads configuration from the global config file.
func Load() (Config, error) {
	cfg := DefaultConfig()
	path := GlobalConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Use defaults if no config file
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// LoadProjectConfig loads project config from .mini-kanban.toml in the given directory or parents.
func LoadProjectConfig(startDir string) (*ProjectConfig, string, error) {
	dir := startDir
	for {
		path := filepath.Join(dir, ".mini-kanban.toml")
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, "", err
			}
			var pc ProjectConfig
			if err := toml.Unmarshal(data, &pc); err != nil {
				return nil, "", err
			}
			return &pc, dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	return nil, "", nil // No project config found
}

// ResolveProject determines the current project name based on priority:
// 1. explicit flag
// 2. MINI_KANBAN_PROJECT env var
// 3. .mini-kanban.toml in cwd or parent
// 4. "default"
func ResolveProject(flagProject string) string {
	// 1. CLI flag
	if flagProject != "" {
		return flagProject
	}

	// 2. Environment variable
	if env := os.Getenv("MINI_KANBAN_PROJECT"); env != "" {
		return env
	}

	// 3. .mini-kanban.toml
	cwd, _ := os.Getwd()
	if pc, _, err := LoadProjectConfig(cwd); err == nil && pc != nil && pc.Project.Name != "" {
		return pc.Project.Name
	}

	// 4. Default
	return "default"
}

// GetAIContext loads AI context for the current project.
func GetAIContext(projectConfigDir string, pc *ProjectConfig) string {
	if pc == nil {
		return ""
	}

	// Try context_file first
	if pc.AI.ContextFile != "" {
		contextPath := pc.AI.ContextFile
		if !filepath.IsAbs(contextPath) {
			contextPath = filepath.Join(projectConfigDir, contextPath)
		}
		if data, err := os.ReadFile(contextPath); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	// Fall back to inline context
	return pc.AI.Context
}

// GenerateDefaultConfig returns a template config.toml content.
func GenerateDefaultConfig() string {
	return `# mini-kanban configuration

[ai]
mode = "auto"        # auto | on | off
max_questions = 3

[llm]
provider = "openai"  # openai | anthropic | gemini | vertex
model = "gpt-4o-mini"

[openai]
api_key = ""

[anthropic]
api_key = ""

[gemini]
api_key = ""

[vertex]
project_id = ""
location = "us-central1"
`
}
