package crypt

import (
	"testing"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
)

var evening = time.Date(2026, 9, 23, 20, 0, 0, 0, time.UTC)

func hasBadge(u Unlocks, id string) bool {
	for _, b := range u.Badges {
		if b.ID == id {
			return true
		}
	}
	return false
}

func TestCaptureAfterEnoughCorrectAnswers(t *testing.T) {
	p := store.New()
	all := monsters.GetAllMonsters()

	u := Apply(p, all, Outcome{Kind: "quiz", Correct: map[string]int{"dracula": 2}, Score: 20, Total: 10, Completed: true, Now: evening})
	if len(u.Captured) != 0 {
		t.Fatal("2 correct answers should not capture yet")
	}
	if !hasBadge(u, "first-fright") {
		t.Fatal("first completed quiz should earn First Fright")
	}

	u = Apply(p, all, Outcome{Kind: "quiz", Correct: map[string]int{"dracula": 1}, Score: 10, Total: 10, Completed: true, Now: evening})
	if len(u.Captured) != 1 || u.Captured[0].ID != "dracula" {
		t.Fatalf("expected Dracula captured, got %+v", u.Captured)
	}
	if !hasBadge(u, "grave-robber") || hasBadge(u, "first-fright") {
		t.Fatalf("expected only new badges, got %+v", u.Badges)
	}
	if p.QuizzesPlayed != 2 {
		t.Fatalf("quizzes played = %d", p.QuizzesPlayed)
	}
}

func TestBadgeRules(t *testing.T) {
	all := monsters.GetAllMonsters()
	p := store.New()
	if u := Apply(p, all, Outcome{Kind: "quiz", Score: 100, Total: 5, Completed: true, Now: evening}); !hasBadge(u, "flawless") {
		t.Error("perfect 5-question quiz should earn Flawless Fiend")
	}
	if u := Apply(p, all, Outcome{Kind: "mash", Now: time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)}); !hasBadge(u, "promoter") || !hasBadge(u, "night-owl") {
		t.Error("2am mash should earn Fight Promoter and Night Owl")
	}

	correct := map[string]int{}
	var seen []string
	for _, m := range all {
		correct[m.ID] = CaptureAt
		seen = append(seen, m.ID)
	}
	u := Apply(p, all, Outcome{Kind: "explore", Seen: seen, Correct: correct, Now: evening})
	for _, id := range []string{"scholar", "silent-scholar", "monster-kid", "folklorist", "master"} {
		if !hasBadge(u, id) {
			t.Errorf("expected badge %s", id)
		}
	}
}

func TestShortQuizIsNotFlawless(t *testing.T) {
	p := store.New()
	if u := Apply(p, monsters.GetAllMonsters(), Outcome{Kind: "quiz", Score: 100, Total: 3, Completed: true, Now: evening}); hasBadge(u, "flawless") {
		t.Error("a 3-question quiz is too short for Flawless Fiend")
	}
}

func TestCalendarBadges(t *testing.T) {
	p := store.New()
	all := monsters.GetAllMonsters()
	// Halloween 2025 was a Friday, but not the 13th; 13 Feb 2026 was a Friday.
	u := Apply(p, all, Outcome{Kind: "mash", Now: time.Date(2025, 10, 31, 21, 0, 0, 0, time.UTC)})
	if !hasBadge(u, "halloween") || hasBadge(u, "friday-13") {
		t.Errorf("Halloween badges: %+v", u.Badges)
	}
	u = Apply(p, all, Outcome{Kind: "quiz", Total: 1, Completed: true, Now: time.Date(2026, 2, 13, 21, 0, 0, 0, time.UTC)})
	if !hasBadge(u, "friday-13") {
		t.Errorf("Friday 13th badges: %+v", u.Badges)
	}
	u = Apply(p, all, Outcome{Kind: "guess", Now: time.Date(2025, 10, 7, 3, 0, 0, 0, time.UTC)})
	if !hasBadge(u, "full-moon") {
		t.Errorf("full moon badges: %+v", u.Badges)
	}
	u = Apply(store.New(), all, Outcome{Kind: "explore", Now: time.Date(2025, 10, 31, 21, 0, 0, 0, time.UTC)})
	if hasBadge(u, "halloween") {
		t.Error("browsing isn't playing a game")
	}
}
