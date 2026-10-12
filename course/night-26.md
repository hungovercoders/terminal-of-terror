# Night 26 · Speaking to Machines

> 📺 *"Not every viewer is a person. Some are scripts, some are other programs, and one, I'm told, is a spreadsheet. Tonight every command learns to speak JSON, the whole program learns to take a `--pack` flag, and a single environment variable makes every random thing in it repeatable. Boring? Perhaps. But the machines have been very patient."*

**Tonight you'll learn**
- `--json` on every command, and one helper to print it
- `json.Encoder` versus `json.Marshal`, and `SetEscapeHTML`
- Shaping output types for machines, not screens
- Persistent flags and `PersistentPreRunE`
- `TERMINAL_OF_TERROR_SEED`: reproducible randomness
- Plain output: `--no-color`, `NO_COLOR` and a pipe
- Error codes with hints, and the guard in front of every screen
- A "did you mean" for typos, and shell completion
- What "the Unix way" means for a CLI

**Where we are:** every command prints for people. Tonight, for programs too.

## One printer

`--json` is a persistent flag on the root command, so every command takes it, and they all end up in one place, [`cmd/output.go`](../cmd/output.go):

```go
// stdout is where printJSON writes; tests point it at a buffer.
var stdout io.Writer = os.Stdout

// printJSON writes v to stdout as indented JSON.
func printJSON(v any) error {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
```

`any` is Go's "anything" type (an alias for `interface{}`). `stdout` is a package variable rather than `os.Stdout` written inline so the `cmd` tests can point it at a buffer and read every command's JSON back. `json.NewEncoder` writes straight to a stream, where `json.Marshal` returns bytes for you to print; for output the encoder is the natural fit and adds the trailing newline. **`SetEscapeHTML(false)`** matters: by default Go's JSON escapes `<`, `>` and `&` as `<` and friends, for safety when JSON is embedded in HTML. In a terminal that turns "Abbott & Costello" into "Abbott & Costello", which is correct JSON and ugly. Nothing here goes into HTML, so it's off.

The command side is a two-line early return, in `list`, `monster`, `random`, `quiz`, `guess`, `mash`, `crypt`, `tonight`, `countdown` and `packs`:

```go
		if jsonOut {
			return printJSON(all)
		}
```

Errors go the same way: under `--json`, `Execute` prints one object on stderr, `{"error": {"code": "unknown-monster", "message": "...", "hint": "..."}}`, instead of Cobra's `Error:` line, and the code is what a script branches on. `fail(code, hint, format, ...)` in `output.go` builds such an error; the codes are listed below.

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

`--pack` limits any command to some packs: `quiz --pack folklore`, `mash --pack universal`. Adding it to eleven commands would be eleven copies, so it's a **persistent flag** on the root, inherited by every subcommand, in [`cmd/root.go`](../cmd/root.go):

```go
var rootCmd = &cobra.Command{
	...
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Plain output for a pipe, a log, or anyone who asked: --no-color,
		// NO_COLOR (https://no-color.org) or stdout not being a terminal.
		ui.SetPlain(noColor || os.Getenv("NO_COLOR") != "" || !isTerminal(os.Stdout))
		ensureCommunityPacks()
		if len(packFilter) > 0 {
			if err := monsters.UsePacks(packFilter); err != nil {
				return fail(codeBadPack, "", "%v", err)
			}
		}
		return nil
	},
}

var communityOnce sync.Once

// ensureCommunityPacks loads the community packs once, whether the first
// to ask is a command or a shell completion.
func ensureCommunityPacks() { communityOnce.Do(loadCommunityPacks) }
```

```go
func init() {
	...
	rootCmd.PersistentFlags().StringSliceVar(&packFilter, "pack", nil, "Only use monsters from these packs, e.g. --pack folklore")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Print the command's data as JSON instead of drawing it (errors become a JSON object on stderr)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Plain output: no colour, ASCII borders, no emoji (also set by NO_COLOR, or by piping)")
	_ = rootCmd.RegisterFlagCompletionFunc("pack", completePacks)
}
```

**`PersistentPreRunE`** runs before *any* command's `RunE`, which makes it the place for setup every command needs: deciding whether output is plain (below), loading the community packs (tomorrow) and applying the filter. `ensureCommunityPacks` wraps the loader in a `sync.Once` because shell completion, at the end of tonight, needs the packs too and never runs `PersistentPreRunE`: whoever asks first loads them, and nobody loads them twice. `StringSliceVar` accepts `--pack a,b` or `--pack a --pack b`; `--json` and `--no-color` are registered beside it, persistent for the same reason, and the last line tells the shell what `--pack` can take. `UsePacks` narrows what `GetAllMonsters` returns and errors on an unknown id, so `quiz --pack hammer` says so instead of running an empty quiz.

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
		fmt.Fprintf(os.Stderr, "warning: ignoring %s=%q: it should be a whole number\n", envSeed, s)
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

## Plain output

Colour codes in a log file are noise, and an emoji in `list | grep` is a character nobody typed. So `PersistentPreRunE` decides once, at startup, whether output is *plain*: `--no-color`, the `NO_COLOR` environment variable (a convention many tools share, written up at [no-color.org](https://no-color.org)), or stdout not being a terminal. [`internal/ui/plain.go`](../internal/ui/plain.go) holds the switch:

```go
// SetPlain switches plain output on or off. The root command decides once
// at startup from --no-color, NO_COLOR and whether stdout is a terminal.
func SetPlain(on bool) {
	if on == plain {
		return
	}
	plain = on
	if on {
		colourProfile = lipgloss.ColorProfile()
		lipgloss.SetColorProfile(termenv.Ascii)
		return
	}
	lipgloss.SetColorProfile(colourProfile)
}
```

Setting Lip Gloss's colour profile to ASCII switches every style's colour off at once, so Night 9's styles need no change. The rest is four small helpers in the same file that every renderer calls instead of writing an emoji or a border directly: `glyph("🎃")` is the emoji and a space, or nothing; `label(m.Emoji, m.Name)` is "🦇 Dracula" or "Dracula"; `pick(fancy, plain)` chooses between two titles; and `border(lipgloss.DoubleBorder())` becomes `lipgloss.ASCIIBorder()`, so a card in a log is still a card. The Bubble Tea screens are untouched: they need a terminal, and a terminal can show an emoji.

## No terminal at all

A screen can't go plain; it needs a terminal or nothing. Run from a pipe or CI, the explorer would hang waiting for keys or spray escape codes into the pipe, so each screen command starts with the guard in [`cmd/tty.go`](../cmd/tty.go):

```go
func requireTerminal(what, alternative string) error {
	if isTerminal(os.Stdout) && isTerminal(os.Stdin) {
		return nil
	}
	return fail(codeNoTerminal, alternative, "%s needs a terminal to play on (stdout or stdin is not one)", what)
}
```

`monster`, `quiz` and `guess` call it with what they are and what to use instead, so `quiz < /dev/null` says `the Midnight Quiz needs a terminal to play on (stdout or stdin is not one); use --json for the questions and answers, or play from a shell`. `crypt reset` checks stdin the same way, because it asks a question, and its hint is `-y`. The hint is the second half of every coded error, and the codes are a short list in `output.go`:

```go
// The error codes a script can switch on. They are stable; the messages
// are not.
const (
	codeUnknownMonster   = "unknown-monster"
	codeAmbiguousMonster = "ambiguous-monster"
	codeBadPack          = "bad-pack"
	codeBadDate          = "bad-date"
	codeBadFlag          = "bad-flag"
	codeNoTerminal       = "no-terminal"
	codeStore            = "store"
	codeUsage            = "usage" // anything else: a bad argument, an empty pack
)
```

Eight codes, one for each thing a script might handle differently: retry with a better name, pass a real date, run with a terminal. The comment is a promise. A script switches on `code` and shows `message` to a person, and the wording can improve without breaking anyone. Night 29's smoke test checks `unknown-monster` and the terminal guard on every push.

## A typo gets a guess

`monster draculla` finds nothing by Night 8's substring search. Before giving up, `resolveMonster` in [`cmd/monster.go`](../cmd/monster.go) asks [`internal/monsters/suggest.go`](../internal/monsters/suggest.go) for names within a few edits of the query: `monsters.DidYouMean(monsters.Suggest(query))` returns "Did you mean Dracula?", or "" when nothing is close, and the `unknown-monster` error carries it. `Suggest` compares the query with every id, name and alias by edit distance (`levenshtein`, the twenty-odd-line classic you'll find in most text tools), allows one edit for a short query and up to three for a long one, and offers at most three names. The hard part of a suggestion is knowing when *not* to make one: a one-letter query is inside half the names, so it gets no guess.

## Completing the names

Cobra gives every program a `completion` command that prints a script for bash, zsh, fish or PowerShell, and the shell then calls back into the binary as `terminal-of-terror __complete monster drac` for the choices. The program only has to say what those are, in [`cmd/complete.go`](../cmd/complete.go):

```go
// completePacks offers pack ids to --pack, built-in and community alike.
func completePacks(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ensureCommunityPacks()
	var out []string
	for _, p := range monsters.AllPacks() {
		if strings.HasPrefix(p.ID, strings.ToLower(toComplete)) {
			out = append(out, p.ID+"\t"+p.Name)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
```

`completeMonsters` does the same for monster ids and aliases, with the monster's name after the tab as a description, and every command that takes a monster sets it as its `ValidArgsFunction`. Ids and aliases are single words, so they complete without quoting; the names themselves, with their spaces and apostrophes, would not. Both call `ensureCommunityPacks` first, which is why the loader is wrapped in a `sync.Once`: a completion request skips `PersistentPreRunE`, and a community monster should complete like a built-in one. Night 31 ships the generated scripts in the release archives.

## The Unix way

A well-mannered command-line tool follows a few habits, all of which the program now has:

- **stdout is for output, stderr is for everything else.** Warnings, progress errors, the bad-seed notice: all stderr. `terminal-of-terror list --json | jq` works.
- **Exit status means something.** `RunE` errors exit 1 (Night 8). A script can check.
- **Colour only in a terminal.** The root decides once at startup, above; pauses only in a terminal (Night 23).
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
- Add a `--seed` flag to the root that beats `TERMINAL_OF_TERROR_SEED`. `seed()` in `cmd/seed.go` is the only place to change. Then reread the paragraph on flags versus environment and decide whether the flag earned its place.
- Set `TERMINAL_OF_TERROR_SEED=abc` and run anything. Read the warning. Then make it fatal and decide which you prefer.

## 💀 Terrifying fact

`encoding/json` marshals a `time.Time` as RFC 3339 with nanoseconds and the zone, and *unmarshals* it back exactly, which is how the crypt's `captured` dates survive the round trip. But a `map[string]time.Time` sorts its keys on output, so the JSON is stable across runs, while a `map[time.Time]int` can't be marshalled at all: JSON object keys must be strings. Choose map key types with the file format in mind.

## 🕯️ Before dawn

Write a shell script (or a Python one) that runs `quiz --json -n 5` and prints only the prompts, one per line. Then run it with a seed, twice, and diff the output. That's your first automated check of the program's data format.

> 📺 *"JSON for the machines, a seed for the scripts, and a pack flag for the picky. Tomorrow night we open the doors and let strangers in. Strangers with their own monsters."*

[← Night 25](night-25.md) · [Index](README.md) · [Night 27 →](night-27.md)
