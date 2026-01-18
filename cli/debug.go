package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"mini-kanban/ai"
)

var debugCmd = &cobra.Command{
	Use:                "debug [command]...",
	Short:              "Run a command in debug mode (auto-fix with AI)",
	Long:               "Executes the specified command. If it fails, the output and error are analyzed by AI to suggest a fix.",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runDebug,
}

func init() {
	// We don't add flags here because DisableFlagParsing is true.
	// Any flags intended for debugCmd would need to be parsed manually or handled differently.
	// But simple usage is `mini-kanban debug <cmd> <args>...`
	rootCmd.AddCommand(debugCmd)
}

func runDebug(cmd *cobra.Command, args []string) error {
	// 1. Execute the command
	commandName := args[0]
	commandArgs := args[1:]

	fmt.Printf("Running: %s %s\n", commandName, strings.Join(commandArgs, " "))

	execCmd := exec.Command(commandName, commandArgs...)

	// Capture output
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// Also pipe to real stdout/stderr so user sees it live?
	// Or maybe just capture. Cursor usually runs it and shows output.
	// If I pipe, I can't easily capture cleanly without a MultiWriter.
	// Let's use MultiWriter to show progress and capture.
	// But standard os/exec with bytes.Buffer captures it.
	// Let's just run it. If it fails, we show the output then.
	// Actually, for interactive commands, this might be tricky.
	// Assuming non-interactive for now as per "Debug Mode" usually fixing build errors or script errors.

	err := execCmd.Run()

	// Print output regardless of success/fail
	if stdout.Len() > 0 {
		fmt.Println(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprintln(os.Stderr, stderr.String())
	}

	if err == nil {
		return nil
	}

	// 2. Command failed
	fmt.Printf("\nCommand failed with error: %v\n", err)
	fmt.Println("🤖 Analyzing failure with AI...")

	// 3. Get AI Provider
	provider, _, err := getAIProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize AI: %w", err)
	}

	// 4. Get Project Context
	aiContext, _, err := getProjectAIContext()
	if err != nil {
		// Proceed without context if it fails, but warn
		fmt.Printf("Warning: could not load project context: %v\n", err)
	}

	// 5. Construct Prompt
	systemPrompt := fmt.Sprintf(`You are an expert software engineer debugging a command failure. 
The user is working in the following project context:
%s

Analyze the command, its output, and the error.
Provide a concise explanation of why it failed.
Suggest a fix. 
If it is a code error, provide the fixed code snippet.
If it is a command usage error, provide the correct command.
Be direct and helpful.`, aiContext)

	userPrompt := fmt.Sprintf(`Command:
%s %s

Stdout:
%s

Stderr:
%s
`, commandName, strings.Join(commandArgs, " "), stdout.String(), stderr.String())

	// 6. Call AI
	opts := ai.GenerateOptions{
		MaxTokens:    2048,
		Temperature:  0.2, // Low temperature for factual debugging
		SystemPrompt: systemPrompt,
	}

	response, err := provider.Generate(context.Background(), userPrompt, opts)
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	// 7. Display Result
	fmt.Println("\n--- AI Diagnosis ---")
	fmt.Println(response)

	return fmt.Errorf("command failed") // Return error to exit with non-zero code
}
