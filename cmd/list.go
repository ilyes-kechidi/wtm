package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ilyes-kechidi/wtm/internal/git"
	"github.com/spf13/cobra"
)

type listEntry struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	Status string `json:"status"`
}

var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List worktrees of the current repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.RepoRoot(cwd())
		if err != nil {
			return err
		}
		wts, err := git.ListWorktrees(repo)
		if err != nil {
			return err
		}
		entries := make([]listEntry, 0, len(wts))
		for _, w := range wts {
			status := "clean"
			if w.Bare {
				status = "bare"
			} else if dirty, derr := git.IsDirty(w.Path); derr == nil && dirty {
				status = "dirty"
			}
			entries = append(entries, listEntry{Branch: w.Branch, Path: w.Path, Status: status})
		}
		if listJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(entries)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-60s %s\n", "BRANCH", "PATH", "STATUS")
		for _, e := range entries {
			br := e.Branch
			if br == "" {
				br = "(detached)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-60s %s\n",
				truncate(br, 24), truncate(e.Path, 60), e.Status)
		}
		return nil
	},
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

var _ = strings.TrimSpace // keep import if unused in future edits

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output JSON")
	rootCmd.AddCommand(listCmd)
}
