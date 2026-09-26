# Night 4 · A Random Terror

> 📺 *"Every good horror host has a segment where the audience never knows what's coming. Tonight, neither will your program. We're going to teach it to pick a monster at random, and, more importantly, to be honest about how random it really is."*

**Tonight you'll learn**
- `math/rand`: random numbers and why they need a *seed*
- Functions that return two values
- Passing a random source into a function instead of hiding it inside
- The `random` command

**Where we are:** `list` prints three monsters.

## Randomness, and the trick behind it

Computers can't do random. What `math/rand` gives you is a very long, very scrambled sequence of numbers that *looks* random, starting from a **seed**. Same seed, same sequence, every time. That sounds like a flaw and it's actually a gift: on Night 26 we'll use a fixed seed to make the whole program replay exactly, which is how you record demos and reproduce bugs.

For a genuine surprise, seed with something that changes: the current time in nanoseconds.

## A function that returns two things

Add to the bottom of `internal/monsters/monsters.go`:

```go
// RandomFact picks a random monster and one of its facts using r.
func RandomFact(r *rand.Rand) (Monster, string) {
	m := monsters[r.Intn(len(monsters))]
	return m, m.Facts[r.Intn(len(m.Facts))]
}
```

and import the package at the top of the file, under the `package` line:

```go
import "math/rand"
```

Read the function's signature slowly: it takes `r`, a pointer to a random source, and returns *two* values, a `Monster` and a `string`. Multiple return values are everywhere in Go; `value, err` is the most common pair, and you'll write hundreds of them.

`r.Intn(n)` gives a whole number from 0 up to but not including `n`. `len(monsters)` is the number of monsters, so `monsters[r.Intn(len(monsters))]` is a random one, and the same trick picks a fact.

Why take `r` as a parameter instead of making one inside? Because whoever calls `RandomFact` then controls the randomness. A test can pass a source with a known seed and check the exact answer. The demo recorder can pass a fixed seed. The command passes a clock-seeded one. The function itself doesn't need to know or care.

## The `random` command

Create `cmd/random.go`:

```go
package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Get a random monster fact",
	Run: func(cmd *cobra.Command, args []string) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		m, fact := monsters.RandomFact(r)
		fmt.Println("🎃 Random Terror Fact 🎃")
		fmt.Println()
		fmt.Println(m.Name)
		fmt.Println(fact)
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
}
```

`rand.NewSource(seed)` makes the scrambled sequence; `rand.New(source)` wraps it in the handy methods like `Intn`. `time.Now().UnixNano()` is the seed: nanoseconds since 1970, different on every run.

`m, fact := monsters.RandomFact(r)` receives both return values at once.

## Run it

```bash
go run . random
go run . random
```

```
🎃 Random Terror Fact 🎃

Frankenstein's Monster
The creature is never named in the novel; Frankenstein is his creator
```

Run it a few times; the fact changes. Now change the seed to a constant, `rand.NewSource(13)`, and run it again. Same fact every time. Put the clock back afterwards.

## Try it

- Print the monster's `Origin` on a third line.
- Import order: Go groups standard-library imports first, then a blank line, then everything else. `gofmt` sorts each group alphabetically but won't move things between groups; the blank line is yours to add. Get it wrong on purpose and see how it looks.
- `go vet ./...` checks every package in the module for common mistakes (the `...` means "and everything below"). Run it now. It should say nothing, which is good news.

## 💀 Terrifying fact

A slice is three numbers: a pointer to some storage, a length and a capacity. Handing a slice to a function copies those three numbers, not the elements. So `GetAllMonsters()` returns a *view* of the vault's own storage, and anything that sorts or overwrites the returned slice changes the vault for everyone. It's fast and it's a trap. Last night's homework sorted the vault by accident. When that matters, copy first: `out := append([]Monster{}, monsters...)`.

## 🕯️ Before dawn

Give `random` a `--monster` flag so `random --monster dracula` picks a random fact from Dracula only. In `init`, before `AddCommand`, add:

```go
randomCmd.Flags().StringVar(&which, "monster", "", "Only facts about this monster")
```

with `var which string` declared above. In `Run`, if `which` isn't empty, loop over `GetAllMonsters()` for a monster whose name matches (`strings.EqualFold` ignores case) and pick from its facts. Print a helpful message if there's no such monster. On Night 8 this becomes a proper argument with fuzzy matching.

> 📺 *"Chance has entered the building. Tomorrow night the monsters leave the code entirely and move somewhere they can be edited without a compiler. Coffins, essentially."*

[← Night 3](night-03.md) · [Index](README.md) · [Night 5 →](night-05.md)
