package cli

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"mini-kanban/config"
	"mini-kanban/srv"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start local web server",
	RunE:  runWeb,
}

var (
	webPort  int
	webOpen  bool
	webDev   bool
	webDebug bool
)

func init() {
	webCmd.Flags().IntVar(&webPort, "port", 0, "port to listen on (0 = auto-select from 9000-9999)")
	webCmd.Flags().BoolVar(&webOpen, "open", false, "open browser after starting")
	webCmd.Flags().BoolVar(&webDev, "dev", false, "run in development mode (serve static files from web/dist)")
	webCmd.Flags().BoolVar(&webDebug, "debug", false, "enable debug logging")
}

func runWeb(cmd *cobra.Command, args []string) error {
	// Set log level
	if webDebug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
	// Find available port
	port := webPort
	if port == 0 {
		var err error
		port, err = findAvailablePort()
		if err != nil {
			return fmt.Errorf("find available port: %w", err)
		}
	}

	projectName := config.ResolveProject(flagProject)

	server, err := srv.New(config.DBPath(), projectName, webDev)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s/?project=%s", addr, projectName)

	fmt.Printf("Starting mini-kanban web server...\n")
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  Press Ctrl+C to stop\n\n")

	if webOpen {
		go openBrowser(url)
	}

	// Save port to file so other CLI commands can notify the server
	_ = os.WriteFile(config.PortFilePath(), []byte(fmt.Sprintf("%d", port)), 0644)

	return server.Serve(addr)
}

func findAvailablePort() (int, error) {
	// Try ports 9000-9999
	for port := 9000; port <= 9999; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return port, nil
		}
	}

	// Fall back to OS selection
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run()
}
