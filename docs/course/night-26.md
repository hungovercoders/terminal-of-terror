# Night 26 · Speaking to Machines

> 📺 *"Not every viewer is a person. Some are scripts, some are other programs, and one, I'm told, is a spreadsheet. Tonight every command learns to speak JSON, the whole program learns to take a `--pack` flag, and a single environment variable makes every random thing in it repeatable. Boring? Perhaps. But the machines have been very patient."*

**Tonight you'll learn**
- `--json` on every command, and one helper to print it
- `json.Encoder` versus `json.Marshal`, and `SetEscapeHTML`
- Shaping output types for machines, not screens
- Persistent flags and `PersistentPreRunE`
- `TERMINAL_OF_TERROR_SEED`: reproducible randomness
- What "the Unix way" means for a CLI

**Where we are:** every command prints for people. Tonight, for programs too.

## One printer

Every command with data to offer has a `--json` flag, and they all end up in one place, [`cmd/output.go`](../../cmd/output.go):

```go
// printJSON writes v to stdout as indented JSON.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
```

`any` is Go's "anything" type (an alias for `interface{}`). `json.NewEncoder` writes straight to a stream, where `json.Marshal` returns bytes for you to print; for output the encoder is the natural fit and adds the trailing newline. **`SetEscapeHTML(false)`** matters: by default Go's JSON escapes `<`, `>` and `&` as `<` and friends, for safety when JSON is embedded in HTML. In a terminal that turns "Abbott & Costello" into "Abbott & Costello", which is correct JSON and ugly. Nothing here goes into HTML, so it's off.

The command side is a two-line early return, in `list`, `monster`, `random`, `quiz`, `crypt`, `packs`:

```go
		if listJSON {
			return printJSON(all)
		}
```

## Shapes for machines

Some commands print their internal type as-is: `list --json` gives the full `Monster` records, with every JSON tag you've been adding since Night 5. That's why the tags were worth getting right: `hostIntro`, `releaseDate`, `omitempty` on the optional fields, all of it is now the program's public data format.

Others define a small type just for output:

```go
// factJSON is the machine-readable shape of a single fact.
type factJSON struct {
	ID      string   `json:"id"`
	Monster string   `json:"monster"`
	Fact    string   `json:"fact"`
	Notes   []string `json:"notes,omitempty"`
}
```

`random --json` prints `{"id": "dracula", "monster": "Dracula", "fact": "...", "notes": [...]}`, not the entire monster. A machine that asked for a fact wants a fact. Deciding the *shape* of output is the design work; the encoding is free. Similarly `packs --json` lists monster *ids* per pack, and `quiz --json` prints the questions with answers, which is how Night 18 let you inspect the generator, and how CI smoke-tests the quiz without a terminal.

## A flag on every command

`--pack` limits any command to some packs: `quiz --pack folklore`, `mash --pack universal`. Adding it to eleven commands would be eleven copies, so it's a **persistent flag** on the root, inherited by every subcommand, in [`cmd/root.go`](../../cmd/root.go):

```go
var rootCmd = &cobra.Command{
	...
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loadCommunityPacks()
		if len(packFilter) > 0 {
			return monsters.UsePacks(packFilter)
		}
		return nil
	},
}

func init() {
	...
	rootCmd.PersistentFlags().StringSliceVar(&packFilter, "pack", nil, "Only use monsters from these packs, e.g. --pack folklore")
}
```

**`PersistentPreRunE`** runs before *any* command's `RunE`, which makes it the place for setup every command needs: loading the community packs (tomorrow) and applying the filter. `StringSliceVar` accepts `--pack a,b` or `--pack a --pack b`. `UsePacks` narrows what `GetAllMonsters` returns and errors on an unknown id, so `quiz --pack hammer` says so instead of running an empty quiz.

There's a subtlety in what the filter *doesn't* apply to: the crypt's badges use `monsters.EveryMonster()`, which ignores the filter, so "capture every monster" means every monster, not every monster in tonight's selection. A global filter needs a way to see past it; that's the second function.

## Repeatable randomness

```go
// envSeed fixes every random choice (questions, portraits, fights, host
// lines) so a run can be repeated exactly. docs/demos/render.sh sets it so
// re-recording the README demos gives the same games every time.
const envSeed = "TERMINAL_OF_TERROR_SEED"

// seed returns the seed from TERMINAL_OF_TERROR_SEED, or the clock.
func seed() int64 {
	if s := os.Getenv(envSeed); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		fmt.Fprintf(os.Stderr, "⚠️  ignoring %s=%q: it should be a whole number\n", envSeed, s)
	}
	return time.Now().UnixNano()
}

// newRand returns a random source seeded by seed().
func newRand() *rand.Rand {
	return rand.New(rand.NewSource(seed()))
}
```

This is where Night 4's decision, "every random function takes a `*rand.Rand`", becomes a feature. Because the quiz, the fog, the fight and the host all draw from a `newRand()`, one environment variable makes the *entire program* deterministic. `TERMINAL_OF_TERROR_SEED=13 terminal-of-terror quiz` asks the same questions every time. Night 30's demo recordings depend on it; a bug report with a seed in it can be reproduced exactly.

A bad value is *warned about and ignored*, not fatal: someone with a typo in their environment should still get a working program. The warning goes to stderr, so `--json` output stays clean.

Why an environment variable and not a `--seed` flag? Because it's not something a person chooses per command; it's something a *script* sets once for a whole recording or test run. Flags for choices, environment for context.

## The Unix way

A well-mannered command-line tool follows a few habits, all of which the program now has:

- **stdout is for output, stderr is for everything else.** Warnings, progress errors, the bad-seed notice: all stderr. `terminal-of-terror list --json | jq` works.
- **Exit status means something.** `RunE` errors exit 1 (Night 8). A script can check.
- **Colour only in a terminal.** Lip Gloss handles it (Night 9); pauses only in a terminal (Night 23).
- **Read the environment for context**, flags for choices, arguments for the thing itself.
- **`--json` when there's data**, so the tool can be a building block.

None of it is hard. All of it is easy to forget, and the difference shows the first time someone tries to use your tool from a script.

## Run it

```bash
go run . random --json
go run . list --json | head -30
go run . quiz --json --pack folklore -n 2
TERMINAL_OF_TERROR_SEED=13 go run . random
TERMINAL_OF_TERROR_SEED=13 go run . random
```

The last two print the same card. The CI smoke test on Night 29 runs most of these.

## Try it

- Pipe `list --json` through `jq '.[] | select(.film.year < 1930) | .name'` (or a Python one-liner) and get the silent-era names.
- Add `--json` to `countdown`. Decide the shape first: nights, moon phase, next anniversary.
- Set `TERMINAL_OF_TERROR_SEED=abc` and run anything. Read the warning. Then make it fatal and decide which you prefer.

## 💀 Terrifying fact

`encoding/json` marshals a `time.Time` as RFC 3339 with nanoseconds and the zone, and *unmarshals* it back exactly, which is how the crypt's `captured` dates survive the round trip. But a `map[string]time.Time` sorts its keys on output, so the JSON is stable across runs, while a `map[time.Time]int` can't be marshalled at all: JSON object keys must be strings. Choose map key types with the file format in mind.

## 🕯️ Before dawn

Write a shell script (or a Python one) that runs `quiz --json -n 5` and prints only the prompts, one per line. Then run it with a seed, twice, and diff the output. That's your first automated check of the program's data format.

> 📺 *"JSON for the machines, a seed for the scripts, and a pack flag for the picky. Tomorrow night we open the doors and let strangers in. Strangers with their own monsters."*

[← Night 25](night-25.md) · [Index](README.md) · [Night 27 →](night-27.md)
