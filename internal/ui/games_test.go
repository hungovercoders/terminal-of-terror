package ui

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	if got := bigNumber(38); got != "▀▀█ █▀█\n ▀█ █▀█\n▄▄█ █▄█" {
		t.Errorf("got %q", got)
	}
}

func TestGuessNeedsEnoughMonsters(t *testing.T) {
	all := monsters.GetAllMonsters()
	r := rand.New(rand.NewSource(1))
	if got := NewGuessRounds(r, all[:2], 3); got != nil {
		t.Errorf("two monsters can't make a fair guess, got %d rounds", len(got))
	}
	// One portrait is enough when other monsters can supply wrong answers.
	few := []monsters.Monster{all[0], all[1], all[2]}
	few[1].ASCII, few[2].ASCII = "", ""
	rounds := NewGuessRounds(r, few, 3)
	if len(rounds) != 1 || len(rounds[0].Options) != 3 {
		t.Fatalf("want 1 round with 3 options, got %+v", rounds)
	}
	for _, g := range NewGuessRounds(r, all, 5) {
		if len(g.Options) != 4 {
			t.Errorf("full roster should give 4 options, got %d", len(g.Options))
		}
	}
}

func TestTilde(t *testing.T) {
	p := filepath.FromSlash
	cases := map[string]string{
		p("/home/ghoul/.config/terminal-of-terror/progress.json"): p("~/.config/terminal-of-terror/progress.json"),
		p("/home/ghoul"):         "~",
		p("/home/ghoul/"):        "~",
		p("/home/ghoulish/x"):    p("/home/ghoulish/x"),
		p("/tmp/somewhere/else"): p("/tmp/somewhere/else"),
	}
	// A trailing separator on the home directory must not stop the match.
	for _, home := range []string{p("/home/ghoul"), p("/home/ghoul/")} {
		// os.UserHomeDir reads HOME on Unix and USERPROFILE on Windows.
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		for in, want := range cases {
			if runtime.GOOS == "windows" {
				want = in // cmd.exe can't expand ~, so Windows paths stay whole
			}
			if got := Tilde(in); got != want {
				t.Errorf("home %q: Tilde(%q) = %q, want %q", home, in, got, want)
			}
		}
	}
}

// TestGuessLayoutFits plays every monster at several widths: no line may be
// wider than the terminal, and clearing the fog must never move the side
// panel (the layout is decided once per round).
func TestGuessLayoutFits(t *testing.T) {
	all := monsters.GetAllMonsters()
	for _, width := range []int{40, 50, 60, 70, 80, 90, 100} {
		for seed := int64(1); seed <= 5; seed++ {
			r := rand.New(rand.NewSource(seed))
			rounds := NewGuessRounds(r, all, len(all))
			if len(rounds) == 0 {
				t.Fatal("no rounds to play")
			}
			var m tea.Model = newGuessModel(rounds, r)
			m, _ = m.Update(size(width, 60))
			for range rounds {
				g := m.(guessModel).res.Rounds[m.(guessModel).i]
				check := func(when string) {
					for _, line := range strings.Split(m.View(), "\n") {
						if lw := lipgloss.Width(line); lw > width {
							t.Fatalf("width %d, %s, %s: line is %d wide: %q", width, g.Monster.ID, when, lw, line)
						}
					}
				}
				panelAt := panelPosition(m.View())
				for stage := 0; stage < len(fogReveal); stage++ {
					check(fmt.Sprintf("stage %d", stage))
					if got := panelPosition(m.View()); got != panelAt {
						t.Fatalf("width %d, %s: side panel moved from %v to %v at stage %d", width, g.Monster.ID, panelAt, got, stage)
					}
					m = step(m, " ")
				}
				m = step(m, "1")
				check("answered")
				m = step(m, "enter")
			}
			check := m.View()
			for _, line := range strings.Split(check, "\n") {
				if lipgloss.Width(line) > width {
					t.Fatalf("width %d: final screen line too wide: %q", width, line)
				}
			}
		}
	}
}

// panelPosition is the line and on-screen column where the side panel starts.
func panelPosition(view string) [2]int {
	for i, line := range strings.Split(view, "\n") {
		if j := strings.Index(line, "Who lurks in the fog?"); j >= 0 {
			return [2]int{i, lipgloss.Width(line[:j])}
		}
	}
	return [2]int{-1, -1}
}

func TestFogStatusAfterGuessing(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	var m tea.Model = newGuessModel(NewGuessRounds(r, monsters.GetAllMonsters(), 1), r)
	m, _ = m.Update(size(100, 40))
	m = step(m, "1")
	if v := m.View(); !strings.Contains(v, "You guessed with 15% of the fog cleared") || strings.Contains(v, "worth 100 pts") {
		t.Errorf("after guessing through the thickest fog the panel should say so:\n%s", v)
	}
}

func TestGuessCluesReadNaturally(t *testing.T) {
	for _, m := range monsters.GetAllMonsters() {
		g := GuessRound{Monster: m}
		clue := g.clues(2)[0]
		if m.Debut.Year == 0 && strings.HasPrefix(clue, "First appeared in") {
			t.Errorf("%s: folklore clue reads %q", m.ID, clue)
		}
		if m.Debut.Year != 0 && clue != fmt.Sprintf("First appeared in %d", m.Debut.Year) {
			t.Errorf("%s: clue %q", m.ID, clue)
		}
	}
}

func TestWrongAnswerHasOneFullStop(t *testing.T) {
	q := quiz.Question{Prompt: "Who played the Wolf Man?", Options: []string{"Boris Karloff", "Lon Chaney Jr."}, Answer: 1}
	var m tea.Model = newQuizModel([]quiz.Question{q}, rand.New(rand.NewSource(1)))
	m, _ = m.Update(size(80, 24))
	m = step(m, "1")
	if v := m.View(); !strings.Contains(v, "The answer is Lon Chaney Jr.") || strings.Contains(v, "Jr..") {
		t.Errorf("unexpected wrong-answer line:\n%s", v)
	}
}
