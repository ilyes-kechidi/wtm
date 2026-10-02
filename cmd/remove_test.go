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

// setupRepoWithWorktree builds a repo whose worktree holds an untracked file
// and an ignored build-output dir — the layout that triggers git's
// "Directory not empty" failure on `git worktree remove`.
func setupRepoWithWorktree(t *testing.T) (repo, wt string) {
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
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("vendor/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(repo, "add", "-A")
	run(repo, "commit", "-qm", "init")
	wt = filepath.Join(t.TempDir(), "wt")
	run(repo, "worktree", "add", wt, "-b", "feature")
	if err := os.WriteFile(filepath.Join(wt, "notes.txt"), []byte("scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(wt, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "vendor", "bundle.js"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo, wt
}

func TestRemoveRefusesUntrackedWithoutForce(t *testing.T) {
	_, wt := setupRepoWithWorktree(t)
	err := refuseReasons(wt)
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected --force refusal, got %v", err)
	}
}

func TestRemoveForceClearsIgnoredLeftovers(t *testing.T) {
	repo, wt := setupRepoWithWorktree(t)
	var buf bytes.Buffer
	if err := removeWorktree(&buf, repo, wt, "feature", true, false); err != nil {
		t.Fatalf("removeWorktree --force: %v", err)
	}
	if _, serr := os.Stat(wt); !os.IsNotExist(serr) {
		t.Fatalf("worktree dir still exists: %s", wt)
	}
	if stillRegistered(repo, wt) {
		t.Fatal("worktree still registered after --force remove")
	}
}

// TestRemoveForceFallbackWhenGitFails exercises the RemoveAll+prune fallback
// deterministically: git cannot remove a plain directory, so the fallback
// must clear it (this is what saves old-git users from "Directory not empty").
func TestRemoveForceFallbackWhenGitFails(t *testing.T) {
	repo, _ := setupRepoWithWorktree(t)
	stubborn := filepath.Join(t.TempDir(), "stubborn")
	if err := os.MkdirAll(filepath.Join(stubborn, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stubborn, "vendor", "bundle.js"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := git.Run(repo, "worktree", "remove", "--force", stubborn); err == nil {
		t.Fatal("expected plain git to reject a non-worktree path")
	}
	var buf bytes.Buffer
	if err := removeWorktree(&buf, repo, stubborn, "feature", true, false); err != nil {
		t.Fatalf("removeWorktree --force fallback: %v", err)
	}
	if _, serr := os.Stat(stubborn); !os.IsNotExist(serr) {
		t.Fatalf("dir still exists: %s", stubborn)
	}
}

func writeTeardown(t *testing.T, repo, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, ".wtm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveRunsTeardownBeforeDelete(t *testing.T) {
	repo, wt := setupRepoWithWorktree(t)
	marker := filepath.Join(t.TempDir(), "teardown-ran")
	writeTeardown(t, repo, "teardown:\n  - touch "+marker+"\n  - 'test \"$WTM_BRANCH\" = feature'\n  - 'test \"$WTM_WORKTREE\" = "+wt+"'\n")
	var buf bytes.Buffer
	if err := removeWorktree(&buf, repo, wt, "feature", true, false); err != nil {
		t.Fatalf("removeWorktree: %v\n%s", err, buf.String())
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("teardown did not run: %v", err)
	}
	if _, serr := os.Stat(wt); !os.IsNotExist(serr) {
		t.Fatalf("worktree dir still exists after teardown+remove: %s", wt)
	}
}

func TestRemoveSkipsMissingTeardownCommand(t *testing.T) {
	repo, wt := setupRepoWithWorktree(t)
	writeTeardown(t, repo, "teardown:\n  - scripts/worktree-down.sh\n")
	var buf bytes.Buffer
	if err := removeWorktree(&buf, repo, wt, "feature", true, false); err != nil {
		t.Fatalf("removeWorktree: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "not found, skipping") {
		t.Fatalf("expected skip warning, got:\n%s", buf.String())
	}
	if _, serr := os.Stat(wt); !os.IsNotExist(serr) {
		t.Fatalf("worktree dir still exists: %s", wt)
	}
}

func TestRemoveAbortsWhenTeardownFails(t *testing.T) {
	repo, wt := setupRepoWithWorktree(t)
	writeTeardown(t, repo, "teardown:\n  - exit 7\n")
	var buf bytes.Buffer
	err := removeWorktree(&buf, repo, wt, "feature", true, false)
	if err == nil {
		t.Fatal("expected teardown failure")
	}
	if _, serr := os.Stat(wt); os.IsNotExist(serr) {
		t.Fatal("worktree was removed despite teardown failure")
	}
	if stillRegistered(repo, wt) == false {
		t.Fatal("worktree unregistered despite teardown failure")
	}
}
