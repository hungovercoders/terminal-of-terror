package ui

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// Fog stages: how much of the portrait shows, and what a right answer is worth.
var (
	fogReveal = []float64{0.15, 0.3, 0.5, 0.75, 1.0}
	fogPoints = []int{100, 80, 60, 40, 20}
)

// GuessRound is one portrait to identify.
type GuessRound struct {
	Monster monsters.Monster
	Options []string
	Answer  int
	Chosen  int // -1 until answered
	Stage   int // fog stage when answered
	order   []int
}

// Correct reports whether the player named the monster.
func (g GuessRound) Correct() bool { return g.Chosen == g.Answer }

// Points scored this round.
func (g GuessRound) Points() int {
	if !g.Correct() {
		return 0
	}
	return fogPoints[g.Stage]
}

// GuessResult is how a game of Guess the Monster went.
type GuessResult struct {
	Rounds    []GuessRound
	Completed bool
}

// Score is the total points.
func (r GuessResult) Score() int {
	s := 0
	for _, g := range r.Rounds {
		s += g.Points()
	}
	return s
}

// CorrectBy counts correct guesses per monster id.
func (r GuessResult) CorrectBy() map[string]int {
	out := map[string]int{}
	for _, g := range r.Rounds {
		if g.Chosen >= 0 && g.Correct() {
			out[g.Monster.ID]++
		}
	}
	return out
}

// EagleEye reports whether any monster was named through the thickest fog.
func (r GuessResult) EagleEye() bool {
	for _, g := range r.Rounds {
		if g.Chosen >= 0 && g.Correct() && g.Stage == 0 {
			return true
		}
	}
	return false
}

// NewGuessRounds picks n different monsters that have portraits.
func NewGuessRounds(r *rand.Rand, all []monsters.Monster, n int) []GuessRound {
	var pool []monsters.Monster
	for _, m := range all {
		if m.ASCII != "" {
			pool = append(pool, m)
		}
	}
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	n = min(n, len(pool))
	rounds := make([]GuessRound, n)
	for i := 0; i < n; i++ {
		m := pool[i]
		opts := []string{m.Name}
		for _, j := range r.Perm(len(pool)) {
			if len(opts) == 4 {
				break
			}
			if pool[j].ID != m.ID {
				opts = append(opts, pool[j].Name)
			}
		}
		r.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		g := GuessRound{Monster: m, Options: opts, Chosen: -1}
		for k, o := range opts {
			if o == m.Name {
				g.Answer = k
			}
		}
		grid := portraitCells(m.ASCII)
		g.order = r.Perm(len(grid) * len(grid[0]))
		rounds[i] = g
	}
	return rounds
}

// portraitCells lays the art out as a padded grid of runes.
func portraitCells(art string) [][]rune {
	lines := strings.Split(art, "\n")
	w := 0
	for _, l := range lines {
		w = max(w, len([]rune(l)))
	}
	grid := make([][]rune, len(lines))
	for i, l := range lines {
		row := []rune(l)
		for len(row) < w {
			row = append(row, ' ')
		}
		grid[i] = row
	}
	return grid
}

// fogged draws the portrait with only a fraction of its cells revealed.
func (g GuessRound) fogged(stage int) string {
	grid := portraitCells(g.Monster.ASCII)
	if len(grid) == 0 {
		return ""
	}
	w := len(grid[0])
	total := len(grid) * w
	show := make([]bool, total)
	for _, idx := range g.order[:int(float64(total)*fogReveal[stage])] {
		show[idx] = true
	}
	fog := []rune("░▒░ ░")
	var b strings.Builder
	for y, row := range grid {
		for x, c := range row {
			idx := y*w + x
			if show[idx] {
				b.WriteRune(c)
			} else {
				b.WriteRune(fog[(idx*7+y)%len(fog)])
			}
		}
		if y < len(grid)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

type guessModel struct {
	res      GuessResult
	i        int
	stage    int
	cursor   int
	answered bool
	finished bool
	quip     string
	rng      *rand.Rand
	width    int
}

func newGuessModel(rounds []GuessRound, r *rand.Rand) guessModel {
	return guessModel{res: GuessResult{Rounds: rounds}, rng: r}
}

func (m guessModel) Init() tea.Cmd { return nil }

func (m guessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" || k == "q" || k == "esc" || m.finished {
			return m, tea.Quit
		}
		g := &m.res.Rounds[m.i]
		if m.answered {
			if k == "enter" || k == " " || k == "n" || k == "right" {
				if m.i == len(m.res.Rounds)-1 {
					m.finished, m.res.Completed = true, true
					return m, nil
				}
				m.i++
				m.stage, m.cursor, m.answered = 0, 0, false
			}
			return m, nil
		}
		switch k {
		case " ", "f":
			m.stage = min(m.stage+1, len(fogReveal)-1)
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(g.Options)) % len(g.Options)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(g.Options)
		case "enter":
			m.choose(m.cursor)
		default:
			if n := optionKey(k); n >= 0 && n < len(g.Options) {
				m.choose(n)
			}
		}
	}
	return m, nil
}

func (m *guessModel) choose(n int) {
	g := &m.res.Rounds[m.i]
	g.Chosen, g.Stage = n, m.stage
	m.answered, m.cursor = true, n
	m.quip = host.Quip(m.rng)
}

func (m guessModel) View() string {
	w := 80
	if m.width > 0 {
		w = min(m.width, 90)
	}
	var b strings.Builder
	b.WriteString(gameHeader("🔍 GUESS THE MONSTER", fmt.Sprintf("Round %d of %d · %d pts", m.i+1, len(m.res.Rounds), m.res.Score()), w) + "\n\n")

	if m.finished {
		best := len(m.res.Rounds) * fogPoints[0]
		b.WriteString(headingStyle.Render(fmt.Sprintf("Final score: %d of a possible %d points", m.res.Score(), best)) + "\n\n")
		b.WriteString(hostStyle.Render(wrap("📺 "+host.Name+": “"+finalWord(m.res.Score()*100/max(best, 1))+"”", w)) + "\n\n")
		b.WriteString(helpStyle.Render("Press any key to leave the studio") + "\n")
		return b.String()
	}

	g := m.res.Rounds[m.i]
	p := paletteFor(g.Monster)
	stage := m.stage
	if m.answered {
		stage = len(fogReveal) - 1
	}
	art := g.fogged(stage)
	artStyle := lipgloss.NewStyle().Foreground(colorDim)
	if m.answered {
		artStyle = lipgloss.NewStyle().Foreground(p.primary)
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.accent).Padding(0, 1).Render(artStyle.Render(art))

	var side strings.Builder
	side.WriteString(lipgloss.NewStyle().Bold(true).Render("Who lurks in the fog?") + "\n")
	side.WriteString(helpStyle.Render(fmt.Sprintf("Fog cleared: %d%% · worth %d pts", int(fogReveal[m.stage]*100), fogPoints[m.stage])) + "\n\n")
	for i, o := range g.Options {
		side.WriteString(optionLine(i, o, m.cursor == i, m.answered, i == g.Answer, i == g.Chosen) + "\n")
	}
	side.WriteString("\n")
	for _, c := range g.clues(m.stage) {
		side.WriteString(metaStyle.Render("Clue: "+c) + "\n")
	}

	var bottom strings.Builder
	if m.answered {
		if g.Correct() {
			bottom.WriteString(trueStyle.Render(fmt.Sprintf("✔ It's %s! +%d points", g.Monster.Name, g.Points())) + "\n")
		} else {
			bottom.WriteString(mythStyle.Render("✘ It was "+g.Monster.Name+".") + "\n")
		}
		bottom.WriteString(factStyle.Render(wrap(g.Monster.Description+". First appearance: "+g.Monster.FirstAppearance()+".", w)) + "\n")
		bottom.WriteString(hostStyle.Render("📺 "+m.quip) + "\n\n")
		bottom.WriteString(helpStyle.Render("enter next round · q quit") + "\n")
	} else {
		bottom.WriteString(helpStyle.Render(joinFit([]string{"space clear the fog", "1-4 guess", "↑/↓ + enter", "q quit"}, " · ", w)) + "\n")
	}

	sideText := side.String()
	if lipgloss.Width(box)+lipgloss.Width(sideText)+3 <= w {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, box, "   ", sideText) + "\n\n")
	} else {
		b.WriteString(box + "\n\n" + sideText + "\n")
	}
	b.WriteString(bottom.String())
	return b.String()
}

// clues grow as the fog lifts.
func (g GuessRound) clues(stage int) []string {
	var out []string
	if stage >= 2 {
		out = append(out, fmt.Sprintf("First appeared in %s", debutYear(g.Monster)))
	}
	if stage >= 3 {
		out = append(out, "Origin: "+g.Monster.Origin)
	}
	return out
}

// RunGuess plays Guess the Monster in the terminal.
func RunGuess(rounds []GuessRound, r *rand.Rand) (GuessResult, error) {
	final, err := tea.NewProgram(newGuessModel(rounds, r)).Run()
	if err != nil {
		return GuessResult{}, err
	}
	return final.(guessModel).res, nil
}
