# Night 7 · Check Under the Bed

> 📺 *"Every child knows the drill: before you sleep, check under the bed, check the wardrobe, check the data files. Tonight we write tests, which is the grown-up version. They look under every bed, every time, and they never get tired of it."*

**Tonight you'll learn**
- `go test`, and what makes a file a test
- `t.Errorf` versus `t.Fatalf`
- Subtests with `t.Run`
- A data test that checks every monster is complete
- Faking a filesystem with `testing/fstest`

**Where we are:** monsters load from a pack of files. One week done.

## Tests are just functions

A Go test is a function in a file ending `_test.go`, named `TestSomething`, taking `t *testing.T`. `go test` finds them all, runs them, and reports. There's no framework to install; it's in the standard library, and it has been since the beginning.

Tests live next to the code they test, in the same package, so they can see everything, including the private `monsters` slice. Create `internal/monsters/monsters_test.go`:

```go
package monsters

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestMonsterData guards the quality of every built-in monster.
func TestMonsterData(t *testing.T) {
	all := GetAllMonsters()
	if len(all) < 3 {
		t.Fatalf("expected at least 3 monsters, got %d", len(all))
	}
	seen := map[string]bool{}
	for _, m := range all {
		t.Run(m.ID, func(t *testing.T) {
			if seen[m.ID] {
				t.Errorf("duplicate id %q", m.ID)
			}
			seen[m.ID] = true
			for field, v := range map[string]string{
				"name": m.Name, "description": m.Description, "origin": m.Origin, "ascii": m.ASCII,
			} {
				if strings.TrimSpace(v) == "" {
					t.Errorf("%s is empty", field)
				}
			}
			if len(m.Facts) < 3 {
				t.Errorf("want at least 3 facts, got %d", len(m.Facts))
			}
		})
	}
}
```

- **`t.Fatalf`** reports a failure and stops this test now. Use it when carrying on makes no sense (no monsters at all).
- **`t.Errorf`** reports a failure and carries on, so one run shows *every* problem, not just the first.
- **`t.Run(name, func)`** makes a **subtest** per monster. Failures are reported as `TestMonsterData/wolf-man`, which tells you exactly whose bed the monster was under.
- The `map[string]string{...}` literal is a compact way to check several fields with one loop. Ranging over a map visits the pairs in random order, which doesn't matter here.

This is a *data* test. It doesn't test code so much as protect the JSON files: every contributor who adds a monster runs it, and a missing fact or misnamed art file fails before anyone sees it.

## Testing the loader with a fake filesystem

The loader takes an `fs.FS`, and `testing/fstest` provides one you can build from a map in a few lines. No files on disk, no cleanup.

```go
func TestLoadPacksOrderAndArt(t *testing.T) {
	fsys := fstest.MapFS{
		"test/pack.json": {Data: []byte(`{"id":"test","name":"Test","order":["b","a"]}`)},
		"test/a.json":    {Data: []byte(`{"name":"A"}`)},
		"test/b.json":    {Data: []byte(`{"name":"B"}`)},
		"test/c.json":    {Data: []byte(`{"name":"C"}`)},
		"test/b.txt":     {Data: []byte("art\n\n")},
	}
	packs, err := LoadPacks(fsys)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, m := range packs[0].Monsters {
		ids = append(ids, m.ID)
	}
	if got := strings.Join(ids, ","); got != "b,a,c" {
		t.Errorf("order = %s, want b,a,c", got)
	}
	if b := packs[0].Monsters[0]; b.ASCII != "art" || b.Pack != "test" {
		t.Errorf("art or pack not loaded: %+v", b)
	}
}

func TestBrokenPackIsAnError(t *testing.T) {
	fsys := fstest.MapFS{
		"bad/pack.json": {Data: []byte(`{"id":"bad"}`)},
		"bad/oops.json": {Data: []byte(`{"name": "Oops",`)},
	}
	_, err := LoadPacks(fsys)
	if err == nil {
		t.Fatal("expected an error for broken JSON")
	}
	if !strings.Contains(err.Error(), `pack "bad"`) || !strings.Contains(err.Error(), "oops.json") {
		t.Errorf("error should say which pack and file: %v", err)
	}
}
```

The first test checks three promises from last night: listed monsters come first in the given order, unlisted ones follow alphabetically, and art and pack ids are filled in. The second checks that a broken file produces an error naming the pack *and* the file, which is what `%w` wrapping bought us.

The backtick strings are **raw strings**: no escape sequences, so JSON with its double quotes can be written straight in. `%+v` prints a struct with its field names, handy in failure messages.

## Run them

```bash
go test ./...
```

```
ok  	github.com/hungovercoders/terminal-of-terror/internal/monsters	0.003s
```

Now break something and watch the test find it. Rename `wolf-man.txt` to `Wolf-Man.txt` (last night's terrifying fact), then:

```bash
go test ./internal/monsters
```

```
--- FAIL: TestMonsterData (0.00s)
    --- FAIL: TestMonsterData/wolf-man (0.00s)
        monsters_test.go:26: ascii is empty
```

Rename it back. Run `go test -v ./internal/monsters` to see every subtest listed as it passes: `-v` is *verbose*.

## Try it

- Add a check that every fact ends without a full stop (the facts are written as sentence fragments). Find a monster that breaks the rule, or add one that does, and see it fail.
- Run `go test -run Broken ./...`. `-run` takes a pattern and runs only matching tests.
- Run `go test -cover ./internal/monsters`. Coverage says how much of the package the tests exercised.

## 💀 Terrifying fact

`go test` caches results. Run the same tests on unchanged code and you'll see `(cached)` next to `ok`: the tests didn't run at all, because Go knows the answer can't have changed. Touch any file the package depends on, including the JSON, and they run for real. `go test -count=1` forces a fresh run.

## 🕯️ Before dawn

Write a test for `RandomFact`. Pass it `rand.New(rand.NewSource(1))` twice and check both calls return the same monster and fact: same seed, same answer. Then check that the fact it returns really belongs to the monster it returns (loop over `m.Facts` and look for it). That's a test that would catch the classic slip of using the wrong index.

> 📺 *"Under the bed: nothing. In the wardrobe: nothing. In the data: nothing missing. Sleep well. Next week, the séance."*

[← Night 6](night-06.md) · [Index](README.md) · [Night 8 →](night-08.md)
