package ui

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
)

// Intro timeline, in ticks.
const (
	staticFrames = 18 // TV static while the set warms up
	flashFrames  = 2  // a lightning flash
	typeSpeed    = 3  // characters of the greeting typed per tick
)

// A tiny two-row block font, just big enough for the title.
var glyphs = map[rune][2]string{
	'T': {"▀█▀", " █ "},
	'E': {"█▀▀", "██▄"},
	'R': {"█▀█", "█▀▄"},
	'M': {"█▀▄▀█", "█ ▀ █"},
	'I': {"█", "█"},
	'N': {"█▄ █", "█ ▀█"},
	'A': {"▄▀█", "█▀█"},
	'L': {"█  ", "█▄▄"},
	'O': {"█▀█", "█▄█"},
	'F': {"█▀▀", "█▀ "},
}

// bigText renders words in the block font.
func bigText(words ...string) string {
	var top, bottom []string
	for _, w := range words {
		var t, b []string
		for _, r := range w {
			g := glyphs[r]
			t, b = append(t, g[0]), append(b, g[1])
		}
		top, bottom = append(top, strings.Join(t, " ")), append(bottom, strings.Join(b, " "))
	}
	return strings.Join(top, "   ") + "\n" + strings.Join(bottom, "   ")
}

// logo is the big title, stacked onto two lines on narrow terminals.
func logo(width int) string {
	full := bigText("TERMINAL", "OF", "TERROR")
	if width == 0 || width >= artWidth(full)+4 {
		return full
	}
	if width >= artWidth(bigText("TERMINAL"))+4 {
		return lipgloss.JoinVertical(lipgloss.Center, bigText("TERMINAL"), "", bigText("OF", "TERROR"))
	}
	return "TERMINAL OF TERROR"
}

// noise draws a screenful of TV static.
func noise(r *rand.Rand, w, h int) string {
	chars := []rune("  ..::░░▒▓")
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b.WriteRune(chars[r.Intn(len(chars))])
		}
		if y < h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func introDoneFrame(greeting string) int {
	n := len([]rune(greeting))
	return staticFrames + flashFrames + (n+typeSpeed-1)/typeSpeed
}

func (m model) introDone() bool {
	return m.frame >= introDoneFrame(m.greeting)
}

func (m model) viewIntro() string {
	w := m.contentWidth()
	center := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	if m.frame < staticFrames {
		s := lipgloss.NewStyle().Foreground(colorDim).Render(m.noise)
		if m.frame > staticFrames/2 {
			label := titleStyle.Render("📺 " + host.Channel + " · CREATURE FEATURE")
			s = lipgloss.JoinVertical(lipgloss.Center, s, "", label)
		}
		return center.Render(s)
	}

	logoStyle := lipgloss.NewStyle().Foreground(colorBlood).Bold(true)
	if m.frame < staticFrames+flashFrames {
		// Lightning: the whole card flashes white for a moment.
		logoStyle = logoStyle.Foreground(colorBlack).Background(colorWhite)
	}

	typed := m.frame - staticFrames - flashFrames
	greeting := []rune(m.greeting)
	shown := min(len(greeting), max(typed, 0)*typeSpeed)
	text := string(greeting[:shown])
	if shown < len(greeting) {
		text += "▌"
	}

	parts := []string{
		"",
		helpStyle.Render("⚡  " + host.Channel + " presents  ⚡"),
		"",
		logoStyle.Render(logo(w)),
		"",
		headingStyle.Render("~ The Creature Feature ~"),
		"",
		hostStyle.Width(min(w-4, 70)).Render("📺 " + host.Name + ": " + text),
	}
	if m.introDone() {
		parts = append(parts, "", metaStyle.Render("Press any key to enter the vault..."))
	}
	return center.Render(lipgloss.JoinVertical(lipgloss.Center, parts...))
}
