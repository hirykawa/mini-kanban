package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"mini-kanban/config"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database utilities",
}

var dbPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show database file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.DBPath())
	},
}

func init() {
	dbCmd.AddCommand(dbPathCmd)
}
