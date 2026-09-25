# Night 18 · Writing the Questions

> 📺 *"I've told you a hundred facts and you've nodded politely at every one. Tonight we find out what stuck. But a quiz needs questions, and I refuse to write them by hand, because then I'd have to write more every time a new monster checked in. So tonight the program writes its own questions, out of the facts it already has."*

**Tonight you'll learn**
- Generating content from data instead of authoring it
- Hiding the answer: masking names with a regular expression
- Distractors: wrong answers that are plausible and never accidentally right
- Shuffling with a seeded source
- Keeping a mix with a per-kind quota
- Property tests: fifty seeds, no bad question

**Where we are:** the explorer is complete. Tonight is pure logic, no screen.

## Questions for free

Every monster has facts, quotes, a debut, myths and film credits. Each of those is a question waiting to be asked:

| From | Question |
|------|----------|
| a fact | "Which monster is this about?" with the name hidden |
| a quote | "Whose story does this line come from?" |
| the debut | "Where did X first appear?" book, film or folklore |
| a myth | "True or false?" |
| the film | who played it, who directed it, what year |

That's the entire design of [`internal/quiz/quiz.go`](../../internal/quiz/quiz.go). Add a monster on Night 27 and the quiz has a dozen new questions about it, without anyone writing one.

```go
// Question is one multiple-choice question.
type Question struct {
	Kind        string   `json:"kind"`
	Prompt      string   `json:"prompt"`
	Clue        string   `json:"clue,omitempty"` // quoted text shown under the prompt
	Options     []string `json:"options"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"`
	MonsterID   string   `json:"monsterId"`
}
```

`Answer` is an index into `Options`. `MonsterID` is what Night 22 uses to give credit for a right answer. The JSON tags are for Night 26's `quiz --json`.

## Hiding the name

"Which monster is this about? *Dracula* can transform into a bat" is not a question. `mask` blanks the monster's name and aliases out of a clue:

```go
// mask hides a monster's names in text so the clue doesn't give it away.
func mask(text string, m monsters.Monster) string {
	terms := append([]string{m.Name, strings.TrimPrefix(m.Name, "The ")}, m.Aliases...)
	sort.Slice(terms, func(i, j int) bool { return len(terms[i]) > len(terms[j]) })
	for _, t := range terms {
		if len(t) < 3 {
			continue
		}
		pattern := `(?i)\b` + regexp.QuoteMeta(t) + `('s)?\b`
		if !strings.Contains(t, " ") {
			// Single words also hide compounds like "Krampusnacht".
			pattern = `(?i)\b` + regexp.QuoteMeta(t) + `[\p{L}']*`
		}
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "▒▒▒▒")
	}
	return text
}
```

This is the project's one **regular expression**, and it earns its place. `(?i)` makes it case-insensitive, `\b` is a word boundary, `regexp.QuoteMeta` escapes anything in the name that means something in a regex (the `.` in "Jr."), and `('s)?` swallows a possessive so "Dracula's cape" becomes "▒▒▒▒ cape" rather than "▒▒▒▒'s cape", which is a hint. Single-word names also eat any letters glued on after them, so "Krampusnacht" doesn't leak "Krampus".

The terms are sorted longest first so "Count Dracula" is masked as one blank before "Dracula" alone gets a chance to leave "Count ▒▒▒▒". And names shorter than three letters are skipped, or a monster called "Ra" would blank out every "ra" in "Transylvania".

`regexp.MustCompile` panics on a bad pattern, which is the right choice here: the pattern is built from data and if it's ever wrong you want to know at test time, not to silently skip a mask.

## Distractors

A multiple-choice question is only as good as its wrong answers. They must be plausible (other monsters' names, other directors), never duplicate each other, and never *be* the right answer under a different spelling. `choiceOK` does all three:

```go
func choiceOK(r *rand.Rand, q Question, answer string, distractors []string) (Question, bool) {
	d := unique(without(distractors, []string{answer}))
	r.Shuffle(len(d), func(i, j int) { d[i], d[j] = d[j], d[i] })
	if len(d) > 3 {
		d = d[:3]
	}
	opts := append([]string{answer}, d...)
	r.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	q.Options = opts
	for i, o := range opts {
		if o == answer {
			q.Answer = i
		}
	}
	return q, len(d) >= 2
}
```

Remove the answer (case-insensitively; `without` lowercases), remove duplicates, shuffle, keep three, add the answer, shuffle again, and find where the answer ended up. The second return value says whether there were enough distractors for a fair question. James Whale directed three of the vault's films, so "who directed Frankenstein?" has fewer distinct wrong directors than it looks; if a pack ever has fewer than two, the question is dropped rather than asked with one option.

**`r.Shuffle(n, swap)`** is Go's Fisher-Yates: you give it the length and a function that swaps two elements, and it calls the swap in the right pattern. It takes a *function* because it doesn't know what it's shuffling; the closure does.

The years question uses `nearbyYears` instead: offsets of ±1 to ±8 in random order, so the wrong years for 1931 are things like 1929, 1933 and 1936, not 1874.

## What not to ask

There's a comment in `questionsAbout` worth reading:

```go
	// No "which is its weakness?" questions: weaknesses written for different
	// monsters overlap in meaning ("Sunlight", "The coming of dawn"), so a
	// wrong option could be just as right as the answer.
```

An earlier version asked about weaknesses, and a test caught that "Sunlight" and "The coming of dawn" could both be offered for a vampire, one as the answer and one as a distractor. There's no automatic way to know two strings mean the same thing, so the question kind was removed, and a test, `TestNoAmbiguousWeaknessQuestions`, makes sure it stays removed. Deleting a feature is sometimes the fix.

## A mix, not a pile

If every fact became a question, a ten-question quiz drawn at random would be nine "which monster is this about?" and one "true or false?". `Generate` shuffles the pool and then fills the quiz with a quota per kind:

```go
	// Keep a mix: no kind may take more than a third of the quiz, unless
	// we would otherwise run short.
	limit := max((n+2)/3, 1)
	var out, spare []Question
	perKind := map[string]int{}
	for _, q := range pool {
		if len(out) == n {
			break
		}
		if perKind[q.Kind] < limit {
			perKind[q.Kind]++
			out = append(out, q)
		} else {
			spare = append(spare, q)
		}
	}
	for _, q := range spare {
		if len(out) == n {
			break
		}
		out = append(out, q)
	}
```

`perKind` is a map used as a counter; reading a missing key gives 0, so there's no need to initialise anything. Questions over the quota go to `spare`, and if the first pass comes up short (a quiz about one monster, say, with `focus` set), the spares fill in. Two passes and a map: that's most quota problems.

`(n+2)/3` is Night 15's round-up again: a third of 10 is 4, not 3.

## Ranks

```go
// Rank names a quiz score (0-100).
func Rank(percent int) string {
	switch {
	case percent >= 90:
		return "Master of Horror"
	case percent >= 70:
		return "Monster Scholar"
	case percent >= 40:
		return "Creature Feature Fan"
	default:
		return "Ghoul-in-Training"
	}
}
```

A bare `switch` with descending thresholds is the idiom for banding a number. The first true case wins, so the order matters.

## Fifty seeds

The test, in [`quiz_test.go`](../../internal/quiz/quiz_test.go), doesn't check any specific question. It checks *properties* that every question must have, across fifty different random seeds:

```go
func TestGeneratedQuestionsAreWellFormed(t *testing.T) {
	all := monsters.GetAllMonsters()
	for seed := int64(1); seed <= 50; seed++ {
		qs := Generate(rand.New(rand.NewSource(seed)), all, 10, "")
		if len(qs) != 10 {
			t.Fatalf("seed %d: got %d questions", seed, len(qs))
		}
		prompts := map[string]bool{}
		for _, q := range qs {
			if q.Answer < 0 || q.Answer >= len(q.Options) || len(q.Options) < 2 {
				t.Fatalf("seed %d: bad answer index: %+v", seed, q)
			}
			seen := map[string]bool{}
			for _, o := range q.Options {
				if seen[o] {
					t.Fatalf("seed %d: duplicate option %q in %+v", seed, o, q)
				}
				seen[o] = true
			}
			key := q.Prompt + q.Clue
			if prompts[key] {
				t.Fatalf("seed %d: repeated question %q", seed, key)
			}
			prompts[key] = true
			if q.Kind == "fact" || q.Kind == "quote" {
				m := monsters.GetMonsterByName(q.MonsterID)
				if strings.Contains(strings.ToLower(q.Clue), strings.ToLower(strings.TrimPrefix(m.Name, "The "))) {
					t.Errorf("clue gives away the answer: %q", q.Clue)
				}
			}
		}
	}
}
```

The answer index is valid, no option repeats, no question repeats, no clue contains the name. Five hundred questions checked in a few milliseconds. When a random generator has a bug, it usually shows on *some* seed; running fifty is how you find it before a player does. `%+v` prints a struct with its field names, which makes the failure message readable.

`TestMask` checks the possessive and the compound case directly, because those are the fiddly bits.

## Run it

```bash
go test ./internal/quiz -v
go run . quiz --json -n 3
```

Fifty seeds of questions checked in a blink, then three real questions as data, answers and all. The `--json` flag is Night 26's, but it's already there.

## Try it

- Add a question kind: "Which monster is afraid of *garlic*?" from `Weaknesses`. Then read the comment about weaknesses again and decide whether it's safe. Try picking distractors only from monsters whose weaknesses don't contain the same word.
- Set `limit` to `n` and generate a few quizzes. That's the pile.

## 💀 Terrifying fact

`sort.Slice` isn't stable: two terms of equal length can come out in either order. It doesn't matter for `mask`, but it's caught people: Go has `sort.SliceStable` and, since 1.21, `slices.SortFunc` and `slices.SortStableFunc` with a comparison function instead of a less function. New code should use the `slices` package; this file predates the habit.

## 🕯️ Before dawn

Write a table-driven test for `debutKind`: `"novel"` is a book, `"1931 film"` is a film, `"Slavic folklore"` is folklore, `"opera"` is nothing. Then add `"opera"` to the book case, because the Phantom's debut is a 1910 novel but someone will write "opera" one day.

> 📺 *"Two hundred questions and I didn't write one of them. Tomorrow night we put them on screen, with a scoreboard and a rank. Pencils down at midnight."*

[← Night 17](night-17.md) · [Index](README.md) · [Night 19 →](night-19.md)
