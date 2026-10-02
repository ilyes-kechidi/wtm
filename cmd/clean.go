package cmd

import (
	"fmt"

	"github.com/ilyes-kechidi/wtm/internal/git"
	"github.com/spf13/cobra"
)

var cleanDryRun bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Prune stale worktrees and remove ones merged into the default branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.RepoRoot(cwd())
		if err != nil {
			return err
		}
		if _, err := git.Run(repo, "worktree", "prune"); err != nil {
			return err
		}
		main := git.MainBranch(repo)
		merged, err := git.MergedBranches(repo, main)
		if err != nil {
			return err
		}
		wts, err := git.ListWorktrees(repo)
		if err != nil {
			return err
		}
		for _, w := range wts {
			if w.Bare || w.Branch == "" || w.Branch == "(detached)" {
				continue
			}
			if w.Path == repo {
				continue // never touch the main checkout
			}
			if !merged[w.Branch] {
				continue
			}
			if cleanDryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "would remove %s (%s, merged into %s)\n", w.Path, w.Branch, main)
				continue
			}
			if err := removeWorktree(cmd.OutOrStdout(), repo, w.Path, w.Branch, false, false); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: remove %s: %v\n", w.Path, err)
			}
		}
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "preview without removing")
	rootCmd.AddCommand(cleanCmd)
}
