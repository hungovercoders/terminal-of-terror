// Package crypt turns play into progress: monsters are captured by
// answering questions about them, and badges reward milestones.
package crypt

import (
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
)

// CaptureAt is how many correct answers about a monster capture it.
const CaptureAt = 3

// Badge is an achievement.
type Badge struct {
	ID          string `json:"id"`
	Emoji       string `json:"emoji"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Badges lists every achievement in display order.
var Badges = []Badge{
	{"first-fright", "🎃", "First Fright", "Finish your first quiz"},
	{"flawless", "💀", "Flawless Fiend", "Score 100% on a quiz of 5 or more questions"},
	{"grave-robber", "⚰️", "Grave Robber", "Capture your first monster"},
	{"eagle-eye", "🔍", "Eagle Eye", "Name a monster in Guess the Monster before the fog lifts"},
	{"promoter", "🥊", "Fight Promoter", "Stage your first Monster Mash"},
	{"scholar", "📚", "Midnight Scholar", "Visit every monster's page in the explorer"},
	{"night-owl", "🦉", "Night Owl", "Play between midnight and 4am"},
	{"full-moon", "🌕", "Survived the Full Moon", "Play a game on a full-moon night"},
	{"friday-13", "🐈", "Unlucky for Some", "Play a game on Friday the 13th"},
	{"halloween", "👻", "Halloween Spirit", "Play a game on Halloween"},
	{"silent-scholar", "🎞️", "Silent Era Scholar", "Capture every silent-film monster"},
	{"monster-kid", "📺", "Monster Kid", "Capture every Universal Classic"},
	{"master", "👑", "Master of the Crypt", "Capture every monster"},
}

// Outcome describes one session of play.
type Outcome struct {
	Kind      string         // quiz, guess, mash or explore
	Correct   map[string]int // correct answers per monster id
	Score     int            // percent for quizzes, points for guessing
	Total     int            // questions or rounds played
	Completed bool           // played to the end
	Seen      []string       // monster pages visited
	EagleEye  bool           // guessed a monster through the thickest fog
	Now       time.Time
}

// Unlocks is what a session newly earned.
type Unlocks struct {
	Captured []monsters.Monster
	Badges   []Badge
}

// Empty reports whether nothing new was earned.
func (u Unlocks) Empty() bool { return len(u.Captured) == 0 && len(u.Badges) == 0 }

// Apply records an outcome in p and reports anything newly unlocked.
func Apply(p *store.Progress, all []monsters.Monster, o Outcome) Unlocks {
	if o.Now.IsZero() {
		o.Now = time.Now()
	}
	var u Unlocks
	award := func(id string) {
		if _, ok := p.Badges[id]; ok {
			return
		}
		for _, b := range Badges {
			if b.ID == id {
				p.Badges[id] = o.Now
				u.Badges = append(u.Badges, b)
			}
		}
	}

	switch o.Kind {
	case "quiz":
		if o.Completed && o.Total > 0 {
			p.QuizzesPlayed++
			p.BestQuizScore = max(p.BestQuizScore, o.Score)
			award("first-fright")
			if o.Score == 100 && o.Total >= 5 {
				award("flawless")
			}
		}
	case "guess":
		if o.Completed && o.Total > 0 {
			p.GuessesPlayed++
			p.BestGuessScore = max(p.BestGuessScore, o.Score)
		}
		if o.EagleEye {
			award("eagle-eye")
		}
	case "mash":
		p.MashesPlayed++
		award("promoter")
	}
	if played := o.Kind == "quiz" || o.Kind == "guess" || o.Kind == "mash"; played {
		if o.Now.Hour() < 4 {
			award("night-owl")
		}
		if calendar.IsFullMoon(o.Now) {
			award("full-moon")
		}
		if calendar.IsFriday13(o.Now) {
			award("friday-13")
		}
		if calendar.IsHalloween(o.Now) {
			award("halloween")
		}
	}

	for _, id := range o.Seen {
		p.Seen[id] = true
	}
	for id, n := range o.Correct {
		p.Knowledge[id] += n
	}
	for _, m := range all {
		if _, done := p.Captured[m.ID]; !done && p.Knowledge[m.ID] >= CaptureAt {
			p.Captured[m.ID] = o.Now
			u.Captured = append(u.Captured, m)
		}
	}

	if len(p.Captured) > 0 {
		award("grave-robber")
	}
	if every(all, func(m monsters.Monster) bool { return p.Seen[m.ID] }) {
		award("scholar")
	}
	captured := func(m monsters.Monster) bool { _, ok := p.Captured[m.ID]; return ok }
	if allWhere(all, monsters.Monster.IsSilent, captured) {
		award("silent-scholar")
	}
	if allWhere(all, func(m monsters.Monster) bool { return m.Pack == "universal" }, captured) {
		award("monster-kid")
	}
	if every(all, captured) {
		award("master")
	}
	return u
}

// every reports whether every monster satisfies ok (false for none).
func every(all []monsters.Monster, ok func(monsters.Monster) bool) bool {
	return allWhere(all, func(monsters.Monster) bool { return true }, ok)
}

// allWhere reports whether every monster matching where satisfies ok,
// and that at least one matches.
func allWhere(all []monsters.Monster, where, ok func(monsters.Monster) bool) bool {
	n := 0
	for _, m := range all {
		if !where(m) {
			continue
		}
		n++
		if !ok(m) {
			return false
		}
	}
	return n > 0
}

// Earned reports whether the badge has been won, and when.
func Earned(p *store.Progress, id string) (time.Time, bool) {
	t, ok := p.Badges[id]
	return t, ok
}
