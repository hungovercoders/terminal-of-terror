# Night 23 · Monster Mash

> 📺 *"It was a graveyard smash. Tonight, two monsters, three rounds, a six-sided die, and a commentary track by yours truly. The stat cards you copied on Night 13 finally do something. Place your bets; the house always wins, and the house is a crypt."*

**Tonight you'll learn**
- Simulating a game with stats plus dice
- Tables of contests with a function per stat
- Filling narration templates, with grammar
- A best-of-three with the loser named
- Pausing for drama, but only in a real terminal
- A hundred seeds again

**Where we are:** monsters have stat cards nobody has used.

## Stats plus dice

Each monster has Strength, Speed, Cunning and Dread from 1 to 10. A round picks one stat, adds a die roll to each side, and the higher total wins. Three rounds, three different stats. [`internal/mash/mash.go`](../../internal/mash/mash.go):

```go
type contest struct {
	stat  string
	value func(monsters.Stats) int
	line  string // {W} winner, {L} loser
}

var contests = []contest{
	{"Strength", func(s monsters.Stats) int { return s.Strength },
		"{A} and {B} lock arms in the graveyard fog. {W} hurls {L} clean through a mausoleum door!"},
	{"Speed", func(s monsters.Stats) int { return s.Speed },
		"A chase across the moonlit moors! {W} is a blur, and {L} is left gasping at the cemetery gates."},
	{"Cunning", func(s monsters.Stats) int { return s.Cunning },
		"A battle of wits by candlelight. {W} lays a trap, and {L} walks straight into it."},
	{"Dread", func(s monsters.Stats) int { return s.Dread },
		"A staring contest of pure terror. {L} blinks first, and {W} lets out a bloodcurdling laugh."},
}
```

The `value` field is a function that pulls one number out of a `Stats`, which is how a table can say "the Strength stat" without reflection or a string switch. Each contest also carries its narration, with placeholders.

```go
func Fight(r *rand.Rand, a, b monsters.Monster) Bout {
	bout := Bout{A: a, B: b}
	order := r.Perm(len(contests))[:3]
	wins := [2]int{}
	for n, ci := range order {
		c := contests[ci]
		sa, sb := c.value(a.Stats), c.value(b.Stats)
		ra, rb := sa+r.Intn(6)+1, sb+r.Intn(6)+1
		for tries := 0; ra == rb && tries < 5; tries++ {
			ra, rb = sa+r.Intn(6)+1, sb+r.Intn(6)+1
		}
		winner := 0
		if rb > ra || (rb == ra && sb > sa) {
			winner = 1
		}
		wins[winner]++
		...
```

`r.Perm(4)[:3]` picks three of the four stats in random order, so no stat repeats. `r.Intn(6)+1` is a d6. A tie is re-rolled up to five times, and if it's still a tie the higher raw stat wins, so a bout always has a winner and the loop can't spin. `wins` is an **array**, `[2]int`, not a slice: fixed size, two counters, indexed by the winner. Arrays are rare in Go code, but a small fixed pair is exactly what they're for.

Stats matter but don't decide: Dracula's Cunning 10 against the Creature's 5 is a five-point edge on a six-sided die, so the Creature wins that round sometimes. That's the tuning that makes it a game rather than a lookup.

## Grammar in templates

The narration lines have `{A}`, `{B}`, `{W}` and `{L}`. "The Wolf Man" at the start of a sentence is fine; "{W} hurls {L}" with The Mummy in the middle produces "The Wolf Man hurls The Mummy", which reads like a headline. So `fill` checks whether each placeholder starts a sentence:

```go
// fill replaces placeholders, using "the" rather than "The" mid-sentence.
func fill(line string, a, b, w, l monsters.Monster) string {
	repl := map[string]string{"{A}": a.Name, "{B}": b.Name, "{W}": w.Name, "{L}": l.Name}
	out := line
	for k, v := range repl {
		for {
			i := strings.Index(out, k)
			if i < 0 {
				break
			}
			name := v
			if !sentenceStart(out[:i]) {
				name = MidSentence(v)
			}
			out = out[:i] + name + out[i+len(k):]
		}
	}
	return out
}

// MidSentence lowercases a leading "The" so names read naturally mid-sentence.
func MidSentence(name string) string {
	if strings.HasPrefix(name, "The ") {
		return "the " + name[4:]
	}
	return name
}
```

`for { ... break }` is Go's `while(true)`. Each placeholder is replaced one occurrence at a time so the "is this a sentence start?" check sees the text *before* it. `MidSentence` is exported because the UI uses it for the host's ring announcement ("In this corner, the Wolf Man!"). `TestMidSentenceNames` covers both positions. The finale does the same for the loser's weakness: `lowerFirst("The coming of dawn")` gives "muttering about the coming of dawn", but "ANANKE" stays as it is, because a word in capitals is a name.

Text generation always ends up needing grammar rules, and the rules always start as one special case. Put them in a function with a name from the first one.

## The bout as data

```go
// Bout is a whole fight.
type Bout struct {
	A      monsters.Monster `json:"-"`
	B      monsters.Monster `json:"-"`
	Rounds []Round          `json:"rounds"`
	Winner int              `json:"winner"`
	Finale string           `json:"finale"`
}
```

`Fight` returns the entire bout, rounds, rolls and narration, *before anything is printed*. The command then plays it back with pauses. Simulating first and presenting second means the simulation is testable, the presentation can be fast or slow, and `mash --json` (Night 26) could print the bout as data. The `json:"-"` tag leaves the two full monsters out of that JSON; the ids would do.

## Drama, but only for people

[`cmd/mash.go`](../../cmd/mash.go):

```go
		bout := mash.Fight(r, fighters[0], fighters[1])
		w := ui.TerminalWidth()
		pause := func(d time.Duration) {
			if !mashFast && term.IsTerminal(os.Stdout.Fd()) {
				time.Sleep(d)
			}
		}
		fmt.Println(ui.MashHeader(bout.A, bout.B, w))
		pause(1200 * time.Millisecond)
		for _, round := range bout.Rounds {
			fmt.Println(ui.MashRound(bout, round, w))
			pause(1500 * time.Millisecond)
		}
		fmt.Print(ui.MashResult(bout, w))
		record(crypt.Outcome{Kind: "mash"})
```

A second and a half between rounds is good theatre for a person and a waste of time for CI, a pipe, or a script. `term.IsTerminal` asks whether stdout is a real terminal; `--fast` lets a person skip the pauses too. This is a plain `fmt.Println` program, not Bubble Tea: nothing needs redrawing, so the simplest tool wins.

The command also handles the two corners:

```go
		if len(fighters) == 2 && fighters[0].ID == fighters[1].ID {
			return fmt.Errorf("%s can't fight itself... pick two different monsters", fighters[0].Name)
		}
		// Fill empty corners from monsters not already fighting. Picking from
		// this pool (not retrying at random) can't loop forever.
		for len(fighters) < 2 {
			var pool []monsters.Monster
			for _, m := range monsters.GetAllMonsters() {
				if len(fighters) == 0 || m.ID != fighters[0].ID {
					pool = append(pool, m)
				}
			}
			...
			fighters = append(fighters, pool[r.Intn(len(pool))])
		}
```

The comment records a fix: an earlier version picked a random monster and *retried* if it matched the first, which with a one-monster `--pack` would loop forever. Building the pool of valid choices and picking from it can't. When you write "pick again if it's wrong", ask what happens when everything is wrong.

## The cards

`ui.MashHeader` draws two `StatCard`s side by side with a `VS` between them, or stacked on a narrow terminal, using the `statBar` you wrote on Night 13. Each round prints the stat, the rolls, the narration and the winner's "signature move", which is one of its `Powers` picked at random. The finale names the loser and one of its `Weaknesses`. Everything the fight says comes from the data files; the mash package has no monster knowledge of its own.

## Run it

```bash
go run . mash dracula "wolf man"
go run . mash --fast
```

## A hundred bouts

```go
func TestFightIsBestOfThree(t *testing.T) {
	a := *monsters.GetMonsterByName("dracula")
	b := *monsters.GetMonsterByName("wolf-man")
	for seed := int64(0); seed < 100; seed++ {
		bout := Fight(rand.New(rand.NewSource(seed)), a, b)
		if len(bout.Rounds) != 3 {
			t.Fatalf("want 3 rounds, got %d", len(bout.Rounds))
		}
		wa, wb := bout.Tally()
		if (wa > wb) != (bout.Winner == 0) {
			t.Fatalf("winner %d disagrees with tally %d-%d", bout.Winner, wa, wb)
		}
		stats := map[string]bool{}
		for _, r := range bout.Rounds {
			if stats[r.Stat] {
				t.Fatalf("stat %s used twice", r.Stat)
			}
			stats[r.Stat] = true
			if strings.Contains(r.Narration, "{") {
				t.Fatalf("unfilled placeholder: %s", r.Narration)
			}
		}
	}
}
```

Night 18's idea again: properties over many seeds. Three rounds, the winner agrees with the tally, no stat repeats, no `{` survives into the text. `*monsters.GetMonsterByName("dracula")` dereferences the pointer to get a copy.

## Try it

- Add a fifth contest, "Legacy", scored by `len(m.Legacy)`. The `value` function makes it a one-liner. Does `r.Perm(5)[:3]` still hold?
- Make a round's die a d10 and play a few bouts. What does it do to how much stats matter?
- Add `--best-of 5`. `wins` and the `[:3]` are the places to look, and the test's `3` becomes a parameter.

## 💀 Terrifying fact

Ranging over a map, as `fill` does with `repl`, visits keys in random order, and here that's *fine*, because each placeholder is independent. But it took a moment's thought to be sure: if `{W}`'s replacement could contain `{L}`, the order would matter and the output would vary between runs. Whenever you range over a map to *build* something, ask whether the order could show.

## 🕯️ Before dawn

Write a `mash --json` that prints the `Bout`. `printJSON` from Night 26 is already in `cmd/output.go`. Then look at the `json:"-"` tags and decide what a script reading the output actually needs about each fighter. Add it.

> 📺 *"Three rounds, one winner, and somebody limping home muttering about garlic. Tomorrow night we look up. The moon is a calculation, and I'm going to show you the sum."*

[← Night 22](night-22.md) · [Index](README.md) · [Night 24 →](night-24.md)
