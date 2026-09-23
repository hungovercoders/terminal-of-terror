// Package quiz builds multiple-choice questions from the monster data, so
// every new fact, film credit or myth check becomes quiz material for free.
package quiz

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

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

// Correct returns the text of the right answer.
func (q Question) Correct() string { return q.Options[q.Answer] }

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

// Generate builds up to n varied questions. If focus is non-empty, every
// question is about that monster.
func Generate(r *rand.Rand, all []monsters.Monster, n int, focus string) []Question {
	var pool []Question
	for _, m := range all {
		if focus != "" && m.ID != focus {
			continue
		}
		pool = append(pool, questionsAbout(r, m, all)...)
	}
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

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
	return out
}

// questionsAbout builds every question we can ask about m.
func questionsAbout(r *rand.Rand, m monsters.Monster, all []monsters.Monster) []Question {
	var qs []Question
	names := func(except string) []string {
		var out []string
		for _, o := range all {
			if o.ID != except {
				out = append(out, o.Name)
			}
		}
		return out
	}

	// Which monster is this fact about?
	for _, f := range m.Facts {
		if q, ok := choiceOK(r, Question{
			Kind:        "fact",
			Prompt:      "Which monster is this about?",
			Clue:        mask(f, m),
			Explanation: m.Name + ": " + f,
			MonsterID:   m.ID,
		}, m.Name, names(m.ID)); ok {
			qs = append(qs, q)
		}
	}

	// Whose story does this line come from?
	for _, q := range m.Quotes {
		if len(q.Text) < 25 {
			continue
		}
		if qq, ok := choiceOK(r, Question{
			Kind:        "quote",
			Prompt:      "Whose story does this line come from?",
			Clue:        "“" + mask(q.Text, m) + "”",
			Explanation: "— " + q.Speaker + ", " + q.Source,
			MonsterID:   m.ID,
		}, m.Name, names(m.ID)); ok {
			qs = append(qs, qq)
		}
	}

	// Book, film or folklore?
	if medium := debutKind(m.Debut.Medium); medium != "" {
		q := Question{
			Kind:        "debut",
			Prompt:      fmt.Sprintf("Where did %s first appear?", m.Name),
			Options:     []string{"A book", "A film", "Folklore and legend"},
			Explanation: "First appearance: " + m.FirstAppearance(),
			MonsterID:   m.ID,
		}
		for i, o := range q.Options {
			if o == medium {
				q.Answer = i
			}
		}
		qs = append(qs, q)
	}

	// True or false?
	for _, my := range m.Myths {
		q := Question{
			Kind:        "myth",
			Prompt:      "True or false?",
			Clue:        my.Claim,
			Options:     []string{"True", "False"},
			Explanation: my.Explanation,
			MonsterID:   m.ID,
		}
		if !my.True {
			q.Answer = 1
		}
		qs = append(qs, q)
	}

	// Film credits.
	if f := m.Film; f != nil {
		var stars, directors []string
		for _, o := range all {
			if o.Film != nil && o.ID != m.ID {
				stars = append(stars, o.Film.Star)
				directors = append(directors, o.Film.Director)
			}
		}
		title := fmt.Sprintf("%s (%d)", f.Title, f.Year)
		if q, ok := choiceOK(r, Question{
			Kind:        "film",
			Prompt:      fmt.Sprintf("Who played the monster in %s?", title),
			Explanation: f.Star + " played the monster in " + title + ".",
			MonsterID:   m.ID,
		}, f.Star, stars); ok {
			qs = append(qs, q)
		}
		if q, ok := choiceOK(r, Question{
			Kind:        "film",
			Prompt:      fmt.Sprintf("Who directed %s?", title),
			Explanation: f.Director + " directed " + title + ".",
			MonsterID:   m.ID,
		}, f.Director, directors); ok {
			qs = append(qs, q)
		}
		qs = append(qs, choice(r, Question{
			Kind:        "year",
			Prompt:      fmt.Sprintf("In what year was %s released?", f.Title),
			Explanation: fmt.Sprintf("%s came out in %d.", f.Title, f.Year),
			MonsterID:   m.ID,
		}, fmt.Sprint(f.Year), nearbyYears(r, f.Year)))
	}

	// Which of these is its weakness?
	var others []string
	for _, o := range all {
		if o.ID != m.ID {
			others = append(others, o.Weaknesses...)
		}
	}
	others = without(others, m.Weaknesses)
	if len(m.Weaknesses) > 0 {
		w := m.Weaknesses[r.Intn(len(m.Weaknesses))]
		if q, ok := choiceOK(r, Question{
			Kind:        "weakness",
			Prompt:      fmt.Sprintf("According to its stat card, which of these is a weakness of %s?", m.Name),
			Explanation: m.Name + "'s weaknesses: " + strings.Join(m.Weaknesses, ", ") + ".",
			MonsterID:   m.ID,
		}, w, others); ok {
			qs = append(qs, q)
		}
	}
	return qs
}

// choice fills in four shuffled options: the answer and three distractors.
func choice(r *rand.Rand, q Question, answer string, distractors []string) Question {
	q, _ = choiceOK(r, q, answer, distractors)
	return q
}

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

func nearbyYears(r *rand.Rand, year int) []string {
	offsets := r.Perm(16)
	var out []string
	for _, o := range offsets {
		d := o/2 + 1
		if o%2 == 0 {
			d = -d
		}
		out = append(out, fmt.Sprint(year+d))
		if len(out) == 3 {
			break
		}
	}
	return out
}

func debutKind(medium string) string {
	m := strings.ToLower(medium)
	switch {
	case strings.Contains(m, "film"):
		return "A film"
	case strings.Contains(m, "novel"), strings.Contains(m, "story"), strings.Contains(m, "book"),
		strings.Contains(m, "poem"), strings.Contains(m, "play"):
		return "A book"
	case strings.Contains(m, "folklore"), strings.Contains(m, "legend"), strings.Contains(m, "myth"):
		return "Folklore and legend"
	}
	return ""
}

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

func unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func without(in, drop []string) []string {
	skip := map[string]bool{}
	for _, d := range drop {
		skip[strings.ToLower(d)] = true
	}
	var out []string
	for _, s := range in {
		if !skip[strings.ToLower(s)] {
			out = append(out, s)
		}
	}
	return out
}
