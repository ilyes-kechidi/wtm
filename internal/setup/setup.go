// Package setup runs post-create and teardown commands inside a worktree.
package setup

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ilyes-kechidi/wtm/internal/config"
)

// Run executes steps sequentially, streaming output, stopping at the first
// failure. Each step runs via `sh -c` in root or root/Dir. Extra env entries
// (KEY=value) are appended to the process environment.
func Run(root string, steps []config.SetupStep, out io.Writer, env ...string) error {
	return run(root, steps, out, "setup", false, env)
}

// RunSkipMissing is Run for teardown. A step whose command is not found
// (exit 127) is skipped with a warning. That happens when an older worktree
// does not contain a script the current repo config now calls. A command that
// exists and fails still stops the sequence.
func RunSkipMissing(root string, steps []config.SetupStep, out io.Writer, env ...string) error {
	return run(root, steps, out, "teardown", true, env)
}

func run(root string, steps []config.SetupStep, out io.Writer, kind string, skipMissing bool, env []string) error {
	for i, s := range steps {
		if s.Run == "" {
			continue
		}
		dir := root
		if s.Dir != "" {
			dir = filepath.Join(root, s.Dir)
		}
		fmt.Fprintf(out, "$ %s (in %s)\n", s.Run, dir)
		cmd := exec.Command("sh", "-c", s.Run)
		cmd.Dir = dir
		cmd.Stdout = out
		cmd.Stderr = out
		if len(env) > 0 {
			cmd.Env = append(os.Environ(), env...)
		}
		if err := cmd.Run(); err != nil {
			if skipMissing && commandNotFound(err) {
				fmt.Fprintf(out, "warning: %s step %d (%q) not found, skipping\n", kind, i+1, s.Run)
				continue
			}
			return fmt.Errorf("%s step %d (%q) failed: %w", kind, i+1, s.Run, err)
		}
	}
	return nil
}

func commandNotFound(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 127
}
