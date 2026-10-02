package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree holds one entry from `git worktree list --porcelain`.
type Worktree struct {
	Path   string
	HEAD   string
	Branch string // may be "(detached)" or ""
	Bare   bool
}

// Run runs git -C dir with args, returning trimmed combined output.
func Run(dir string, args ...string) (string, error) {
	full := []string{}
	if dir != "" {
		full = append(full, "-C", dir)
	}
	full = append(full, args...)
	cmd := exec.Command("git", full...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(out.String()), nil
}

// RepoRoot resolves the current repo via git rev-parse (works from any worktree).
func RepoRoot(cwd string) (string, error) {
	out, err := Run(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not inside a git repo: %w", err)
	}
	// In a linked worktree, --show-toplevel returns the worktree path.
	// Normalize to the main worktree via --git-common-dir.
	common, cerr := Run(out, "rev-parse", "--git-common-dir")
	if cerr != nil {
		return out, nil
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(out, common)
	}
	// common dir is <main>/.git (or <main>.git for bare); parent is main root.
	main := filepath.Dir(filepath.Clean(common))
	if filepath.Base(main) == ".git" {
		main = filepath.Dir(main)
	}
	// Sanity: must contain a .git entry or be bare; else fall back.
	if _, err := Run(main, "rev-parse", "--git-dir"); err != nil {
		return out, nil
	}
	return main, nil
}

// IsDirty reports uncommitted tracked changes in dir.
// Untracked files (e.g. .env files waiting to be copied) do not count.
func IsDirty(dir string) (bool, error) {
	out, err := Run(dir, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// UnpushedCount returns commits in HEAD not on upstream; -1 when no upstream.
func UnpushedCount(dir string) (int, error) {
	out, err := Run(dir, "rev-list", "--count", "@{u}..HEAD")
	if err != nil {
		return -1, nil // no upstream — not an error for our purposes
	}
	var n int
	if _, serr := fmt.Sscanf(strings.TrimSpace(out), "%d", &n); serr != nil {
		return -1, nil
	}
	return n, nil
}

// BranchExists checks local branch existence.
func BranchExists(repo, branch string) bool {
	_, err := Run(repo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

// MergedBranches returns branches merged into ref.
func MergedBranches(repo, ref string) (map[string]bool, error) {
	out, err := Run(repo, "branch", "--merged", ref, "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, b := range strings.Split(out, "\n") {
		b = strings.TrimSpace(b)
		if b != "" {
			set[b] = true
		}
	}
	return set, nil
}

// MainBranch guesses the default branch: origin/HEAD, else origin/main|master, else main.
func MainBranch(repo string) string {
	if out, err := Run(repo, "symbolic-ref", "refs/remotes/origin/HEAD"); err == nil {
		if i := strings.LastIndex(out, "/"); i >= 0 {
			return out[i+1:]
		}
	}
	for _, b := range []string{"main", "master"} {
		if _, err := Run(repo, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+b); err == nil {
			return b
		}
		if BranchExists(repo, b) {
			return b
		}
	}
	return "main"
}

// ListWorktrees parses `git worktree list --porcelain`.
func ListWorktrees(repo string) ([]Worktree, error) {
	out, err := Run(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var wts []Worktree
	var cur Worktree
	flush := func() {
		if cur.Path != "" {
			wts = append(wts, cur)
			cur = Worktree{}
		}
	}
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			cur.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			cur.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case line == "bare":
			cur.Bare = true
		case line == "detached":
			cur.Branch = "(detached)"
		case line == "":
			flush()
		}
	}
	flush()
	return wts, nil
}

// HasUntracked reports untracked files in dir.
func HasUntracked(dir string) (bool, error) {
	out, err := Run(dir, "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "??") {
			return true, nil
		}
	}
	return false, nil
}

// BranchOf resolves the branch checked out at path (for go/remove lookup).
func BranchOf(path string) string {
	out, err := Run(path, "branch", "--show-current")
	if err != nil || strings.TrimSpace(out) == "" {
		return ""
	}
	return strings.TrimSpace(out)
}
