package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ilyes-kechidi/wtm/internal/config"
)

func TestRunStepsInDirs(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	steps := []config.SetupStep{
		{Run: "mkdir -p front/app && touch front/app/installed.txt"},
		{Run: "touch created.txt", Dir: "front/app"},
	}
	if err := Run(root, steps, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, f := range []string{"front/app/installed.txt", "front/app/created.txt"} {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
}

func TestRunStopsOnFailure(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	err := Run(root, []config.SetupStep{
		{Run: "exit 3"},
		{Run: "touch should-not-exist.txt"},
	}, &out)
	if err == nil {
		t.Fatal("expected error from failing step")
	}
	if _, serr := os.Stat(filepath.Join(root, "should-not-exist.txt")); !os.IsNotExist(serr) {
		t.Fatal("second step ran after failure")
	}
}

func TestRunSkipMissingContinuesWhenCommandNotFound(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	err := RunSkipMissing(root, []config.SetupStep{
		{Run: "scripts/worktree-down.sh"},
		{Run: "touch after-skip.txt"},
	}, &out)
	if err != nil {
		t.Fatalf("RunSkipMissing: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "not found, skipping") {
		t.Fatalf("expected skip warning, got:\n%s", out.String())
	}
	if _, serr := os.Stat(filepath.Join(root, "after-skip.txt")); serr != nil {
		t.Fatalf("step after a missing command did not run: %v", serr)
	}
}

func TestRunSkipMissingStopsOnRealFailure(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	err := RunSkipMissing(root, []config.SetupStep{
		{Run: "exit 7"},
		{Run: "touch should-not-exist.txt"},
	}, &out)
	if err == nil {
		t.Fatal("expected error from failing teardown step")
	}
	if _, serr := os.Stat(filepath.Join(root, "should-not-exist.txt")); !os.IsNotExist(serr) {
		t.Fatal("step after a real failure still ran")
	}
}

func TestRunPassesExtraEnv(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "env.txt")
	var out bytes.Buffer
	err := Run(root, []config.SetupStep{
		{Run: `printf '%s\n%s\n' "$WTM_BRANCH" "$WTM_WORKTREE" > env.txt`},
	}, &out, "WTM_BRANCH=feature/x", "WTM_WORKTREE="+root)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(data))
	want := "feature/x\n" + root
	if got != want {
		t.Fatalf("env contents:\n got %q\nwant %q", got, want)
	}
}
