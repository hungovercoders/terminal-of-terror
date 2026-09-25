# Night 20 · Who Lurks in the Fog

> 📺 *"A portrait, mostly hidden. Four names. Guess now and score big, or clear a little fog and play it safe. This is the game every horror host has run since the dawn of television, and it's the first one where the ASCII art earns its keep."*

**Tonight you'll learn**
- Treating ASCII art as a grid of cells
- A random reveal order with `rand.Perm`
- Stages, and points that fall as clues rise
- Reusing the quiz's option line and key mapping
- Deciding a layout for the whole round, not per frame
- Sorting map keys so a seeded game is repeatable

**Where we are:** the Midnight Quiz works. Tonight, the second game.

## Art as data

Every monster has a portrait, twenty or so lines of text. For the game, that's a grid of characters, some of which are shown and some hidden. [`internal/ui/guess.go`](../../internal/ui/guess.go):

```go
// portraitCells lays the art out as a padded grid of runes.
func portraitCells(art string) [][]rune {
	lines := strings.Split(art, "\n")
	w := 0
	for _, l := range lines {
		w = max(w, len([]rune(l)))
	}
	grid := make([][]rune, len(lines))
	for i, l := range lines {
		row := []rune(l)
		for len(row) < w {
			row = append(row, ' ')
		}
		grid[i] = row
	}
	return grid
}
```

`[][]rune` is a slice of slices: a rectangle of characters, every row padded to the same width so a cell can be numbered `y*w + x`. That numbering is the trick: the whole portrait is `total` cells, and a permutation of `0..total-1` is a random order to reveal them in.

## The fog

```go
// Fog stages: how much of the portrait shows, and what a right answer is worth.
var (
	fogReveal = []float64{0.15, 0.3, 0.5, 0.75, 1.0}
	fogPoints = []int{100, 80, 60, 40, 20}
)
```

Five stages. At stage 0 fifteen percent of cells show and a right answer is worth 100; each press of space lifts more fog and costs twenty points. Two slices indexed by the same stage, rather than a slice of structs, because they're read in different places and never together.

When a round is made, its reveal order is fixed once:

```go
		grid := portraitCells(m.ASCII)
		g.order = r.Perm(len(grid) * len(grid[0]))
```

**`r.Perm(n)`** returns the numbers `0..n-1` shuffled. Storing the permutation in the round means that clearing fog *adds* cells to what's showing rather than re-rolling: stage 1 is stage 0 plus more. Drawing takes the first `total * fogReveal[stage]` entries of the order and shows those cells; the rest get a fog character:

```go
func (g GuessRound) fogged(stage int) string {
	grid := portraitCells(g.Monster.ASCII)
	w := len(grid[0])
	total := len(grid) * w
	show := make([]bool, total)
	for _, idx := range g.order[:int(float64(total)*fogReveal[stage])] {
		show[idx] = true
	}
	fog := []rune("░▒░ ░")
	var b strings.Builder
	for y, row := range grid {
		for x, c := range row {
			idx := y*w + x
			if show[idx] {
				b.WriteRune(c)
			} else {
				b.WriteRune(fog[(idx*7+y)%len(fog)])
			}
		}
		if y < len(grid)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
```

`int(float64(total)*fogReveal[stage])` is the one place the code mixes integers and floats, and Go makes you say so at every step: no implicit conversion, ever. The fog pattern `(idx*7+y)%len(fog)` is deterministic noise, textured but not random, so the fog doesn't shimmer between frames; the game has no ticks at all.

`order` is a lower-case field on an exported struct, so a caller outside the package can see a `GuessRound`'s `Monster` and `Options` but not its reveal order. Exporting is per field.

## Choosing the rounds

```go
func NewGuessRounds(r *rand.Rand, all []monsters.Monster, n int) []GuessRound {
	var pool []monsters.Monster
	names := map[string]bool{}
	for _, m := range all {
		names[m.Name] = true
		if m.ASCII != "" {
			pool = append(pool, m)
		}
	}
	if len(names) < minGuessOptions {
		return nil
	}
	var everyName []string
	for name := range names {
		everyName = append(everyName, name)
	}
	sort.Strings(everyName) // map order is random; keep seeded games repeatable
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	n = min(n, len(pool))
	...
```

Only monsters with a portrait can be the answer, but any monster's name can be a wrong option. The `names` map deduplicates; then, because Night 13's terrifying fact said map order is random, the names are sorted before the seeded shuffle picks from them. Without that `sort.Strings`, the same seed would give a different game every run, and Night 30's demo recordings would never match. A test, `TestGuessNeedsEnoughMonsters`, covers the `nil` return: with two monsters there's no game, and the command says so instead of offering a one-option round.

## Points and clues

```go
// Points scored this round.
func (g GuessRound) Points() int {
	if !g.Correct() {
		return 0
	}
	return fogPoints[g.Stage]
}
```

`Stage` is recorded at the moment of the guess, so the score is what the fog was worth *then*, even though the view then clears the fog fully to show the answer. Clues arrive with the fog: from stage 2 the debut year (or the era, for folklore), from stage 3 the origin. `EagleEye()` reports a correct guess at stage 0, which Night 22 turns into a badge.

## Same keys, same lines

The quiz's `optionKey` and `optionLine` are reused unchanged, so `1`-`4` and arrows work the same in both games, and a marked-up answer looks the same. Two games that behave alike are easier to learn than two that are each clever. Space, which the quiz used for "choose", here means "clear fog", because that's the move a player makes most.

## Deciding the layout once

The portrait sits left, the options right, if they fit. But "if they fit" is subtle: the side panel's width changes as clues appear and as options get ticks and crosses. If the layout were decided per frame, clearing fog could flip the screen from side-by-side to stacked mid-round, which is disorienting. So `View` measures the *widest the panel could ever get*:

```go
	// Put the side panel beside the portrait only if its widest possible
	// line fits: every option and the longest status line. Judging the whole
	// round up front means revealing clues can never flip the layout.
	sideW := w - lipgloss.Width(box) - 3
	need := max(lipgloss.Width(fogStatus(len(fogReveal)-1, false)), lipgloss.Width(fogStatus(len(fogReveal)-1, true)))
	for i, o := range g.Options {
		need = max(need, lipgloss.Width(optionLine(i, o, false, true, true, false)))
	}
	sideBySide := sideW >= max(need, 24)
```

It renders the longest status line and every option *in its widest state* just to measure them. Rendering something you'll throw away, to measure it, is normal in terminal layout; strings are cheap. `TestGuessLayoutFits` in [`games_test.go`](../../internal/ui/games_test.go) plays rounds at several widths and checks no line overflows, the same idea as Night 12's test.

## The command

[`cmd/guess.go`](../../cmd/guess.go) is the quiz command's twin: a `--rounds`/`-n` flag, a sanity check, `NewGuessRounds`, `RunGuess`, and `record(...)` with the outcome, including `EagleEye`. It returns early if no round was played, so quitting on the first portrait doesn't count as a game.

## Run it

```bash
go run . guess
```

Fifteen percent of a portrait. Guess, or press space. A right answer through thick fog is 100 points; wait for the origin clue and it's 40. Five rounds.

## The test

```go
func TestGuessPlaythrough(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	rounds := NewGuessRounds(r, monsters.GetAllMonsters(), 3)
	var m tea.Model = newGuessModel(rounds, r)
	m, _ = m.Update(size(100, 40))

	first := m.View()
	m = step(m, " ", " ")
	if m.View() == first {
		t.Error("clearing the fog should change the portrait")
	}
	// Round 1 at stage 2 correct, round 2 at stage 0 correct, round 3 wrong.
	m = step(m, string(rune('1'+rounds[0].Answer)), "enter")
	m = step(m, string(rune('1'+rounds[1].Answer)), "enter")
	m = step(m, string(rune('1'+(rounds[2].Answer+1)%4)), "enter")
	gm := m.(guessModel)
	if got, want := gm.res.Score(), fogPoints[2]+fogPoints[0]; got != want {
		t.Errorf("score = %d, want %d", got, want)
	}
	...
```

Space twice, guess right at stage 2 (60), guess right at stage 0 (100), guess wrong: 160 points, and the test says so in terms of `fogPoints` rather than a magic 160, so retuning the points doesn't break it. `TestFogHidesThePortrait` checks that stage 0 really does hide most cells: it counts them.

## Try it

- Make the fog lift in a spiral from the centre instead of at random: replace `r.Perm` with an order you compute. The rest of the game won't notice.
- Add a stage 5 at `1.0` reveal worth 0 points, so a player can always see the whole thing before guessing.
- What happens with a portrait whose lines contain tabs? `portraitCells` counts a tab as one cell; the terminal draws it as up to eight. Fix `portraitCells` or the data.

## 💀 Terrifying fact

`fogReveal` and `fogPoints` are `var`, not `const`, because Go has no constant slices: a `const` can only be a number, string or boolean. That means any code in the package could modify them at runtime. Nothing does, but the language won't stop it. Some projects wrap such tables in a function that returns a copy; here, the comment and the tests are the guard.

## 🕯️ Before dawn

Add a "hard mode" flag, `--hard`, where the options are eight names instead of four. `optionKey` already understands `1`-`9`, but check `optionLine`'s width assumption in the layout comment ("✔ 1. " takes five columns) and the test at several widths.

> 📺 *"Through the fog, a face. And a score, which is the important thing. Tomorrow night we build the crypt's memory: everything you've earned, written to disk, so it's still there when the sun comes up."*

[← Night 19](night-19.md) · [Index](README.md) · [Night 21 →](night-21.md)
