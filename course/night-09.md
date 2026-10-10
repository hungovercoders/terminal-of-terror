# Night 9 · A Splash of Blood

> 📺 *"The great horror films of the thirties were black and white, and that's fine for them. We have a terminal that can show sixteen million colours, and I intend to use at least four. Tonight, style."*

**Tonight you'll learn**
- How terminals do colour, and what Lip Gloss does about it
- Styles: colour, bold, italic, padding, and `Render`
- A `ui` package that owns the look
- Measuring the terminal, and measuring strings correctly
- Truncating text to fit

**Where we are:** `list` prints a plain table; `monster <name>` prints a plain page.

## How terminals do colour

A terminal is a stream of characters, and colour is done with *escape sequences*: `\x1b[31m` switches to red, `\x1b[0m` switches back. Every terminal since the 1970s speaks some version of this. You could write those sequences yourself, and for the next 22 nights you'd regret it.

[Lip Gloss](https://github.com/charmbracelet/lipgloss) is a library for describing how text should look, then rendering it. It knows how many colours the terminal supports, degrades gracefully, and, crucially, drops the colours entirely when output isn't a terminal at all, so `list > monsters.txt` gives you clean text.

```bash
go get github.com/charmbracelet/lipgloss@v1.1.0 github.com/charmbracelet/x/term@v0.2.1
```

The second package measures the terminal.

## The `ui` package

Everything about how things look goes in one place. Create `internal/ui/styles.go`:

```go
// Package ui draws everything the user sees.
package ui

import "github.com/charmbracelet/lipgloss"

// Base palette, shared by every screen.
var (
	colorBlood  = lipgloss.Color("#FF0000")
	colorGold   = lipgloss.Color("#FFD700")
	colorOrange = lipgloss.Color("#FFA500")
	colorSky    = lipgloss.Color("#87CEEB")
	colorMint   = lipgloss.Color("#90EE90")
	colorDim    = lipgloss.Color("#808080")
	colorBlack  = lipgloss.Color("#000000")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlood).
			Background(colorBlack).
			Padding(0, 2)

	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(colorOrange)
	nameStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorGold)
	factStyle    = lipgloss.NewStyle().Foreground(colorSky)
	metaStyle    = lipgloss.NewStyle().Foreground(colorMint).Italic(true)
	helpStyle    = lipgloss.NewStyle().Foreground(colorDim)
)

// truncate shortens s to fit in w columns, ending with an ellipsis.
func truncate(s string, w int) string {
	if w <= 1 || lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
```

A **style** is built by chaining methods: `NewStyle().Bold(true).Foreground(...)`. Each call returns a new style, so the chain reads like a sentence. `Padding(0, 2)` is CSS-style: vertical then horizontal. Nothing is drawn until you call `style.Render(text)`.

Colours are hex strings, like the web. Terminals that can't show them get the nearest of their 256 or 16 colours.

Naming the colours (`colorBlood`) and then the styles (`titleStyle`) is deliberate. Screens will use *styles*, never raw colours, so changing the look of every heading is a one-line edit.

### Measuring strings

`truncate` uses `lipgloss.Width`, not `len`. Width counts terminal *columns*: an emoji is two, an accent is zero, and escape sequences are nothing at all. `len` counts bytes and would be wrong three ways at once. Whenever you're fitting text on a screen, measure with `lipgloss.Width`.

## The list, in colour

Create `internal/ui/cards.go`:

```go
package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
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
func RenderList(packs []monsters.Pack, width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🎃 TERMINAL OF TERROR · The Vault 🎃") + "\n")

	n := 0
	for _, p := range packs {
		b.WriteString("\n" + headingStyle.Render(fmt.Sprintf("── %s (%d) ──", p.Name, len(p.Monsters))) + "\n")
		for _, m := range p.Monsters {
			n++
			line := fmt.Sprintf("%s %s ", helpStyle.Render(fmt.Sprintf("%3d.", n)), nameStyle.Render(fmt.Sprintf("%-24s", m.Name)))
			desc := m.Description
			if width > 0 {
				desc = truncate(desc, width-lipgloss.Width(line))
			}
			b.WriteString(line + metaStyle.Render(desc) + "\n")
		}
	}
	b.WriteString("\n" + metaStyle.Render(fmt.Sprintf("%d monsters lurk in the vault.", n)))
	b.WriteString("\n" + helpStyle.Render("Meet one: terminal-of-terror monster <name>") + "\n")
	return b.String()
}
```

`RenderList` *returns* a string rather than printing. That's a habit worth forming now: functions that build text are easy to test (compare the string) and easy to reuse (the same text can go to a screen, a file or a test). The command does the printing.

`term.GetSize` asks the terminal how wide it is. When output is a pipe or a file, there's no terminal and it returns an error; we turn that into 0, and `RenderList` treats 0 as "don't truncate".

`%-*s` is a new verb: the `*` takes the width from the argument list, so `fmt.Sprintf("%-*s", 24, m.Name)` pads to 24. Here the width is a constant, but the repo's version measures the longest name first.

## Wire it up

Replace `cmd/list.go`:

```go
package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List every monster in the vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(ui.RenderList(monsters.GetPacks(), ui.TerminalWidth()))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
```

## Run it

```bash
go mod tidy
go run . list
```

A red-on-black title, an orange pack heading, gold names, green descriptions. Now narrow your terminal to 60 columns and run it again: the descriptions end in `…` instead of wrapping. Then:

```bash
go run . list | cat
```

Plain text, no colour codes. Lip Gloss noticed the pipe.

## Try it

- Change `colorGold` to a nicer gold and watch every name change.
- Add `Underline(true)` to `headingStyle`.
- Wrap the whole list in a border: `lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(text)`. Tomorrow's fact card is built this way.

## 💀 Terrifying fact

`lipgloss.Width("🎃")` is 2, but some terminals draw certain emoji one column wide, especially ones that need an invisible "variation selector" character to look like an emoji at all (⚰️ is one). Text after them then shifts a column and box borders break. This project avoids those emoji and sticks to ones that are wide everywhere. A rule to steal: if an emoji has a plain-text twin, it will misbehave somewhere.

## 🕯️ Before dawn

Style the `monster` command's page: name in `nameStyle`, description in `metaStyle`, the art in `factStyle`, the facts as `factStyle` bullets. Put `RenderMonster(m monsters.Monster) string` in `cards.go` and call it from the command. Tomorrow the monster chooses its own colours.

> 📺 *"Red on black. Gold for names. Green for the small print. Tomorrow night, each monster picks its own wardrobe, and we build them a little card to sleep in."*

[← Night 8](night-08.md) · [Index](README.md) · [Night 10 →](night-10.md)
