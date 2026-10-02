package layout

import "testing"

func TestSanitizeBranch(t *testing.T) {
	cases := map[string]string{
		"feat/foo":   "feat-foo",
		"a/b/c":      "a-b-c",
		"main":       "main",
		"":           "worktree",
		"my feature": "my-feature",
	}
	for in, want := range cases {
		if got := SanitizeBranch(in); got != want {
			t.Errorf("SanitizeBranch(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDestFor(t *testing.T) {
	got := DestFor("/work/demo", "feat/foo")
	want := "/work/demo-worktrees/feat-foo"
	if got != want {
		t.Errorf("DestFor = %q, want %q", got, want)
	}
}
