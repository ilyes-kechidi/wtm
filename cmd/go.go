package cmd

import (
	"fmt"

	"github.com/ilyes-kechidi/wtm/internal/git"
	"github.com/ilyes-kechidi/wtm/internal/layout"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:   "go [<branch>]",
	Short: "Print a worktree path for shell cd (picker when no branch given)",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := ""
		if len(args) > 0 {
			branch = args[0]
		} else {
			picked, err := pickWorktree("switch")
			if err != nil {
				return err
			}
			branch = picked
		}
		repo, err := git.RepoRoot(cwd())
		if err != nil {
			return err
		}
		if wts, lerr := git.ListWorktrees(repo); lerr == nil {
			for _, w := range wts {
				if w.Branch == branch {
					fmt.Fprintln(cmd.OutOrStdout(), w.Path)
					return nil
				}
			}
		}
		// Fall back to conventional path (lets shell report a clean error).
		fmt.Fprintln(cmd.OutOrStdout(), layout.DestFor(repo, branch))
		return nil
	},
}

func init() {
	goCmd.ValidArgsFunction = completeWorktreeBranches
	rootCmd.AddCommand(goCmd)
}
