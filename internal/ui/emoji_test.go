package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Emoji that need a variation selector (U+FE0F) to look like emoji, such as
// the coffin and the film frames, are drawn one column wide by some terminals and two by others,
// which breaks box alignment. docs/ui.md forbids them in anything the
// program draws; this test holds the Go sources and the monster data to it.
// Markdown is exempt: GitHub renders it, not a terminal.
func TestNoVariationSelectorEmoji(t *testing.T) {
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() && (name == ".git" || name == "node_modules" || name == "dist" || name == ".goreleaser") {
			return filepath.SkipDir
		}
		if d.IsDir() || !(strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".txt")) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.ContainsRune(line, '\uFE0F') {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("%s:%d uses an emoji with a variation selector; pick one that is wide everywhere (docs/ui.md)", rel, i+1)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
