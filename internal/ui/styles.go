package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// Base palette, shared by every screen.
var (
	colorBlood  = lipgloss.Color("#FF0000")
	colorGold   = lipgloss.Color("#FFD700")
	colorOrange = lipgloss.Color("#FFA500")
	colorSky    = lipgloss.Color("#87CEEB")
	colorMint   = lipgloss.Color("#90EE90")
	colorDim    = lipgloss.Color("#808080")
	colorBlack  = lipgloss.Color("#000000")
	colorWhite  = lipgloss.Color("#FFFFFF")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlood).
			Background(colorBlack).
			Padding(0, 2)

	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(colorOrange)
	factStyle    = lipgloss.NewStyle().Foreground(colorSky)
	metaStyle    = lipgloss.NewStyle().Foreground(colorMint).Italic(true)
	helpStyle    = lipgloss.NewStyle().Foreground(colorDim)
	hostStyle    = lipgloss.NewStyle().Foreground(colorDim).Italic(true)
	trueStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorMint)
	mythStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorBlood)
)

// palette is the set of colours used to draw one monster.
type palette struct {
	primary lipgloss.Color
	accent  lipgloss.Color
	body    lipgloss.Color
	silent  bool
}

// paletteFor picks a monster's colours. Silent-era monsters are always
// drawn in black and white, like the films they starred in.
func paletteFor(m monsters.Monster) palette {
	if m.IsSilent() {
		return palette{
			primary: lipgloss.Color("#F5F5F5"),
			accent:  lipgloss.Color("#BDBDBD"),
			body:    lipgloss.Color("#D6D6D6"),
			silent:  true,
		}
	}
	p := palette{primary: colorGold, accent: lipgloss.Color("#FF6347"), body: colorSky}
	if m.Theme.Primary != "" {
		p.primary = lipgloss.Color(m.Theme.Primary)
	}
	if m.Theme.Accent != "" {
		p.accent = lipgloss.Color(m.Theme.Accent)
	}
	return p
}

// intertitle frames text like a silent-film title card.
func intertitle(text string, width int, p palette, ornaments bool) string {
	inner := lipgloss.NewStyle().
		Width(max(width-6, 10)).
		Align(lipgloss.Center).
		Foreground(p.primary).
		Italic(true).
		Render(text)
	if ornaments {
		inner = "❦\n" + inner + "\n❦"
	}
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(p.accent).
		Padding(0, 1).
		Render(inner)
}

// wrap word-wraps text to width.
func wrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	return lipgloss.NewStyle().Width(width).Render(text)
}

// artWidth is the width of the widest line of ASCII art.
func artWidth(art string) int {
	w := 0
	for _, l := range strings.Split(art, "\n") {
		w = max(w, lipgloss.Width(l))
	}
	return w
}

// Tilde shortens a path under the home directory to ~/..., the way people
// usually write it.
func Tilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || home == "/" {
		return path
	}
	if path == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(path, home+string(os.PathSeparator)); ok {
		return "~" + string(os.PathSeparator) + rest
	}
	return path
}

// Dim renders text in the muted help colour.
func Dim(s string) string { return helpStyle.Render(s) }
