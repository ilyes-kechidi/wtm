package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wtm",
	Short: "Manage git worktrees as siblings of the current repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Bare `wtm` opens the switch picker.
		return runPicker("switch")
	},
}

// Execute runs the CLI.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return err
	}
	return nil
}

func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
