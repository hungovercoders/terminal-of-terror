# Night 22 · Captures and Badges

> 📺 *"Week four. The finale. You've got a memory now; tonight you get something to remember. Answer three questions about a monster and it's yours, locked in the crypt. Do something clever, or merely stay up too late, and there's a badge for it. Fourteen badges. I designed the cat one myself."*

**Tonight you'll learn**
- Separating the *rules* of progress from the storage and the screens
- One `Outcome` type for every kind of play
- Badges as data, and a small type that awards them
- Small predicate helpers: `every` and `allWhere`
- Functions as values: `monsters.Monster.IsSilent` as an argument
- A subcommand with a confirmation prompt

**Where we are:** games save their scores. Tonight the scores mean something.

## Three packages, three jobs

The progress feature is split across three packages, and the split is the lesson:

- `store` (Night 21) knows how to *save and load* a `Progress`. It knows nothing about quizzes.
- `crypt` (tonight) knows the *rules*: what captures a monster, what earns a badge. It knows nothing about files.
- `ui` draws the crypt. It knows nothing about either, beyond the struct.

Because the rules are a pure function from an outcome to a change in `Progress`, they're tested with no files and no screens, in a few milliseconds. And because the store doesn't know the rules, it never needs changing when a badge is added.

## One outcome type

Every game ends in the same call, `record(crypt.Outcome{...})`, with the fields that apply filled in, from [`internal/crypt/crypt.go`](../internal/crypt/crypt.go):

```go
// Outcome describes one session of play.
type Outcome struct {
	Kind      string         // quiz, guess, mash or explore
	Correct   map[string]int // correct answers per monster id
	Score     int            // percent for quizzes, points for guessing
	Total     int            // questions or rounds played
	Completed bool           // played to the end
	Seen      []string       // monster pages visited
	EagleEye  bool           // guessed a monster through the thickest fog
	Now       time.Time
}
```

A quiz sets `Correct`, `Score`, `Total`, `Completed`. The explorer sets `Seen`. A Monster Mash sets only `Kind`. One type with optional fields, rather than four types, means one `Apply` function and one `record`. `Now` is a field so tests can say it's Halloween.

## Badges are data

```go
// Badge is an achievement.
type Badge struct {
	ID          string `json:"id"`
	Emoji       string `json:"emoji"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Badges lists every achievement in display order.
var Badges = []Badge{
	{"first-fright", "🎃", "First Fright", "Finish your first quiz"},
	{"flawless", "🏆", "Flawless Fiend", "Score 100% on a quiz of 5 or more questions"},
	{"grave-robber", "🦴", "Grave Robber", "Capture your first monster"},
	{"eagle-eye", "🔍", "Eagle Eye", "Name a monster in Guess the Monster before the fog lifts"},
	...
	{"master", "👑", "Master of the Crypt", "Capture every monster"},
}
```

The crypt screen lists these in order, locked or earned, with no code per badge. The *rule* for each one lives in `Apply`; the *description* lives here, and the two are joined by the id. When the two disagree, the description is what the player reads, so it's the one to get right.

## Apply

`Apply` takes the progress, the full monster list and an outcome, changes the progress, and returns what was *newly* earned so the command can announce it. It reads as a table of contents:

```go
func Apply(p *store.Progress, all []monsters.Monster, o Outcome) Unlocks {
	if o.Now.IsZero() {
		o.Now = time.Now()
	}
	a := &awarder{p: p, now: o.Now}
	a.recordGame(o)
	if o.Kind == "quiz" || o.Kind == "guess" || o.Kind == "mash" {
		a.calendarBadges()
	}
	a.recordKnowledge(all, o)
	a.collectionBadges(all)
	return a.u
}
```

The work happens on an `awarder`, a small struct holding the progress, the time and the unlocks collected so far:

```go
type awarder struct {
	p   *store.Progress
	now time.Time
	u   Unlocks
}

// award gives badge id unless it was already won.
func (a *awarder) award(id string) {
	if _, ok := a.p.Badges[id]; ok {
		return
	}
	for _, b := range Badges {
		if b.ID == id {
			a.p.Badges[id] = a.now
			a.u.Badges = append(a.u.Badges, b)
		}
	}
}

// awardIf gives badge id when cond holds.
func (a *awarder) awardIf(cond bool, id string) {
	if cond {
		a.award(id)
	}
}
```

`award` is idempotent: awarding a badge you already have does nothing, so the rules can be careless about calling it and read as a plain list. The methods have *pointer* receivers, `*awarder`, because they change it: each one appends to the same `a.u`. Then the rules:

```go
func (a *awarder) recordGame(o Outcome) {
	p := a.p
	finished := o.Completed && o.Total > 0
	switch o.Kind {
	case "quiz":
		if finished {
			p.QuizzesPlayed++
			p.BestQuizScore = max(p.BestQuizScore, o.Score)
			a.award("first-fright")
			a.awardIf(o.Score == 100 && o.Total >= 5, "flawless")
		}
	case "guess":
		...
	case "mash":
		p.MashesPlayed++
		a.award("promoter")
	}
}

func (a *awarder) calendarBadges() {
	a.awardIf(a.now.Hour() < 4, "night-owl")
	a.awardIf(calendar.IsFullMoon(a.now), "full-moon")
	...
}
```

Every rule is a line, and each maps to one line of `Badges`. `Flawless Fiend` needs five questions, because a 100% on a one-question quiz isn't flawless, it's lucky; `TestShortQuizIsNotFlawless` pins that. The calendar badges call into Night 24's package, so `crypt` knows nothing about moon maths either.

Then the captures, in `recordKnowledge`:

```go
	for id, n := range o.Correct {
		p.Knowledge[id] += n
	}
	for _, m := range all {
		if _, done := p.Captured[m.ID]; !done && p.Knowledge[m.ID] >= CaptureAt {
			p.Captured[m.ID] = a.now
			a.u.Captured = append(a.u.Captured, m)
		}
	}
```

Knowledge accumulates across sessions; three right answers about Dracula, tonight or over a month, capture him. `CaptureAt` is a named constant, used by the crypt screen for its `●●○` progress pips too, so changing the rule to four is one edit.

Splitting `Apply` this way isn't decoration. The first version was one long function with an `award` closure, and it grew a branch for every badge until it was the most tangled function in the project. Each method now does one job and fits on a screen, and the order they're called in is the order badges are announced.

## Small predicates

The collection badges ask questions like "is every silent-film monster captured?" and "is every Universal monster captured?". Rather than a loop each, `collectionBadges` asks them with predicates:

```go
	captured := func(m monsters.Monster) bool { _, ok := p.Captured[m.ID]; return ok }
	inPack := func(pack string) func(monsters.Monster) bool {
		return func(m monsters.Monster) bool { return m.Pack == pack }
	}
	a.awardIf(allWhere(all, monsters.Monster.IsSilent, captured), "silent-scholar")
	a.awardIf(allWhere(all, inPack("universal"), captured), "monster-kid")
	...
	a.awardIf(every(all, captured), "master")
```

with two helpers:

```go
// every reports whether every monster satisfies ok (false for none).
func every(all []monsters.Monster, ok func(monsters.Monster) bool) bool {
	return allWhere(all, func(monsters.Monster) bool { return true }, ok)
}

// allWhere reports whether every monster matching where satisfies ok,
// and that at least one matches.
func allWhere(all []monsters.Monster, where, ok func(monsters.Monster) bool) bool {
	n := 0
	for _, m := range all {
		if !where(m) {
			continue
		}
		n++
		if !ok(m) {
			return false
		}
	}
	return n > 0
}
```

**Functions are values.** `allWhere` takes two functions as arguments. `captured` is a closure; `inPack("universal")` calls a function that *returns* one, a different one per pack; and `monsters.Monster.IsSilent` is a *method expression*: the `IsSilent` method, detached from any particular monster, as a function that takes the monster as its first argument. All three have the type `func(monsters.Monster) bool`, so all three fit.

The `n > 0` at the end is the subtle bit: "every silent monster is captured" must be *false* when there are no silent monsters, or a `--pack folklore` player would earn Silent Era Scholar for free. "All of nothing" is a classic bug in rule engines; the comment names it.

## Announcing

`record` (Night 21) passes the `Unlocks` to `ui.RenderUnlocks`, which prints a line per new capture and badge after the game's output:

```
💀 CAPTURED: 🧛 Dracula has joined your crypt
🏅 BADGE: 🦴 Grave Robber · Capture your first monster
See your collection: terminal-of-terror crypt
```

Only *new* things are announced, because `award` and the capture loop only add to `a.u` when something changed. That's why `Apply` returns unlocks instead of the command comparing before and after.

## The crypt and its reset

[`cmd/crypt.go`](../cmd/crypt.go) renders the collection. It also has a **subcommand**, `crypt reset`, which is a `cobra.Command` added to `cryptCmd` instead of `rootCmd`:

```go
var cryptResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Empty your crypt and start again",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !resetYes {
			fmt.Print("This releases every captured monster and forgets all your badges. Type 'yes' to confirm: ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) != "yes" {
				fmt.Println("Phew. Your crypt is untouched.")
				return nil
			}
		}
		if err := store.Reset(); err != nil {
			return err
		}
		fmt.Println("The crypt doors swing open... every monster has escaped. Your progress has been reset.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cryptCmd)
	cryptCmd.AddCommand(cryptResetCmd)
	cryptResetCmd.Flags().BoolVarP(&resetYes, "yes", "y", false, "Don't ask for confirmation")
}
```

A destructive command asks first, reading a line from standard input with `bufio`, and requires the whole word "yes". The `--yes` flag skips the question for scripts. That pair, prompt by default and a flag to skip it, is the convention for anything that deletes.

## Run it

```bash
export TERMINAL_OF_TERROR_HOME=/tmp/tot-scratch
go run . quiz dracula -n 5
go run . crypt
```

Get three right and Dracula is captured, with Grave Robber and First Fright. `crypt` shows him ticked, everyone else with pips, and the badge list. `go run . crypt reset` asks; type `no`.

## Testing the rules

[`crypt_test.go`](../internal/crypt/crypt_test.go) never touches a file: `store.New()` gives an empty progress in memory.

```go
func TestCaptureAfterEnoughCorrectAnswers(t *testing.T) {
	p := store.New()
	all := monsters.GetAllMonsters()

	u := Apply(p, all, Outcome{Kind: "quiz", Correct: map[string]int{"dracula": 2}, Score: 20, Total: 10, Completed: true, Now: evening})
	if len(u.Captured) != 0 {
		t.Fatal("2 correct answers should not capture yet")
	}
	if !hasBadge(u, "first-fright") {
		t.Fatal("first completed quiz should earn First Fright")
	}

	u = Apply(p, all, Outcome{Kind: "quiz", Correct: map[string]int{"dracula": 1}, Score: 10, Total: 10, Completed: true, Now: evening})
	if len(u.Captured) != 1 || u.Captured[0].ID != "dracula" {
		t.Fatalf("expected Dracula captured, got %+v", u.Captured)
	}
	if !hasBadge(u, "grave-robber") || hasBadge(u, "first-fright") {
		t.Fatalf("expected only new badges, got %+v", u.Badges)
	}
}
```

Two sessions, 2 + 1 answers, a capture on the second and no *repeat* of First Fright. `evening` is a fixed `time.Time` at 8pm, so Night Owl can't sneak in depending on when the tests run; `TestCalendarBadges` uses specific dates for the moon and Friday the 13th.

## Try it

- Add a badge: "Regular", for playing ten quizzes. One line in `Badges`, one `awardIf` in `recordGame`, one case in a test.
- `Knowledge` never goes down. Should a wrong answer cost a point? Try it and see how it changes the feel of the quiz.
- Run `go run . crypt --json` and read the same data the screen draws.

## 💀 Terrifying fact

`monsters.Monster.IsSilent` works because `IsSilent` has a *value* receiver. A method with a pointer receiver, `func (m *Monster) X()`, has the method expression `(*monsters.Monster).X` and takes a pointer as its first argument, which wouldn't fit `allWhere`. Value receivers on small, read-only methods keep this kind of composition easy, which is one more reason the project uses them everywhere it can.

## 🕯️ Before dawn

Design a badge that's *hard*: it needs something across several sessions that `Progress` doesn't currently record. Add the field, the rule and the test. Notice the store didn't need to know.

> 📺 *"Captured, catalogued, and awarded a small cat. Tomorrow night the monsters stop answering questions and start throwing each other through mausoleum doors."*

[← Night 21](night-21.md) · [Index](README.md) · [Night 23 →](night-23.md)
