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
		year := helpStyle.Render(fmt.Sprintf("%-5s", debutYear(m)))
		line := fmt.Sprintf("%s %s %s %s ", num, m.Emoji, name, year)
		desc := m.Description
		if width > 0 {
			desc = truncate(desc, width-lipgloss.Width(line))
		}
		b.WriteString(line + hostStyle.Render(desc) + "\n")
	}
	b.WriteString("\n" + metaStyle.Render(fmt.Sprintf("%d monsters lurk in the vault.", len(list))))
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
