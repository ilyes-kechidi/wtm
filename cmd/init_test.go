package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// Bare `wtm` opens the switch picker, so every shell wrapper must intercept
// the no-args case and cd to the selection (regression: it used to print the
// branch name and stay put).
func TestInitWrappersInterceptBareWtm(t *testing.T) {
	needles := map[string]string{
		"bash": "[ $# -eq 0 ]",
		"zsh":  "[ $# -eq 0 ]",
		"fish": "(count $argv) -eq 0",
	}
	for shell, needle := range needles {
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)
		rootCmd.SetArgs([]string{"init", shell})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init %s: %v", shell, err)
		}
		out := buf.String()
		if !strings.Contains(out, needle) {
			t.Errorf("init %s does not intercept bare wtm:\n%s", shell, out)
		}
		if !strings.Contains(out, "command wtm go") {
			t.Errorf("init %s does not resolve the selection to a path:\n%s", shell, out)
		}
	}
}
