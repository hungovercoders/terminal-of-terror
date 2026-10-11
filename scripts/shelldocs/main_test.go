package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hungovercoders/terminal-of-terror/cmd"
)

// Every command gets a page that starts like a man page, and every shell
// gets a completion script that names the program.
func TestWriteCoversEveryCommandAndShell(t *testing.T) {
	dir := t.TempDir()
	if err := Write(cmd.Root(), dir); err != nil {
		t.Fatal(err)
	}
	cmds := Commands(cmd.Root())
	if len(cmds) < 12 {
		t.Fatalf("only %d commands found", len(cmds))
	}
	for _, c := range cmds {
		page, err := os.ReadFile(filepath.Join(dir, "man", ManName(c)+".1"))
		if err != nil {
			t.Errorf("%s: %v", c.CommandPath(), err)
			continue
		}
		if !strings.HasPrefix(string(page), ".TH ") || !strings.Contains(string(page), ".SH NAME") || !strings.Contains(string(page), c.Short) {
			t.Errorf("%s: not a man page:\n%s", c.CommandPath(), page)
		}
	}
	for _, ext := range []string{"bash", "zsh", "fish", "ps1"} {
		data, err := os.ReadFile(filepath.Join(dir, "completions", "terminal-of-terror."+ext))
		if err != nil || !strings.Contains(string(data), "terminal-of-terror") {
			t.Errorf("completions/terminal-of-terror.%s: %v", ext, err)
		}
	}
}

// The Homebrew cask lists the man pages one by one; a command added
// without its line would be installed without its page.
func TestCaskListsEveryManPage(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".goreleaser.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	listed := regexp.MustCompile(`(?m)^\s+- man/(terminal-of-terror[\w-]*\.1)$`).FindAllStringSubmatch(string(data), -1)
	var got []string
	for _, m := range listed {
		got = append(got, m[1])
	}
	var want []string
	for _, c := range Commands(cmd.Root()) {
		want = append(want, ManName(c)+".1")
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf(".goreleaser.yaml manpages:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// A leading dot or quote would be read as a roff macro.
func TestRoffEscapes(t *testing.T) {
	if roff(".hidden") != `\&.hidden` || roff(`a\b`) != `a\\b` || roff("plain") != "plain" {
		t.Error("roff escaping is off")
	}
}
