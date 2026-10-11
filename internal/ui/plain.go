package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// plain is set when the output is for a pipe, a log or someone who asked
// for no colour: no colour codes, ASCII borders, and none of the emoji
// that dress the titles and the host's lines. The monsters' own emoji go
// too, so `list | grep` sees names. The Bubble Tea screens are untouched:
// they need a terminal, and a terminal can show an emoji.
var plain bool

// colourProfile is what lipgloss was using before plain switched colour
// off, so switching it back on in a test restores it.
var colourProfile termenv.Profile

// SetPlain switches plain output on or off. The root command decides once
// at startup from --no-color, NO_COLOR and whether stdout is a terminal.
func SetPlain(on bool) {
	if on == plain {
		return
	}
	plain = on
	if on {
		colourProfile = lipgloss.ColorProfile()
		lipgloss.SetColorProfile(termenv.Ascii)
		return
	}
	lipgloss.SetColorProfile(colourProfile)
}

// Plain reports whether output is plain.
func Plain() bool { return plain }

// glyph is a decorative emoji and the space after it, or nothing when
// plain: glyph("🎃") + "TITLE".
func glyph(e string) string {
	if plain {
		return ""
	}
	return e + " "
}

// label is an emoji and a name, "🦇 Dracula", or the name alone when plain
// or there is no emoji.
func label(emoji, name string) string {
	if plain || emoji == "" {
		return name
	}
	return emoji + " " + name
}

// pick is fancy, or plainText when plain, for a title that carries emoji
// at both ends.
func pick(fancy, plainText string) string {
	if plain {
		return plainText
	}
	return fancy
}

// unglyph drops a leading emoji (and the space after it) from a line of
// text when plain: the calendar's notes arrive as "🎃 It's Halloween!".
func unglyph(s string) string {
	if !plain {
		return s
	}
	for _, r := range s {
		if r < 0x1F000 {
			return s
		}
		return strings.TrimLeft(s[len(string(r)):], " ")
	}
	return s
}

// border is b, or the ASCII border when plain, so a card in a log is
// still a card.
func border(b lipgloss.Border) lipgloss.Border {
	if plain {
		return lipgloss.ASCIIBorder()
	}
	return b
}
