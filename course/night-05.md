# Night 5 · Data in the Coffin

> 📺 *"A vampire keeps his native soil in a box so he can rest anywhere. Tonight our monsters get boxes of their own: a data file they can sleep in, separate from the code, so that adding a monster never again means recompiling the program. Well. Almost never."*

**Tonight you'll learn**
- JSON, and why data belongs in files
- Struct tags: telling Go how JSON names map to fields
- `encoding/json` and `Unmarshal`
- `//go:embed`: baking a file into the binary
- `init` for loading, and `log.Fatalf` for the unloadable

**Where we are:** three monsters written out in Go code.

## Code is a bad place for data

Right now, adding a monster means editing a Go file, keeping every brace and comma straight, and rebuilding. That's fine for three monsters and miserable for nineteen. Worse, the person adding a monster has to understand Go. A horror fan with a great Baba Yaga fact shouldn't need to.

**JSON** is a text format for data that every language reads. It looks like this. Create `internal/monsters/monsters.json`:

```json
[
  {
    "name": "Dracula",
    "description": "The legendary vampire count from Transylvania",
    "origin": "Bram Stoker's novel (1897)",
    "facts": [
      "He casts no reflection, as Jonathan Harker discovers while shaving",
      "Stoker took the name from Vlad III of Wallachia, but little else",
      "Bela Lugosi played the Count on Broadway in 1927 before the film"
    ]
  },
  {
    "name": "Frankenstein's Monster",
    "description": "The tragic creature created by Dr. Victor Frankenstein",
    "origin": "Mary Shelley's novel (1818)",
    "facts": [
      "The creature is never named in the novel; Frankenstein is his creator",
      "He teaches himself to read by secretly watching a cottage family",
      "Boris Karloff was credited only as '?' in the 1931 film's opening titles"
    ]
  },
  {
    "name": "The Wolf Man",
    "description": "A man cursed to transform into a werewolf",
    "origin": "The Wolf Man (1941 film)",
    "facts": [
      "Lon Chaney Jr. played Larry Talbot in all five of his Universal films",
      "In the 1941 film he is killed with a silver-headed cane, not a silver bullet",
      "The film opened just five days after the attack on Pearl Harbor"
    ]
  }
]
```

`[ ]` is a list, `{ }` is an object with `"key": value` pairs, and strings use double quotes. No trailing commas, which is the one way JSON is stricter than Go.

## Teaching the struct about JSON

Replace the whole of `internal/monsters/monsters.go`:

```go
// Package monsters knows every creature in the vault.
package monsters

import (
	_ "embed"
	"encoding/json"
	"log"
	"math/rand"
)

// Monster is a single creature and what we know about it.
type Monster struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Origin      string   `json:"origin"`
	Facts       []string `json:"facts"`
}

//go:embed monsters.json
var monstersJSON []byte

var monsters []Monster

func init() {
	if err := json.Unmarshal(monstersJSON, &monsters); err != nil {
		log.Fatalf("Failed to load monsters data: %v", err)
	}
}

// GetAllMonsters returns every monster in the vault.
func GetAllMonsters() []Monster {
	return monsters
}

// RandomFact picks a random monster and one of its facts using r.
func RandomFact(r *rand.Rand) (Monster, string) {
	m := monsters[r.Intn(len(monsters))]
	return m, m.Facts[r.Intn(len(m.Facts))]
}
```

Four new ideas, one per paragraph.

**Struct tags.** The backticked `` `json:"name"` `` after each field is a tag: a note for other packages to read. `encoding/json` reads it to learn that the JSON key `name` fills the field `Name`. Without tags, it would look for a key called `Name`, and by convention JSON keys are lowercase.

**`//go:embed monsters.json`.** This comment is an instruction to the compiler: read that file at build time and put its contents in the variable on the next line. The result is that the binary carries the data inside it. Ship one file, and it still works. The `_ "embed"` import is required to use the directive even though we never call anything in the `embed` package; the underscore says "import this for its side effect".

**`json.Unmarshal(data, &monsters)`.** *Unmarshal* means turn bytes into Go values. It needs to *write into* `monsters`, so we pass `&monsters`, a **pointer** to it. Pass `monsters` by itself and the function would get a copy, fill the copy, and throw it away. Any function that needs to change your variable takes a pointer; `&` makes one.

**`log.Fatalf`.** If the JSON is broken, there's nothing sensible the program can do, so `Fatalf` prints the message and exits with status 1. This runs in `init`, before `main`, so a bad data file fails before any command runs. Try it: delete a comma in the JSON and run `go run . list`:

```
2026/09/25 19:54:10 Failed to load monsters data: invalid character '"' after object key:value pair
exit status 1
```

Put the comma back.

## Run it

```bash
go run . list
go run . random
```

Identical output to last night, and now you can add a monster by editing a text file and rebuilding. On Night 27, community packs will load from disk *without* rebuilding, using the same `Unmarshal` call.

## Try it

- Add The Mummy to the JSON. No Go changes needed.
- Add a field the JSON doesn't have (`Year int \`json:"year"\``) and run it. Missing keys are fine: the field stays at its zero value, `0` for a number, `""` for a string, `nil` for a slice.
- Add a key the struct doesn't have (`"colour": "red"`) and run it. Unknown keys are silently ignored. Both behaviours are why JSON is forgiving to work with, and why Night 7's tests will check that every monster is complete.

## 💀 Terrifying fact

`json.Unmarshal` only fills *exported* fields. Lowercase a field name, `name string`, and Unmarshal skips it without a word, because it's outside the struct's package and can't see in. If a field mysteriously stays empty, check its capital letter first.

## 🕯️ Before dawn

Marshal goes the other way: `json.MarshalIndent(monsters, "", "  ")` turns the slice into pretty-printed JSON bytes. Add a `--json` flag to `list` that prints the vault as JSON instead of the table. Two things to find out: what the second return value of `MarshalIndent` is, and what `os.Stdout.Write` does with a `[]byte`. Night 26 does this properly for every command.

> 📺 *"The monsters are in their boxes, and the boxes are in the program. Tomorrow night, one box each, and a whole crate of them."*

[← Night 4](night-04.md) · [Index](README.md) · [Night 6 →](night-06.md)
