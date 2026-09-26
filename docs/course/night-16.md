# Night 16 · Silent Picture

> 📺 *"Before the talkies there was Lon Chaney, the Man of a Thousand Faces, and pictures that told their stories in title cards and organ music. The Phantom of the Opera, 1925. The Hunchback of Notre Dame, 1923. Tonight, when one of them is on screen, the terminal goes black and white, the title cards come out, and the film grain flickers. Slowly. It's a long night and the projector gets hot."*

**Tonight you'll learn**
- Letting data drive presentation: one boolean, a whole look
- Extending a palette without touching its callers
- A silent-film intertitle with `DoubleBorder` and centred text
- Animating only when something on screen moves
- Two tick speeds, and a test that checks the slow one

**Where we are:** the explorer has an animated intro and a host.

## One flag in the data

Two of the vault's films are silent. Their JSON says so, in the `film` block:

```json
  "film": {
    "title": "The Phantom of the Opera",
    "year": 1925,
    "silent": true,
    ...
```

and `Monster` has a method for it, in [`monsters.go`](../../internal/monsters/monsters.go):

```go
// IsSilent reports whether the monster's classic film is a silent picture.
func (m Monster) IsSilent() bool {
	return m.Film != nil && m.Film.Silent
}
```

The `m.Film != nil &&` guard matters: `m.Film.Silent` on a folklore monster with no film would dereference a nil pointer and crash. Go evaluates `&&` left to right and stops at the first false, so the guard is enough. Every place that wants to know asks `IsSilent()`; nobody else repeats the nil check.

## The palette goes monochrome

Night 10's `paletteFor` started from defaults and applied the monster's theme. It now has a first branch, in [`styles.go`](../../internal/ui/styles.go):

```go
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
	...
```

Three greys, and a `silent` flag so the drawing code can do more than change colour. Because every screen already draws through `paletteFor`, the list, the fact card, the gallery, the explorer, all went monochrome for the Phantom the moment this branch existed. That's the payoff of Night 10's discipline: styles come from one function, so a new look is one edit.

## Title cards

Silent films told their stories in *intertitles*, text on black between shots, often in a decorated frame. The Phantom's description is drawn as one:

```go
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
```

`Align(lipgloss.Center)` with a `Width` centres each wrapped line. The `❦` (a *fleuron*, the printer's ornament) goes above and below unless the terminal is short, in which case `detailHeader` in [`render.go`](../../internal/ui/render.go) passes `ornaments = false` to save two lines. The same function adds a `▶ SILENT PICTURE · 1925` badge after the name when `p.silent` is set, and reads the year from `mo.Film.Year`, which is safe because `silent` implies a film.

## Grain

A worn reel flickers. `artBox` draws a line of Night 15's `noise` above and below the portrait, when there is one:

```go
func (m model) artBox(mo monsters.Monster, p palette) string {
	art := lipgloss.NewStyle().Foreground(p.primary).Render(mo.ASCII)
	if p.silent && m.grain != "" {
		// Flickering film grain above and below, like a worn reel.
		g := helpStyle.Render(m.grain)
		art = g + "\n" + art + "\n" + g
	}
	...
```

and the tick handler regenerates `m.grain` each frame on a silent page:

```go
		if m.screen == screenDetail && m.current().IsSilent() {
			m.grain = noise(m.rng, artWidth(m.current().ASCII)+4, 1)
		}
```

## Animate only what moves

Here's the design question of the night. The intro ticks every 70ms. If the explorer kept ticking at that rate forever, an idle terminal would redraw fourteen times a second doing nothing, warming a laptop for no reason. And a page that isn't silent has nothing to animate at all.

So the model tracks whether a tick is in flight and whether one is *needed*, in [`ui.go`](../../internal/ui/ui.go):

```go
// needsTick reports whether anything on screen is animated.
func (m model) needsTick() bool {
	if m.screen == screenIntro {
		return !m.introDone()
	}
	return m.screen == screenDetail && m.current().IsSilent()
}

func (m *model) ensureTick() tea.Cmd {
	if m.ticking || !m.needsTick() {
		return nil
	}
	m.ticking = true
	return tick(m.tickEvery())
}
```

`ensureTick` is called after any change that might start an animation: entering a screen, moving to another monster, a tick arriving. It returns a command only if nothing's already scheduled *and* the screen wants one. The `ticking` flag prevents two chains at once, which would double the frame rate; the tick handler sets it back to `false` on arrival, so the chain is exactly one tick long at any moment.

And the two speeds:

```go
const (
	introTick = 70 * time.Millisecond
	grainTick = 600 * time.Millisecond
)

// tickEvery is how often the current screen animates.
func (m model) tickEvery() time.Duration {
	if m.screen == screenIntro {
		return introTick
	}
	return grainTick
}
```

Grain at 600ms is a flicker, not a strobe, and it's under two redraws a second. Move from the Phantom to the Mummy and `needsTick` goes false, the chain ends, and the program goes quiet until the next key. A TUI that's idle should cost nothing; this is how.

## Run it

```bash
go run . monster phantom
```

Greys, a title card, a silent-picture badge, and grain flickering above and below Erik's portrait. `→` to the Hunchback, silent too, then `→` again to the Bride and colour comes back. `go run . list` and notice the Phantom and the Hunchback are grey there too.

## Testing a frame rate

You can't test that something "feels slow", but you can test the numbers:

```go
func TestSilentGrainTicksSlowly(t *testing.T) {
	m := newModel(Options{StartID: "phantom", Seed: 1})
	if !m.needsTick() || m.tickEvery() != grainTick || grainTick < 500*time.Millisecond {
		t.Errorf("silent pages should flicker slowly, got %v", m.tickEvery())
	}
	if i := newModel(Options{Seed: 1}); i.tickEvery() != introTick {
		t.Error("intro should animate at intro speed")
	}
}
```

It pins the design decision: silent pages tick, they tick at the slow rate, and the slow rate is at least half a second. If someone later "fixes" the grain to be smoother by ticking at 50ms, this test explains why not.

## Try it

- Make silent pages sepia instead of grey: `#D2B48C`, `#8B7355`, `#C4A882`.
- Give the intertitle a `ThickBorder()`. Which looks more 1925?
- Set `grainTick` to `50 * time.Millisecond`, open the Phantom, and watch `top` or Activity Monitor. Then put it back.

## 💀 Terrifying fact

Bubble Tea coalesces renders: if `View` produces the same string as last frame, nothing is written to the terminal. So a tick that changes nothing is cheap-ish, but not free, since `View` still runs and still builds the whole screen string. The expensive part of most TUIs is rendering, not logic. Ticking only when something moves is the difference between a tool people leave open and one they close.

## 🕯️ Before dawn

The Wolf Man's film opened in December 1941. Make the *fact card* from Night 10 put a `🎬 1941` in its corner when the monster has a film, using `Film != nil`, and nothing when it doesn't. Small, but it's the nil-pointer habit, and you'll use it on every monster with optional data from here on.

> 📺 *"Black and white, title cards, and a flicker. Tomorrow night we hang every portrait in one long corridor and let you walk down it. Mind the Gill-man; he drips."*

[← Night 15](night-15.md) · [Index](README.md) · [Night 17 →](night-17.md)
