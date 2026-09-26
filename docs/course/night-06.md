# Night 6 · Packs of Monsters

> 📺 *"One big coffin is fine for one vampire. Tonight we build a crypt: a folder of coffins, one per monster, arranged in packs. It's the biggest night of the week, so pour something warm. Blood is traditional; cocoa is acceptable."*

**Tonight you'll learn**
- One file per monster, and why
- `embed.FS`: embedding a whole folder
- `io/fs`: reading directories and files from any filesystem
- Maps, `delete`, and sorting
- Wrapping errors with `%w`
- Loading ASCII art from a text file

**Where we are:** monsters load from one embedded JSON file.

## Why one file each?

One big `monsters.json` has a problem you only meet later: two people adding monsters at the same time edit the same file, and their changes collide. A folder with `dracula.json`, `mummy.json` and so on lets everyone work on their own monster. It also lets each monster keep its ASCII art in a plain text file, `dracula.txt`, where the backslashes and quotes don't need escaping.

A **pack** is a folder of monsters with a `pack.json` describing it. Tonight there's one pack; on Night 27 the program will load packs that other people write.

## The files

Delete `internal/monsters/monsters.json`. Make a folder `internal/monsters/packs/universal/` and, in it, `pack.json`:

```json
{
  "id": "universal",
  "name": "Universal Classics",
  "description": "The monsters that haunted Universal Pictures from the silent era to the 1950s.",
  "order": ["dracula", "frankenstein", "wolf-man"]
}
```

Then one file per monster. `dracula.json` is last night's Dracula object on its own:

```json
{
  "name": "Dracula",
  "description": "The legendary vampire count from Transylvania",
  "origin": "Bram Stoker's novel (1897)",
  "facts": [
    "He casts no reflection, as Jonathan Harker discovers while shaving",
    "Stoker took the name from Vlad III of Wallachia, but little else",
    "Bela Lugosi played the Count on Broadway in 1927 before the film"
  ]
}
```

Do the same for `frankenstein.json` and `wolf-man.json`. The file name, without `.json`, becomes the monster's **id**: a short, lowercase, dash-separated name that we'll use in commands and to find the art file.

Finally, the art. Copy [`dracula.txt`](../../internal/monsters/packs/universal/dracula.txt), [`frankenstein.txt`](../../internal/monsters/packs/universal/frankenstein.txt) and [`wolf-man.txt`](../../internal/monsters/packs/universal/wolf-man.txt) from this repository into the same folder, or draw your own. Plain text, one line per row.

## The loader

Replace `internal/monsters/monsters.go`. It's long, so it's broken into pieces here; the whole file is [in the repository](../../internal/monsters/monsters.go) if you'd rather copy it (the repo version has more fields and features, which later nights add).

### Types and the embedded folder

```go
// Package monsters knows every creature in the vault.
package monsters

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"path"
	"sort"
	"strings"
)

// Monster is a single creature and what we know about it.
type Monster struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Origin      string   `json:"origin"`
	Facts       []string `json:"facts"`
	ASCII       string   `json:"ascii,omitempty"`
	Pack        string   `json:"pack"`
}

// Pack is a themed collection of monsters.
type Pack struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       []string  `json:"order"`
	Monsters    []Monster `json:"-"`
}

//go:embed packs
var builtinFS embed.FS
```

Two new tag tricks: `omitempty` leaves a field out when *writing* JSON if it's empty, and `json:"-"` means "never read or write this field". `Monsters` is filled by our loader, not by the JSON.

`//go:embed packs` embeds an entire folder, and the variable's type is `embed.FS`: a **filesystem** value. Go has an interface, `fs.FS`, for "anything you can read files from". An embedded folder is one. A real directory is another (`os.DirFS`). A fake one in a test is a third. Write the loader against `fs.FS` and it works with all of them, which is exactly what we'll need on Nights 7 and 27.

### Loading at startup

```go
var (
	packs    []Pack
	monsters []Monster
)

func init() {
	sub, err := fs.Sub(builtinFS, "packs")
	if err != nil {
		log.Fatalf("Failed to open built-in packs: %v", err)
	}
	packs, err = LoadPacks(sub)
	if err != nil {
		log.Fatalf("Failed to load monsters data: %v", err)
	}
	for _, p := range packs {
		monsters = append(monsters, p.Monsters...)
	}
}
```

`fs.Sub` gives a filesystem rooted at `packs/`, so the loader sees `universal/` at its top level rather than `packs/universal/`. `append(monsters, p.Monsters...)` adds a whole slice to another; the `...` spreads it out.

### Every pack in the folder

```go
// LoadPacks reads every pack directory in fsys. Each directory holds a
// pack.json, one <id>.json per monster and an optional <id>.txt of ASCII art.
func LoadPacks(fsys fs.FS) ([]Pack, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var out []Pack
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := loadPack(fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("pack %q: %w", e.Name(), err)
		}
		out = append(out, p)
	}
	return out, nil
}
```

`fmt.Errorf` with `%w` **wraps** an error: the message gains context (`pack "universal": ...`) and the original error stays inside, where `errors.Is` can still find it. Wrap at every layer and a failure reads like a trail of breadcrumbs: `pack "universal": dracula.json: invalid character ...`.

### One pack

```go
// loadPack reads one pack directory.
func loadPack(fsys fs.FS, dir string) (Pack, error) {
	var p Pack
	raw, err := fs.ReadFile(fsys, path.Join(dir, "pack.json"))
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("pack.json: %w", err)
	}
	if p.ID == "" {
		p.ID = dir
	}

	files, err := fs.Glob(fsys, path.Join(dir, "*.json"))
	if err != nil {
		return p, err
	}
	byID := map[string]Monster{}
	for _, f := range files {
		if path.Base(f) == "pack.json" {
			continue
		}
		raw, err := fs.ReadFile(fsys, f)
		if err != nil {
			return p, err
		}
		var m Monster
		if err := json.Unmarshal(raw, &m); err != nil {
			return p, fmt.Errorf("%s: %w", path.Base(f), err)
		}
		if m.ID == "" {
			m.ID = strings.TrimSuffix(path.Base(f), ".json")
		}
		if art, err := fs.ReadFile(fsys, path.Join(dir, m.ID+".txt")); err == nil {
			m.ASCII = strings.TrimRight(string(art), "\n")
		}
		m.Pack = p.ID
		byID[m.ID] = m
	}

	// Listed monsters first, in the pack's chosen order, then any extras.
	for _, id := range p.Order {
		if m, ok := byID[id]; ok {
			p.Monsters = append(p.Monsters, m)
			delete(byID, id)
		}
	}
	var rest []string
	for id := range byID {
		rest = append(rest, id)
	}
	sort.Strings(rest)
	for _, id := range rest {
		p.Monsters = append(p.Monsters, byID[id])
	}
	return p, nil
}
```

Take it in three parts.

**Reading the files.** `fs.Glob` finds every `*.json` in the pack. We skip `pack.json`, unmarshal each monster, and fill in its id from the file name. Then we *try* to read `<id>.txt`; if that fails, there's simply no art, so we ignore the error, which is the one time it's fine to. Note `path.Join`, not `filepath.Join`: an `fs.FS` always uses forward slashes, whatever the operating system.

**A map.** `byID` is a **map** from string to Monster: `byID["dracula"]` looks one up, `byID[id] = m` stores one, `delete(byID, id)` removes one, and `m, ok := byID[id]` tells you whether it was there. Maps are Go's dictionaries.

**Ordering.** `pack.json` says which monsters come first. We take those in order, deleting each from the map as we go, then add whatever's left in alphabetical order so that a monster nobody listed still appears. Ranging over a map gives keys in a *deliberately random* order, which is why `rest` is sorted before use.

### The doors

```go
// GetAllMonsters returns every monster in the vault.
func GetAllMonsters() []Monster {
	return monsters
}

// GetPacks returns every loaded pack.
func GetPacks() []Pack {
	return packs
}

// RandomFact picks a random monster and one of its facts using r.
func RandomFact(r *rand.Rand) (Monster, string) {
	m := monsters[r.Intn(len(monsters))]
	return m, m.Facts[r.Intn(len(m.Facts))]
}
```

## Run it

```bash
go vet ./...
go run . list
```

Same three monsters as before, in the order `pack.json` asked for. Nothing visible changed tonight, which is what a good refactor looks like. Everything underneath is new.

## Try it

- Swap the order in `pack.json`. Run `list`.
- Remove `wolf-man` from the order. It still appears, at the end.
- Make a second pack: `packs/folklore/pack.json` and one monster, say `golem.json`. The Golem is a clay giant from Jewish folklore, brought to life with sacred words. `list` now shows four. That's Night 27's community packs, minus the loading-from-disk part.

## 💀 Terrifying fact

`fs.ReadFile(fsys, path.Join(dir, m.ID+".txt"))` is looked up by *exact* name, and embedded filesystems are case-sensitive even on Windows and macOS, where the real disk usually isn't. Name the art `Dracula.txt` and the monster loads without a face, silently. Night 7 writes a test that would catch it.

## 🕯️ Before dawn

Make `list` group monsters by pack, with the pack's name as a heading, using `GetPacks()`. Then print each pack's description under its heading. The finished version is in [`internal/ui/cards.go`](../../internal/ui/cards.go), `RenderList`, styled with tomorrow-week's colours.

> 📺 *"A crypt, properly organised. The undead are filed alphabetically, which would upset them if they knew. Tomorrow night, we make sure nothing is missing. Check under the bed."*

[← Night 5](night-05.md) · [Index](README.md) · [Night 7 →](night-07.md)
