package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilyes-kechidi/wtm/internal/git"
)

type pickerModel struct {
	items  []git.Worktree
	cursor int
	action string
	chosen string
	quit   bool
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.items) > 0 {
				m.chosen = m.items[m.cursor].Branch
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickerModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "wtm %s — ↑/↓ select, enter confirm, q quit\n\n", m.action)
	for i, w := range m.items {
		mark := " "
		if i == m.cursor {
			mark = ">"
		}
		br := w.Branch
		if br == "" {
			br = "(detached)"
		}
		fmt.Fprintf(&b, "%s %-24s %s\n", mark, br, w.Path)
	}
	return b.String()
}

// pickWorktree runs the built-in fuzzy-less picker and returns a branch name.
func pickWorktree(action string) (string, error) {
	repo, err := git.RepoRoot(cwd())
	if err != nil {
		return "", err
	}
	wts, err := git.ListWorktrees(repo)
	if err != nil {
		return "", err
	}
	// Filter out bare entries for switching.
	var items []git.Worktree
	for _, w := range wts {
		if action == "switch" && w.Bare {
			continue
		}
		items = append(items, w)
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}
	// Render the TUI on stderr so the selection on stdout stays capturable
	// by the shell wrapper (which cds to it) without breaking interactive use.
	p := tea.NewProgram(pickerModel{items: items, action: action}, tea.WithInput(os.Stdin), tea.WithOutput(os.Stderr))
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	m, ok := final.(pickerModel)
	if !ok || m.quit || m.chosen == "" {
		return "", fmt.Errorf("no selection")
	}
	return m.chosen, nil
}

func runPicker(action string) error {
	branch, err := pickWorktree(action)
	if err != nil {
		return err
	}
	fmt.Println(branch)
	return nil
}
