# Night 28 · Haunting the Tests

> 📺 *"You've written tests every night since the seventh. Tonight we step back and look at the whole graveyard: what's buried where, which tests earn their keep, and how to test the things that only go wrong in a real terminal. Then the race detector, which finds the ghosts that only appear when two things happen at once."*

**Tonight you'll learn**
- The shape of the project's test suite, package by package
- Testing a Bubble Tea model: `send`, `step` and the size sweep
- Seeing what a test sees: `DUMP=1`
- What *not* to unit test, and how to check it anyway
- The race detector
- Table stakes: `gofmt`, `go vet`, `go mod tidy`

**Where we are:** the program is complete. Three nights of making it *shippable* begin.

## The graveyard

Run the whole suite and read the packages:

```bash
go test ./...
```

```
ok   internal/calendar   known moons, countdowns, Friday the 13th, notes
ok   internal/crypt      capture rules, every badge
ok   internal/mash       a hundred bouts
ok   internal/monsters   every monster's data, the loader, search, packs
ok   internal/quiz       fifty seeds of questions, the mask
ok   internal/store      round trip, corrupt file
ok   internal/ui         every screen at every size, both games played through
```

Each package tests its own job at its own level. Nothing in `crypt` touches a file; nothing in `quiz` renders a screen; nothing in `ui` reads the disk except through `TERMINAL_OF_TERROR_HOME`. That layering is the reason the suite runs in about a second and the reason a failure points at a package rather than at "the program".

Three kinds of test carry most of the weight, and you've written all three:

1. **Data tests** (Night 7). Every monster has what the screens and games need. The most valuable test in a data-driven program is the one that validates the data, because most contributions *are* data.
2. **Property tests over seeds** (Nights 18, 23). Random generators checked for invariants across many seeds, so "some seed produces a bad question" is found before a player finds it.
3. **Model tests** (Nights 11 to 20). Send messages, look at strings.

## Testing a Bubble Tea model, properly

The explorer tests in [`ui_test.go`](../../internal/ui/ui_test.go) grew two helpers that every test uses:

```go
func send(m model, msgs ...tea.Msg) model {
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(model)
	}
	return m
}
```

`send` takes the *concrete* model and returns it, so tests can read `m.screen` and `m.cursor` directly instead of scraping the view. Assert on state when the state is the point (which screen are we on?), and on the rendered string when the rendering is the point (is the name visible?). Both are legitimate; mixing them up makes brittle tests.

The size sweep from Night 12 became the suite's workhorse:

```go
func TestEveryPageRenders(t *testing.T) {
	for _, sz := range [][2]int{{40, 20}, {80, 24}, {140, 45}, {0, 0}} {
		m := send(newModel(Options{NoIntro: true, Seed: 1}), size(sz[0], sz[1]))
		for range monsters.GetAllMonsters() {
			mo := m.current()
			for range tabsFor(mo) {
				v := m.View()
				if !strings.Contains(v, strings.ToUpper(mo.Name)) {
					t.Fatalf("%dx%d %s tab %d: name missing", sz[0], sz[1], mo.ID, m.tab)
				}
				if sz[1] > 0 && strings.Count(v, "\n")+1 > sz[1] {
					t.Errorf("%dx%d %s tab %d: %d lines overflow the screen", ...)
				}
				m = send(m, key("r"), key("end"), key("tab"))
			}
			m = send(m, key("right"))
		}
	}
}
```

Every monster, every page, revealed and scrolled to the end, at four sizes including "unknown". A new monster with a long film note, a new page type, a wider portrait: all covered without a new test. When you add a screen, add it to a sweep like this before writing a test for its specific behaviour.

## Seeing what the test sees

A failing layout test prints a screen, but sometimes you want to *look* before anything fails. `TestDumpScreens` is skipped unless asked:

```go
func TestDumpScreens(t *testing.T) {
	if os.Getenv("DUMP") == "" {
		t.Skip("set DUMP=1 to print screens")
	}
	w, h := 120, 40
	m := send(newModel(Options{Seed: 3}), size(w, h))
	for i := 0; i < 12; i++ {
		m = send(m, tickMsg{})
	}
	t.Log("\n" + m.View())
	...
```

```bash
DUMP=1 go test ./internal/ui -run Dump -v
```

prints the intro mid-static, the detail page, the gallery, the Phantom with grain, and a revealed myths page, as the test renders them: no terminal, no colours. It's the fastest way to check a layout change across several screens, and it's how the layouts in this course were checked while writing. A skipped-by-default test that dumps output is a debugging tool that lives with the code instead of in someone's shell history.

`t.Log` output only shows with `-v`, and `-run Dump` selects tests by name with a regular expression.

## What the tests can't see

Model tests never touch a terminal, which is their strength and their gap. Three things only show up for real:

- **Emoji widths.** `lipgloss.Width("🎃")` is 2, but a terminal with an old Unicode table draws some emoji one column wide, and a border shifts. (Night 9's terrifying fact, and the reason `render.sh` on Night 30 forces Unicode 11 in the recorder.)
- **Alt-screen and restore.** Does `q` leave the terminal exactly as it found it? Does `ctrl+c` mid-intro?
- **Timing.** Does the intro *feel* right at 70ms?

For these, the check is a person at a terminal, and the project makes that cheap: `go run . monster --no-intro`, `TERMINAL_OF_TERROR_SEED=13` for a repeatable run, and `--date` for any night of the year. When you must automate it, `script` (on Linux and macOS) or a Python `pty` can run the real binary under a pseudo-terminal and capture what it wrote; Charm's own `teatest` package (in `charmbracelet/x/exp/teatest`) drives a Bubble Tea program end to end and compares golden output. Neither is in this project, because the model tests cover the logic and a human covers the feel. Know that the tools exist for the day the balance changes.

## The race detector

```bash
go test -race ./...
```

Night 15's terrifying fact: commands run in goroutines. A `tea.Cmd` that touched the model, or two tests sharing a package-level variable, would be a *data race*: two goroutines using the same memory, one of them writing, with nothing ordering them. Races don't fail reliably; they fail on the CI machine at 3am. The race detector instruments every memory access and reports the first race it sees, with both stack traces. It makes tests slower, which is why CI runs it on one platform and not three, but it should run *somewhere* on every change.

The project's models are values, its commands return messages, and its package-level state (`monsters`, `allPacks`) is written once at startup and in `AddPacks`, before any goroutine exists. `-race` passes. Keep it that way: if a future test runs the loader in parallel with `t.Parallel()`, the detector will say so.

## Table stakes

Three checks that aren't tests but that CI runs before them:

```bash
gofmt -l .        # lists files that aren't formatted; empty is good
go vet ./...      # suspicious constructs: wrong Printf verbs, unreachable code, copied locks
go mod tidy       # then: git diff --exit-code go.mod go.sum
```

`gofmt` ends every argument about formatting by having none. `go vet` catches the class of bug the compiler allows and nobody means: `fmt.Printf("%d", "three")`. `go mod tidy` keeps the module file honest so a fresh clone builds the same thing. All three take a second, and all three are in the pre-push habit of anyone who's been bitten once.

## Run it

```bash
gofmt -l . && go vet ./... && go test -race ./...
DUMP=1 go test ./internal/ui -run Dump -v | head -60
```

## Try it

- Make a data change that should fail the data test (delete a monster's `description`) and read the failure. Is it clear which file to fix?
- Introduce a race on purpose: a package-level counter incremented inside a `tea.Cmd`. Run with and without `-race`.
- Add `{30, 10}` to the size sweep. Something will overflow. Decide whether the fix belongs in the layout or in the test's list of supported sizes. (The README promises 40 columns, not 30.)

## 💀 Terrifying fact

`go test` caches results. Run the same tests twice on unchanged code and the second run prints `(cached)` and takes no time. It's keyed on the source, the flags and the *environment variables the test reads*, which is why `DUMP=1` re-runs the dump test but doesn't invalidate the others. If a test reads a file the cache doesn't know about, it can be cached wrongly; `-count=1` forces a real run when you're suspicious.

## 🕯️ Before dawn

Pick the test you'd delete first if you had to delete one, and the one you'd keep if you could keep only one. Write a sentence for each about why. Tomorrow, the whole suite runs on three operating systems whether you like it or not.

> 📺 *"Every ghost accounted for, every race run down. Tomorrow night the tests leave your machine and go to work in the cloud, on three operating systems, every time anyone touches the code. The Night Watch."*

[← Night 27](night-27.md) · [Index](README.md) · [Night 29 →](night-29.md)
