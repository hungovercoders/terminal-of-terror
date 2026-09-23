// Package mash stages a best-of-three bout between two monsters, decided by
// their stat cards and a roll of the dice.
package mash

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// Round is one exchange in the bout.
type Round struct {
	Number    int    `json:"number"`
	Stat      string `json:"stat"`
	RollA     int    `json:"rollA"`
	RollB     int    `json:"rollB"`
	Winner    int    `json:"winner"` // 0 for A, 1 for B
	Narration string `json:"narration"`
	Move      string `json:"move"`
}

// Bout is a whole fight.
type Bout struct {
	A      monsters.Monster `json:"-"`
	B      monsters.Monster `json:"-"`
	Rounds []Round          `json:"rounds"`
	Winner int              `json:"winner"`
	Finale string           `json:"finale"`
}

// WinnerMonster returns the victor.
func (b Bout) WinnerMonster() monsters.Monster {
	if b.Winner == 0 {
		return b.A
	}
	return b.B
}

// LoserMonster returns the defeated monster.
func (b Bout) LoserMonster() monsters.Monster {
	if b.Winner == 0 {
		return b.B
	}
	return b.A
}

// Tally returns rounds won by A and B.
func (b Bout) Tally() (a, bWins int) {
	for _, r := range b.Rounds {
		if r.Winner == 0 {
			a++
		} else {
			bWins++
		}
	}
	return a, bWins
}

type contest struct {
	stat  string
	value func(monsters.Stats) int
	line  string // {W} winner, {L} loser
}

var contests = []contest{
	{"Strength", func(s monsters.Stats) int { return s.Strength },
		"{A} and {B} lock arms in the graveyard fog. {W} hurls {L} clean through a mausoleum door!"},
	{"Speed", func(s monsters.Stats) int { return s.Speed },
		"A chase across the moonlit moors! {W} is a blur, and {L} is left gasping at the cemetery gates."},
	{"Cunning", func(s monsters.Stats) int { return s.Cunning },
		"A battle of wits by candlelight. {W} lays a trap, and {L} walks straight into it."},
	{"Dread", func(s monsters.Stats) int { return s.Dread },
		"A staring contest of pure terror. {L} blinks first, and {W} lets out a bloodcurdling laugh."},
}

// Fight stages a bout of three rounds. Each round tests a different stat:
// stat plus a six-sided die, highest wins.
func Fight(r *rand.Rand, a, b monsters.Monster) Bout {
	bout := Bout{A: a, B: b}
	order := r.Perm(len(contests))[:3]
	wins := [2]int{}
	for n, ci := range order {
		c := contests[ci]
		sa, sb := c.value(a.Stats), c.value(b.Stats)
		ra, rb := sa+r.Intn(6)+1, sb+r.Intn(6)+1
		for tries := 0; ra == rb && tries < 5; tries++ {
			ra, rb = sa+r.Intn(6)+1, sb+r.Intn(6)+1
		}
		winner := 0
		if rb > ra || (rb == ra && sb > sa) {
			winner = 1
		}
		wins[winner]++
		w, l := pair(a, b, winner)
		bout.Rounds = append(bout.Rounds, Round{
			Number:    n + 1,
			Stat:      c.stat,
			RollA:     ra,
			RollB:     rb,
			Winner:    winner,
			Narration: fill(c.line, a, b, w, l),
			Move:      fmt.Sprintf("%s's signature move: %s!", w.Name, pickOr(r, w.Powers, "sheer terror")),
		})
	}
	if wins[1] > wins[0] {
		bout.Winner = 1
	}
	_, l := pair(a, b, bout.Winner)
	bout.Finale = fmt.Sprintf("%s limps off into the night, muttering about %s.", l.Name, lowerFirst(pickOr(r, l.Weaknesses, "bad luck")))
	return bout
}

func pair(a, b monsters.Monster, winner int) (w, l monsters.Monster) {
	if winner == 0 {
		return a, b
	}
	return b, a
}

// fill replaces placeholders, using "the" rather than "The" mid-sentence.
func fill(line string, a, b, w, l monsters.Monster) string {
	repl := map[string]string{"{A}": a.Name, "{B}": b.Name, "{W}": w.Name, "{L}": l.Name}
	out := line
	for k, v := range repl {
		for {
			i := strings.Index(out, k)
			if i < 0 {
				break
			}
			name := v
			if !sentenceStart(out[:i]) {
				name = MidSentence(v)
			}
			out = out[:i] + name + out[i+len(k):]
		}
	}
	return out
}

// sentenceStart reports whether text ending at this point starts a new sentence.
func sentenceStart(before string) bool {
	t := strings.TrimRight(before, " ")
	return t == "" || strings.HasSuffix(t, ".") || strings.HasSuffix(t, "!") || strings.HasSuffix(t, "?")
}

// MidSentence lowercases a leading "The" so names read naturally mid-sentence.
func MidSentence(name string) string {
	if strings.HasPrefix(name, "The ") {
		return "the " + name[4:]
	}
	return name
}

func lowerFirst(s string) string {
	r := []rune(s)
	if len(r) > 1 && unicode.IsUpper(r[0]) && unicode.IsLower(r[1]) {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}

func pickOr(r *rand.Rand, xs []string, fallback string) string {
	if len(xs) == 0 {
		return fallback
	}
	return xs[r.Intn(len(xs))]
}
