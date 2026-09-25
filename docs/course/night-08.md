# Night 8 · Speak Its Name

> 📺 *"Every monster answers to more than one name. Dracula is the Count; the Wolf Man is poor Larry Talbot; and I've been called things I won't repeat on air. Tonight we teach the program to find a monster from whatever a viewer types, misspellings and all."*

**Tonight you'll learn**
- Normalising text so "The Wolf-Man" and "wolfman" match
- Aliases in the data
- Exact lookup, then partial matching, then candidates
- Command arguments with `cobra.ExactArgs`
- `RunE`: returning errors that Cobra prints for you
- Pointers, for real this time

**Where we are:** Week 2 begins. The vault loads from a pack; tests pass.

## What people actually type

`terminal-of-terror monster dracula` is easy. But people will type `Dracula`, `"the count"`, `wolfman`, `wolf-man`, `Wolf Man` and `frank`. A good tool meets them halfway. Our plan:

1. **Normalise** both the query and every name: lowercase, drop a leading "the", keep only letters and digits, squash the rest to single spaces.
2. Look for an **exact** match against each monster's id, name and aliases.
3. Failing that, look for names that **contain** the query. One hit is a match; several is ambiguous, and we say so.

## Aliases

Add an `Aliases` field to `Monster` in `internal/monsters/monsters.go`, right after `Name`:

```go
	Aliases     []string `json:"aliases,omitempty"`
```

and give each monster some, in its JSON file. For `wolf-man.json`:

```json
{
  "name": "The Wolf Man",
  "aliases": ["Wolfman", "Larry Talbot"],
  ...
```

Dracula gets `["Count Dracula", "The Count"]`; Frankenstein's Monster gets `["Frankenstein", "The Monster"]`.

## Normalising

Add to the bottom of `monsters.go`:

```go
// normalize lowercases and drops a leading "the" and punctuation so that
// "the wolf-man" and "Wolf Man" match.
func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "the ")
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
```

Ranging over a string gives you its characters one at a time as **runes** (Go's name for a Unicode code point). The `switch` with no value is Go's way of writing an if-else chain: the first `case` whose condition is true runs. Letters and digits are kept; spaces, dashes and underscores become spaces; everything else (apostrophes, dots) vanishes. `strings.Fields` splits on any run of spaces, and joining with a single space tidies up. `"The Wolf-Man"` becomes `wolf man`; so does `"wolf   man"`.

`strings.Builder` is how you build a string piece by piece. Adding to a string with `+` in a loop makes a new string each time; a Builder grows in place.

## Exact lookup

```go
// GetMonsterByName returns a monster by id, name or alias (case-insensitive)
func GetMonsterByName(name string) *Monster {
	key := normalize(name)
	for i := range monsters {
		m := &monsters[i]
		if normalize(m.ID) == key || normalize(m.Name) == key {
			return m
		}
		for _, a := range m.Aliases {
			if normalize(a) == key {
				return m
			}
		}
	}
	return nil
}
```

This returns `*Monster`, a **pointer** to a monster: the address of the real one in the vault rather than a copy. Two reasons. First, the function needs a way to say "not found", and `nil` (no pointer) is that way; a plain `Monster` has no such value. Second, callers can read `m.Name` through a pointer exactly as if it were the struct; Go dereferences for you.

Notice `for i := range monsters` with `m := &monsters[i]`. Ranging with `_, m` would give a copy, and `&m` would be the address of the copy. Taking `&monsters[i]` points into the slice itself.

## Partial matching

```go
// Find looks a monster up by exact name first, then by partial match. When
// the query is ambiguous or unknown it returns nil plus any candidates.
func Find(query string) (*Monster, []Monster) {
	if m := GetMonsterByName(query); m != nil {
		return m, nil
	}
	key := normalize(query)
	if key == "" {
		return nil, nil
	}
	var matches []Monster
	for _, m := range monsters {
		hay := []string{m.ID, m.Name}
		hay = append(hay, m.Aliases...)
		for _, h := range hay {
			if strings.Contains(normalize(h), key) {
				matches = append(matches, m)
				break
			}
		}
	}
	if len(matches) == 1 {
		return &matches[0], nil
	}
	return nil, matches
}
```

`Find` has three outcomes, and the two return values express them: a monster and no candidates (found), no monster and several candidates (ambiguous), or neither (unknown). `break` leaves the inner loop as soon as one of a monster's names matches, so a monster is never added twice.

## The `monster` command

Create `cmd/monster.go`. For tonight it prints a plain page; on Night 11 the same command opens the interactive explorer.

```go
package cmd

import (
	"fmt"
	"strings"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var monsterCmd = &cobra.Command{
	Use:   "monster <name>",
	Short: "Meet a monster",
	Long: `Show everything the vault knows about one monster.

Partial names and nicknames work too, e.g. "dracula", "wolfman" or "the count".`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := resolveMonster(args[0])
		if err != nil {
			return err
		}
		fmt.Println(strings.ToUpper(m.Name))
		fmt.Println(m.Description)
		fmt.Println()
		fmt.Println(m.ASCII)
		fmt.Println()
		fmt.Println("Terrifying facts:")
		for _, f := range m.Facts {
			fmt.Println("  •", f)
		}
		fmt.Println()
		fmt.Println("First appearance:", m.Origin)
		return nil
	},
}

// resolveMonster turns a user-supplied name into a monster, with a helpful
// error listing candidates when the name is ambiguous or unknown.
func resolveMonster(query string) (*monsters.Monster, error) {
	m, candidates := monsters.Find(query)
	if m != nil {
		return m, nil
	}
	if len(candidates) > 1 {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = c.Name
		}
		return nil, fmt.Errorf("%q could be any of: %s", query, strings.Join(names, ", "))
	}
	return nil, fmt.Errorf("no monster called %q lurks here; try 'terminal-of-terror list'", query)
}

func init() {
	rootCmd.AddCommand(monsterCmd)
}
```

Three Cobra features:

- **`Args: cobra.ExactArgs(1)`** makes Cobra insist on exactly one argument and produce the error itself otherwise. There's also `MaximumNArgs`, `NoArgs` and more.
- **`RunE`** instead of `Run`: the function returns an error, and Cobra prints it as `Error: ...` and makes `Execute` return it, so the program exits with status 1. Your command code never calls `os.Exit` or prints errors itself; it just returns them.
- **`fmt.Errorf`** builds an error from a format string. `%q` prints a string in quotes, which makes `"an"` stand out in the message.

`make([]string, len(candidates))` creates a slice of a known length up front, so `names[i] = ...` can fill it by position.

## Run it

```bash
go run . monster "the count"
go run . monster wolfman
go run . monster fran
go run . monster an
go run . monster zombie
```

The first three find Dracula, the Wolf Man and Frankenstein's Monster. Then:

```
Error: "an" could be any of: Frankenstein's Monster, The Wolf Man
Error: no monster called "zombie" lurks here; try 'terminal-of-terror list'
```

Both exit with status 1. Because of Night 2's `SilenceUsage`, that one line is all you see.

## Try it

- `go run . monster` with no name. Cobra's `ExactArgs` complains for you.
- Write a test for `Find` in `monsters_test.go`: a table of queries and the id you expect, looped with `t.Run`. Include one ambiguous query and check it returns two candidates. The repo's version is `TestFind` in [`monsters_test.go`](../../internal/monsters/monsters_test.go).
- What does `normalize("Dr. Jekyll & Mr. Hyde")` return? Work it out, then check with a test.

## 💀 Terrifying fact

A `string` in Go is bytes, and one character can be several bytes: `é` is two, `🧛` is four. `len("🧛")` is 4. Ranging over a string, as `normalize` does, walks it rune by rune, which is what you want. Indexing it, `s[0]`, gives you a byte, which usually isn't. When in doubt, convert: `[]rune(s)` is one element per character.

## 🕯️ Before dawn

Make `random` from Night 4 take an optional monster name, `random [monster]`, with `cobra.MaximumNArgs(1)`, resolved with `resolveMonster`. When a name is given, pick from that monster's facts only. `resolveMonster` is in the same package, so `random.go` can call it directly.

> 📺 *"Say its name and it appears. That's how it works with monsters, and with search functions. Tomorrow night: colour. Mostly red."*

[← Night 7](night-07.md) · [Index](README.md) · [Night 9 →](night-09.md)
