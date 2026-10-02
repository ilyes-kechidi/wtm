package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func completeOutput(t *testing.T, args ...string) string {
	t.Helper()
	var out, errb bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errb)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, errb.String())
	}
	return out.String()
}

func completionNames(out string) []string {
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		name, _, _ := strings.Cut(line, "\t")
		names = append(names, name)
	}
	return names
}

func hasName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

func TestCompleteRemoveIncludesWorktreeBranches(t *testing.T) {
	repo, _ := setupRepoWithWorktree(t)
	t.Chdir(repo)

	out := completeOutput(t, "__complete", "remove", "--force", "")
	names := completionNames(out)
	if !hasName(names, "main") || !hasName(names, "feature") {
		t.Fatalf("remove completions = %q\n%s", names, out)
	}
	last := ""
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		last = line
	}
	if last != fmt.Sprintf(":%d", cobra.ShellCompDirectiveNoFileComp) {
		t.Fatalf("directive = %q, want no-file-comp", last)
	}
}

func TestCompleteAddOmitsCheckedOutWorktreeBranch(t *testing.T) {
	repo, _ := setupRepoWithWorktree(t)
	t.Chdir(repo)

	names := completionNames(completeOutput(t, "__complete", "add", ""))
	if !hasName(names, "main") {
		t.Fatalf("add completions missing main: %q", names)
	}
	if hasName(names, "feature") {
		t.Fatalf("add completions include checked-out feature: %q", names)
	}
}

func TestCompleteRemoveSecondArgEmpty(t *testing.T) {
	repo, _ := setupRepoWithWorktree(t)
	t.Chdir(repo)

	names := completionNames(completeOutput(t, "__complete", "remove", "feature", ""))
	if len(names) != 0 {
		t.Fatalf("second positional completions = %q", names)
	}
}
