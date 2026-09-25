# Night 12 · Fitting the Screen

> 📺 *"The Invisible Man had one great advantage: he fitted in any room. Our explorer doesn't, yet. Make the terminal narrow and the facts spill off the edge; make it short and the footer vanishes into the floor. Tonight we teach it the size of the room, and how to behave in one."*

**Tonight you'll learn**
- `tea.WindowSizeMsg`: learning the terminal's size, and hearing when it changes
- Wrapping text to a width
- Side-by-side layout with `lipgloss.JoinHorizontal`
- A viewport: showing part of something taller than the screen
- Clamping, and why `Update` calls it
- Tests that render at several sizes

**Where we are:** the explorer moves between monsters but ignores the terminal's size.

## The terminal tells you its size

When a Bubble Tea program starts, and whenever the window is resized, it sends a `tea.WindowSizeMsg` with the width and height in characters. Store them in the model and every `View` can lay itself out to fit.

Until that first message arrives, the model doesn't know the size, so we assume 80 columns: the width terminals have defaulted to since punched cards, and still a sensible guess.

## The plan

The page has three parts: a **header** (title, name, description) that must always be visible, a **body** (portrait and facts) that might be taller than the screen, and a **footer** (help). The body scrolls; the other two don't. On a wide terminal the portrait sits beside the facts; on a narrow one it goes underneath; on a very narrow one it's dropped.

Replace `internal/ui/ui.go`. The whole file is [in the repository's history](../../internal/ui/ui.go) in a bigger form; here it is in pieces.

### The model, with a size and a scroll position

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// model is everything the explorer needs to draw itself.
type model struct {
	monsters []monsters.Monster
	index    int // which monster is on screen
	scroll   int // how far the body is scrolled
	quitting bool

	width, height int // the terminal, once we've been told
}

func newModel(startID string) model {
	m := model{monsters: monsters.GetAllMonsters()}
	for i, mo := range m.monsters {
		if mo.ID == startID {
			m.index = i
		}
	}
	return m
}

func (m model) current() monsters.Monster { return m.monsters[m.index] }

// Init runs once when the program starts. Nothing to do yet.
func (m model) Init() tea.Cmd { return nil }
```

### Update: size, scrolling, and clamping

```go
// Update handles one message and returns the new model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampScroll()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "right", "l", "n":
			m.index = (m.index + 1) % len(m.monsters)
			m.scroll = 0
		case "left", "h", "p":
			m.index = (m.index - 1 + len(m.monsters)) % len(m.monsters)
			m.scroll = 0
		case "down", "j":
			m.scroll++
		case "up", "k":
			m.scroll--
		}
		m.clampScroll()
	}
	return m, nil
}

// clampScroll keeps the scroll offset within the body.
func (m *model) clampScroll() {
	maxScroll := 0
	if h := m.bodyHeight(); h > 0 {
		maxScroll = max(len(m.bodyLines())-h, 0)
	}
	m.scroll = min(max(m.scroll, 0), maxScroll)
}
```

`j` and `k` change `scroll` without checking anything; `clampScroll` then pulls it back into range. Doing the check in one place, after *every* change, means no key handler can forget it. A resize can also make a scroll position invalid (the screen got taller, so there's less to scroll), which is why the size message clamps too.

**`func (m *model) clampScroll()`** has a *pointer* receiver, unlike every other method so far. It needs to change `m.scroll` on the caller's copy, not a fresh one. Inside `Update`, `m` is a local variable, so `m.clampScroll()` changes that local, which is then returned. Rule of thumb: pointer receiver to modify, value receiver to read.

### Measuring the room

```go
// contentWidth is the usable width, assuming 80 columns until told otherwise.
func (m model) contentWidth() int {
	if m.width <= 0 {
		return 80
	}
	return m.width
}

// bodyHeight is how many body lines fit between the header and footer.
func (m model) bodyHeight() int {
	if m.height <= 0 {
		return 0
	}
	return max(m.height-lipgloss.Height(m.header())-lipgloss.Height(m.footer())-3, 3)
}
```

`bodyHeight` renders the header and footer to see how tall they are (they wrap, so it depends on the width), subtracts them and three lines of spacing, and never goes below 3. A height of 0 means "unknown, show everything", which keeps the old tests working.

### View: three parts

```go
// View draws the whole screen from the model.
func (m model) View() string {
	if m.quitting {
		return ""
	}
	body := viewport(m.bodyLines(), m.scroll, m.bodyHeight())
	return m.header() + "\n" + body + "\n\n" + m.footer()
}

func (m model) header() string {
	mo := m.current()
	p := paletteFor(mo)
	w := m.contentWidth()
	return titleStyle.Render("🎃 TERMINAL OF TERROR 🎃") + "\n\n" +
		lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(mo.Emoji+"  "+strings.ToUpper(mo.Name)) + "\n" +
		lipgloss.NewStyle().Italic(true).Foreground(p.accent).Render(wrap(mo.Description, w))
}

func (m model) footer() string {
	return helpStyle.Render(wrap(fmt.Sprintf("Monster %d of %d · ←/→ next · ↑/↓ scroll · q quit", m.index+1, len(m.monsters)), m.contentWidth()))
}
```

### The body: beside, below, or alone

```go
// bodyLines renders the scrollable part of the page: the portrait beside the
// facts on a wide terminal, or above them on a narrow one.
func (m model) bodyLines() []string {
	mo := m.current()
	p := paletteFor(mo)
	w := m.contentWidth()

	var body string
	if mo.ASCII != "" && w >= 100 {
		box := artBox(mo, p)
		textW := w - lipgloss.Width(box) - 3
		body = lipgloss.JoinHorizontal(lipgloss.Top, box, "   ", m.facts(mo, p, textW))
	} else if mo.ASCII != "" && w >= artWidth(mo.ASCII)+6 {
		body = m.facts(mo, p, w) + "\n\n" + artBox(mo, p)
	} else {
		body = m.facts(mo, p, w)
	}
	return strings.Split("\n"+body, "\n")
}

func (m model) facts(mo monsters.Monster, p palette, w int) string {
	var b strings.Builder
	b.WriteString(headingStyle.Render("Terrifying Facts") + "\n\n")
	for _, f := range mo.Facts {
		b.WriteString(bullet(f, w) + "\n")
	}
	b.WriteString("\n" + metaStyle.Render(wrap("First appearance: "+mo.Origin, w)))
	return b.String()
}
```

**`lipgloss.JoinHorizontal(lipgloss.Top, a, b, c)`** places multi-line strings side by side, top-aligned, padding shorter ones so the columns stay straight. It's the one function that makes two-column terminal layouts easy. `lipgloss.Width(box)` measures the *widest* line of the box, so the facts get exactly the room that's left.

The body is returned as a slice of *lines*, because scrolling works in lines.

### Helpers

```go
// artBox frames the portrait in the monster's colours.
func artBox(mo monsters.Monster, p palette) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.accent).
		Padding(1, 2).
		Render(lipgloss.NewStyle().Foreground(p.primary).Render(mo.ASCII))
}

// bullet wraps text to w with a hanging indent under a bullet.
func bullet(text string, w int) string {
	lines := strings.Split(wrap(text, max(w-4, 10)), "\n")
	for i := range lines {
		prefix := "    "
		if i == 0 {
			prefix = "  • "
		}
		lines[i] = prefix + factStyle.Render(strings.TrimRight(lines[i], " "))
	}
	return strings.Join(lines, "\n")
}

// artWidth is the width of the widest line of ASCII art.
func artWidth(art string) int {
	w := 0
	for _, l := range strings.Split(art, "\n") {
		w = max(w, lipgloss.Width(l))
	}
	return w
}

// viewport shows the slice of lines starting at offset that fits height.
func viewport(lines []string, offset, height int) string {
	if height <= 0 || len(lines) <= height {
		return strings.Join(lines, "\n")
	}
	offset = min(max(offset, 0), len(lines)-height)
	visible := append([]string{}, lines[offset:offset+height]...)
	if offset > 0 {
		visible[0] = helpStyle.Render("  ▲ more above")
	}
	if offset+height < len(lines) {
		visible[len(visible)-1] = helpStyle.Render("  ▼ more below")
	}
	return strings.Join(visible, "\n")
}

// RunUI starts the interactive explorer, opening on startID if given.
func RunUI(startID string) error {
	p := tea.NewProgram(newModel(startID), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

`viewport` is the whole scrolling mechanism: take `height` lines starting at `offset`, and replace the first or last with a marker if there's more. `lines[offset:offset+height]` is a *slice expression*: a view of part of the slice. `append([]string{}, ...)` copies it, so overwriting `visible[0]` doesn't scribble on the original.

`bullet` does a hanging indent by hand: wrap the text four columns narrower, then prefix the first line with the bullet and the rest with spaces.

## Run it

```bash
go run . monster
```

Now resize the terminal while it's running. Wide: the portrait sits left, facts right. Narrower than 100 columns: facts on top, portrait below. Shorter than the page: `▼ more below` appears, and `j`/`k` scroll. The program redraws on every resize because each one is just another message.

## Testing every size

Add to `ui_test.go`, and import `lipgloss`:

```go
func size(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }

// TestFitsEverySize renders every monster at several terminal sizes: no line
// may be wider than the terminal, and no screen taller than it.
func TestFitsEverySize(t *testing.T) {
	for _, sz := range [][2]int{{40, 20}, {80, 24}, {140, 45}} {
		var m tea.Model = newModel("")
		m, _ = m.Update(size(sz[0], sz[1]))
		for range 3 {
			v := m.View()
			if h := strings.Count(v, "\n") + 1; h > sz[1] {
				t.Errorf("%dx%d: %d lines is taller than the screen", sz[0], sz[1], h)
			}
			for _, line := range strings.Split(v, "\n") {
				if lipgloss.Width(line) > sz[0] {
					t.Errorf("%dx%d: line too wide: %q", sz[0], sz[1], line)
				}
			}
			m, _ = m.Update(key("right"))
		}
	}
}

func TestScrolling(t *testing.T) {
	var m tea.Model = newModel("")
	m, _ = m.Update(size(60, 16))
	top := m.View()
	if !strings.Contains(top, "▼ more below") {
		t.Fatalf("a 16-line screen should need scrolling:\n%s", top)
	}
	m, _ = m.Update(key("j"))
	if v := m.View(); v == top || !strings.Contains(v, "▲ more above") {
		t.Fatalf("j should scroll down:\n%s", v)
	}
}
```

This test is the most valuable one in the project. A resize message is just a struct, so the test can pretend to be any terminal, and "no line wider than the screen" catches an enormous class of layout bugs before a human ever sees them. The finished project runs this over every monster, every section and seven widths.

`for range 3` loops three times, a Go 1.22 addition.

## Try it

- Add `space` and `pgdown` that scroll by a screenful: `m.scroll += max(m.bodyHeight()-2, 1)`.
- Change the 100-column threshold to 60 and see the layout squeeze.
- Make the test loop over `len(m.monsters)` instead of 3 and add `{30, 10}` to the sizes. Something will probably break. Fix it.

## 💀 Terrifying fact

`lipgloss.Height(s)` is `strings.Count(s, "\n") + 1`, which means a string that ends in a newline is one line taller than it looks. The trailing newline you add for `fmt.Print` is a *bug* inside a `View`. Every helper in the explorer returns text without one, and `View` adds separators itself.

## 🕯️ Before dawn

Add `home` and `end` keys that scroll to the top and bottom. `end` can set `m.scroll` to a huge number and let `clampScroll` sort it out, which is the whole point of clamping in one place.

> 📺 *"It fits the room now, whatever the room. Tomorrow is Friday the 13th, or near enough, and we're turning pages. Several pages. Per monster."*

[← Night 11](night-11.md) · [Index](README.md) · [Night 13 →](night-13.md)
