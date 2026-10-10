# Night 14 · Whispers in the Dark

> 📺 *"A viewer writes: 'Which one was played by Lugosi?' Another: 'Who's afraid of silver?' Until tonight the program couldn't answer either. Tonight it can. Press slash, whisper a word, and every monster in the vault that knows something about it steps forward."*

**Tonight you'll learn**
- Searching text across many fields, ranked by where it matched
- Byte offsets versus character offsets, and why it matters
- A *mode* in a TUI: the same keys meaning different things
- Handling typed text: runes, space, backspace, escape, enter
- Table-driven tests
- The end of the séance: what you've built in two weeks

**Where we are:** the explorer has five pages per monster. Tonight it gets a search.

## Where to look

Searching is more than "does the name contain the word". A viewer typing `lugosi` wants Dracula, but Lugosi is in Dracula's *film* field, in a Wolf Man *fact* (he played the werewolf who bites Larry Talbot) and in a Frankenstein *legacy* line. All three monsters should appear, each once, and the reason should be shown: "(film)", "(fact)". And when a monster matches in several places, show the most important one: name beats alias beats description beats a fact buried in the legend.

So the design is: for each monster, a list of *fields* in order of importance; take the first that matches; remember which it was and a snippet of text around the match.

Create `internal/monsters/search.go`:

```go
package monsters

import "strings"

// SearchHit is a monster matching a search, with where and why it matched.
type SearchHit struct {
	Monster Monster
	Field   string
	Snippet string
}

// Search finds monsters whose name, aliases or lore mention the query. Each
// monster appears at most once, at its best-ranked matching field.
func Search(query string) []SearchHit {
	return searchIn(monsters, query)
}

func searchIn(list []Monster, query string) []SearchHit {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	var hits []SearchHit
	for _, m := range list {
		for _, f := range searchFields(m) {
			lower := strings.ToLower(f.text)
			if i := strings.Index(lower, q); i >= 0 {
				text := f.text
				if len(lower) != len(text) {
					// Lowercasing changed the byte length, so i would mis-slice text.
					text = lower
				}
				hits = append(hits, SearchHit{Monster: m, Field: f.name, Snippet: snippet(text, i, len(q))})
				break
			}
		}
	}
	return hits
}

type field struct{ name, text string }

// searchFields lists what we search, most important first.
func searchFields(m Monster) []field {
	fs := []field{{"name", m.Name}}
	for _, a := range m.Aliases {
		fs = append(fs, field{"alias", a})
	}
	fs = append(fs, field{"description", m.Description}, field{"origin", m.Origin})
	if m.Film != nil {
		fs = append(fs, field{"film", m.Film.Title + " · " + m.Film.Director + " · " + m.Film.Star})
	}
	for _, f := range m.Facts {
		fs = append(fs, field{"fact", f})
	}
	fs = append(fs, field{"legend", m.Legend})
	for _, my := range m.Myths {
		fs = append(fs, field{"myth", my.Claim + " " + my.Explanation})
	}
	return fs
}
```

`Search` is a one-liner over `searchIn`, the same trick as `LoadPacks` on Night 6: the public function uses the real vault, and a test can hand `searchIn` a small list of its own. `searchFields` is the ranking, written as an ordered list, so changing what counts as important is a matter of moving a line.

**`strings.Index`** returns where the query starts in the text, or -1. The `if i := ...; i >= 0` form declares `i` for just that `if`. The `break` leaves the fields loop at the first hit, which is what "at most once, at its best field" means.

## Bytes and characters, again

The snippet is a window of text around the match. `strings.Index` returned a *byte* offset, but the window should be measured in *characters*, and Night 8's terrifying fact was that those differ: "…" is three bytes, `é` is two. Slicing a string by bytes in the middle of a character produces garbage.

```go
// snippet trims long text to a window of characters around the match, which
// starts at byte offset at and is n bytes long.
func snippet(text string, at, n int) string {
	const pad = 30
	r := []rune(text)
	start := len([]rune(text[:at]))
	end := start + len([]rune(text[at:at+n]))
	from, to := start-pad, end+pad
	prefix, suffix := "…", "…"
	if from <= 0 {
		from, prefix = 0, ""
	}
	if to >= len(r) {
		to, suffix = len(r), ""
	}
	return prefix + string(r[from:to]) + suffix
}
```

The conversion is `len([]rune(text[:at]))`: the number of characters *before* byte `at`. Once `start` and `end` are in characters, the rest is slicing `r`, the rune slice, and adding an ellipsis on whichever sides were cut. `const pad = 30` is a constant local to the function.

The comment in `searchIn` about byte length covers a real edge: `strings.ToLower` can change the *length* of a string (a few letters have multi-byte lowercase forms), and then an offset found in the lowered text would be wrong in the original. When that happens we snippet the lowered text, which is safe if slightly less pretty. Knowing that this edge exists is the difference between search that works and search that panics on Night 27's community pack with an İ in it.

## A table of tests

Create `internal/monsters/search_test.go`:

```go
package monsters

import (
	"strings"
	"testing"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		query string
		ids   []string
		field string
	}{
		{"browning", []string{"dracula"}, "film"},
		{"LARRY", []string{"wolf-man"}, "alias"},
		{"silver", []string{"wolf-man"}, "myth"},
		{"", nil, ""},
		{"zombie", nil, ""},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			hits := Search(tc.query)
			if len(hits) != len(tc.ids) {
				t.Fatalf("Search(%q) = %d hits, want %d", tc.query, len(hits), len(tc.ids))
			}
			for i, h := range hits {
				if h.Monster.ID != tc.ids[i] {
					t.Errorf("hit %d is %s, want %s", i, h.Monster.ID, tc.ids[i])
				}
				if h.Field != tc.field {
					t.Errorf("hit %d matched in %q, want %q", i, h.Field, tc.field)
				}
				if !strings.Contains(strings.ToLower(h.Snippet), strings.ToLower(tc.query)) {
					t.Errorf("snippet %q should contain the query", h.Snippet)
				}
			}
		})
	}
}

func TestSnippet(t *testing.T) {
	text := strings.Repeat("a", 50) + "match" + strings.Repeat("b", 50)
	got := snippet(text, 50, 5)
	want := "…" + strings.Repeat("a", 30) + "match" + strings.Repeat("b", 30) + "…"
	if got != want {
		t.Errorf("snippet = %q, want %q", got, want)
	}
	if got := snippet("short", 0, 5); got != "short" {
		t.Errorf("a short snippet should be untouched, got %q", got)
	}
}
```

This is a **table-driven test**, the most common shape of Go test: a slice of anonymous structs (Night 13's idiom again), one row per case, and a loop that runs each as a subtest. Adding a case is adding a row. `t.Run(tc.query, ...)` names each subtest after its query, so a failure reads `TestSearch/silver`.

Why does `silver` match the Wolf Man in a *myth* and not a fact? Because his facts don't mention silver; the myths do ("killed with a silver bullet": false, it was a silver-headed cane). The test found that out for me; I'd assumed otherwise. Write down what you expect, let the test correct you.

## A mode

Now the explorer. Until tonight every key meant one thing. Once the viewer is typing a search, `q` has to be the letter q, not quit, and `j` is a letter, not scroll. The explorer needs a **mode**: a flag that says which set of meanings is in force. Add to the model in `ui.go`:

```go
	searching bool                 // the / prompt is open
	query     string               // what has been typed so far
	hits      []monsters.SearchHit // matches for query
	hitCursor int                  // which hit is chosen
```

At the top of the `tea.KeyMsg` case in `Update`, before anything else:

```go
	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
```

and a way in, next to the `q` case:

```go
		case "/":
			m.searching, m.query, m.hits, m.hitCursor = true, "", nil, 0
```

Every key while searching goes to its own function, and the normal keys never see it. That single early `return` is the whole mode mechanism, and it's how the finished explorer also handles the gallery, the quiz and the help overlay: one `screen` field and a dispatch at the top of `Update`.

## Reading typed text

```go
// updateSearch handles keys while the / prompt is open.
func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
		return m, nil
	case tea.KeyEnter:
		if len(m.hits) > 0 {
			id := m.hits[m.hitCursor].Monster.ID
			m.searching = false
			for i, mo := range m.monsters {
				if mo.ID == id {
					m.open(i)
				}
			}
		}
		return m, nil
	case tea.KeyUp:
		m.hitCursor = max(m.hitCursor-1, 0)
		return m, nil
	case tea.KeyDown:
		m.hitCursor = min(m.hitCursor+1, len(m.hits)-1)
		return m, nil
	case tea.KeyBackspace:
		if r := []rune(m.query); len(r) > 0 {
			m.query = string(r[:len(r)-1])
		}
	case tea.KeyRunes:
		m.query += string(msg.Runes)
	case tea.KeySpace:
		m.query += " "
	default:
		return m, nil
	}
	m.hits = monsters.Search(m.query)
	m.hitCursor = 0
	return m, nil
}
```

This switches on `msg.Type` rather than `msg.String()`, because now we care about *kinds* of key. `tea.KeyRunes` is "the user typed some characters", and `msg.Runes` is what they were: a rune slice, so pasting `Bela Lugosi` arrives as one message. Space is its own type in Bubble Tea, hence the separate case.

The shape of the function is worth studying. Escape, enter and the arrows change the mode or the cursor and **return early**. The three cases that change the *query* fall out of the switch to the bottom, where the search is re-run and the cursor reset. Any other key hits `default` and returns without searching. So there's exactly one place the search runs, and it runs only when the text changed. Search-as-you-type on a vault of a few hundred fields is instant; if it weren't, this is the one line you'd optimise.

Backspace removes a *rune*, not a byte, via the `[]rune` conversion, so deleting the last character of `Nosferatü` deletes the `ü` and not half of it.

## Showing the results

In `View`, after the `body :=` line:

```go
	if m.searching {
		body = m.searchResults(m.contentWidth())
	}
```

Replace `footer` and add `searchResults`:

```go
func (m model) footer() string {
	w := m.contentWidth()
	if m.searching {
		prompt := headingStyle.Render("/") + " " + m.query + "▌"
		return prompt + "\n" + helpStyle.Render(joinFit([]string{"type to search", "↑/↓ choose", "enter open", "esc cancel"}, " · ", w))
	}
	keys := []string{fmt.Sprintf("Monster %d of %d", m.index+1, len(m.monsters)), "←/→ monster", "tab/1-5 section", "↑/↓ scroll", "r reveal", "/ search", "q quit"}
	return helpStyle.Render(joinFit(keys, " · ", w))
}

// searchResults draws the hits for the current query, or a hint.
func (m model) searchResults(w int) string {
	if strings.TrimSpace(m.query) == "" {
		return "\n" + helpStyle.Render(wrap("Search names, nicknames, facts, films and legends. Try \"lugosi\", \"silver\" or \"1931\".", w))
	}
	if len(m.hits) == 0 {
		return "\n" + helpStyle.Render("Nothing stirs in the dark... try another search.")
	}
	lines := []string{""}
	for i, h := range m.hits {
		marker := "  "
		if i == m.hitCursor {
			marker = lipgloss.NewStyle().Foreground(colorBlood).Render("▶ ")
		}
		name := lipgloss.NewStyle().Foreground(paletteFor(h.Monster).primary).Render(h.Monster.Emoji + " " + h.Monster.Name)
		lines = append(lines, marker+name, "    "+helpStyle.Render(truncate("("+h.Field+") "+h.Snippet, w-6)))
	}
	return strings.Join(lines, "\n")
}
```

The prompt lives in the footer, where a `vim` or `less` user expects a `/` prompt, with a `▌` block as a fake cursor: the real cursor is hidden in a Bubble Tea program, since the program draws everything. The results replace the body. Both the empty state and the no-results state say something, because a blank screen after typing is the worst feedback of all.

## Run it

```bash
go run . monster
```

`/`, then `lugosi`: three monsters, each with the field and snippet that matched. `↓` `↓` `enter` and you're on the Wolf Man. `/` `1931`: one hit, although two of our films came out that year, because a film's year is a number and `searchFields` only looks at text. The hint promises more than it delivers; that's yours to fix below. `/` `zzz`: nothing stirs. `esc` puts everything back. Try `silver`, `fog`, `stoker`, `yak`.

## The test

Extend `key()` in `ui_test.go` with enter, escape and backspace:

```go
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
```

And a test that types:

```go
func TestSearch(t *testing.T) {
	var m tea.Model = newModel("")
	m, _ = m.Update(size(80, 24))
	m, _ = m.Update(key("/"))
	if v := m.View(); !strings.Contains(v, "/ ▌") || !strings.Contains(v, "esc cancel") {
		t.Fatalf("/ should open the search prompt:\n%s", v)
	}
	for _, r := range "larry" {
		m, _ = m.Update(key(string(r)))
	}
	v := m.View()
	if !strings.Contains(v, "/ larry▌") || !strings.Contains(v, "▶ 🐺 The Wolf Man") || strings.Contains(v, "Dracula") {
		t.Fatalf("typing should filter to the Wolf Man:\n%s", v)
	}
	m, _ = m.Update(key("enter"))
	if v := m.View(); !strings.Contains(v, "WOLF MAN") || strings.Contains(v, "esc cancel") {
		t.Fatalf("enter should open the chosen monster:\n%s", v)
	}
	m, _ = m.Update(key("/"))
	for _, r := range "zzz" {
		m, _ = m.Update(key(string(r)))
	}
	if v := m.View(); !strings.Contains(v, "Nothing stirs") {
		t.Fatalf("a miss should say so:\n%s", v)
	}
	m, _ = m.Update(key("esc"))
	if v := m.View(); strings.Contains(v, "Nothing stirs") || !strings.Contains(v, "WOLF MAN") {
		t.Fatalf("esc should cancel the search:\n%s", v)
	}
}
```

`for _, r := range "larry"` sends one rune message per letter, like a person typing. The whole interaction, open, type, choose, cancel, is a dozen lines and runs in a millisecond.

```bash
gofmt -l . && go vet ./... && go test ./...
```

## Try it

- Make `q` while searching type a q (it does already; check why) and make `ctrl+c` still quit (it doesn't; fix it in `updateSearch`).
- Make `searchFields` include the film year, so `1931` finds both films. What type is `m.Film.Year`, and how do you get a string from it? (`strconv.Itoa`.)
- Highlight the matched word inside the snippet. `strings.Index` again, and a style around the slice.

## 💀 Terrifying fact

`strings.ToLower("İ")` (a capital I with a dot, used in Turkish) is `"i̇"`: two characters and three bytes, from one character and two bytes. Most string code that "just lowercases and compares" has a bug it will never see, until a name from the wrong alphabet arrives. The `len(lower) != len(text)` check is a small price. The bigger lesson: any time you compute an offset in one string and use it in another, ask whether they're really the same length.

## 🕯️ Before dawn

Two weeks in. Look at what you have: a CLI with four commands, a data vault loaded from embedded JSON, a styled and colour-themed listing, a fact card, and an interactive explorer with pages, scrolling, resizing and search, and tests for all of it. That's a real program; people ship less.

Tag it: `git tag night-14`. Then, no code tonight. Run `go run . monster`, press `/`, and go looking for something you didn't know about a monster. Tomorrow, someone arrives to introduce them.

> 📺 *"The séance is over. The spirits answer to their names, their nicknames, and any word you care to whisper. Next week is the double feature: an introduction from your host, a silent picture, and the midnight quiz. Same channel. Bring a blanket."*

[← Night 13](night-13.md) · [Index](README.md) · [Night 15 →](night-15.md)
