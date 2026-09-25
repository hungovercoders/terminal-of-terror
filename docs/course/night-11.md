# Night 11 · The Séance Begins

> 📺 *"Until now the program has spoken and then fallen silent, like a ghost with one message. Tonight it starts listening. Press a key and something happens. This is the night the terminal becomes a room you can walk around in."*

**Tonight you'll learn**
- What a TUI is, and how Bubble Tea structures one
- The three methods: `Init`, `Update`, `View`
- Messages, and a type switch
- Wrapping around a list with `%`
- The alternate screen
- Testing an interactive program without a terminal

**Where we are:** the `monster` command prints a static page.

## Programs that listen

A **TUI**, a text user interface, is a program that redraws the terminal as you press keys: `vim`, `htop`, `lazygit`. Writing one from scratch means putting the terminal into raw mode, reading key codes, tracking what's on screen, and redrawing without flicker. [Bubble Tea](https://github.com/charmbracelet/bubbletea) does all of that and leaves you three functions to write.

```bash
go get github.com/charmbracelet/bubbletea@v1.3.10
```

## The Elm architecture

Bubble Tea follows a pattern from the Elm language, and it's worth understanding before you type anything:

- A **model** is a struct holding everything the screen needs: which monster is showing, whether we're quitting.
- **`Update(msg)`** takes a *message*, a key press or a window resize or a timer, and returns a new model. It's the only place state changes.
- **`View()`** turns the model into a string. That string *is* the screen. Bubble Tea diffs it against the last one and redraws only what changed.
- **`Init()`** returns any work to start with (a timer, say). Ours does nothing yet.

The loop is: a message arrives, `Update` produces a new model, `View` draws it, repeat. There's no "when this key is pressed, change that widget". You never touch the screen; you describe it.

## The model

Create `internal/ui/ui.go`:

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
	quitting bool
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

`tea "github.com/charmbracelet/bubbletea"` gives the package a short name, `tea`, because `bubbletea.KeyMsg` gets tiring.

## Update

```go
// Update handles one message (a key press, for now) and returns the new model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "right", "l", "n":
			m.index = (m.index + 1) % len(m.monsters)
		case "left", "h", "p":
			m.index = (m.index - 1 + len(m.monsters)) % len(m.monsters)
		}
	}
	return m, nil
}
```

**`switch msg := msg.(type)`** is a *type switch*: `tea.Msg` can be any type, and each `case` matches one. Inside `case tea.KeyMsg`, `msg` *is* a `KeyMsg`, so `msg.String()` works. Tomorrow adds `case tea.WindowSizeMsg`.

**Wrapping around.** `(m.index + 1) % len(m.monsters)` moves right and wraps from the last monster to the first; `%` is remainder. Going left adds `len` before the remainder so the result never goes negative (`-1 % 3` is `-1` in Go, not `2`).

**`tea.Quit`** is a *command*: something for Bubble Tea to do after this update. Returning it ends the program. `nil` means nothing to do.

Because `Update` has a *value* receiver, `m` is a copy; we change the copy and return it, and Bubble Tea keeps the returned one. The model is never mutated in place, which is exactly why this pattern is easy to test.

## View

```go
// View draws the whole screen from the model. It is called after every Update.
func (m model) View() string {
	if m.quitting {
		return ""
	}
	mo := m.current()
	p := paletteFor(mo)
	var b strings.Builder
	b.WriteString(titleStyle.Render("🎃 TERMINAL OF TERROR 🎃") + "\n\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(mo.Emoji+"  "+strings.ToUpper(mo.Name)) + "\n")
	b.WriteString(lipgloss.NewStyle().Italic(true).Foreground(p.accent).Render(mo.Description) + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(p.primary).Render(mo.ASCII) + "\n\n")
	b.WriteString(headingStyle.Render("Terrifying Facts") + "\n")
	for _, f := range mo.Facts {
		b.WriteString(factStyle.Render("  • "+f) + "\n")
	}
	b.WriteString("\n" + metaStyle.Render("First appearance: "+mo.Origin) + "\n\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("Monster %d of %d · ←/→ next · q quit", m.index+1, len(m.monsters))) + "\n")
	return b.String()
}

// RunUI starts the interactive explorer, opening on startID if given.
func RunUI(startID string) error {
	p := tea.NewProgram(newModel(startID), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

`View` looks like last night's card code, and that's the point: drawing is just building a string. Returning `""` when quitting leaves a clean terminal behind.

**`tea.WithAltScreen()`** switches to the terminal's *alternate screen*, the one `vim` and `less` use. Your program gets a blank canvas, and when it exits, the terminal restores whatever was there before. Without it, the explorer would draw over your shell history.

## The command

Change `cmd/monster.go`: `Use: "monster [name]"`, `Args: cobra.MaximumNArgs(1)`, and a `RunE` that resolves the optional name and calls `ui.RunUI(startID)`:

```go
	RunE: func(cmd *cobra.Command, args []string) error {
		startID := ""
		if len(args) == 1 {
			m, err := resolveMonster(args[0])
			if err != nil {
				return err
			}
			startID = m.ID
		}
		return ui.RunUI(startID)
	},
```

Import `ui` and drop the `Println`s. `resolveMonster` stays exactly as it was.

## Run it

```bash
go mod tidy
go run . monster
```

The screen clears, Dracula appears. Press `→` for Frankenstein's Monster, `→` again for the Wolf Man, `→` once more to wrap round. `q` leaves, and your terminal is exactly as you left it. Try `go run . monster wolfman` to open on the Wolf Man.

## Testing a screen without a screen

The best thing about the Elm pattern: `Update` and `View` are ordinary functions, so a test can call them. Create `internal/ui/ui_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestExplorerMovesBetweenMonsters(t *testing.T) {
	var m tea.Model = newModel("")
	if v := m.View(); !strings.Contains(v, "DRACULA") {
		t.Fatalf("should open on Dracula:\n%s", v)
	}
	m, _ = m.Update(key("right"))
	if v := m.View(); !strings.Contains(v, "FRANKENSTEIN") {
		t.Fatalf("right should show the next monster:\n%s", v)
	}
	m, _ = m.Update(key("left"))
	m, _ = m.Update(key("left"))
	if v := m.View(); !strings.Contains(v, "WOLF MAN") {
		t.Fatalf("left from the first should wrap to the last:\n%s", v)
	}
	m, cmd := m.Update(key("q"))
	if cmd == nil || m.View() != "" {
		t.Fatal("q should quit and clear the screen")
	}
}

func TestStartOnANamedMonster(t *testing.T) {
	if v := newModel("wolf-man").View(); !strings.Contains(v, "WOLF MAN") {
		t.Fatalf("expected the Wolf Man:\n%s", v)
	}
}
```

`var m tea.Model = newModel("")` declares `m` as the *interface* type, so the `m, _ = m.Update(...)` line type-checks: `Update` returns a `tea.Model`. A `tea.KeyMsg` is just a struct, so the test makes them by hand. No terminal, no timing, runs in milliseconds:

```bash
go test ./...
```

## Try it

- Add `"home"` and `"end"` keys that jump to the first and last monster.
- Press an unknown key. Nothing happens, because the inner `switch` has no matching case and `Update` returns the model unchanged. Add a `default:` that stores the key name in the model and shows it in the footer, a handy trick when you're not sure what a key is called.
- Remove `tea.WithAltScreen()` and run it. See why it's there.

## 💀 Terrifying fact

`View` runs after *every* message, and Bubble Tea renders at up to 60 frames a second. That's fine, because `View` is pure: same model, same string, and Bubble Tea only writes the lines that differ. What you must never do in `View` is anything slow or with side effects. Reading a file in `View` would read it sixty times a second. Do work in `Update`; only draw in `View`.

## 🕯️ Before dawn

Add a `showArt bool` to the model, toggled by `a`, that hides and shows the portrait. Then write a test for it. Notice how little the test needs to know: send a key, look at the string.

> 📺 *"The séance is under way. The spirits respond to arrow keys, which is more than I can say for most spirits. Tomorrow night we teach them about the size of the room."*

[← Night 10](night-10.md) · [Index](README.md) · [Night 12 →](night-12.md)
