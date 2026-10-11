package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hungovercoders/terminal-of-terror/internal/store"
)

// go test runs with stdout and stdin on pipes, which is exactly the case
// the guard is for.
func TestRequireTerminalRefusesAPipe(t *testing.T) {
	err := requireTerminal("the Midnight Quiz", "use --json")
	if err == nil || !strings.Contains(err.Error(), "needs a terminal") || !strings.Contains(err.Error(), "use --json") {
		t.Fatalf("requireTerminal under go test = %v, want a refusal naming the alternative", err)
	}
}

// Without a terminal to confirm on, crypt reset does nothing unless -y
// says so.
func TestCryptResetWithoutATerminalNeedsYes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("TERMINAL_OF_TERROR_HOME", home)
	p, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	p.Seen = map[string]bool{"dracula": true}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	resetYes = false
	err = cryptResetCmd.RunE(cryptResetCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "-y") {
		t.Fatalf("crypt reset on a pipe = %v, want a refusal that names -y", err)
	}
	after, err := store.Load()
	if err != nil || !after.Seen["dracula"] {
		t.Errorf("progress was touched: %+v %v", after, err)
	}
	resetYes = true
	defer func() { resetYes = false }()
	if err := cryptResetCmd.RunE(cryptResetCmd, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, "progress.json")); err == nil {
		after, _ := store.Load()
		if after.Seen["dracula"] {
			t.Error("-y did not reset the crypt")
		}
	}
}
