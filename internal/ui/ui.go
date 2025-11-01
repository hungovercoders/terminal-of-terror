package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#000000")).
			Padding(1, 2).
			MarginBottom(1)

	monsterNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700")).
				MarginTop(1)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6347")).
				Italic(true).
				MarginBottom(1)

	factStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#87CEEB")).
			MarginLeft(2)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#90EE90")).
			Italic(true).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")).
			MarginTop(2)
)

type model struct {
	monsters        []monsters.Monster
	currentIndex    int
	quitting        bool
	width           int
	height          int
	showAllMonsters bool
}

func InitialModel(showAll bool) model {
	allMonsters := monsters.GetAllMonsters()
	return model{
		monsters:        allMonsters,
		currentIndex:    0,
		showAllMonsters: showAll,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "right", "l", "n":
			m.currentIndex = (m.currentIndex + 1) % len(m.monsters)
		case "left", "h", "p":
			m.currentIndex = (m.currentIndex - 1 + len(m.monsters)) % len(m.monsters)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for exploring the monsters! 👻\n"
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render("🎃 TERMINAL OF TERROR 🎃")
	b.WriteString(title + "\n\n")

	if m.showAllMonsters {
		// Show all monsters
		for i, monster := range m.monsters {
			if i == m.currentIndex {
				b.WriteString("▶ ")
			} else {
				b.WriteString("  ")
			}
			b.WriteString(monsterNameStyle.Render(monster.Name) + "\n")
			b.WriteString("  " + descriptionStyle.Render(monster.Description) + "\n")
			b.WriteString("  " + metaStyle.Render(fmt.Sprintf("Origin: %s | First Appearance: %s", monster.Origin, monster.FirstApp)) + "\n\n")
		}
	} else {
		// Show current monster with details
		monster := m.monsters[m.currentIndex]

		// Monster name
		b.WriteString(monsterNameStyle.Render(monster.Name) + "\n")

		// Description
		b.WriteString(descriptionStyle.Render(monster.Description) + "\n\n")

		// ASCII Art
		if monster.ASCII != "" {
			asciiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF69B4")).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("#666666")).
				Padding(1, 2)
			b.WriteString(asciiStyle.Render(monster.ASCII) + "\n\n")
		}

		// Facts
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500")).Render("Terrifying Facts:") + "\n")
		for _, fact := range monster.Facts {
			b.WriteString(factStyle.Render("• "+fact) + "\n")
		}

		// Meta information
		b.WriteString("\n")
		b.WriteString(metaStyle.Render(fmt.Sprintf("Origin: %s", monster.Origin)) + "\n")
		b.WriteString(metaStyle.Render(fmt.Sprintf("First Appearance: %s", monster.FirstApp)) + "\n")

		// Navigation info
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(fmt.Sprintf("Monster %d of %d", m.currentIndex+1, len(m.monsters))) + "\n")
	}

	// Help text
	b.WriteString(helpStyle.Render("Navigation: ← → or h l or p n | Quit: q or Ctrl+C") + "\n")

	return b.String()
}

func RunUI(showAll bool) error {
	p := tea.NewProgram(InitialModel(showAll))
	_, err := p.Run()
	return err
}
