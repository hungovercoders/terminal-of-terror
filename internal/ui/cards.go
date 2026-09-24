package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// TerminalWidth is stdout's width, or 0 when it isn't a terminal.
func TerminalWidth() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 0
	}
	return w
}

// RenderList draws every monster, grouped by pack.
func RenderList(list []monsters.Monster, width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🎃 TERMINAL OF TERROR · The Vault 🎃") + "\n")

	nameW := 0
	for _, m := range list {
		nameW = max(nameW, lipgloss.Width(m.Name))
	}
	pack := ""
	for i, m := range list {
		if m.Pack != pack {
			pack = m.Pack
			count := 0
			for _, o := range list {
				if o.Pack == pack {
					count++
				}
			}
			b.WriteString("\n" + headingStyle.Render(fmt.Sprintf("── %s (%d) ──", packName(pack), count)) + "\n")
		}
		p := paletteFor(m)
		num := helpStyle.Render(fmt.Sprintf("%3d.", i+1))
		name := lipgloss.NewStyle().Foreground(p.primary).Bold(true).Render(fmt.Sprintf("%-*s", nameW, m.Name))
		year := helpStyle.Render(fmt.Sprintf("%-5s", shortYear(m)))
		line := fmt.Sprintf("%s %s %s %s ", num, m.Emoji, name, year)
		desc := m.Description
		if width > 0 {
			desc = truncate(desc, width-lipgloss.Width(line))
		}
		b.WriteString(line + hostStyle.Render(desc) + "\n")
	}
	verb := " lurk"
	if len(list) == 1 {
		verb = " lurks"
	}
	b.WriteString("\n" + metaStyle.Render(countOf(len(list), "monster")+verb+" in the vault."))
	b.WriteString("\n" + helpStyle.Render("Meet one: terminal-of-terror monster <name>") + "\n")
	return b.String()
}

// FactCard frames a single fact about a monster, with a word from the host.
type FactCard struct {
	Title   string // e.g. "RANDOM TERROR FACT"
	Monster monsters.Monster
	Fact    string
	Host    string   // host's comment; empty for none
	Extras  []string // extra lines shown under the fact (calendar notes etc.)
}

// Render draws the card at most width columns wide (0 means 64).
func (c FactCard) Render(width int) string {
	w := 64
	if width > 0 {
		w = min(width-2, 64)
	}
	inner := max(w-4, 20)
	p := paletteFor(c.Monster)

	parts := []string{
		headingStyle.Render("🎃 " + c.Title),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(c.Monster.Emoji + " " + c.Monster.Name),
		lipgloss.NewStyle().Foreground(p.body).Render(wrap(c.Fact, inner)),
	}
	for _, e := range c.Extras {
		parts = append(parts, "", metaStyle.Render(wrap(e, inner)))
	}
	if c.Host != "" {
		parts = append(parts, "", hostStyle.Render(wrap("📺 "+host.Name+": “"+c.Host+"”", inner)))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.accent).
		Padding(0, 1).
		Render(strings.Join(parts, "\n")) + "\n"
}

// RenderPacks lists every pack and where community packs live.
func RenderPacks(packs []monsters.Pack, communityDir string, width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📦 MONSTER PACKS") + "\n")
	for _, p := range packs {
		b.WriteString("\n" + headingStyle.Render(fmt.Sprintf("%s (%s) · %s", p.Name, p.ID, countOf(len(p.Monsters), "monster"))) + "\n")
		if p.Description != "" {
			b.WriteString(hostStyle.Render(wrap(p.Description, textWidth(width))) + "\n")
		}
		var names []string
		for _, m := range p.Monsters {
			names = append(names, m.Emoji+" "+m.Name)
		}
		b.WriteString(factStyle.Render(joinFit(names, " · ", textWidth(width))) + "\n")
		if p.Source != "built-in" && p.Source != "" {
			b.WriteString(helpStyle.Render("from "+Tilde(p.Source)) + "\n")
		}
	}
	b.WriteString("\n" + metaStyle.Render("Use --pack to choose, e.g. terminal-of-terror quiz --pack folklore") + "\n")
	if communityDir != "" {
		b.WriteString(helpStyle.Render("Community packs live in "+Tilde(communityDir)) + "\n")
	}
	b.WriteString(helpStyle.Render("Make your own: terminal-of-terror packs new <pack-id>") + "\n")
	return b.String()
}

func countOf(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
