package mash

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

func TestFightIsBestOfThree(t *testing.T) {
	a := *monsters.GetMonsterByName("dracula")
	b := *monsters.GetMonsterByName("wolf-man")
	for seed := int64(0); seed < 100; seed++ {
		bout := Fight(rand.New(rand.NewSource(seed)), a, b)
		if len(bout.Rounds) != 3 {
			t.Fatalf("want 3 rounds, got %d", len(bout.Rounds))
		}
		wa, wb := bout.Tally()
		if (wa > wb) != (bout.Winner == 0) {
			t.Fatalf("winner %d disagrees with tally %d-%d", bout.Winner, wa, wb)
		}
		stats := map[string]bool{}
		for _, r := range bout.Rounds {
			if stats[r.Stat] {
				t.Fatalf("stat %s used twice", r.Stat)
			}
			stats[r.Stat] = true
			if strings.Contains(r.Narration, "{") {
				t.Fatalf("unfilled placeholder: %s", r.Narration)
			}
		}
	}
}

func TestMidSentenceNames(t *testing.T) {
	a := *monsters.GetMonsterByName("wolf-man")
	b := *monsters.GetMonsterByName("mummy")
	got := fill("{A} faces {B}", a, b, a, b)
	if got != "The Wolf Man faces the Mummy" {
		t.Errorf("got %q", got)
	}
	if got := fill("A chase! {W} wins.", a, b, a, b); got != "A chase! The Wolf Man wins." {
		t.Errorf("sentence start: got %q", got)
	}
	if lowerFirst("The goddess Isis") != "the goddess Isis" || lowerFirst("ANANKE") != "ANANKE" {
		t.Error("lowerFirst mangled text")
	}
}
