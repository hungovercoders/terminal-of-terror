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
	{"flawless", "🏆", "Flawless Fiend", "Score 100% on a quiz of 5 or more questions"},
	{"grave-robber", "🦴", "Grave Robber", "Capture your first monster"},
	{"eagle-eye", "🔍", "Eagle Eye", "Name a monster in Guess the Monster before the fog lifts"},
	{"promoter", "🥊", "Fight Promoter", "Stage your first Monster Mash"},
	{"scholar", "📚", "Midnight Scholar", "Visit every monster's page in the explorer"},
	{"night-owl", "🦉", "Night Owl", "Play between midnight and 4am"},
	{"full-moon", "🌕", "Survived the Full Moon", "Play a game on a full-moon night"},
	{"friday-13", "🐈", "Unlucky for Some", "Play a game on Friday the 13th"},
	{"halloween", "👻", "Halloween Spirit", "Play a game on Halloween"},
	{"silent-scholar", "🎬", "Silent Era Scholar", "Capture every silent-film monster"},
	{"monster-kid", "📺", "Monster Kid", "Capture every Universal Classic"},
	{"folklorist", "🌍", "Folklorist", "Capture 5 monsters from World Folklore"},
	{"cryptozoologist", "🔭", "Cryptozoologist", "Capture every cryptid"},
	{"bookworm", "📖", "Bookworm", "Capture every literary monster"},
	{"hero-of-legend", "🏹", "Hero of Legend", "Capture every monster of classical mythology"},
	{"master", "👑", "Master of the Crypt", "Capture every monster"},
}

// packBadges are awarded for capturing every monster in a pack, in the
// order they are announced.
var packBadges = []struct{ pack, badge string }{
	{"cryptids", "cryptozoologist"},
	{"literary", "bookworm"},
	{"mythology", "hero-of-legend"},
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
	a := &awarder{p: p, now: o.Now}
	a.recordGame(o)
	if o.Kind == "quiz" || o.Kind == "guess" || o.Kind == "mash" {
		a.calendarBadges()
	}
	a.recordKnowledge(all, o)
	a.collectionBadges(all)
	return a.u
}

// awarder applies one outcome to progress, collecting what it unlocks.
type awarder struct {
	p   *store.Progress
	now time.Time
	u   Unlocks
}

// award gives badge id unless it was already won.
func (a *awarder) award(id string) {
	if _, ok := a.p.Badges[id]; ok {
		return
	}
	for _, b := range Badges {
		if b.ID == id {
			a.p.Badges[id] = a.now
			a.u.Badges = append(a.u.Badges, b)
		}
	}
}

// awardIf gives badge id when cond holds.
func (a *awarder) awardIf(cond bool, id string) {
	if cond {
		a.award(id)
	}
}

// recordGame updates the per-game counters and badges.
func (a *awarder) recordGame(o Outcome) {
	p := a.p
	finished := o.Completed && o.Total > 0
	switch o.Kind {
	case "quiz":
		if finished {
			p.QuizzesPlayed++
			p.BestQuizScore = max(p.BestQuizScore, o.Score)
			a.award("first-fright")
			a.awardIf(o.Score == 100 && o.Total >= 5, "flawless")
		}
	case "guess":
		if finished {
			p.GuessesPlayed++
			p.BestGuessScore = max(p.BestGuessScore, o.Score)
		}
		a.awardIf(o.EagleEye, "eagle-eye")
	case "mash":
		p.MashesPlayed++
		a.award("promoter")
	}
}

// calendarBadges rewards playing at a memorable time.
func (a *awarder) calendarBadges() {
	a.awardIf(a.now.Hour() < 4, "night-owl")
	a.awardIf(calendar.IsFullMoon(a.now), "full-moon")
	a.awardIf(calendar.IsFriday13(a.now), "friday-13")
	a.awardIf(calendar.IsHalloween(a.now), "halloween")
}

// recordKnowledge adds pages seen and correct answers, capturing any
// monster that reaches CaptureAt.
func (a *awarder) recordKnowledge(all []monsters.Monster, o Outcome) {
	p := a.p
	for _, id := range o.Seen {
		p.Seen[id] = true
	}
	for id, n := range o.Correct {
		p.Knowledge[id] += n
	}
	for _, m := range all {
		if _, done := p.Captured[m.ID]; !done && p.Knowledge[m.ID] >= CaptureAt {
			p.Captured[m.ID] = a.now
			a.u.Captured = append(a.u.Captured, m)
		}
	}
}

// collectionBadges rewards what the crypt holds and what has been read.
func (a *awarder) collectionBadges(all []monsters.Monster) {
	p := a.p
	captured := func(m monsters.Monster) bool { _, ok := p.Captured[m.ID]; return ok }
	inPack := func(pack string) func(monsters.Monster) bool {
		return func(m monsters.Monster) bool { return m.Pack == pack }
	}
	a.awardIf(len(p.Captured) > 0, "grave-robber")
	a.awardIf(every(all, func(m monsters.Monster) bool { return p.Seen[m.ID] }), "scholar")
	a.awardIf(allWhere(all, monsters.Monster.IsSilent, captured), "silent-scholar")
	a.awardIf(allWhere(all, inPack("universal"), captured), "monster-kid")
	a.awardIf(countWhere(all, inPack("folklore"), captured) >= 5, "folklorist")
	for _, pb := range packBadges {
		a.awardIf(allWhere(all, inPack(pb.pack), captured), pb.badge)
	}
	a.awardIf(every(all, captured), "master")
}

// countWhere counts the monsters matching where that satisfy ok.
func countWhere(all []monsters.Monster, where, ok func(monsters.Monster) bool) int {
	n := 0
	for _, m := range all {
		if where(m) && ok(m) {
			n++
		}
	}
	return n
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
