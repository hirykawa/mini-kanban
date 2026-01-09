// Package main is the CLI entrypoint for mini-kanban.
package main

import (
	"fmt"
	"os"

	"mini-kanban/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
