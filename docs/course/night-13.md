# Night 13 · Friday the 13th: Turn the Page

> 📺 *"Thirteen. Unlucky for some, mostly for anyone who sat through House of Dracula. Tonight we admit that a monster is more than a portrait and a list of facts. It has a legend, a film, a stat card, and a set of myths that need putting straight. So we give each monster several pages, and a way to turn them."*

**Tonight you'll learn**
- Growing a data model: nested structs, pointers for optional data
- A method on a struct that formats its own fields
- `iota`: Go's way of numbering a set of names
- Maps, and a `switch` on a value
- Closures: little functions that remember their surroundings
- Anonymous structs in a loop
- Turning key presses into numbers with `strconv`

**Where we are:** the explorer fits any screen but shows only facts.

## The full monster

So far a monster has a name, a description, an origin, facts, aliases, an emoji, colours and a portrait. The finished project knows a great deal more. Open `internal/monsters/monsters.go` and grow `Monster` into its final shape:

```go
// Monster is a single creature and everything we know about it.
type Monster struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Aliases     []string    `json:"aliases,omitempty"`
	Emoji       string      `json:"emoji"`
	Description string      `json:"description"`
	Origin      string      `json:"origin"`
	Debut       Debut       `json:"debut"`
	Film        *Film       `json:"film,omitempty"`
	Legend      string      `json:"legend"`
	Myths       []MythCheck `json:"myths,omitempty"`
	Quotes      []Quote     `json:"quotes,omitempty"`
	Powers      []string    `json:"powers"`
	Weaknesses  []string    `json:"weaknesses"`
	Stats       Stats       `json:"stats"`
	Legacy      []string    `json:"legacy,omitempty"`
	Facts       []string    `json:"facts"`
	HostIntro   string      `json:"hostIntro,omitempty"`
	Theme       Theme       `json:"theme"`
	ASCII       string      `json:"ascii,omitempty"`
	Pack        string      `json:"pack"`
}

// Debut describes where a monster first appeared.
type Debut struct {
	Medium  string `json:"medium"` // novel, film, folklore, ...
	Title   string `json:"title"`
	Creator string `json:"creator,omitempty"`
	Year    int    `json:"year,omitempty"`
	Era     string `json:"era,omitempty"` // for folklore without a single year
}

// Film describes a monster's classic screen outing.
type Film struct {
	Title       string `json:"title"`
	Year        int    `json:"year"`
	ReleaseDate string `json:"releaseDate,omitempty"` // YYYY-MM-DD
	Studio      string `json:"studio,omitempty"`
	Director    string `json:"director"`
	Star        string `json:"star"`
	Makeup      string `json:"makeup,omitempty"`
	Silent      bool   `json:"silent,omitempty"`
	Note        string `json:"note,omitempty"`
}

// MythCheck is a popular belief with a verdict and the real story.
type MythCheck struct {
	Claim       string `json:"claim"`
	True        bool   `json:"true"`
	Explanation string `json:"explanation"`
}

// Quote is a line from a public-domain source text.
type Quote struct {
	Text    string `json:"text"`
	Speaker string `json:"speaker"`
	Source  string `json:"source"`
}

// Stats power the Monster Mash. Each is 1-10 and purely for fun.
type Stats struct {
	Strength int `json:"strength"`
	Speed    int `json:"speed"`
	Cunning  int `json:"cunning"`
	Dread    int `json:"dread"`
}
```

`Theme` stays as it was. Two things to notice.

**`Film *Film`** is a pointer, and the other nested structs aren't. Every monster has a `Debut` and `Stats`, so plain values are right: if the JSON leaves them out, they're just zero. But not every monster has a film (Baba Yaga never got her Universal picture), and a program needs to tell "no film" from "a film with empty fields". A pointer can be `nil`, and that's the difference. When the key is missing from the JSON, `encoding/json` leaves the pointer `nil`; when it's present, it allocates a `Film` and points at it.

**`Quotes`** come with a `Source`, and the project's rule is that every quote is from a public-domain text: the 1897 novel, the 1818 novel, folklore. Film dialogue is copyrighted, and a project that teaches people about monsters shouldn't copy its lines. Describe a scene in your own words instead.

### A method that formats

`Origin` is now a *setting* ("Eastern European vampire folklore") and `Debut` is the first *work*. Since a debut has a title, maybe a creator, a year or an era, it deserves a method that formats it consistently:

```go
// FirstAppearance renders the debut as a readable line.
func (m Monster) FirstAppearance() string {
	d := m.Debut
	s := d.Title
	if d.Creator != "" {
		s += " by " + d.Creator
	}
	switch {
	case d.Year != 0 && d.Medium != "":
		s += fmt.Sprintf(" (%d %s)", d.Year, d.Medium)
	case d.Year != 0:
		s += fmt.Sprintf(" (%d)", d.Year)
	case d.Era != "":
		s += " (" + d.Era + ")"
	}
	return s
}
```

Dracula's becomes `Dracula by Bram Stoker (1897 novel)`. Anywhere that wants a first appearance calls this, so the format is decided once.

### The data

Now the JSON. The three files need a lot of new fields, and typing them out by hand is a night's work in itself, so copy the finished ones: [`dracula.json`](../../internal/monsters/packs/universal/dracula.json), [`frankenstein.json`](../../internal/monsters/packs/universal/frankenstein.json) and [`wolf-man.json`](../../internal/monsters/packs/universal/wolf-man.json) over your own. Read one as you go: `debut`, `film`, `legend`, five `myths` each with a verdict, `powers`, `weaknesses`, `stats`, `legacy`, seven `facts`, a `hostIntro` for Night 15. Every field maps to a struct field by its tag.

Run `go test ./...` before touching the UI. The data test still passes, because the fields it checks are all still there. That's the pleasure of a loader that tolerates extra keys: the data can grow ahead of the code.

## Naming the pages

Create `internal/ui/tabs.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

type tab int

const (
	tabFacts tab = iota
	tabLegend
	tabFilm
	tabMyths
	tabStats
)

var tabNames = map[tab]string{
	tabFacts:  "Facts",
	tabLegend: "Legend",
	tabFilm:   "The Film",
	tabMyths:  "Myth vs Movie",
	tabStats:  "Stat Card",
}

// tabsFor lists the sections a monster has something to say in.
func tabsFor(m monsters.Monster) []tab {
	tabs := []tab{tabFacts}
	if m.Legend != "" {
		tabs = append(tabs, tabLegend)
	}
	if m.Film != nil {
		tabs = append(tabs, tabFilm)
	}
	if len(m.Myths) > 0 {
		tabs = append(tabs, tabMyths)
	}
	return append(tabs, tabStats)
}
```

**`type tab int`** makes a new type that is an integer underneath but is *not* an `int`: you can't pass a plain `3` where a `tab` is wanted, which stops you mixing up a tab with a monster index. Go has no `enum` keyword; this is how you make one.

**`iota`** is a counter that starts at 0 in each `const` block and goes up one per line. Only the first line says `tab = iota`; the others repeat it implicitly. So `tabFacts` is 0, `tabLegend` 1 and so on, and you never write the numbers. Adding `tabQuotes` in the middle renumbers everything correctly.

**`map[tab]string`** is a map, Go's dictionary: a key type in the brackets, a value type after. `tabNames[t]` looks one up. Maps are the second workhorse collection after slices.

`tabsFor` decides which pages a monster *has*: everyone has Facts and a Stat Card, but the Film page only appears when `m.Film != nil`, the pointer check from earlier. The tab bar is built from this list, so a folklore monster simply has fewer pages, and the number keys renumber themselves.

## Drawing the bar

```go
// tabBar draws the numbered section names, highlighting the current one.
func (m model) tabBar(p palette) string {
	active := lipgloss.NewStyle().Bold(true).Foreground(colorBlack).Background(p.primary).Padding(0, 1)
	inactive := lipgloss.NewStyle().Foreground(colorDim).Padding(0, 1)
	var parts []string
	for i, t := range tabsFor(m.current()) {
		label := fmt.Sprintf("%d %s", i+1, tabNames[t])
		if i == m.tab {
			parts = append(parts, active.Render(label))
		} else {
			parts = append(parts, inactive.Render(label))
		}
	}
	return joinFit(parts, " ", m.contentWidth())
}
```

The current tab is drawn as black text on the monster's own colour; the rest are dim. The bar is `1 Facts  2 Legend  3 The Film ...`, numbered from the *position* in `tabsFor`, not the constant, so the number on screen is always the key to press.

`joinFit` goes at the bottom of the file. It's a word-wrap for a list of items that must not be split in the middle:

```go
// joinFit joins items with sep, starting a new line rather than splitting an item.
func joinFit(items []string, sep string, width int) string {
	var lines []string
	line := ""
	for _, it := range items {
		switch {
		case line == "":
			line = it
		case lipgloss.Width(line+sep+it) <= width:
			line += sep + it
		default:
			lines = append(lines, line)
			line = it
		}
	}
	return strings.Join(append(lines, line), "\n")
}
```

Ordinary `wrap` would happily break `Myth vs` / `Movie` across lines, and a styled label broken in half loses its background colour on one side. `joinFit` only ever breaks *between* items.

## Drawing a page

The heart of tonight. One function, one `switch`, one case per page:

```go
// tabContent renders one section of a monster's page, w columns wide.
func (m model) tabContent(t tab, mo monsters.Monster, p palette, w int) string {
	body := lipgloss.NewStyle().Foreground(p.body)
	var b strings.Builder
	heading := func(s string) { b.WriteString(headingStyle.Render(s) + "\n\n") }

	switch t {
	case tabFacts:
		heading("Terrifying Facts")
		for _, f := range mo.Facts {
			b.WriteString(bullet(f, w) + "\n")
		}
		b.WriteString("\n" + metaStyle.Render(wrap("Origin: "+mo.Origin, w)))
		b.WriteString("\n" + metaStyle.Render(wrap("First appearance: "+mo.FirstAppearance(), w)))

	case tabLegend:
		heading("The Legend Behind the Monster")
		b.WriteString(body.Render(wrap(mo.Legend, w)))

	case tabFilm:
		f := mo.Film
		heading(fmt.Sprintf("🎬 %s (%d)", f.Title, f.Year))
		row := func(label, value string) {
			if value != "" {
				b.WriteString(metaStyle.Render(fmt.Sprintf("%-19s", label)) + body.Render(wrap(value, max(w-19, 10))) + "\n")
			}
		}
		row("Studio", f.Studio)
		row("Director", f.Director)
		row("Monster played by", f.Star)
		row("Makeup", f.Makeup)
		if f.Note != "" {
			b.WriteString("\n" + body.Render(wrap(f.Note, w)))
		}

	case tabMyths:
		heading("Myth vs Movie")
		if !m.revealed {
			b.WriteString(helpStyle.Render(wrap("True or false? Make your guesses, then press r to reveal the verdicts.", w)) + "\n\n")
		}
		for i, my := range mo.Myths {
			b.WriteString(body.Render(wrap(fmt.Sprintf("%d. %s", i+1, my.Claim), w)) + "\n")
			switch {
			case !m.revealed:
				b.WriteString(helpStyle.Render("   ? ? ?") + "\n\n")
			case my.True:
				b.WriteString("   " + trueStyle.Render("✔ TRUE") + " " + helpStyle.Render(wrap(my.Explanation, max(w-10, 10))) + "\n\n")
			default:
				b.WriteString("   " + mythStyle.Render("✘ MYTH") + " " + helpStyle.Render(wrap(my.Explanation, max(w-10, 10))) + "\n\n")
			}
		}

	case tabStats:
		heading("Stat Card")
		s := mo.Stats
		for _, st := range []struct {
			name string
			v    int
		}{{"Strength", s.Strength}, {"Speed", s.Speed}, {"Cunning", s.Cunning}, {"Dread", s.Dread}} {
			b.WriteString(metaStyle.Render(fmt.Sprintf("%-9s", st.name)) + statBar(st.v, p) + "\n")
		}
		b.WriteString("\n" + headingStyle.Render("Powers") + "\n")
		for _, pw := range mo.Powers {
			b.WriteString(bullet(pw, w) + "\n")
		}
		b.WriteString("\n" + headingStyle.Render("Weaknesses") + "\n")
		for _, wk := range mo.Weaknesses {
			b.WriteString(bullet(wk, w) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// statBar draws a 10-block meter, e.g. ███████░░░  7
func statBar(v int, p palette) string {
	v = min(max(v, 0), 10)
	return lipgloss.NewStyle().Foreground(p.primary).Render(strings.Repeat("█", v)) +
		helpStyle.Render(strings.Repeat("░", 10-v)) +
		fmt.Sprintf(" %2d", v)
}
```

Several new things are packed in here.

**`switch t`** compares a value against each `case`. Unlike the type switch on Night 11 or the bare `switch {` on Night 8, this one is the plain form: `case tabFacts:` runs when `t == tabFacts`. Go's cases don't fall through, so there's no `break` after each one.

**`heading := func(s string) { ... }`** is a **closure**: a function with no name, stored in a variable, that can see the variables around it. `heading` writes to `b`, the builder declared just above it, without `b` being passed in. `row` in the film case does the same and also skips empty values, so a film with no known makeup artist just has no Makeup line. Closures are how Go avoids a dozen tiny helper functions that each need five arguments.

**The stat loop** ranges over a slice of an **anonymous struct**: a struct type declared right where it's used, with no name, because it's used nowhere else. `[]struct{ name string; v int }{{"Strength", 8}, ...}` is four labelled numbers, and the loop draws each the same way. It's a common Go idiom for tables of test cases too.

**`m.revealed`** hides the myth verdicts behind `? ? ?` until the viewer presses `r`. The page is a quiz before it's an answer key, which is more fun and better teaching. Each verdict is a `✔ TRUE` in mint or a `✘ MYTH` in red; add the two styles to `styles.go`:

```go
	trueStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorMint)
	mythStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorBlood)
```

**`statBar`** is ten blocks, `v` filled and the rest hollow, `strings.Repeat` making each run. The `min(max(v, 0), 10)` clamp means a typo of 12 in the data draws a full bar instead of crashing on a negative repeat count.

## Keys

In `ui.go`, add two fields to the model:

```go
	tab      int  // which section, as an index into tabsFor(current)
	revealed bool // myth verdicts shown
```

`tab` is an *index into* `tabsFor(current)`, not a `tab` value. That's deliberate: `tab` 1 means "the second page this monster has", which is what the tab bar and the number keys work in.

Replace the `tea.KeyMsg` case in `Update`, and import `strconv`:

```go
	case tea.KeyMsg:
		tabs := tabsFor(m.current())
		switch key := msg.String(); key {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "right", "l", "n":
			m.open((m.index + 1) % len(m.monsters))
		case "left", "h", "p":
			m.open((m.index - 1 + len(m.monsters)) % len(m.monsters))
		case "tab":
			m.setTab((m.tab + 1) % len(tabs))
		case "shift+tab":
			m.setTab((m.tab - 1 + len(tabs)) % len(tabs))
		case "down", "j":
			m.scroll++
		case "up", "k":
			m.scroll--
		case "r":
			m.revealed = !m.revealed
		default:
			if n, err := strconv.Atoi(key); err == nil && n >= 1 && n <= len(tabs) {
				m.setTab(n - 1)
			}
		}
		m.clampScroll()
```

**`switch key := msg.String(); key {`** declares `key` and switches on it in one line, so the `default` case can use it. **`strconv.Atoi`** turns a string into an integer, or returns an error if it isn't one: `"4"` becomes 4; `"r"` fails and the key is ignored. The range check keeps `7` harmless on a monster with five pages.

Then the two small methods that the keys call:

```go
// open shows monster i, keeping the same kind of section where it has one.
func (m *model) open(i int) {
	prev := tabsFor(m.current())[m.tab]
	m.index, m.tab, m.scroll, m.revealed = i, 0, 0, false
	for j, t := range tabsFor(m.current()) {
		if t == prev {
			m.tab = j
		}
	}
}

func (m *model) setTab(i int) {
	m.tab, m.scroll = i, 0
}
```

`open` has a subtlety worth the eight lines. If you're reading Dracula's Film page and press `→`, you want the Wolf Man's Film page, not his Facts. But "Film" might be the third page for one monster and the second for another, or missing entirely. So `open` remembers which *kind* of page was showing, moves, and looks for the same kind in the new monster's list, falling back to Facts. Small touches like this are what make an interface feel like it's on your side.

Both use pointer receivers, like `clampScroll`, because they change the model.

## Wiring the pages in

`header` gains the bar as its last line:

```go
		lipgloss.NewStyle().Italic(true).Foreground(p.accent).Render(wrap(mo.Description, w)) + "\n\n" +
		m.tabBar(p)
```

`footer` lists the keys with `joinFit`, so it wraps between items on a narrow screen:

```go
func (m model) footer() string {
	keys := []string{fmt.Sprintf("Monster %d of %d", m.index+1, len(m.monsters)), "←/→ monster", "tab/1-5 section", "↑/↓ scroll", "r reveal", "q quit"}
	return helpStyle.Render(joinFit(keys, " · ", m.contentWidth()))
}
```

And `bodyLines` draws the current page instead of the facts. Delete the `facts` method and change the middle of `bodyLines` to:

```go
	t := tabsFor(mo)[m.tab]

	var body string
	if mo.ASCII != "" && w >= 100 {
		box := artBox(mo, p)
		textW := w - lipgloss.Width(box) - 3
		body = lipgloss.JoinHorizontal(lipgloss.Top, box, "   ", m.tabContent(t, mo, p, textW))
	} else if mo.ASCII != "" && t == tabFacts && w >= artWidth(mo.ASCII)+6 {
		body = m.tabContent(t, mo, p, w) + "\n\n" + artBox(mo, p)
	} else {
		body = m.tabContent(t, mo, p, w)
	}
```

The one change of behaviour: on a narrow screen the portrait only appears under the *Facts* page. Under the Legend it would be a screenful of scrolling before the text.

## Run it

```bash
go run . monster
```

`2` for the legend of how Stoker built his Count from library books. `3` for the film: Tod Browning, Bela Lugosi, Jack Pierce. `4` for five claims about Dracula and a row of `? ? ?`; decide, then `r`. (Sunlight killing him? That's Nosferatu's invention.) `5` for the stat card. `→` while on the Stat Card and you're on Frankenstein's Monster's stat card. `tab` cycles.

## Tests

Teach `key()` in `ui_test.go` about tab:

```go
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
```

Make `TestFitsEverySize` visit every page of every monster, revealed, by replacing the `v := m.View()` line with:

```go
			for range 5 {
				m, _ = m.Update(key("tab"))
			}
			m, _ = m.Update(key("r"))
			v := m.View()
```

Five tabs is a full lap, and pressing `tab` on a monster with fewer pages just wraps sooner. The test now checks the widest content, the film note and the myth explanations, at 40 columns. Then a test for tonight's keys:

```go
func TestTabsAndReveal(t *testing.T) {
	var m tea.Model = newModel("dracula")
	m, _ = m.Update(size(80, 40))
	m, _ = m.Update(key("4"))
	v := m.View()
	if !strings.Contains(v, "Myth vs Movie") || !strings.Contains(v, "? ? ?") || strings.Contains(v, "✘ MYTH") {
		t.Fatalf("4 should open Myth vs Movie with hidden verdicts:\n%s", v)
	}
	m, _ = m.Update(key("r"))
	if v := m.View(); !strings.Contains(v, "✘ MYTH") {
		t.Fatalf("r should reveal verdicts:\n%s", v)
	}
	m, _ = m.Update(key("tab"))
	if v := m.View(); !strings.Contains(v, "Stat Card") || !strings.Contains(v, "██") {
		t.Fatalf("tab should move to the stat card:\n%s", v)
	}
}
```

```bash
go vet ./... && go test ./...
```

## Try it

- Add a `tabQuotes` page between Film and Myths, showing each quote in italics with its speaker and source under it. `tabsFor` should only include it when there are quotes. Watch the number keys renumber themselves. The finished project has this, and a Legacy page: [`render.go`](../../internal/ui/render.go).
- Folklore monsters have no film, so "Myth vs Movie" is the wrong name for them. The project's `tabBar` calls it "Myth or Fact" when `m.Film == nil`. Try it.
- Press `r` on the Facts page. Nothing visible happens, but `revealed` flips, so the next visit to Myths is already revealed. Is that a bug? Decide, and fix it if you think so.

## 💀 Terrifying fact

A map in Go is unordered. `for k, v := range tabNames` visits the entries in a *deliberately random* order, different on every run, so that nobody accidentally relies on it. That's why `tabsFor` returns a slice, and the map is only ever used for lookups. If you ever need a map in order, collect the keys into a slice and `sort` them.

## 🕯️ Before dawn

Add a `hostIntro` line to the Facts page, in `helpStyle` and quoted, when the monster has one. Every monster you copied tonight does. Night 15 gives it a voice.

> 📺 *"Five pages per monster, and a verdict on every myth. Tomorrow night, the last night of the séance, we let the viewer ask a question and the whole vault answers. Type a slash."*

[← Night 12](night-12.md) · [Index](README.md) · [Night 14 →](night-14.md)
