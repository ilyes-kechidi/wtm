package cmd

import (
	"strings"

	"github.com/ilyes-kechidi/wtm/internal/git"
	"github.com/spf13/cobra"
)

// completeWorktreeBranches suggests branches checked out in a worktree.
// Descriptions (after a tab) are the worktree paths.
func completeWorktreeBranches(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	repo, err := git.RepoRoot(cwd())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	wts, err := git.ListWorktrees(repo)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, w := range wts {
		if w.Bare || w.Branch == "" || w.Branch == "(detached)" {
			continue
		}
		out = append(out, w.Branch+"\t"+w.Path)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completeUnusedLocalBranches suggests local branches that are not already
// checked out in a linked worktree. The primary checkout stays in the list.
func completeUnusedLocalBranches(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return localBranches(false), cobra.ShellCompDirectiveNoFileComp
}

// completeBaseRefs suggests every local branch as a base ref.
func completeBaseRefs(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return localBranches(true), cobra.ShellCompDirectiveNoFileComp
}

// localBranches lists local branch names. When includeCheckedOut is false,
// branches checked out in a linked worktree are omitted. The primary
// worktree's branch is always included.
func localBranches(includeCheckedOut bool) []string {
	repo, err := git.RepoRoot(cwd())
	if err != nil {
		return nil
	}
	raw, err := git.Run(repo, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil
	}
	checkedOut := map[string]bool{}
	if !includeCheckedOut {
		if wts, lerr := git.ListWorktrees(repo); lerr == nil {
			for _, w := range wts {
				if w.Path == repo || w.Branch == "" || w.Branch == "(detached)" {
					continue
				}
				checkedOut[w.Branch] = true
			}
		}
	}
	var out []string
	for _, b := range strings.Split(raw, "\n") {
		b = strings.TrimSpace(b)
		if b == "" || checkedOut[b] {
			continue
		}
		out = append(out, b)
	}
	return out
}
