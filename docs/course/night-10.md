# Night 10 · Every Monster in Its Colours

> 📺 *"Dracula wears black and red. The Creature is swamp green. The Mummy, a tasteful sand. Tonight every monster chooses its own colours, and we build a little card to show them off in. It has a border. The border is the best part."*

**Tonight you'll learn**
- Putting presentation choices in the data
- A `palette` type and a function that picks one
- Borders, word wrapping and nested structs
- Methods on a struct
- Building a reusable card

**Where we are:** `list` is styled with house colours.

## Colours in the data

Add two fields to `Monster` in `internal/monsters/monsters.go`, `Emoji` after `Aliases` and `Theme` before `ASCII`, plus a new type:

```go
	Emoji       string   `json:"emoji"`
	...
	Theme       Theme    `json:"theme"`
```

```go
// Theme holds a monster's colours as hex strings.
type Theme struct {
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
}
```

A struct inside a struct, and JSON maps it naturally. In `dracula.json` add:

```json
  "emoji": "🧛",
  "theme": {"primary": "#C1121F", "accent": "#F1E4E8"},
```

Frankenstein's Monster: `🧟`, `#7CB518` and `#DDE5B6`. The Wolf Man: `🐺`, `#B08968` and `#EDE0D4`. (The repo's data files have every monster's theme if you'd rather copy.)

Why put colours in the *data*? Because the person adding Baba Yaga knows she should be blood-red and bone-white, and shouldn't have to find the Go file where colours live. Data that describes a thing belongs with the thing.

## Palettes

Add to `internal/ui/styles.go`, and change its import to bring in the monsters package:

```go
import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)
```

```go
// palette is the set of colours used to draw one monster.
type palette struct {
	primary lipgloss.Color
	accent  lipgloss.Color
	body    lipgloss.Color
}

// paletteFor picks a monster's colours, falling back to the house style.
func paletteFor(m monsters.Monster) palette {
	p := palette{primary: colorGold, accent: lipgloss.Color("#FF6347"), body: colorSky}
	if m.Theme.Primary != "" {
		p.primary = lipgloss.Color(m.Theme.Primary)
	}
	if m.Theme.Accent != "" {
		p.accent = lipgloss.Color(m.Theme.Accent)
	}
	return p
}

// wrap word-wraps text to width.
func wrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	return lipgloss.NewStyle().Width(width).Render(text)
}
```

`paletteFor` starts from defaults and overrides what the monster specifies. A monster with no theme still looks fine. That pattern, *defaults then overrides*, comes back on Night 27 for community monsters.

`wrap` is Lip Gloss doing word-wrap: a style with a `Width` breaks lines at spaces to fit. We'll use it constantly.

## The fact card

Add to `internal/ui/cards.go`:

```go
// FactCard frames a single fact about a monster.
type FactCard struct {
	Title   string // e.g. "RANDOM TERROR FACT"
	Monster monsters.Monster
	Fact    string
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
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.accent).
		Padding(0, 1).
		Render(strings.Join(parts, "\n")) + "\n"
}
```

**`func (c FactCard) Render(...)`** is a **method**: a function attached to a type. `c` is the *receiver*, like `self` or `this` elsewhere, and you call it as `card.Render(80)`. Any type can have methods, not just structs. The `Monster` type in the repo has `FirstAppearance()` and `IsSilent()` methods for the same reason: behaviour that belongs to the data.

The card is at most 64 columns, narrower on a narrow terminal, with a 2-column margin (`width-2`). The text inside is wrapped to `inner`, which leaves room for the border and padding. `min` and `max` are built into Go.

Lip Gloss borders: `RoundedBorder`, `DoubleBorder`, `ThickBorder` and more, coloured with `BorderForeground`. The border is drawn around the *rendered* text, so it fits whatever `parts` came to.

## Use it

In `cmd/random.go`, replace the four `Println` calls with:

```go
		card := ui.FactCard{Title: "RANDOM TERROR FACT", Monster: m, Fact: fact}
		fmt.Print(card.Render(ui.TerminalWidth()))
```

and import `"github.com/hungovercoders/terminal-of-terror/internal/ui"`. Then give the list each monster's colours. In `RenderList`, replace the `line :=` statement with:

```go
			name := lipgloss.NewStyle().Bold(true).Foreground(paletteFor(m).primary).Render(fmt.Sprintf("%-24s", m.Name))
			line := fmt.Sprintf("%s %s %s ", helpStyle.Render(fmt.Sprintf("%3d.", n)), m.Emoji, name)
```

`nameStyle` is now unused; delete it from `styles.go`, or Go's compiler will let it slide (unused *variables* inside functions are errors; unused package-level ones aren't) but `staticcheck` would grumble.

## Run it

```bash
go run . random
```

```
╭──────────────────────────────────────────────────────────────╮
│ 🎃 RANDOM TERROR FACT                                        │
│                                                              │
│ 🐺 The Wolf Man                                              │
│ The film opened just five days after the attack on Pearl     │
│ Harbor                                                       │
╰──────────────────────────────────────────────────────────────╯
```

The name is in the Wolf Man's brown, the border in his cream. Run it a few times and watch the card change colours with the monster. Narrow the terminal to 50 columns; the card shrinks and the fact re-wraps.

## Try it

- Add a `Host string` field to `FactCard` and, when it's set, a final line in `helpStyle` reading `📺 Count Cathode: “...”`. The repo's card has exactly this; Night 15 gives the host something to say.
- Try `lipgloss.DoubleBorder()`.
- Make `paletteFor` return a *third* colour, `body`, from the theme too. What should the JSON key be?

## 💀 Terrifying fact

Every Lip Gloss style method returns a *new* style; the original is untouched. So `s := helpStyle; s.Bold(true)` does nothing, and `s = helpStyle.Bold(true)` is what you meant. It's a common first-week mistake, and the compiler can't catch it, because ignoring a return value is legal Go.

## 🕯️ Before dawn

Give the `monster` page from Night 9's homework the monster's own colours through `paletteFor`, and put the whole thing in a card with a `DoubleBorder`. Tomorrow that page becomes interactive, and you'll keep most of the styling.

> 📺 *"Every monster in its own colours, framed and hung. Tomorrow night, the frame starts moving. Hold on to something."*

[← Night 9](night-09.md) · [Index](README.md) · [Night 11 →](night-11.md)
