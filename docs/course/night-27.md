# Night 27 · Open the Doors

> 📺 *"Nineteen monsters is a start. But somewhere out there is someone who knows everything about the Jersey Devil, or the Mothman, or a creature from their grandmother's village that no film ever touched. Tonight we let them in. A folder in the config directory, a JSON file or two, and their monster is in the quiz by morning."*

**Tonight you'll learn**
- Loading the same format from disk that you embed at build time
- Validation with helpful messages, and defaults for what's missing
- Skipping the broken without stopping the working
- Merging, and refusing duplicates
- A scaffolding command: writing files for the user to edit
- Pointers to fields: `[]*int`

**Where we are:** monsters come from the embedded packs. Tonight, from the user's disk too.

## The same loader, a different filesystem

Night 6's `LoadPacks(fsys fs.FS, ...)` took an `fs.FS` rather than reading `embed.FS` directly, and the reason arrives tonight. A directory on disk is also an `fs.FS`:

```go
// LoadUserPacks reads community packs from dir, where each subdirectory is
// a pack. Broken or incomplete monsters are skipped and reported; a missing
// dir simply means no community packs.
func LoadUserPacks(dir string) ([]Pack, []error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []error{err}
	}
	fsys := os.DirFS(dir)
	var out []Pack
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, perrs := loadPack(fsys, e.Name())
		...
```

`os.DirFS(dir)` wraps a directory as an `fs.FS`, and the same `loadPack` that reads `packs/universal` out of the binary reads `~/.config/terminal-of-terror/packs/cryptids` off the disk. The format is identical: a `pack.json`, one JSON per monster, an optional `.txt` portrait. The user's monster gets everything the built-in ones get, because it *is* one, as far as the rest of the program can tell. Design the built-in data as if it were user data and the door is already half open.

A missing directory is not an error; it's a user without community packs, which is nearly everyone. The `errors.Is` pattern from Night 21.

## Errors, plural

`LoadUserPacks` returns `[]error`, a *list*. One bad file in a pack of ten monsters shouldn't take the other nine down, and a user who's typing JSON by hand needs to hear about *all* their mistakes, not one per run. So each problem is collected and the loop continues:

```go
		for _, m := range p.Monsters {
			if problems := Validate(m); len(problems) > 0 {
				errs = append(errs, fmt.Errorf("pack %q: skipping %q: %s", p.ID, m.ID, strings.Join(problems, "; ")))
				continue
			}
			applyDefaults(&m)
			ok = append(ok, m)
		}
```

and `loadCommunityPacks` in [`cmd/root.go`](../../cmd/root.go) prints each one to stderr with a ⚠️ and carries on. The program starts; the message says exactly which pack, which monster, and what's wrong.

## Validate and default

Two functions, and the split between them is a policy:

```go
// Validate lists what a community monster is missing to be playable.
func Validate(m Monster) []string {
	var problems []string
	if strings.TrimSpace(m.Name) == "" {
		problems = append(problems, "needs a name")
	}
	if strings.TrimSpace(m.Description) == "" {
		problems = append(problems, "needs a description")
	}
	if len(m.Facts) == 0 {
		problems = append(problems, "needs at least one fact")
	}
	for _, v := range []int{m.Stats.Strength, m.Stats.Speed, m.Stats.Cunning, m.Stats.Dread} {
		if v < 0 || v > 10 {
			problems = append(problems, "stats must be between 1 and 10")
			break
		}
	}
	return problems
}
```

Three required fields. Everything a monster *must* have to appear in the explorer and the quiz, and nothing more. Then:

```go
// applyDefaults fills in optional fields so community monsters work everywhere.
func applyDefaults(m *Monster) {
	if m.Emoji == "" {
		m.Emoji = "👹"
	}
	if m.Origin == "" {
		m.Origin = "Unknown"
	}
	if m.Debut.Title == "" {
		m.Debut.Title = m.Name
	}
	if m.Debut.Year == 0 && m.Debut.Era == "" {
		m.Debut.Era = "date unknown"
	}
	for _, s := range []*int{&m.Stats.Strength, &m.Stats.Speed, &m.Stats.Cunning, &m.Stats.Dread} {
		if *s == 0 {
			*s = 5
		}
	}
	if len(m.Powers) == 0 {
		m.Powers = []string{"Sheer terror"}
	}
	if len(m.Weaknesses) == 0 {
		m.Weaknesses = []string{"A good night's sleep"}
	}
}
```

Everything else gets a default, so a three-field monster still has stats for the Mash and powers for a signature move. The rule: *require what can't be guessed, default what can*. A missing name is a real gap; a missing emoji is a 👹.

`[]*int{&m.Stats.Strength, ...}` is a slice of **pointers to fields**, so one loop can set four different fields: `*s = 5` writes through the pointer to wherever it points. Night 8 said pointers let you say "not found"; this is the other use, letting one piece of code modify something it was handed.

## Merging without collisions

```go
// AddPacks merges extra packs into the collection. Packs or monsters whose
// ids are already taken are skipped and reported.
func AddPacks(extra []Pack) []error {
	...
	for _, p := range extra {
		if packIDs[p.ID] {
			errs = append(errs, fmt.Errorf("pack %q: a pack with that id already exists", p.ID))
			continue
		}
		var ok []Monster
		for _, m := range p.Monsters {
			if monsterIDs[m.ID] {
				errs = append(errs, fmt.Errorf("pack %q: skipping %q: a monster with that id already exists", p.ID, m.ID))
				continue
			}
			monsterIDs[m.ID] = true
			ok = append(ok, m)
		}
		...
```

A community pack called `universal` or a monster with id `dracula` would collide with the built-ins, and the program's lookups assume ids are unique. Two sets, `packIDs` and `monsterIDs`, are built from what's loaded and checked as extras arrive. The built-ins win; the duplicate is reported and skipped. Sets are `map[string]bool` in Go; there's no separate set type, and this is the idiom.

## Scaffolding

Nobody should have to read the docs to make their first pack. `packs new <id>` in [`cmd/packs.go`](../../cmd/packs.go) writes a working example to edit:

```go
		monsterID := id + "-monster"
		monsterName := "My " + id + " monster"
		files := map[string]any{
			"pack.json": map[string]any{
				"id":          id,
				"name":        "My " + id + " pack",
				"description": "Describe your pack here.",
				"order":       []string{monsterID},
			},
			monsterID + ".json": exampleMonster(monsterID, monsterName),
		}
		for name, v := range files {
			raw, _ := json.MarshalIndent(v, "", "  ")
			if err := os.WriteFile(filepath.Join(packDir, name), append(raw, '\n'), 0o644); err != nil {
				return err
			}
		}
```

`map[string]any` is JSON's shape in Go without declaring a type: string keys, any values, nested as needed. `exampleMonster` fills every field with a placeholder that *explains itself*: `"description": "A one-line description of your monster"`, `"hostIntro": "What Count Cathode says to introduce your monster."`. The user opens the file and the file is the documentation. The example monster's id is derived from the pack id (`cryptids-monster`) so two scaffolded packs never collide with each other.

The command also validates the pack id against `^[a-z0-9][a-z0-9-]*$`, refuses to overwrite an existing directory, and ends by printing the exact next commands to run. A scaffolding command's job is to get the user to their first success in one minute.

## Run it

```bash
export TERMINAL_OF_TERROR_HOME=/tmp/tot-scratch
go run . packs new cryptids
go run . packs
go run . monster "my cryptids monster"
go run . quiz --pack cryptids
```

Then edit `/tmp/tot-scratch/packs/cryptids/cryptids-monster.json`: rename it Mothman, write three facts, save. `monster mothman` finds it by Night 8's search, `quiz` asks about it by Night 18's generator, `guess` fogs its portrait if you draw one. Now break it: delete the `name` line. The next run warns `pack "cryptids": skipping "cryptids-monster": needs a name` and everything else still works.

## Try it

- `packs new Cryptids` with a capital. Read the error. Now try `packs new cryptids` twice.
- Make a pack with a monster whose id is `dracula`. Who wins?
- Give a community monster `"quotes"` with a `source`. It shows on the Quotes page; but the rules say public-domain only. Should `Validate` check anything about that? (It can't. Some rules are for people.)

## 💀 Terrifying fact

`pack.json` is optional for a community pack: `loadPack` treats a missing one as a pack named after its folder, so a directory with nothing but `mothman.json` in it works. And a monster file that isn't listed in `order` isn't lost; it's appended after the listed ones, alphabetically by id. Both are deliberate leniency for people editing JSON by hand, and both are the kind of thing that must be *tested*, or the next refactor will quietly make `pack.json` required and nobody will notice until a user's pack vanishes. `TestLoadPacksOrderAndArt` from Night 6 is where that check lives.

## 🕯️ Before dawn

Make a real pack. Three monsters from your own country's folklore, with a portrait each, the legend, and a myth or two. Play the quiz on it. If it's good, the project would like it: see [CONTRIBUTING.md](../../CONTRIBUTING.md) for how packs get in.

> 📺 *"The doors are open. Whatever's in your grandmother's stories can be in the quiz by morning. Tomorrow night we test everything, including the things that only break in a real terminal."*

[← Night 26](night-26.md) · [Index](README.md) · [Night 28 →](night-28.md)
