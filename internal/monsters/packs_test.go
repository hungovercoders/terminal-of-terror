package monsters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func restorePacks(t *testing.T) {
	saved, savedAll := packs, allPacks
	t.Cleanup(func() { allPacks = savedAll; setPacks(saved) })
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadUserPacks(t *testing.T) {
	restorePacks(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "cryptids", "pack.json"), `{"id":"cryptids","name":"Cryptids"}`)
	writeFile(t, filepath.Join(dir, "cryptids", "mothman.json"), `{"name":"Mothman","description":"Red-eyed winged figure of West Virginia","facts":["Sighted in Point Pleasant in 1966"]}`)
	writeFile(t, filepath.Join(dir, "cryptids", "mothman.txt"), " (o o)\n")
	writeFile(t, filepath.Join(dir, "cryptids", "nameless.json"), `{"description":"no name","facts":["x"]}`)
	writeFile(t, filepath.Join(dir, "cryptids", "broken.json"), `{nope`)
	writeFile(t, filepath.Join(dir, "clash", "dracula.json"), `{"name":"Dracula Again","description":"dupe","facts":["x"]}`)

	loaded, errs := LoadUserPacks(dir)
	if len(errs) != 2 {
		t.Fatalf("want 2 load errors (broken + nameless), got %v", errs)
	}
	errs = AddPacks(loaded)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "dracula") {
		t.Fatalf("want 1 clash error, got %v", errs)
	}

	m := GetMonsterByName("mothman")
	if m == nil {
		t.Fatal("mothman not loaded")
	}
	if m.Pack != "cryptids" || m.Emoji == "" || m.Stats.Strength != 5 || m.ASCII != " (o o)" || m.Debut.Era == "" {
		t.Errorf("defaults not applied: %+v", m)
	}

	if err := UsePacks([]string{"cryptids"}); err != nil {
		t.Fatal(err)
	}
	if len(GetAllMonsters()) != 1 {
		t.Errorf("filter left %d monsters", len(GetAllMonsters()))
	}
	if len(EveryMonster()) <= 1 {
		t.Error("EveryMonster should ignore the filter")
	}
	if err := UsePacks([]string{"nope"}); err == nil || !strings.Contains(err.Error(), "cryptids") {
		t.Errorf("unknown pack should list choices, got %v", err)
	}
}

func TestMissingUserDirIsFine(t *testing.T) {
	p, errs := LoadUserPacks(filepath.Join(t.TempDir(), "nothing-here"))
	if len(p) != 0 || len(errs) != 0 {
		t.Errorf("got %v %v", p, errs)
	}
}

func TestUsePacksIgnoresRepeats(t *testing.T) {
	restorePacks(t)
	before := len(GetAllMonsters())
	universal := 0
	for _, m := range GetAllMonsters() {
		if m.Pack == "universal" {
			universal++
		}
	}
	if err := UsePacks([]string{"universal", "universal"}); err != nil {
		t.Fatal(err)
	}
	if got := len(GetAllMonsters()); got != universal || got >= before {
		t.Errorf("--pack universal,universal gave %d monsters, want %d", got, universal)
	}
}

func TestCRLFArtIsNormalised(t *testing.T) {
	restorePacks(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "win", "ghost.json"), `{"name":"Windows Ghost","description":"Saved on Windows","facts":["x"]}`)
	writeFile(t, filepath.Join(dir, "win", "ghost.txt"), " (o o)\r\n  | |\r\n")
	loaded, errs := LoadUserPacks(dir)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	if art := loaded[0].Monsters[0].ASCII; art != " (o o)\n  | |" {
		t.Errorf("art = %q, want CRLF stripped", art)
	}
}
