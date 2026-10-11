package ui

import (
	"strings"
	"testing"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// Plain output has no colour codes, no emoji and ASCII borders; switching
// it off brings them back.
func TestPlainOutputHasNoColourOrEmoji(t *testing.T) {
	SetPlain(true)
	defer SetPlain(false)
	all := monsters.GetAllMonsters()
	list := RenderList(all, 0)
	card := FactCard{Title: "RANDOM TERROR FACT", Monster: all[0], Fact: "A fact.", Host: "Chilling.", Extras: []string{"🎃 It's Halloween!"}}.Render(0)
	for name, out := range map[string]string{"list": list, "card": card} {
		if strings.Contains(out, "\x1b[") {
			t.Errorf("%s has colour codes when plain:\n%s", name, out)
		}
		for _, r := range out {
			if r >= 0x1F000 {
				t.Errorf("%s has an emoji (%q) when plain:\n%s", name, r, out)
				break
			}
		}
	}
	if !strings.Contains(list, "TERMINAL OF TERROR · The Vault") || !strings.Contains(list, all[0].Name) {
		t.Errorf("plain list lost its title or names:\n%s", list)
	}
	if !strings.Contains(card, "+") || !strings.Contains(card, "|") || strings.Contains(card, "╭") {
		t.Errorf("plain card should have an ASCII border:\n%s", card)
	}
	if !strings.Contains(card, "It's Halloween!") || !strings.Contains(card, "Count Cathode: “Chilling.”") {
		t.Errorf("plain card should keep the host's line, without the glyph:\n%s", card)
	}
	SetPlain(false)
	if fancy := RenderList(all, 0); !strings.Contains(fancy, "🎃") || !strings.Contains(fancy, all[0].Emoji) {
		t.Errorf("fancy list lost its emoji:\n%s", fancy)
	}
}
