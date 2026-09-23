package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

var digits = map[rune][2]string{
	'0': {"█▀█", "█▄█"},
	'1': {"▄█ ", " █ "},
	'2': {"▀▀█", "█▄▄"},
	'3': {"▀▀█", "▄▄█"},
	'4': {"█ █", "▀▀█"},
	'5': {"█▀▀", "▄▄█"},
	'6': {"█▄▄", "█▄█"},
	'7': {"▀▀█", "  █"},
	'8': {"█▄█", "█▄█"},
	'9': {"█▀█", "▀▀█"},
}

// bigNumber renders n in the block font.
func bigNumber(n int) string {
	var top, bottom []string
	for _, r := range fmt.Sprint(n) {
		top, bottom = append(top, digits[r][0]), append(bottom, digits[r][1])
	}
	return strings.Join(top, " ") + "\n" + strings.Join(bottom, " ")
}

const pumpkin = `         )
    .-"""(""-.
  .'  ^  |  ^  '.
 /   /_\ | /_\   \
|        ^        |
|  \/\/\/\/\/\/\  |
 \  \/\/\/\/\/\/ /
  '.           .'
    '-._____.-'`

// RenderCountdown shows the nights until Halloween and what's coming up.
func RenderCountdown(now time.Time, all []monsters.Monster, width int) string {
	w := textWidth(width)
	orange := lipgloss.NewStyle().Foreground(colorOrange).Bold(true)
	var parts []string

	if calendar.IsHalloween(now) {
		parts = append(parts,
			orange.Render(pumpkin),
			"",
			titleStyle.Render("🎃 HAPPY HALLOWEEN! 🎃"),
			"",
			hostStyle.Render(wrap("📺 "+host.Name+": “It's the big night, fiends! Every monster in the vault is out tonight. Lock the doors, light the lantern and don't answer the door to anyone in a cape.”", w)),
		)
	} else {
		n := calendar.NightsUntilHalloween(now)
		label := "NIGHTS UNTIL HALLOWEEN"
		if n == 1 {
			label = "NIGHT UNTIL HALLOWEEN"
		}
		parts = append(parts,
			lipgloss.JoinHorizontal(lipgloss.Center, orange.Render(pumpkin), "    ",
				lipgloss.JoinVertical(lipgloss.Left, orange.Render(bigNumber(n)), "", headingStyle.Render(label))),
			"",
		)
		if night, ok := calendar.NightOfFright(now); ok {
			m := all[(night-1)%len(all)]
			fact := m.Facts[(night-1)/len(all)%len(m.Facts)]
			parts = append(parts,
				titleStyle.Render(fmt.Sprintf("31 NIGHTS OF FRIGHT · NIGHT %d", night)),
				"",
				lipgloss.NewStyle().Bold(true).Foreground(paletteFor(m).primary).Render("Tonight's monster: "+m.Emoji+" "+m.Name),
				factStyle.Render(wrap(fact, w)),
			)
		} else {
			parts = append(parts, hostStyle.Render(wrap(fmt.Sprintf("📺 %s: “The 31 Nights of Fright begin on 1 October. A different monster every night until Halloween. Don't miss it!”", host.Name), w)))
		}
	}

	moon := calendar.Moon(now)
	parts = append(parts, "", metaStyle.Render(fmt.Sprintf("%s Tonight's moon: %s (%.0f%% lit)", moon.Emoji, moon.Name, moon.Illumination*100)))
	if a, when, ok := calendar.NextAnniversary(now, all); ok {
		parts = append(parts, metaStyle.Render(fmt.Sprintf("🎬 Next anniversary: %s turns %d on %s", a.Monster.Film.Title, a.Years, when.Format("2 January"))))
	}
	return strings.Join(parts, "\n") + "\n"
}

// RenderTicket draws tonight's double-feature ticket.
func RenderTicket(now time.Time, first, second monsters.Monster, width int) string {
	inner := min(textWidth(width)-4, 64)
	gold := lipgloss.NewStyle().Foreground(colorGold).Bold(true)
	moon := calendar.Moon(now)

	feature := func(at string, m monsters.Monster) string {
		p := paletteFor(m)
		title := strings.ToUpper(m.Name)
		credit := "A tale from " + m.Origin
		if f := m.Film; f != nil {
			title = fmt.Sprintf("%s (%d)", strings.ToUpper(f.Title), f.Year)
			credit = fmt.Sprintf("Directed by %s · starring %s", f.Director, f.Star)
			if f.Silent {
				credit += " · silent"
			}
		}
		timeCol := lipgloss.NewStyle().Width(10).Foreground(colorOrange).Bold(true).Render(at)
		text := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(m.Emoji+" "+title),
			helpStyle.Render(wrap(credit, inner-12)),
		)
		return lipgloss.JoinHorizontal(lipgloss.Top, timeCol, text)
	}

	perforation := helpStyle.Render(strings.Repeat("- ", inner/2))
	body := strings.Join([]string{
		gold.Render("📺 " + strings.ToUpper(host.Channel) + " PRESENTS"),
		titleStyle.Render("TONIGHT'S DOUBLE FEATURE"),
		metaStyle.Render(fmt.Sprintf("%s · %s %s", now.Format("Monday 2 January 2006"), moon.Emoji, moon.Name)),
		perforation,
		feature("9:00 PM", first),
		"",
		feature("10:45 PM", second),
		perforation,
		gold.Render("ADMIT ONE") + helpStyle.Render("  ·  No refunds if you're frightened to death"),
	}, "\n")

	ticket := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(colorOrange).
		Padding(0, 2).
		Width(inner + 4).
		Render(body)
	return ticket + "\n"
}
