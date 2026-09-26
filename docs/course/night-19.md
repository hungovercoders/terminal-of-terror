# Night 19 · The Midnight Quiz

> 📺 *"The questions are written. Tonight we build the studio: a question on screen, four answers, a cursor, a buzzer, and at the end a rank you can boast about. No lifelines. The phone-a-friend is a rotary and it's disconnected."*

**Tonight you'll learn**
- A second, separate Bubble Tea program
- A result type that the game fills in and the command reads
- Two-phase interaction: choose, then see feedback, then continue
- Mapping keys to option numbers
- A command with flags, and an error for a bad flag value
- Playing a whole game in a test

**Where we are:** `quiz.Generate` produces questions; nothing shows them yet.

## A program of its own

The explorer is one Bubble Tea program. The quiz is *another*, with its own model, in [`internal/ui/quiz.go`](../../internal/ui/quiz.go). They share the package and the styles, nothing else. Bolting the quiz onto the explorer's model as a fourth screen would have meant every explorer key handler checking "unless we're in a quiz". Separate programs for separate activities keeps each one small.

```go
type quizModel struct {
	res      QuizResult
	i        int  // current question
	cursor   int
	answered bool // chosen, showing feedback
	finished bool
	quip     string
	rng      *rand.Rand
	width    int
}
```

## The result comes out the end

Night 15 showed `p.Run()` returning the final model. The quiz leans on that: everything the command needs to know is in a `QuizResult` that the model fills in as the player goes:

```go
// QuizResult is how a quiz went.
type QuizResult struct {
	Questions []quiz.Question
	Answers   []int // chosen option per question, -1 if unanswered
	Completed bool
}

// Score returns correct answers and questions answered.
func (r QuizResult) Score() (correct, answered int) {
	for i, a := range r.Answers {
		if a < 0 {
			continue
		}
		answered++
		if a == r.Questions[i].Answer {
			correct++
		}
	}
	return correct, answered
}
```

`Answers` starts as all `-1`, meaning unanswered, which is how the command tells "quit on question one" from "got question one wrong". **Named return values** (`(correct, answered int)`) are declared by the signature and start at zero, so the function just counts and returns. `Percent()` and `CorrectBy()` (correct answers per monster, for Night 22) are two more methods on the same struct. Methods on a plain data type, computing views of it, are the Go way to avoid a class hierarchy.

```go
// RunQuiz plays a quiz in the terminal.
func RunQuiz(qs []quiz.Question, r *rand.Rand) (QuizResult, error) {
	final, err := tea.NewProgram(newQuizModel(qs, r)).Run()
	if err != nil {
		return QuizResult{}, err
	}
	return final.(quizModel).res, nil
}
```

No `WithAltScreen` this time: the quiz is short and it's nice to be able to scroll back and see what you got wrong.

## Two phases

A question has two states: *choosing*, where arrows and number keys pick an option, and *answered*, where the right answer is marked and any of enter/space/n moves on. `Update` handles them in order:

```go
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}
		if m.finished {
			return m, tea.Quit
		}
		q := m.res.Questions[m.i]
		if m.answered {
			if k == "enter" || k == " " || k == "n" || k == "right" {
				m.i++
				m.answered, m.cursor = false, 0
				if m.i == len(m.res.Questions) {
					m.i--
					m.finished = true
					m.res.Completed = true
				}
			}
			return m, nil
		}
		switch k {
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(q.Options)) % len(q.Options)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(q.Options)
		case "enter", " ":
			m.answer(m.cursor)
		default:
			if n := optionKey(k); n >= 0 && n < len(q.Options) {
				m.answer(n)
			}
		}
```

The guards at the top are ordered from "always" to "sometimes": quit keys work anywhere; on the results screen any key leaves; while showing feedback only the continue keys do anything; otherwise we're choosing. Each phase returns before the next is considered, so no phase has to know about the others. When you find a screen with several states, this is the shape: a ladder of `if ... return`.

Notice `m.i--` when the last question is passed: the results screen still wants to show "Question 10 of 10" and `m.i` must stay a valid index. Off-by-one at the end of a list is where most crashes in this kind of code live; a comment or a test on the boundary is cheap.

```go
// optionKey maps 1-4 and a-d to an option index, or -1.
func optionKey(k string) int {
	if len(k) != 1 {
		return -1
	}
	switch c := k[0]; {
	case c >= '1' && c <= '9':
		return int(c - '1')
	case c >= 'a' && c <= 'd':
		return int(c - 'a')
	}
	return -1
}
```

Bytes are numbers: `'3' - '1'` is 2. A one-character key that's a digit or a-d becomes an index; the caller checks it's in range for this question, since a true/false question has only two options.

## Feedback

`answer` records the choice, and asks the host for a quip, one of Night 15's random lines:

```go
func (m *quizModel) answer(n int) {
	m.res.Answers[m.i] = n
	m.answered = true
	m.cursor = n
	m.quip = host.Quip(m.rng)
}
```

Then `View` draws the options through `optionLine`, which takes six booleans and turns them into a marker and a style:

```go
func optionLine(i int, text string, selected, answered, isAnswer, isChosen bool) string {
	marker := "  "
	if selected && !answered {
		marker = lipgloss.NewStyle().Foreground(colorBlood).Render("▶ ")
	}
	label := fmt.Sprintf("%d. %s", i+1, text)
	style := lipgloss.NewStyle()
	switch {
	case answered && isAnswer:
		marker, style = trueStyle.Render("✔ "), trueStyle
	case answered && isChosen:
		marker, style = mythStyle.Render("✘ "), mythStyle
	case answered:
		style = helpStyle
	case selected:
		style = style.Bold(true).Foreground(colorGold)
	}
	return marker + style.Render(label)
}
```

Before answering: a red cursor and a gold line. After: the right answer green with a tick, a wrong choice red with a cross, the rest dimmed. The same function draws Night 20's options, which is why it's a free function rather than a method. Under the options, a line of feedback, the explanation, and the quip:

```go
			// Don't double the full stop after answers like "Lon Chaney Jr."
			b.WriteString(mythStyle.Render(wrap("✘ Not quite. The answer is "+strings.TrimSuffix(q.Correct(), ".")+".", w)) + "\n")
```

That comment records a real bug from a code review: "The answer is Lon Chaney Jr.." Data with punctuation in it will always find your string concatenation.

## The command

[`cmd/quiz.go`](../../cmd/quiz.go):

```go
var quizCmd = &cobra.Command{
	Use:   "quiz [monster]",
	Short: "Test your monster knowledge in the Midnight Quiz",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if quizQuestions < 1 {
			return fmt.Errorf("--questions must be at least 1")
		}
		focus := ""
		if len(args) == 1 {
			m, err := resolveMonster(args[0])
			if err != nil {
				return err
			}
			focus = m.ID
		}
		r := newRand()
		qs := quiz.Generate(r, monsters.GetAllMonsters(), quizQuestions, focus)
		if len(qs) == 0 {
			return fmt.Errorf("couldn't find any questions to ask")
		}
		res, err := ui.RunQuiz(qs, r)
		if err != nil {
			return err
		}
		_, answered := res.Score()
		if answered == 0 {
			return nil
		}
		record(crypt.Outcome{Kind: "quiz", Correct: res.CorrectBy(), Score: res.Percent(), Total: len(qs), Completed: res.Completed})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(quizCmd)
	quizCmd.Flags().IntVarP(&quizQuestions, "questions", "n", 10, "Number of questions")
}
```

**`IntVarP`** binds an integer flag to a variable, with a long name, a short name and a default: `--questions 5` or `-n 5`. Cobra parses it before `RunE` runs, so the variable is already set. Cobra checks that it's a number; the command checks that it's a *sensible* number, because `--questions 0` would parse fine and then generate nothing.

`newRand()` (Night 26 explains it) gives every random choice one seeded source, so a demo recording can replay the same quiz. The three lines at the end, `record(...)`, are Night 21 and 22: the score is saved, and captures are announced. Tonight they're a stub, and the command doesn't care.

## Run it

```bash
go run . quiz -n 5
go run . quiz dracula
```

Choose with `1`-`4` or arrows and enter. Read the explanation, press enter, repeat. At the end: a score, a percentage, a rank and a word from the host.

## Playing in a test

Games are the easiest programs to test, because the whole thing is "keys in, score out". [`games_test.go`](../../internal/ui/games_test.go):

```go
func step(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		m, _ = m.Update(key(k))
	}
	return m
}

func TestQuizPlaythrough(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	qs := quiz.Generate(r, monsters.GetAllMonsters(), 4, "")
	var m tea.Model = newQuizModel(qs, r)
	m, _ = m.Update(size(80, 24))

	// Answer the first question correctly, the rest wrongly.
	m = step(m, string(rune('1'+qs[0].Answer)), "enter")
	for _, q := range qs[1:] {
		wrong := (q.Answer + 1) % len(q.Options)
		m = step(m, string(rune('1'+wrong)))
		if !strings.Contains(m.View(), "The answer is") {
			t.Fatal("wrong answer should show the right one")
		}
		m = step(m, "enter")
	}
	qm := m.(quizModel)
	if !qm.finished || !qm.res.Completed {
		t.Fatal("quiz should be finished")
	}
	if c, n := qm.res.Score(); c != 1 || n != 4 || qm.res.Percent() != 25 {
		t.Fatalf("score = %d/%d (%d%%)", c, n, qm.res.Percent())
	}
	if !strings.Contains(m.View(), quiz.Rank(25)) {
		t.Error("results should show the rank")
	}
}
```

The test doesn't know which questions it got; it reads the answer out of each question and presses that key, or the next one along for a wrong answer. `string(rune('1'+qs[0].Answer))` turns an index into the key `"1"`, `"2"`... The score at the end is checked exactly. There's no terminal, no timing and no randomness leaking in, because the seed is fixed.

## Try it

- Add a timer: a `tickMsg` every second, a countdown per question, and an automatic wrong answer at zero. Night 15's `tick` is right there.
- Show a running streak ("3 in a row!") in the header.
- Make `q` during a quiz *finish* it (show results for what was answered) rather than quit outright. What should `Completed` be then?

## 💀 Terrifying fact

`final.(quizModel)` panics if the program was somehow left with a different model type. It can't happen here, since `Update` always returns a `quizModel`, but the explorer's `RunUI` uses the two-value form, `fm, ok := final.(model)`, and returns quietly if `!ok`. Prefer the two-value form whenever the assertion could fail; use the one-value form only where a failure would be a programming error worth crashing on.

## 🕯️ Before dawn

`quiz --json` prints the generated questions *with* the answers. Add `--no-answers` that strips `Answer` and `Explanation` before printing, so a teacher could hand out the questions. The `Question` struct needs `omitempty` on those fields, or the answer index prints as `0`, which is a real answer.

> 📺 *"Ten questions, one rank, no arguing with the judges. Tomorrow night the pictures come back, but you'll barely be able to see them. That's the game."*

[← Night 18](night-18.md) · [Index](README.md) · [Night 20 →](night-20.md)
