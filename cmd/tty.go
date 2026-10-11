package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
)

// requireTerminal is the guard in front of every command that opens a
// Bubble Tea screen. Run from a script, a pipe or CI, such a command would
// otherwise hang waiting for a terminal or spray escape codes into the
// pipe; this says plainly what it needs instead. what names the screen in
// the show's own words; alternative says what a script should use.
func requireTerminal(what, alternative string) error {
	if isTerminal(os.Stdout) && isTerminal(os.Stdin) {
		return nil
	}
	return fmt.Errorf("%s needs a terminal to play on (stdout or stdin is not one); %s", what, alternative)
}

// isTerminal reports whether f is attached to a terminal.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(f.Fd())
}
