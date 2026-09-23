package ui

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/quiz"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
)

func step(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		m, _ = m.Update(key(k))
	}
	return m
}

func TestQuizPlaythrough(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	qs := quiz.Generate(r, monsters.GetAllMonsters(), 4, "")
	var m tea.Model = newQuizModel(qs, r)
	m, _ = m.Update(size(80, 24))

	// Answer the first question correctly, the rest wrongly.
	m = step(m, string(rune('1'+qs[0].Answer)), "enter")
	for _, q := range qs[1:] {
		wrong := (q.Answer + 1) % len(q.Options)
		m = step(m, string(rune('1'+wrong)))
		if !strings.Contains(m.View(), "The answer is") {
			t.Fatal("wrong answer should show the right one")
		}
		m = step(m, "enter")
	}
	qm := m.(quizModel)
	if !qm.finished || !qm.res.Completed {
		t.Fatal("quiz should be finished")
	}
	if c, n := qm.res.Score(); c != 1 || n != 4 || qm.res.Percent() != 25 {
		t.Fatalf("score = %d/%d (%d%%)", c, n, qm.res.Percent())
	}
	if !strings.Contains(m.View(), quiz.Rank(25)) {
		t.Error("results should show the rank")
	}
	if qm.res.CorrectBy()[qs[0].MonsterID] != 1 {
		t.Error("correct answer not credited to its monster")
	}
}

func TestGuessPlaythrough(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	rounds := NewGuessRounds(r, monsters.GetAllMonsters(), 3)
	if len(rounds) != 3 {
		t.Fatalf("got %d rounds", len(rounds))
	}
	var m tea.Model = newGuessModel(rounds, r)
	m, _ = m.Update(size(100, 40))

	first := m.View()
	m = step(m, " ", " ")
	if m.View() == first {
		t.Error("clearing the fog should change the portrait")
	}
	// Round 1 at stage 2 correct, round 2 at stage 0 correct, round 3 wrong.
	m = step(m, string(rune('1'+rounds[0].Answer)), "enter")
	m = step(m, string(rune('1'+rounds[1].Answer)), "enter")
	m = step(m, string(rune('1'+(rounds[2].Answer+1)%4)), "enter")
	gm := m.(guessModel)
	if !gm.finished {
		t.Fatal("game should be over")
	}
	if got, want := gm.res.Score(), fogPoints[2]+fogPoints[0]; got != want {
		t.Errorf("score = %d, want %d", got, want)
	}
	if !gm.res.EagleEye() {
		t.Error("a stage-0 correct guess is an Eagle Eye")
	}
}

func TestFogHidesThePortrait(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	g := NewGuessRounds(r, monsters.GetAllMonsters(), 1)[0]
	full := portraitCells(g.Monster.ASCII)
	var want []string
	for _, row := range full {
		want = append(want, string(row))
	}
	if g.fogged(len(fogReveal)-1) != strings.Join(want, "\n") {
		t.Error("the last stage should show the whole portrait")
	}
	if g.fogged(0) == g.fogged(len(fogReveal)-1) {
		t.Error("the first stage should be foggy")
	}
}

func TestRenderCrypt(t *testing.T) {
	p := store.New()
	p.Knowledge["dracula"] = 2
	out := RenderCrypt(p, monsters.GetAllMonsters(), 100)
	if !strings.Contains(out, "0 of ") || !strings.Contains(out, "●●○") {
		t.Errorf("unexpected crypt:\n%s", out)
	}
}

func TestRenderCountdownAndTicket(t *testing.T) {
	all := monsters.GetAllMonsters()
	for _, d := range []time.Time{
		time.Date(2026, 9, 23, 21, 0, 0, 0, time.Local),
		time.Date(2026, 10, 13, 21, 0, 0, 0, time.Local),
		time.Date(2026, 10, 31, 21, 0, 0, 0, time.Local),
	} {
		out := RenderCountdown(d, all, 80)
		switch {
		case d.Day() == 31 && !strings.Contains(out, "HAPPY HALLOWEEN"):
			t.Errorf("%s: missing Halloween banner", d)
		case d.Month() == time.October && d.Day() == 13 && !strings.Contains(out, "NIGHT 13"):
			t.Errorf("%s: missing Night of Fright", d)
		case d.Month() == time.September && !strings.Contains(out, "NIGHTS UNTIL HALLOWEEN"):
			t.Errorf("%s: missing countdown", d)
		}
	}
	ticket := RenderTicket(time.Date(2026, 9, 23, 21, 0, 0, 0, time.Local), all[0], all[1], 80)
	if !strings.Contains(ticket, "ADMIT ONE") || !strings.Contains(ticket, "DRACULA (1931)") {
		t.Errorf("unexpected ticket:\n%s", ticket)
	}
}

func TestBigNumber(t *testing.T) {
	if got := bigNumber(38); got != "▀▀█ █▄█\n▄▄█ █▄█" {
		t.Errorf("got %q", got)
	}
}
