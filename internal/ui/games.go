package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/mash"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
)

// RenderCrypt shows the player's collection, badges and records.
func RenderCrypt(p *store.Progress, all []monsters.Monster, width int) string {
	var b strings.Builder
	captured := 0
	for _, m := range all {
		if _, ok := p.Captured[m.ID]; ok {
			captured++
		}
	}
	b.WriteString(titleStyle.Render("💀 THE CRYPT") + "\n\n")
	b.WriteString(headingStyle.Render(fmt.Sprintf("%d of %d monsters captured · %d of %d badges", captured, len(all), len(p.Badges), len(crypt.Badges))) + "\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("Answer %d questions about a monster correctly, in quizzes or Guess the Monster, to capture it.", crypt.CaptureAt)) + "\n")

	pack := ""
	for _, m := range all {
		if m.Pack != pack {
			pack = m.Pack
			b.WriteString("\n" + headingStyle.Render("── "+packName(pack)+" ──") + "\n")
		}
		if at, ok := p.Captured[m.ID]; ok {
			name := lipgloss.NewStyle().Bold(true).Foreground(paletteFor(m).primary).Render(m.Name)
			b.WriteString(fmt.Sprintf("  %s %s  %s\n", m.Emoji, name, trueStyle.Render("✔ captured "+at.Format("2 Jan 2006"))))
			continue
		}
		k := min(p.Knowledge[m.ID], crypt.CaptureAt)
		pips := strings.Repeat("●", k) + strings.Repeat("○", crypt.CaptureAt-k)
		b.WriteString(fmt.Sprintf("  🔒 %s  %s\n", helpStyle.Render(m.Name), metaStyle.Render(pips)))
	}

	b.WriteString("\n" + headingStyle.Render("── Badges ──") + "\n")
	for _, bd := range crypt.Badges {
		if at, ok := crypt.Earned(p, bd.ID); ok {
			b.WriteString(fmt.Sprintf("  %s %s  %s\n", bd.Emoji, lipgloss.NewStyle().Bold(true).Foreground(colorGold).Render(bd.Name), hostStyle.Render(bd.Description+" · "+at.Format("2 Jan 2006"))))
		} else {
			b.WriteString(fmt.Sprintf("  🔒 %s  %s\n", helpStyle.Render(bd.Name), helpStyle.Render(bd.Description)))
		}
	}

	b.WriteString("\n" + headingStyle.Render("── Records ──") + "\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("  Quizzes played: %d · best score %d%%", p.QuizzesPlayed, p.BestQuizScore)) + "\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("  Guess the Monster games: %d · best score %d pts", p.GuessesPlayed, p.BestGuessScore)) + "\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("  Monster Mashes staged: %d", p.MashesPlayed)) + "\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("  Monster pages visited: %d of %d", countSeen(p, all), len(all))) + "\n")
	return b.String()
}

func countSeen(p *store.Progress, all []monsters.Monster) int {
	n := 0
	for _, m := range all {
		if p.Seen[m.ID] {
			n++
		}
	}
	return n
}

// RenderUnlocks announces newly captured monsters and badges.
func RenderUnlocks(u crypt.Unlocks) string {
	if u.Empty() {
		return ""
	}
	var b strings.Builder
	for _, m := range u.Captured {
		b.WriteString(trueStyle.Render("💀 CAPTURED: ") + lipgloss.NewStyle().Bold(true).Foreground(paletteFor(m).primary).Render(m.Emoji+" "+m.Name) + helpStyle.Render(" has joined your crypt") + "\n")
	}
	for _, bd := range u.Badges {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGold).Render("🏅 BADGE: "+bd.Emoji+" "+bd.Name) + helpStyle.Render(" · "+bd.Description) + "\n")
	}
	b.WriteString(helpStyle.Render("See your collection: terminal-of-terror crypt") + "\n")
	return b.String()
}

// StatCard is a compact trading-card view of a monster.
func StatCard(m monsters.Monster) string {
	p := paletteFor(m)
	s := m.Stats
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(m.Emoji + " " + m.Name),
		"",
	}
	for _, st := range []struct {
		name string
		v    int
	}{{"Strength", s.Strength}, {"Speed", s.Speed}, {"Cunning", s.Cunning}, {"Dread", s.Dread}} {
		lines = append(lines, metaStyle.Render(fmt.Sprintf("%-9s", st.name))+statBar(st.v, p))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.accent).
		Padding(0, 1).
		Width(34).
		Render(strings.Join(lines, "\n"))
}

// MashHeader shows the two fighters' cards side by side.
func MashHeader(a, b monsters.Monster, width int) string {
	title := titleStyle.Render("🥊 MONSTER MASH")
	ca, cb := StatCard(a), StatCard(b)
	vs := lipgloss.NewStyle().Bold(true).Foreground(colorBlood).Padding(2, 2).Render("VS")
	var cards string
	if width == 0 || lipgloss.Width(ca)*2+lipgloss.Width(vs) <= width {
		cards = lipgloss.JoinHorizontal(lipgloss.Center, ca, vs, cb)
	} else {
		cards = lipgloss.JoinVertical(lipgloss.Center, ca, vs, cb)
	}
	intro := hostStyle.Render(wrap(fmt.Sprintf("📺 %s: “In this corner, %s! And in the other, %s! Three rounds, no rules, no refunds.”", host.Name, mash.MidSentence(a.Name), mash.MidSentence(b.Name)), textWidth(width)))
	return title + "\n\n" + cards + "\n\n" + intro + "\n"
}

// MashRound narrates one round.
func MashRound(bout mash.Bout, r mash.Round, width int) string {
	w := textWidth(width)
	winner := bout.A
	if r.Winner == 1 {
		winner = bout.B
	}
	head := headingStyle.Render(fmt.Sprintf("Round %d · %s", r.Number, r.Stat)) +
		helpStyle.Render(fmt.Sprintf("   %s rolls %d · %s rolls %d", bout.A.Name, r.RollA, bout.B.Name, r.RollB))
	return head + "\n" +
		factStyle.Render(wrap(r.Narration, w)) + "\n" +
		lipgloss.NewStyle().Italic(true).Foreground(paletteFor(winner).primary).Render(wrap(r.Move, w)) + "\n"
}

// MashResult announces the winner.
func MashResult(bout mash.Bout, width int) string {
	w := textWidth(width)
	wa, wb := bout.Tally()
	winner := bout.WinnerMonster()
	score := fmt.Sprintf("%d-%d", max(wa, wb), min(wa, wb))
	banner := lipgloss.NewStyle().Bold(true).Foreground(colorBlack).Background(colorGold).Padding(0, 1).
		Render(fmt.Sprintf("🏆 WINNER: %s %s (%s)", winner.Emoji, strings.ToUpper(winner.Name), score))
	return banner + "\n" + hostStyle.Render(wrap(bout.Finale, w)) + "\n"
}

// textWidth caps prose at 80 columns, assuming 80 when the width is unknown.
func textWidth(width int) int {
	if width <= 0 {
		return 80
	}
	return max(min(width, 80), 30)
}
