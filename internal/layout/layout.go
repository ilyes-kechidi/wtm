package layout

import (
	"os"
	"path/filepath"
	"strings"
)

// SanitizeBranch maps feature/foo -> feature-foo for directory names.
func SanitizeBranch(branch string) string {
	s := strings.ReplaceAll(branch, "/", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "worktree"
	}
	return s
}

// WorktreesRoot returns <repo>-worktrees next to the main repo,
// or $WTM_ROOT/<repo-name> when WTM_ROOT is set.
func WorktreesRoot(mainRepo string) string {
	if override := strings.TrimSpace(os.Getenv("WTM_ROOT")); override != "" {
		return filepath.Join(override, filepath.Base(mainRepo))
	}
	return mainRepo + "-worktrees"
}

// DestFor returns the worktree directory for branch.
func DestFor(mainRepo, branch string) string {
	return filepath.Join(WorktreesRoot(mainRepo), SanitizeBranch(branch))
}
