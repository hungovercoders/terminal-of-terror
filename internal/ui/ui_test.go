package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

func send(m model, msgs ...tea.Msg) model {
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(model)
	}
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func size(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }

// TestEveryPageRenders visits every tab of every monster at several sizes.
func TestEveryPageRenders(t *testing.T) {
	for _, sz := range [][2]int{{40, 20}, {80, 24}, {140, 45}, {0, 0}} {
		m := send(newModel(Options{NoIntro: true, Seed: 1}), size(sz[0], sz[1]))
		for range monsters.GetAllMonsters() {
			mo := m.current()
			for range tabsFor(mo) {
				v := m.View()
				if !strings.Contains(v, strings.ToUpper(mo.Name)) {
					t.Fatalf("%dx%d %s tab %d: name missing", sz[0], sz[1], mo.ID, m.tab)
				}
				if sz[1] > 0 && strings.Count(v, "\n")+1 > sz[1] {
					t.Errorf("%dx%d %s tab %d: %d lines overflow the screen", sz[0], sz[1], mo.ID, m.tab, strings.Count(v, "\n")+1)
				}
				m = send(m, key("r"), key("end"), key("tab"))
			}
			m = send(m, key("right"))
		}
	}
}

func TestIntroSkipsThenEnters(t *testing.T) {
	m := send(newModel(Options{Seed: 1}), size(100, 30))
	if m.screen != screenIntro {
		t.Fatal("expected intro")
	}
	m = send(m, key("x"))
	if !m.introDone() || !strings.Contains(m.View(), "Press any key") {
		t.Fatal("first key should finish the intro")
	}
	m = send(m, key("x"))
	if m.screen != screenDetail {
		t.Fatalf("second key should enter the explorer, got %v", m.screen)
	}
}

func TestIntroShowsTonightsNotes(t *testing.T) {
	halloween := time.Date(2026, 10, 31, 21, 0, 0, 0, time.Local)
	m := send(newModel(Options{Seed: 1, Now: halloween}), size(100, 40), key("x"))
	v := m.View()
	if !strings.Contains(v, "Happy Halloween") || !strings.Contains(v, "It's Halloween!") {
		t.Errorf("Halloween intro missing its greeting or note:\n%s", v)
	}
}

func TestStartIDSkipsIntro(t *testing.T) {
	m := newModel(Options{StartID: "mummy", Seed: 1})
	if m.screen != screenDetail || m.current().ID != "mummy" {
		t.Fatalf("expected mummy detail page, got screen %v %s", m.screen, m.current().ID)
	}
}

func TestSearchOpensMonster(t *testing.T) {
	m := send(newModel(Options{NoIntro: true, Seed: 1}), size(80, 24), key("/"))
	for _, r := range "lugosi" {
		m = send(m, key(string(r)))
	}
	if len(m.hits) == 0 {
		t.Fatal("expected hits for lugosi")
	}
	m = send(m, key("down"), key("enter"))
	if m.searching || m.screen != screenDetail {
		t.Fatal("enter should open the chosen monster")
	}
}

func TestGalleryNavigation(t *testing.T) {
	m := send(newModel(Options{NoIntro: true, ShowAll: true, Seed: 1}), size(120, 30))
	if m.screen != screenGallery {
		t.Fatal("expected gallery")
	}
	m = send(m, key("j"), key("j"), key("enter"))
	if m.screen != screenDetail || m.index != 2 {
		t.Fatalf("expected third monster, got %d", m.index)
	}
	m = send(m, key("esc"))
	if m.screen != screenGallery || m.cursor != 2 {
		t.Fatal("esc should return to gallery with cursor kept")
	}
}

func TestMythVerdictsHiddenUntilRevealed(t *testing.T) {
	m := send(newModel(Options{StartID: "wolf-man", Seed: 1}), size(80, 60))
	for i, tb := range tabsFor(m.current()) {
		if tb == tabMyths {
			m = send(m, key(string(rune('1'+i))))
		}
	}
	if strings.Contains(m.View(), "✘ MYTH") {
		t.Fatal("verdicts should start hidden")
	}
	m = send(m, key("r"))
	if !strings.Contains(m.View(), "✘ MYTH") {
		t.Fatal("r should reveal verdicts")
	}
}

// TestDumpScreens writes sample screens for eyeballing: DUMP=1 go test ./internal/ui -run Dump -v
func TestDumpScreens(t *testing.T) {
	if os.Getenv("DUMP") == "" {
		t.Skip("set DUMP=1 to print screens")
	}
	w, h := 120, 40
	m := send(newModel(Options{Seed: 3}), size(w, h))
	for i := 0; i < 12; i++ {
		m = send(m, tickMsg{})
	}
	t.Log("\n" + m.View())
	m = send(m, key("x"))
	t.Log("\n" + m.View())
	m = send(m, key("x"))
	t.Log("\n" + m.View())
	m = send(m, key("tab"), key("tab"))
	t.Log("\n" + m.View())
	m = send(m, key("g"))
	t.Log("\n" + m.View())
	m2 := send(newModel(Options{StartID: "phantom", Seed: 3}), size(80, 24), tickMsg{})
	t.Log("\n" + m2.View())
	m2 = send(m2, key("4"), key("r"))
	t.Log("\n" + m2.View())
}
