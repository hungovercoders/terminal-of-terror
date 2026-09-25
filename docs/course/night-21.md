# Night 21 · What the Crypt Remembers

> 📺 *"So far, everything you've done vanishes when the program ends, like a dream, or a Universal sequel. Tonight the program gets a memory. A small file, in the right place, written carefully, so a crash halfway through never leaves you with half a crypt."*

**Tonight you'll learn**
- Where a program should keep its files: `os.UserConfigDir`
- An environment variable to override it, and why tests need one
- Reading a file that may not exist: `errors.Is` and `fs.ErrNotExist`
- Writing a file atomically: temp file, then rename
- `json.MarshalIndent`, and maps of times
- `t.TempDir` and `t.Setenv`

**Where we are:** the games produce results that go nowhere. Week 3 ends with somewhere for them to go.

## Where files live

A program should not scatter files around the home directory. Every operating system has a place for an application's configuration: `~/.config/<app>` on Linux, `~/Library/Application Support/<app>` on a Mac, `%AppData%\<app>` on Windows. Go knows all three, in [`internal/store/store.go`](../../internal/store/store.go):

```go
// EnvHome overrides where progress and community packs are kept.
const EnvHome = "TERMINAL_OF_TERROR_HOME"

// Dir is the app's config directory: $TERMINAL_OF_TERROR_HOME if set,
// otherwise <user config dir>/terminal-of-terror.
func Dir() (string, error) {
	if d := os.Getenv(EnvHome); d != "" {
		return d, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "terminal-of-terror"), nil
}
```

**`filepath.Join`** joins path pieces with the right separator for the OS; never build paths with `+ "/" +`. And the environment variable comes first for two reasons. Users get to say where their files go. More importantly, *tests* can point the program at a scratch directory, so running the test suite never touches, or wipes, your real progress. Every command in the project honours it; so should you when trying things: `TERMINAL_OF_TERROR_HOME=/tmp/tot go run . quiz`.

## What's remembered

```go
// Progress is everything we remember about a player.
type Progress struct {
	Version   int                  `json:"version"`
	Knowledge map[string]int       `json:"knowledge"` // correct answers per monster id
	Captured  map[string]time.Time `json:"captured"`  // monster id -> when captured
	Badges    map[string]time.Time `json:"badges"`    // badge id -> when earned
	Seen      map[string]bool      `json:"seen"`      // monster pages visited

	QuizzesPlayed  int `json:"quizzesPlayed"`
	BestQuizScore  int `json:"bestQuizScore"` // percent
	GuessesPlayed  int `json:"guessesPlayed"`
	BestGuessScore int `json:"bestGuessScore"` // points
	MashesPlayed   int `json:"mashesPlayed"`

	path string
}
```

Maps keyed by monster id, so a monster added later just gains an entry. `time.Time` values round-trip through JSON as RFC 3339 strings (`"2026-10-31T21:00:00Z"`) with no work. `Version` is there so a future change to the format can tell old files from new. And `path`, unexported and untagged, is remembered by `Load` so `Save` writes back to the same place; `encoding/json` ignores unexported fields, so it never reaches the file.

## Loading what may not exist

The first time anyone runs the program there is no progress file. That's not an error; it's a new player.

```go
// Load reads saved progress, returning empty progress if there is none yet.
func Load() (*Progress, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	p := &Progress{path: path}
	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return nil, err
	default:
		if err := json.Unmarshal(raw, p); err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
	}
	p.init()
	return p, nil
}
```

**`errors.Is(err, fs.ErrNotExist)`** asks whether the error *is*, or wraps, "file not found". Comparing `err == fs.ErrNotExist` would miss it, because `os.ReadFile` returns a `*PathError` that wraps the underlying one. `errors.Is` walks the chain. The empty `case` for it is deliberate: do nothing, fall through to `init`, which creates the empty maps. Any *other* error, permissions say, is returned. And a file that exists but won't parse is an error with the path in it and the original wrapped by `%w`, so the message reads `reading /home/you/.config/terminal-of-terror/progress.json: invalid character...`.

`init` turns nil maps into empty ones. Reading a nil map is fine in Go (you get the zero value) but *writing* to one panics, so a struct with map fields nearly always needs a constructor or an `init` like this.

## Writing without losing anything

Here's the careful part.

```go
// Save writes progress atomically, creating the directory if needed.
func (p *Progress) Save() error {
	...
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p.path), "progress-*.json")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), p.path)
}
```

If `Save` simply opened `progress.json` and wrote into it, a crash, a full disk or a power cut halfway through would leave a truncated file, and the next `Load` would fail on it, forever. Instead: write the whole thing to a temporary file *in the same directory*, close it, then **rename** it over the real one. On every mainstream filesystem a rename is atomic: at any instant the path names either the old complete file or the new complete file, never a half-written one. Same directory matters, because a rename across filesystems is a copy.

Every failure path removes the temp file. It's four lines of cleanup, and it's the difference between a program that leaves `progress-183726.json` litter and one that doesn't. `0o755` is an octal permission (owner can do anything, others can read and enter) for the directory; `os.CreateTemp` picks a safe mode for the file itself.

`json.MarshalIndent(p, "", "  ")` writes the JSON with two-space indentation, so the file is readable by a human who opens it. It costs a few bytes and saves someone an afternoon.

## The stub becomes real

`record` in [`cmd/progress.go`](../../cmd/progress.go) is what the game commands call:

```go
// record saves an outcome to the player's progress and announces anything
// new. Progress problems are reported but never stop the fun.
func record(o crypt.Outcome) {
	p, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Couldn't read your progress, so this session won't be saved: %v\n", err)
		return
	}
	...
	u := crypt.Apply(p, monsters.EveryMonster(), o)
	if err := p.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Couldn't save your progress: %v\n", err)
		return
	}
	if s := ui.RenderUnlocks(u); s != "" {
		fmt.Print("\n" + s)
	}
}
```

A design decision in the comment: a problem with the save file prints a warning to **stderr** and the game carries on. You just played a quiz; the program refusing to show your score because a directory wasn't writable would be the wrong priority. `os.Stderr` rather than `Println` because warnings aren't output: `quiz --json > questions.json` should still produce clean JSON. Tomorrow's `crypt.Apply` is what turns an outcome into captures and badges.

## Run it

```bash
export TERMINAL_OF_TERROR_HOME=/tmp/tot-scratch
go run . quiz -n 3
cat /tmp/tot-scratch/progress.json
```

Play three questions and there's the file: `knowledge`, `quizzesPlayed`, `bestQuizScore`. Play again and watch the numbers change. Unset the variable (`unset TERMINAL_OF_TERROR_HOME`) and `go run . crypt` shows where your real file lives.

## Testing with a scratch directory

[`store_test.go`](../../internal/store/store_test.go):

```go
func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, dir)

	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Captured) != 0 {
		t.Fatal("fresh progress should be empty")
	}
	p.Knowledge["dracula"] = 2
	p.Captured["mummy"] = time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	p.QuizzesPlayed = 3
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}

	q, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if q.Knowledge["dracula"] != 2 || q.QuizzesPlayed != 3 || q.Captured["mummy"].IsZero() {
		t.Fatalf("round trip lost data: %+v", q)
	}
	...
}

func TestCorruptFileIsAnError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, dir)
	os.WriteFile(filepath.Join(dir, "progress.json"), []byte("{nope"), 0o644)
	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a corrupt file")
	}
}
```

**`t.TempDir()`** gives a fresh empty directory that's deleted when the test ends. **`t.Setenv`** sets an environment variable for the duration of the test and restores it after. Together, and with `Dir()` honouring the variable, the tests can load, save, corrupt and reset without ever seeing the developer's own file. A *round trip*, save then load then compare, is the one test every storage layer needs.

## Try it

- Delete the `os.Remove` lines from `Save`, make `tmp.Write` fail (write to a full disk, or just `return errors.New("boom")` temporarily) and look at the directory afterwards.
- Add a `LastPlayed time.Time` field. Notice you don't need to touch `Load` or `Save`.
- Bump `Version` to 2 and make `Load` print a warning when it reads a version 1 file. What would a real migration look like?

## 💀 Terrifying fact

`os.Rename` is atomic on Linux, macOS and modern Windows for files on the same volume. But on Windows it *fails* if the destination is open by another process, where Linux would happily replace it. Cross-platform "write a file safely" is a surprisingly deep topic, and this is the simple version that works for a single-user tool. If two copies of the program ran at once, they could still race each other; the project doesn't lock the file, and the comment in the code should probably say so.

## 🕯️ Before dawn

Week 3 is done. You've built an animated intro, a monochrome film mode, a gallery, and two games with saved scores. Tag it, `git tag night-21`. Then add a `--path` flag to tomorrow's `crypt` command that prints where the progress file is and exits; it's a one-liner with `store.Path()`, and the kind of flag that saves a support question. (The finished project prints the path at the bottom of `crypt` instead. Which is friendlier?)

> 📺 *"The crypt remembers. Every right answer, every fogbound guess, every page you lingered on. Next week: what it does with all that. Captures, badges, a fight, a moon, and finally, the release. Sleep while you can."*

[← Night 20](night-20.md) · [Index](README.md) · [Night 22 →](night-22.md)
