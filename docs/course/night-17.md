# Night 17 · The Gallery

> 📺 *"Nineteen monsters, and so far you've met them one at a time, like a receiving line at a very strange wedding. Tonight we open the gallery: every portrait in one corridor, a cursor to walk it, and a preview of whoever you're standing in front of. Don't make eye contact with the Metaluna Mutant."*

**Tonight you'll learn**
- A list screen with a cursor
- Keeping the selected row on screen when the list scrolls
- A preview pane that appears when there's room
- Screens that remember where you were
- A test that walks the whole list at tiny heights

**Where we are:** the explorer has an intro and a detail page for each monster.

## The third screen

Night 15 added `screen` to the model. The gallery is `screenGallery`, reached by `g` or `esc` from a monster's page, or straight away with `monster --all`. It has its own key handler, in [`ui.go`](../../internal/ui/ui.go):

```go
func (m model) updateGallery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.cursor = (m.cursor - 1 + len(m.monsters)) % len(m.monsters)
	case "down", "j":
		m.cursor = (m.cursor + 1) % len(m.monsters)
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.monsters) - 1
	case "enter", " ", "right", "l":
		m.open(m.cursor)
		return m, m.ensureTick()
	}
	return m, nil
}
```

`cursor` is a separate field from `index`: the gallery selection and the monster on the detail page. They're usually the same, and `open` sets both, but keeping them apart means pressing `esc` from a monster's page puts the cursor *on that monster*, wherever you'd left it before:

```go
	case "g", "esc", "backspace":
		m.cursor = m.index
		m.screen = screenGallery
```

Small courtesy, and it's the kind that makes an interface feel solid: screens remember.

## Drawing the list

`galleryLines` in [`render.go`](../../internal/ui/render.go) builds the whole list as lines, two per monster plus a heading per pack, and reports which line the cursor is on:

```go
func (m model) galleryLines(width int) ([]string, int) {
	var lines []string
	cursorLine := 0
	pack := ""
	for i, mo := range m.monsters {
		if mo.Pack != pack {
			pack = mo.Pack
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, headingStyle.Render("── "+packName(pack)+" ──"))
		}
		p := paletteFor(mo)
		name := mo.Emoji + " " + mo.Name
		marker := "  "
		nameStyle := lipgloss.NewStyle().Foreground(p.primary)
		if i == m.cursor {
			marker = lipgloss.NewStyle().Foreground(colorBlood).Render("▶ ")
			nameStyle = nameStyle.Bold(true).Underline(true)
			cursorLine = len(lines)
		}
		year := ""
		if mo.Debut.Year != 0 {
			year = helpStyle.Render(fmt.Sprintf(" (%d)", mo.Debut.Year))
		}
		lines = append(lines, marker+nameStyle.Render(name)+year)
		lines = append(lines, "    "+hostStyle.Render(truncate(mo.Description, width-6)))
	}
	return lines, cursorLine
}
```

The `pack` variable is a classic *group-break*: remember the last group seen, and print a heading when it changes. Since monsters arrive in pack order, that's all a grouped list needs. `cursorLine` is captured as `len(lines)` *before* the row is appended, which is the index the row will have.

Returning two values, the lines and the cursor line, is a typical Go signature. The caller needs both, and a struct for two ints would be ceremony.

## Keeping the cursor on screen

Night 12's `viewport` shows `height` lines from an offset. For the detail page the offset was the scroll position. For a list, the offset must follow the cursor: when you press `j` off the bottom of the screen, the list should scroll so the selected row is visible. In `viewGallery`:

```go
		lines, cursorLine := m.galleryLines(listW)
		h := m.availableHeight(header, footer)
		// Scroll so the selected name and its description sit above the
		// "more below" marker, but never under the "more above" one (row 0).
		offset := 0
		if h > 0 && cursorLine+1 >= h-1 {
			offset = min(cursorLine-h+3, cursorLine-1)
		}
		list := viewport(lines, offset, h)
```

Read it slowly, because this is where list UIs go wrong. `viewport` replaces the first visible line with `▲ more above` and the last with `▼ more below`, so the *usable* rows are 1 to `h-2`. The cursor's row and its description row must both land in there. If they fit without scrolling, the offset is 0. Otherwise the offset puts the description on row `h-2`, unless that would push the name onto row 0 (under the marker), in which case the name goes on row 1 instead. The `min` takes whichever is smaller.

A comment like the one in the code is worth more than the arithmetic. Someone will change `viewport`'s markers one day, and the comment tells them what to keep true.

## The preview

On a terminal 100 columns or wider the list takes the left half and a card with the selected monster's name, debut and portrait takes the right:

```go
		listW := w
		showPreview := w >= 100
		if showPreview {
			listW = w / 2
		}
		...
		if showPreview {
			list = lipgloss.NewStyle().Width(listW).Render(list)
			body = lipgloss.JoinHorizontal(lipgloss.Top, list, m.preview(m.monsters[m.cursor], w-listW-2))
		}
```

`preview` is the fact card's cousin: a bordered box in the monster's colours with a `MaxWidth`, so a wide portrait is clipped rather than breaking the layout. The list is rendered with a fixed `Width` first so `JoinHorizontal` has a straight edge to butt the preview against; without it, the preview would wander left and right as the visible names changed length.

## Run it

```bash
go run . monster --all
```

Walk the list with `j`/`k`. Watch the preview change and the list scroll when the cursor reaches the bottom. `enter` opens a monster, `esc` comes back to the same row. Then shrink the terminal under 100 columns and the preview folds away; shrink it to ten lines tall and it still shows the selected row.

## Testing at tiny heights

That last claim has a test, in [`ui_test.go`](../../internal/ui/ui_test.go):

```go
// TestGallerySelectionVisibleOnTinyScreens covers the smallest gallery viewport.
func TestGallerySelectionVisibleOnTinyScreens(t *testing.T) {
	for _, h := range []int{8, 9, 10, 12, 24} {
		m := send(newModel(Options{NoIntro: true, ShowAll: true, Seed: 1}), size(80, h))
		for i := range monsters.GetAllMonsters() {
			if name := m.monsters[m.cursor].Name; !strings.Contains(m.View(), name) {
				t.Fatalf("height %d, cursor %d: selected %q not visible", h, i, name)
			}
			m = send(m, key("j"))
		}
	}
}
```

Five heights, every monster, one assertion: the selected name is somewhere on screen. The offset arithmetic above was rewritten twice, and both times this test was what said whether the rewrite was right. When you find yourself doing careful arithmetic about screen rows, write the test *first*; it's cheaper than squinting at a terminal at five sizes.

And the navigation test, which is the same shape as Night 11's:

```go
func TestGalleryNavigation(t *testing.T) {
	m := send(newModel(Options{NoIntro: true, ShowAll: true, Seed: 1}), size(120, 30))
	if m.screen != screenGallery {
		t.Fatal("expected gallery")
	}
	m = send(m, key("j"), key("j"), key("enter"))
	if m.screen != screenDetail || m.index != 2 {
		t.Fatalf("expected third monster, got %d", m.index)
	}
	m = send(m, key("esc"))
	if m.screen != screenGallery || m.cursor != 2 {
		t.Fatal("esc should return to gallery with cursor kept")
	}
}
```

## Try it

- Add `pgdown`/`pgup` to the gallery, moving the cursor by `h-3` rows at a time.
- Type-ahead: while in the gallery, letters jump the cursor to the next monster whose name starts with that letter. (Careful with `j`, `k`, `g`, `q`.)
- Change `showPreview` to `w >= 80` and see what happens to a 30-column-wide preview of the Creature. `MaxWidth` is doing quiet work.

## 💀 Terrifying fact

`lipgloss.JoinHorizontal` pads every block to the height of the tallest, but it does *not* pad widths: each column is as wide as its own widest line. If the left block's lines vary in width, the right block zigzags. Rendering the left block through a style with a fixed `Width` first is the fix, and it's the mistake behind most "why is my second column jagged" questions.

## 🕯️ Before dawn

The gallery marks nothing about which monsters you've visited. The model has a `seen` map (Night 15's `enter` fills it). Show a small `·` after the name of every monster in `seen`, in `helpStyle`. Night 21 saves that map to disk; tonight, it only lives for the session.

> 📺 *"A corridor of portraits, and none of the eyes follow you. Probably. Tomorrow night we stop showing you facts and start asking about them. Sharpen a pencil."*

[← Night 16](night-16.md) · [Index](README.md) · [Night 18 →](night-18.md)
