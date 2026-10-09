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
	// Community ids must not clash with built-in ones (cryptids and mothman are built in now).
	writeFile(t, filepath.Join(dir, "urban-legends", "pack.json"), `{"id":"urban-legends","name":"Urban Legends"}`)
	writeFile(t, filepath.Join(dir, "urban-legends", "bloody-mary.json"), `{"name":"Bloody Mary","description":"Say her name three times in the mirror","facts":["A chant game recorded by folklorists since the 1970s"]}`)
	writeFile(t, filepath.Join(dir, "urban-legends", "bloody-mary.txt"), " (o o)\n")
	writeFile(t, filepath.Join(dir, "urban-legends", "nameless.json"), `{"description":"no name","facts":["x"]}`)
	writeFile(t, filepath.Join(dir, "urban-legends", "broken.json"), `{nope`)
	writeFile(t, filepath.Join(dir, "clash", "dracula.json"), `{"name":"Dracula Again","description":"dupe","facts":["x"]}`)
	writeFile(t, filepath.Join(dir, "cryptids", "thing.json"), `{"name":"The Thing","description":"a built-in pack id","facts":["x"]}`)

	loaded, errs := LoadUserPacks(dir)
	if len(errs) != 2 {
		t.Fatalf("want 2 load errors (broken + nameless), got %v", errs)
	}
	errs = AddPacks(loaded)
	if len(errs) != 2 {
		t.Fatalf("want 2 clash errors (monster id, pack id), got %v", errs)
	}
	joined := errs[0].Error() + errs[1].Error()
	if !strings.Contains(joined, "dracula") || !strings.Contains(joined, `pack "cryptids"`) {
		t.Fatalf("want a monster clash and a pack clash, got %v", errs)
	}
	if m := GetMonsterByName("the thing"); m != nil {
		t.Error("a community pack with a built-in id must not load")
	}

	m := GetMonsterByName("bloody mary")
	if m == nil {
		t.Fatal("bloody-mary not loaded")
	}
	if m.Pack != "urban-legends" || m.Emoji == "" || m.Stats.Strength != 5 || m.ASCII != " (o o)" || m.Debut.Era == "" {
		t.Errorf("defaults not applied: %+v", m)
	}

	if err := UsePacks([]string{"urban-legends"}); err != nil {
		t.Fatal(err)
	}
	if len(GetAllMonsters()) != 1 {
		t.Errorf("filter left %d monsters", len(GetAllMonsters()))
	}
	if len(EveryMonster()) <= 1 {
		t.Error("EveryMonster should ignore the filter")
	}
	if err := UsePacks([]string{"nope"}); err == nil || !strings.Contains(err.Error(), "urban-legends") {
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
