package quiz

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

func TestGeneratedQuestionsAreWellFormed(t *testing.T) {
	all := monsters.GetAllMonsters()
	for seed := int64(1); seed <= 50; seed++ {
		qs := Generate(rand.New(rand.NewSource(seed)), all, 10, "")
		if len(qs) != 10 {
			t.Fatalf("seed %d: got %d questions", seed, len(qs))
		}
		prompts := map[string]bool{}
		for _, q := range qs {
			if q.Answer < 0 || q.Answer >= len(q.Options) || len(q.Options) < 2 {
				t.Fatalf("seed %d: bad answer index: %+v", seed, q)
			}
			seen := map[string]bool{}
			for _, o := range q.Options {
				if seen[o] {
					t.Fatalf("seed %d: duplicate option %q in %+v", seed, o, q)
				}
				seen[o] = true
			}
			key := q.Prompt + q.Clue
			if prompts[key] {
				t.Fatalf("seed %d: repeated question %q", seed, key)
			}
			prompts[key] = true
			if q.Kind == "fact" || q.Kind == "quote" {
				m := monsters.GetMonsterByName(q.MonsterID)
				if strings.Contains(strings.ToLower(q.Clue), strings.ToLower(strings.TrimPrefix(m.Name, "The "))) {
					t.Errorf("clue gives away the answer: %q", q.Clue)
				}
			}
		}
	}
}

func TestFocusOnOneMonster(t *testing.T) {
	qs := Generate(rand.New(rand.NewSource(7)), monsters.GetAllMonsters(), 8, "mummy")
	if len(qs) == 0 {
		t.Fatal("no questions")
	}
	for _, q := range qs {
		if q.MonsterID != "mummy" {
			t.Errorf("question about %s, want mummy", q.MonsterID)
		}
	}
}

func TestMask(t *testing.T) {
	m := monsters.GetMonsterByName("dracula")
	got := mask("Dracula's cape and Count Dracula's castle", *m)
	if strings.Contains(got, "Dracula") {
		t.Errorf("mask left the name in: %q", got)
	}
	k := monsters.GetMonsterByName("krampus")
	if got := mask("Krampusnacht and Krampuslauf", *k); strings.Contains(got, "Krampus") {
		t.Errorf("mask left a compound in: %q", got)
	}
}

func TestRank(t *testing.T) {
	if Rank(100) != "Master of Horror" || Rank(0) != "Ghoul-in-Training" {
		t.Error("unexpected ranks")
	}
}
