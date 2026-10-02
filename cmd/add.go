package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ilyes-kechidi/wtm/internal/config"
	"github.com/ilyes-kechidi/wtm/internal/envcopy"
	"github.com/ilyes-kechidi/wtm/internal/git"
	"github.com/ilyes-kechidi/wtm/internal/layout"
	"github.com/ilyes-kechidi/wtm/internal/setup"
	"github.com/spf13/cobra"
)

var (
	addBase       string
	addNoEnv      bool
	addNoSetup    bool
	addCopyEnv    []string
	addAllowDirty bool
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a sibling worktree for branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := strings.TrimSpace(args[0])
		if branch == "" {
			return fmt.Errorf("branch name required")
		}
		repo, err := git.RepoRoot(cwd())
		if err != nil {
			return err
		}
		if !addAllowDirty {
			if dirty, derr := git.IsDirty(repo); derr == nil && dirty {
				return fmt.Errorf("main repo is dirty (use --allow-dirty to override)")
			}
		}
		base := addBase
		if base == "" {
			base = "HEAD"
		}
		dest := layout.DestFor(repo, branch)
		if _, err := os.Stat(dest); err == nil {
			return fmt.Errorf("destination already exists: %s", dest)
		}
		var gerr error
		if git.BranchExists(repo, branch) {
			_, gerr = git.Run(repo, "worktree", "add", dest, branch)
		} else {
			_, gerr = git.Run(repo, "worktree", "add", "-b", branch, dest, base)
		}
		if gerr != nil {
			return gerr
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", dest)

		cfg := config.Load(repo)
		if !addNoEnv {
			globs := addCopyEnv
			if len(globs) == 0 {
				globs = cfg.Copy
			}
			copied, cerr := envcopy.CopyGlobs(repo, dest, globs)
			if cerr != nil {
				fmt.Fprintf(os.Stderr, "warning: env copy: %v\n", cerr)
				return nil
			}
			for _, f := range copied {
				fmt.Fprintf(cmd.OutOrStdout(), "copied %s\n", f)
			}
		}
		if !addNoSetup && len(cfg.Setup) > 0 {
			if err := setup.Run(dest, cfg.Setup, cmd.OutOrStdout(),
				"WTM_BRANCH="+branch,
				"WTM_WORKTREE="+dest,
			); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	addCmd.ValidArgsFunction = completeUnusedLocalBranches
	addCmd.Flags().StringVar(&addBase, "base", "", "base ref for new branches (default HEAD)")
	addCmd.Flags().BoolVar(&addNoEnv, "no-env", false, "skip copying env files")
	addCmd.Flags().BoolVar(&addNoSetup, "no-setup", false, "skip post-create setup commands")
	addCmd.Flags().StringSliceVar(&addCopyEnv, "copy-env", nil, "extra env globs (relative to repo root)")
	addCmd.Flags().BoolVar(&addAllowDirty, "allow-dirty", false, "allow creating while main repo is dirty")
	_ = addCmd.RegisterFlagCompletionFunc("base", completeBaseRefs)
	rootCmd.AddCommand(addCmd)
}
