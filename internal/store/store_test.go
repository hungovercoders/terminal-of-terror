package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, dir)

	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Captured) != 0 {
		t.Fatal("fresh progress should be empty")
	}
	p.Knowledge["dracula"] = 2
	p.Captured["mummy"] = time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	p.QuizzesPlayed = 3
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}

	q, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if q.Knowledge["dracula"] != 2 || q.QuizzesPlayed != 3 || q.Captured["mummy"].IsZero() {
		t.Fatalf("round trip lost data: %+v", q)
	}

	if err := Reset(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "progress.json")); !os.IsNotExist(err) {
		t.Fatal("reset should delete the file")
	}
}

func TestCorruptFileIsAnError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, dir)
	os.WriteFile(filepath.Join(dir, "progress.json"), []byte("{nope"), 0o644)
	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a corrupt file")
	}
}
