package ui

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/quiz"
)

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

// Percent is the score out of 100 across all questions.
func (r QuizResult) Percent() int {
	if len(r.Questions) == 0 {
		return 0
	}
	c, _ := r.Score()
	return c * 100 / len(r.Questions)
}

// CorrectBy counts correct answers per monster id.
func (r QuizResult) CorrectBy() map[string]int {
	out := map[string]int{}
	for i, a := range r.Answers {
		if a >= 0 && a == r.Questions[i].Answer {
			out[r.Questions[i].MonsterID]++
		}
	}
	return out
}

type quizModel struct {
	res      QuizResult
	i        int
	cursor   int
	answered bool
	finished bool
	quip     string
	rng      *rand.Rand
	width    int
}

func newQuizModel(qs []quiz.Question, r *rand.Rand) quizModel {
	ans := make([]int, len(qs))
	for i := range ans {
		ans[i] = -1
	}
	return quizModel{res: QuizResult{Questions: qs, Answers: ans}, rng: r}
}

func (m quizModel) Init() tea.Cmd { return nil }

func (m quizModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
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
	}
	return m, nil
}

func (m *quizModel) answer(n int) {
	m.res.Answers[m.i] = n
	m.answered = true
	m.cursor = n
	m.quip = host.Quip(m.rng)
}

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

func (m quizModel) View() string {
	w := min(m.widthOr(80), 90)
	correct, _ := m.res.Score()
	var b strings.Builder
	b.WriteString(gameHeader("🎃 THE MIDNIGHT QUIZ", fmt.Sprintf("Question %d of %d · Score %d", m.i+1, len(m.res.Questions), correct), w) + "\n\n")

	if m.finished {
		pct := m.res.Percent()
		b.WriteString(headingStyle.Render(fmt.Sprintf("Final score: %d of %d (%d%%)", correct, len(m.res.Questions), pct)) + "\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGold).Render("Your rank: "+quiz.Rank(pct)) + "\n\n")
		b.WriteString(hostStyle.Render(wrap("📺 "+host.Name+": “"+finalWord(pct)+"”", w)) + "\n\n")
		b.WriteString(helpStyle.Render("Press any key to leave the studio") + "\n")
		return b.String()
	}

	q := m.res.Questions[m.i]
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(wrap(q.Prompt, w)) + "\n")
	if q.Clue != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorSky).Italic(true).PaddingLeft(2).Render(wrap(q.Clue, w-2)) + "\n")
	}
	b.WriteString("\n")
	for i, o := range q.Options {
		b.WriteString(optionLine(i, o, m.cursor == i, m.answered, i == q.Answer, i == m.res.Answers[m.i]) + "\n")
	}
	b.WriteString("\n")
	if m.answered {
		if m.res.Answers[m.i] == q.Answer {
			b.WriteString(trueStyle.Render("✔ Correct!") + "\n")
		} else {
			b.WriteString(mythStyle.Render("✘ Not quite. The answer is "+q.Correct()+".") + "\n")
		}
		b.WriteString(factStyle.Render(wrap(q.Explanation, w)) + "\n")
		b.WriteString(hostStyle.Render("📺 "+m.quip) + "\n\n")
		b.WriteString(helpStyle.Render("enter next question · q quit") + "\n")
	} else {
		b.WriteString(helpStyle.Render(joinFit([]string{"1-4 or a-d answer", "↑/↓ + enter", "q quit"}, " · ", w)) + "\n")
	}
	return b.String()
}

func (m quizModel) widthOr(d int) int {
	if m.width > 0 {
		return m.width
	}
	return d
}

func finalWord(pct int) string {
	switch {
	case pct >= 90:
		return "Magnificent! You must have been watching the late show since you were in your cradle."
	case pct >= 70:
		return "Very impressive. The monsters are nervous about how much you know."
	case pct >= 40:
		return "Not bad at all. Stay up for a few more features and you'll be an expert."
	default:
		return "A frightful score! But every monster scholar starts somewhere. Try again."
	}
}

// gameHeader is the title bar shared by the games.
func gameHeader(title, right string, w int) string {
	left := titleStyle.Render(title)
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		return left + "\n" + helpStyle.Render(right)
	}
	return left + strings.Repeat(" ", gap) + helpStyle.Render(right)
}

// optionLine draws one multiple-choice option, marking right and wrong
// answers once the player has chosen.
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

// RunQuiz plays a quiz in the terminal.
func RunQuiz(qs []quiz.Question, r *rand.Rand) (QuizResult, error) {
	final, err := tea.NewProgram(newQuizModel(qs, r)).Run()
	if err != nil {
		return QuizResult{}, err
	}
	return final.(quizModel).res, nil
}
