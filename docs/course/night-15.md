# Night 15 · Your Host, Count Cathode

> 📺 *"Every creature feature needs a host. Someone to sit in a cardboard coffin, introduce the picture, and fill the ad breaks with terrible puns. Tonight the program gets one. It also gets a test pattern, a lightning flash and a title card typed out letter by letter, because a show needs an opening."*

**Tonight you'll learn**
- A package for a character: keeping words out of the UI code
- Time in Bubble Tea: `tea.Tick`, and why a timer is a *command*
- Animation as a frame counter
- Random TV static, and a two-row block font
- A typewriter effect with rune slicing
- An `Options` struct instead of a growing list of arguments

**Where we are:** Week 3 begins. From tonight the lessons show the finished project's code and point at the files; build along if you like, or read along.

## The host package

Count Cathode has lines: greetings, sign-offs, reactions to a fact, an introduction for each monster. None of that belongs in the UI code, which should know how to *draw* a quote, not what it says. So there's a tiny package, [`internal/host/host.go`](../../internal/host/host.go):

```go
// Package host gives a voice to Count Cathode, Terminal of Terror's
// late-night horror host, broadcasting from beyond the static.
package host

// Name is the host's on-air name.
const Name = "Count Cathode"

// Channel is the fictional station the host broadcasts on.
const Channel = "Channel 13"

var greetings = []string{
	"Good evening, creatures of the night, and welcome to the Channel 13 Creature Feature. I'm your host, Count Cathode, broadcasting from beyond the static.",
	"Pull up a coffin and dim the lights. The rabbit ears are up, the popcorn is stale, and the monsters are restless.",
	...
}

// Greeting returns a random opening line.
func Greeting(r *rand.Rand) string { return pick(r, greetings) }

// Intro is the host's introduction for a monster.
func Intro(m monsters.Monster) string {
	if m.HostIntro != "" {
		return m.HostIntro
	}
	return fmt.Sprintf("Tonight's feature: %s. Viewer discretion is advised.", m.Name)
}

func pick(r *rand.Rand, lines []string) string {
	if r == nil {
		return lines[rand.Intn(len(lines))]
	}
	return lines[r.Intn(len(lines))]
}
```

Sixty lines, and a whole character. Two habits to notice. **`const`** for the things that never change: a constant can't be reassigned, and the compiler can inline it. And every random choice takes a `*rand.Rand` (Night 4's lesson): the intro test can pass a seeded one and get the same greeting every run.

`Intro` uses the `hostIntro` field you copied into the JSON on Night 13, with a fallback for monsters that don't have one. Every monster page now shows it under the description, in `hostStyle`, a new dim italic in [`styles.go`](../../internal/ui/styles.go).

## Time is a message

Bubble Tea has no timers, loops or threads that you write. What it has is **commands**: a `tea.Cmd` is a function that does some work, possibly slowly, and returns a message. Bubble Tea runs it in the background and feeds the message into `Update` like any key press. `tea.Quit` was one. `tea.Tick` is another:

```go
type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}
```

`tea.Tick(d, f)` waits `d`, then calls `f` with the time and delivers what it returns. `tickMsg` is our own type (`type tickMsg time.Time`) so the type switch in `Update` can recognise it. A `time.Duration` is a number of nanoseconds with a type, and `70 * time.Millisecond` reads as what it is.

One tick is one frame. To animate, `Update` handles `tickMsg` by advancing a counter and asking for the *next* tick:

```go
	case tickMsg:
		m.ticking = false
		m.frame++
		if m.screen == screenIntro && m.frame < staticFrames {
			m.noise = noise(m.rng, min(max(m.contentWidth()-4, 20), 70), 9)
		}
		...
		return m, m.ensureTick()
```

and `Init`, which was `nil` until now, returns the first one:

```go
func (m model) Init() tea.Cmd {
	if m.ticking {
		return tick(m.tickEvery())
	}
	return nil
}
```

Each tick schedules one more, until `ensureTick` decides the screen no longer needs one (tomorrow's topic). Nothing is running in a loop anywhere; there's just a chain of single-shot timers, each delivered as a message. `View` still only draws `m.frame`, which is what makes an animated screen as testable as a still one: a test calls `Update(tickMsg{})` twenty times and looks at the string.

## The timeline

The intro in [`internal/ui/intro.go`](../../internal/ui/intro.go) is a schedule in frames:

```go
// Intro timeline, in ticks.
const (
	staticFrames = 18 // TV static while the set warms up
	flashFrames  = 2  // a lightning flash
	typeSpeed    = 3  // characters of the greeting typed per tick
)
```

At 70ms a tick, that's 1.3 seconds of static, a two-frame flash, then the greeting typed at about 40 characters a second. The intro is done when the greeting is fully typed:

```go
func introDoneFrame(greeting string) int {
	n := len([]rune(greeting))
	return staticFrames + flashFrames + (n+typeSpeed-1)/typeSpeed
}
```

`(n+typeSpeed-1)/typeSpeed` is integer division rounded *up*: 100 characters at 3 per frame is 34 frames, not 33. Integer division in Go truncates, so this trick comes up whenever you're counting pages or frames.

## Static

```go
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
```

A grid of random characters, weighted towards spaces and dots so it looks like static rather than a wall. It's regenerated on each tick for the first 18 frames, so it flickers. That's the whole trick: animation is a picture that's different next frame.

## A two-row font

The title card needs a logo, and there's no font library. There's a map:

```go
var glyphs = map[rune][2]string{
	'T': {"▀█▀", " █ "},
	'E': {"█▀▀", "██▄"},
	'R': {"█▀█", "█▀▄"},
	...
}
```

Ten letters, exactly the ones in TERMINAL OF TERROR, each two rows tall. `bigText` joins the top rows of each letter with a space, then the bottom rows, and `logo(width)` picks the full title, the title stacked on two lines, or plain text, depending on how wide the terminal is. Half-block characters (`▀ █ ▄`) are the pixel art of the terminal: each is one cell, two "pixels" tall.

## The typewriter

```go
	typed := m.frame - staticFrames - flashFrames
	greeting := []rune(m.greeting)
	shown := min(len(greeting), max(typed, 0)*typeSpeed)
	text := string(greeting[:shown])
	if shown < len(greeting) {
		text += "▌"
	}
```

Frame number in, prefix of the greeting out, with a block cursor while it's still going. The greeting is sliced as *runes* (Night 8 again) so a curly apostrophe doesn't come out as half a character mid-type. The lightning flash is two frames where the logo style is swapped for black on white:

```go
	logoStyle := lipgloss.NewStyle().Foreground(colorBlood).Bold(true)
	if m.frame < staticFrames+flashFrames {
		// Lightning: the whole card flashes white for a moment.
		logoStyle = logoStyle.Foreground(colorBlack).Background(colorWhite)
	}
```

Then `viewIntro` centres everything with `lipgloss.NewStyle().Width(w).Align(lipgloss.Center)`, and once `introDone()` says the typing is over, adds "Press any key to enter the vault...". Any key during the typing skips to the end; any key after it enters:

```go
		if m.screen == screenIntro {
			if msg.String() == "q" {
				return m.quit()
			}
			if !m.introDone() {
				m.frame = introDoneFrame(m.greeting)
				return m, nil
			}
			m.enter(m.next)
			return m, m.ensureTick()
		}
```

That's [`ui.go`](../../internal/ui/ui.go), which now has a `screen` field (`screenIntro`, `screenGallery`, `screenDetail`, an `iota` enum like the tabs) and dispatches at the top of `Update` on which screen is showing, the way Night 14's `searching` flag did.

## Options, not arguments

`RunUI(startID string)` was fine with one argument. Now there's a start id, a "show the gallery" flag, "skip the intro", a seed and a clock. Five positional arguments is a bug waiting to happen (`RunUI("", true, false, 0, now)`: which is which?). Go's answer is an options struct:

```go
// Options configures the monster explorer.
type Options struct {
	ShowAll bool   // open on the gallery of every monster
	StartID string // open on this monster's page
	NoIntro bool   // skip the Channel 13 opening
	Seed    int64  // random seed; 0 means use the clock
	Now     time.Time
}
```

Callers name what they set, `ui.RunUI(ui.Options{StartID: m.ID, NoIntro: noIntro})`, and everything else is its zero value, which is always the sensible default. That's why the fields are phrased as `NoIntro` rather than `Intro`: `false` has to mean "the normal thing". The `monster` command gained `--all` and `--no-intro` flags that map straight onto them ([`cmd/monster.go`](../../cmd/monster.go)). And when a monster is named on the command line, the intro is skipped without being asked: you know what you want.

## Sign-off

When the viewer quits, the host gets the last word. `quit()` picks a sign-off, and `RunUI` prints it *after* the program ends, once the alternate screen has gone and the normal terminal is back:

```go
	final, err := p.Run()
	...
	fm, ok := final.(model)
	if fm.signOff != "" {
		fmt.Println(hostStyle.Render("📺 " + fm.signOff))
	}
```

`p.Run()` returns the final model as a `tea.Model`; `final.(model)` is a **type assertion**, the single-case cousin of the type switch, getting our concrete `model` back so its fields can be read. This is how a TUI hands results to the command that ran it: the whole final state comes out the end. Tomorrow's games rely on the same trick to return a score.

## Run it

```bash
go run . monster
```

Static, the Channel 13 label, a flash, the title, and Count Cathode typing. Press a key to skip, another to enter. `q` on the way out gets you a sign-off. `go run . monster --no-intro` skips it; `go run . monster dracula` skips it too.

## Testing time

The intro tests in [`ui_test.go`](../../internal/ui/ui_test.go) never wait 70 milliseconds for anything:

```go
func TestIntroSkipsThenEnters(t *testing.T) {
	m := send(newModel(Options{Seed: 1}), size(100, 30))
	if m.screen != screenIntro {
		t.Fatal("expected intro")
	}
	m = send(m, key("x"))
	if !m.introDone() || !strings.Contains(m.View(), "Press any key") {
		t.Fatal("first key should finish the intro")
	}
	m = send(m, key("x"))
	if m.screen != screenDetail {
		t.Fatalf("second key should enter the explorer, got %v", m.screen)
	}
}
```

`send` is a helper that feeds messages through `Update` and hands back the concrete model. `Seed: 1` fixes the greeting, so the test is the same every run. Ticks are just `tickMsg{}` values a test can send by hand, as many as it likes, instantly.

## Try it

- Change `typeSpeed` to 1 and `introTick` to 30ms. Which feels more like a 1950s TV?
- Add a letter to `glyphs` (`'S'` is `{"█▀▀", "▄██"}`) and change the logo to something of your own.
- The greeting is chosen at random, but on Halloween it's fixed. Find where in `newModel`, and what package decides it's Halloween. That's Night 24.

## 💀 Terrifying fact

A `tea.Cmd` runs in its own goroutine, Go's lightweight thread, and Bubble Tea may run several at once. That's why a command must never touch the model: it returns a *message*, and `Update`, which runs one message at a time, does the changing. Break that rule and you get the class of bug Go's race detector exists for; CI runs the tests with `-race` on Night 29 for exactly this reason.

## 🕯️ Before dawn

Add a *second* animation to the intro: after the greeting, make the "Press any key" line blink by showing it only on even frames (`m.frame%2 == 0`). You'll need `needsTick` to keep ticking after `introDone()`. Then decide whether it's better or just annoying, and set it accordingly.

> 📺 *"And that's how a host gets on the air: static, lightning, and a typewriter. Tomorrow night's picture has no sound at all. It's from 1925, and it flickers."*

[← Night 14](night-14.md) · [Index](README.md) · [Night 16 →](night-16.md)
