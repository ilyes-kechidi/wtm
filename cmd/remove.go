package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ilyes-kechidi/wtm/internal/config"
	"github.com/ilyes-kechidi/wtm/internal/git"
	"github.com/ilyes-kechidi/wtm/internal/layout"
	"github.com/ilyes-kechidi/wtm/internal/setup"
	"github.com/spf13/cobra"
)

var (
	rmForce      bool
	rmWithBranch bool
)

var removeCmd = &cobra.Command{
	Use:   "remove [<branch>]",
	Short: "Remove a sibling worktree (picker when no branch given)",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := ""
		if len(args) > 0 {
			branch = args[0]
		} else {
			picked, err := pickWorktree("remove")
			if err != nil {
				return err
			}
			branch = picked
		}
		repo, err := git.RepoRoot(cwd())
		if err != nil {
			return err
		}
		dest := layout.DestFor(repo, branch)
		registered := false
		if wts, lerr := git.ListWorktrees(repo); lerr == nil {
			for _, w := range wts {
				if w.Branch == branch {
					dest = w.Path
					registered = true
					break
				}
			}
		}
		if !registered {
			if _, serr := os.Stat(dest); serr != nil {
				return fmt.Errorf("no worktree found for branch %q", branch)
			}
		}
		if !rmForce {
			if err := refuseReasons(dest); err != nil {
				return err
			}
		}
		return removeWorktree(cmd.OutOrStdout(), repo, dest, branch, rmForce, rmWithBranch)
	},
}

// refuseReasons reports why a non-force removal is blocked.
func refuseReasons(dest string) error {
	if dirty, derr := git.IsDirty(dest); derr == nil && dirty {
		return fmt.Errorf("worktree has uncommitted changes (use --force to discard): %s", dest)
	}
	if n, uerr := git.UnpushedCount(dest); uerr == nil && n > 0 {
		return fmt.Errorf("worktree has %d unpushed commit(s) (use --force): %s", n, dest)
	}
	if untracked, uerr := git.HasUntracked(dest); uerr == nil && untracked {
		return fmt.Errorf("worktree has untracked files (use --force to delete them too): %s", dest)
	}
	return nil
}

// runTeardown runs configured teardown steps inside the worktree. Failure
// means the caller must not remove the worktree.
func runTeardown(out io.Writer, repo, dest, branch string) error {
	cfg := config.Load(repo)
	if len(cfg.Teardown) == 0 {
		return nil
	}
	return setup.RunSkipMissing(dest, cfg.Teardown, out,
		"WTM_BRANCH="+branch,
		"WTM_WORKTREE="+dest,
	)
}

// removeWorktree removes the worktree at dest. With force, leftover files git
// won't delete (ignored build output, copied .env files) are removed as well
// and stale metadata is pruned.
func removeWorktree(out io.Writer, repo, dest, branch string, force, withBranch bool) error {
	if err := runTeardown(out, repo, dest, branch); err != nil {
		return err
	}
	args := []string{"worktree", "remove", dest}
	if force {
		args = append(args, "--force")
	}
	_, rerr := git.Run(repo, args...)
	if rerr != nil && !force {
		return rerr
	}
	if force {
		if st, serr := os.Stat(dest); serr == nil && st.IsDir() {
			if err := os.RemoveAll(dest); err != nil {
				if rerr != nil {
					return rerr
				}
				return fmt.Errorf("remove leftovers in %s: %w", dest, err)
			}
		}
		_, _ = git.Run(repo, "worktree", "prune")
	}
	if rerr != nil && stillRegistered(repo, dest) {
		return rerr
	}
	if withBranch {
		_, _ = git.Run(repo, "branch", "-D", branch)
	}
	fmt.Fprintf(out, "removed %s\n", dest)
	return nil
}

func stillRegistered(repo, dest string) bool {
	wts, err := git.ListWorktrees(repo)
	if err != nil {
		return true // unknown: report the original error
	}
	for _, w := range wts {
		if w.Path == dest {
			return true
		}
	}
	return false
}

func init() {
	removeCmd.ValidArgsFunction = completeWorktreeBranches
	removeCmd.Flags().BoolVar(&rmForce, "force", false, "remove even when dirty/unpushed; also deletes untracked and ignored files")
	removeCmd.Flags().BoolVar(&rmWithBranch, "with-branch", false, "also delete the branch")
	rootCmd.AddCommand(removeCmd)
}
