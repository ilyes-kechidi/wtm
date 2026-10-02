package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const posixWrapper = `wtm() {
  if [ "$1" = "go" ]; then
    shift
    local dest
    dest="$(command wtm go "$@")" || return $?
    cd "$dest" || return $?
  elif [ $# -eq 0 ]; then
    local branch dest
    branch="$(command wtm)" || return $?
    dest="$(command wtm go "$branch")" || return $?
    cd "$dest" || return $?
  else
    command wtm "$@"
  fi
}
`

const fishWrapper = `function wtm
  if test "$argv[1]" = "go"
    set -e argv[1]
    set dest (command wtm go $argv); or return $status
    cd "$dest"; or return $status
  else if test (count $argv) -eq 0
    set branch (command wtm); or return $status
    set dest (command wtm go $branch); or return $status
    cd "$dest"; or return $status
  else
    command wtm $argv
  end
end
`

var initCmd = &cobra.Command{
	Use:   "init <bash|zsh|fish>",
	Short: "Print shell integration and branch completion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			fmt.Fprint(out, posixWrapper)
			return rootCmd.GenBashCompletion(out)
		case "zsh":
			fmt.Fprint(out, posixWrapper)
			return rootCmd.GenZshCompletion(out)
		case "fish":
			fmt.Fprint(out, fishWrapper)
			return rootCmd.GenFishCompletion(out, true)
		default:
			return fmt.Errorf("unsupported shell %q (bash|zsh|fish)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
