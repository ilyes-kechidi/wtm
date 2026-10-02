package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ilyes-kechidi/wtm/internal/git"
)

// setupMergedWorktree builds a repo with a feature branch worktree whose
// commits are already merged into main (eligible for `wtm clean`).
func setupMergedWorktree(t *testing.T) (repo, wt string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo = t.TempDir()
	run := func(dir string, args ...string) string {
		t.Helper()
		out, err := git.Run(dir, args...)
		if err != nil {
			t.Fatalf("git %s: %v", strings.Join(args, " "), err)
		}
		return out
	}
	run("", "init", "-q", "-b", "main", repo)
	run(repo, "config", "user.email", "t@t.t")
	run(repo, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(repo, "add", "-A")
	run(repo, "commit", "-qm", "init")
	run(repo, "checkout", "-qb", "feature")
	if err := os.WriteFile(filepath.Join(repo, "feat.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(repo, "add", "-A")
	run(repo, "commit", "-qm", "feat")
	run(repo, "checkout", "-q", "main")
	run(repo, "merge", "-q", "--no-ff", "feature", "-m", "merge feature")
	wt = filepath.Join(t.TempDir(), "wt")
	run(repo, "worktree", "add", wt, "feature")
	return repo, wt
}

func TestCleanRunsTeardownBeforeDelete(t *testing.T) {
	repo, wt := setupMergedWorktree(t)
	marker := filepath.Join(t.TempDir(), "teardown-ran")
	if err := os.WriteFile(filepath.Join(repo, ".wtm.yaml"), []byte(
		"teardown:\n  - touch "+marker+"\n  - 'test \"$WTM_BRANCH\" = feature'\n  - 'test \"$WTM_WORKTREE\" = "+wt+"'\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	cleanDryRun = false
	var out, errBuf bytes.Buffer
	cleanCmd.SetOut(&out)
	cleanCmd.SetErr(&errBuf)
	if err := cleanCmd.RunE(cleanCmd, nil); err != nil {
		t.Fatalf("clean: %v\nstdout=%s\nstderr=%s", err, out.String(), errBuf.String())
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("teardown did not run: %v\nstdout=%s\nstderr=%s", err, out.String(), errBuf.String())
	}
	if _, serr := os.Stat(wt); !os.IsNotExist(serr) {
		t.Fatalf("worktree still exists: %s", wt)
	}
}

func TestCleanAbortsWhenTeardownFails(t *testing.T) {
	repo, wt := setupMergedWorktree(t)
	if err := os.WriteFile(filepath.Join(repo, ".wtm.yaml"), []byte("teardown:\n  - exit 7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	cleanDryRun = false
	var out, errBuf bytes.Buffer
	cleanCmd.SetOut(&out)
	cleanCmd.SetErr(&errBuf)
	if err := cleanCmd.RunE(cleanCmd, nil); err != nil {
		t.Fatalf("clean: %v", err)
	}
	if !strings.Contains(errBuf.String(), "warning: remove") {
		t.Fatalf("expected warning on stderr, got %q", errBuf.String())
	}
	if _, serr := os.Stat(wt); os.IsNotExist(serr) {
		t.Fatal("worktree was removed despite teardown failure")
	}
}

func TestCleanDryRunSkipsTeardown(t *testing.T) {
	repo, wt := setupMergedWorktree(t)
	marker := filepath.Join(t.TempDir(), "teardown-ran")
	if err := os.WriteFile(filepath.Join(repo, ".wtm.yaml"), []byte("teardown:\n  - touch "+marker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	cleanDryRun = true
	t.Cleanup(func() { cleanDryRun = false })
	var out, errBuf bytes.Buffer
	cleanCmd.SetOut(&out)
	cleanCmd.SetErr(&errBuf)
	if err := cleanCmd.RunE(cleanCmd, nil); err != nil {
		t.Fatalf("clean --dry-run: %v", err)
	}
	if _, serr := os.Stat(marker); !os.IsNotExist(serr) {
		t.Fatal("teardown ran during --dry-run")
	}
	if _, serr := os.Stat(wt); os.IsNotExist(serr) {
		t.Fatal("worktree was removed during --dry-run")
	}
	if !strings.Contains(out.String(), "would remove") {
		t.Fatalf("expected dry-run message, got %q", out.String())
	}
}
