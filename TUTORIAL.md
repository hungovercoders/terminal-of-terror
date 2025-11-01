# Building Terminal of Terror: A Complete Tutorial

Learn how to build a terminal-based CLI application with an interactive UI from scratch using Go, Cobra, Bubbletea, and Lipgloss.

> **Note:** This tutorial walks you through building the Terminal of Terror application from scratch. You can find the complete, working code in the [hungovercoders/terminal-of-terror](https://github.com/hungovercoders/terminal-of-terror) repository. Feel free to reference the actual code as you work through this tutorial!

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Project Overview](#project-overview)
3. [Setting Up Your Go Project](#setting-up-your-go-project)
4. [Understanding the Core Technologies](#understanding-the-core-technologies)
5. [Building the Foundation](#building-the-foundation)
6. [Creating the Data Layer](#creating-the-data-layer)
7. [Building CLI Commands with Cobra](#building-cli-commands-with-cobra)
8. [Creating an Interactive UI with Bubbletea](#creating-an-interactive-ui-with-bubbletea)
9. [Styling with Lipgloss](#styling-with-lipgloss)
10. [Adding Additional Commands](#adding-additional-commands)
11. [Building and Testing](#building-and-testing)
12. [Next Steps and Enhancements](#next-steps-and-enhancements)

---

## Prerequisites

Before starting this tutorial, you should have:

- **Go 1.19 or higher** installed ([Download Go](https://golang.org/dl/))
- Basic understanding of Go programming language
- Familiarity with command-line interfaces
- A text editor or IDE (VS Code, GoLand, etc.)
- Terminal/command prompt access

Verify your Go installation:
```bash
go version
```

## Project Overview

**Terminal of Terror** is a CLI application that displays information about classic universal monsters. The project demonstrates:

- **CLI Framework**: Using Cobra for command structure
- **TUI (Terminal User Interface)**: Building interactive UIs with Bubbletea
- **Terminal Styling**: Beautiful terminal output with Lipgloss
- **Data Management**: Working with embedded JSON files
- **Go Best Practices**: Clean code organization and package structure

### Key Features
- Interactive monster explorer with keyboard navigation
- List command to display all monsters
- Random fact generator
- Styled terminal output with colors and formatting
- ASCII art for visual appeal

## Setting Up Your Go Project

### Step 1: Initialize Your Project

Create a new directory and initialize a Go module:

```bash
mkdir terminal-of-terror
cd terminal-of-terror
go mod init github.com/yourusername/terminal-of-terror
```

**Note:** Replace `yourusername` with your actual GitHub username throughout this tutorial. For example, if you're studying the original repository, it would be `github.com/hungovercoders/terminal-of-terror`.

This creates a `go.mod` file that manages your project's dependencies.

### Step 2: Project Structure

Create the following directory structure:

```bash
mkdir -p cmd internal/monsters internal/ui
touch main.go
```

Your structure should look like:
```
terminal-of-terror/
├── cmd/                    # CLI command definitions
├── internal/              # Internal packages
│   ├── monsters/          # Monster data and logic
│   └── ui/                # TUI components
├── main.go                # Application entry point
└── go.mod                 # Go module file
```

**Why this structure?**
- `cmd/`: Organizes CLI commands separately
- `internal/`: Packages that are internal to this project (not importable by others)
- `main.go`: Entry point keeps the main package minimal

## Understanding the Core Technologies

### Cobra - CLI Framework

**Cobra** is a powerful library for creating modern CLI applications in Go.

**Key Concepts:**
- **Commands**: Represent actions (e.g., `list`, `random`, `monster`)
- **Flags**: Modify command behavior (e.g., `--all`, `--help`)
- **Subcommands**: Commands can have nested subcommands

**Why Cobra?**
- Automatic help generation
- Flag parsing
- POSIX-compliant flags
- Used by popular tools (kubectl, Hugo, GitHub CLI)

Install Cobra:
```bash
go get -u github.com/spf13/cobra@latest
```

### Bubbletea - TUI Framework

**Bubbletea** is a framework for building terminal user interfaces based on The Elm Architecture.

**The Elm Architecture:**
1. **Model**: Application state
2. **Update**: Handle events and update state
3. **View**: Render the UI based on state

**Why Bubbletea?**
- Declarative UI programming
- Clean separation of concerns
- Event-driven architecture
- Great for interactive applications

Install Bubbletea:
```bash
go get github.com/charmbracelet/bubbletea@latest
```

### Lipgloss - Terminal Styling

**Lipgloss** provides style definitions for terminal output with a simple, chainable API.

**Features:**
- Colors and gradients
- Borders and padding
- Alignment and margins
- Composable styles

Install Lipgloss:
```bash
go get github.com/charmbracelet/lipgloss@latest
```

After installing all dependencies, run:
```bash
go mod tidy
```

## Building the Foundation

### Step 1: Create the Main Entry Point

Create `main.go`:

```go
package main

import "github.com/yourusername/terminal-of-terror/cmd"

func main() {
	cmd.Execute()
}
```

**Why so simple?**
- Keeps main package minimal
- Delegates execution to cmd package
- Makes testing easier

### Step 2: Create the Root Command

Create `cmd/root.go`:

```go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terminal-of-terror",
	Short: "A terminal tool that terrifies you with universal monsters!",
	Long: `Terminal of Terror is a CLI application that brings classic 
universal monsters to your terminal. Explore famous monsters like 
Dracula, Frankenstein's Monster, The Wolf Man, and more!

Learn fascinating and terrifying facts about these legendary creatures 
from horror history.`,
	Version: "1.0.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
```

**Key Points:**
- `rootCmd`: The base command for your CLI
- `Use`: The command name users type
- `Short` and `Long`: Help text descriptions
- `Version`: Version information
- `Execute()`: Entry point for command execution
- `init()`: Runs automatically at package initialization

## Creating the Data Layer

### Step 1: Define the Monster Data Structure

> **Note:** The code examples in this tutorial are simplified for teaching purposes. In production code, you should add additional error handling and validation (e.g., checking for empty slices before using `rand.Intn()`).

Create `internal/monsters/monsters.go`:

```go
package monsters

import (
	_ "embed"
	"encoding/json"
	"log"
	"math/rand"
	"strings"
)

// Monster represents a universal monster with facts
type Monster struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Facts       []string `json:"facts"`
	Origin      string   `json:"origin"`
	FirstApp    string   `json:"firstApp"`
	ASCII       string   `json:"ascii"`
}

//go:embed monsters.json
var monstersJSON []byte

var monsters []Monster

func init() {
	if err := json.Unmarshal(monstersJSON, &monsters); err != nil {
		log.Fatalf("Failed to load monsters data: %v", err)
	}
}

// GetAllMonsters returns all available monsters
func GetAllMonsters() []Monster {
	return monsters
}

// GetRandomMonster returns a random monster
func GetRandomMonster() Monster {
	return monsters[rand.Intn(len(monsters))]
}

// GetMonsterByName returns a monster by name (case-insensitive)
func GetMonsterByName(name string) *Monster {
	for _, monster := range monsters {
		if strings.EqualFold(monster.Name, name) {
			return &monster
		}
	}
	return nil
}

// GetRandomFact returns a random fact from a random monster
func GetRandomFact() (string, string) {
	monster := GetRandomMonster()
	fact := monster.Facts[rand.Intn(len(monster.Facts))]
	return monster.Name, fact
}
```

**Key Concepts:**

1. **Struct Tags**: `json:"name"` tells Go how to map JSON fields
2. **Embed Directive**: `//go:embed monsters.json` embeds the JSON file into the binary
3. **Package-level Variables**: `monsters` is accessible throughout the package
4. **init() Function**: Loads data automatically when package is imported
5. **Exported Functions**: Start with capital letters (GetAllMonsters, etc.)

### Step 2: Create the Monster Data File

Create `internal/monsters/monsters.json`:

```json
[
  {
    "name": "Dracula",
    "description": "The legendary vampire count from Transylvania",
    "origin": "Romanian folklore",
    "firstApp": "Bram Stoker's Dracula (1897)",
    "ascii": "          xxxxxxxxxxxxxxxxxxxxx             \n         xxxxxxxxxxxxxxxxxxxxxxx            \n        xxxxxxxxxxxxxxxxxxxxxxxx            \n       xxxx    xxxxxxx        xxxx          \n      xx          xxx             xxxx      \nxxx  xx      xxx   x     xx          xx  xxx\nx xxxx         xxx     xx             xxxx x\n x  xx        x  xx  xxx  x                x\n x   x  x                        xx        x\n x      xxx         x           xx         x\n x        xxx       x          xx          x\n x xx      xxxx            xxxxx       xxx x\n xxxx       xxxxx        xxxx          x xxx\n  xxxxx      x  xxxxxxxxxx  x         xx   x\n      xx     xx xx       x  x       xxx     \n       xxx    x x        xxxx     xxx       \n         xx   xxx         xx     xx         \n          xxx  xx          x  xxx           \n             xxx            xxx             \n                xxx        xx               \n                  xxxxx xxx                 \n                      xxx                   ",
    "facts": [
      "Dracula can transform into a bat, wolf, or mist",
      "He is repelled by garlic, crosses, and holy water",
      "Dracula must rest in his coffin with Transylvanian soil",
      "He casts no reflection and has no shadow",
      "The character was inspired by Vlad the Impaler",
      "Dracula has superhuman strength and can control weather"
    ]
  },
  {
    "name": "Frankenstein's Monster",
    "description": "The tragic creature created by Dr. Victor Frankenstein",
    "origin": "Literary fiction",
    "firstApp": "Frankenstein by Mary Shelley (1818)",
    "ascii": "    ___________\n   |  _     _  |\n   | |_|   |_| |\n   |     ^     |\n   |  \\-----/  |\n   |___________|  \n   |           |\n   |   [___]   |\n   |           |\n  /|           |\\\n / |___________|_\\\n|               |\n|  |||     |||  |\n|  |||     |||  |\n|___||_____||___|\n    ||     ||    \n   _||_   _||_   \n  |___|  |___|   ",
    "facts": [
      "The monster is never given a name in the original novel",
      "He was created from assembled body parts",
      "The monster is highly intelligent and teaches himself to read",
      "He seeks love and acceptance but is rejected by society",
      "Mary Shelley wrote Frankenstein when she was only 18 years old",
      "The monster's first request was for a female companion"
    ]
  }
]
```

**Note**: Add more monsters as needed! The structure is the same for each monster.

**Why Embed JSON?**
- Single binary distribution (no external files needed)
- Simpler deployment
- Data is available at compile time

## Building CLI Commands with Cobra

### Step 1: Create the List Command

Create `cmd/list.go`:

```go
package cmd

import (
	"fmt"

	"github.com/yourusername/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available monsters",
	Long:  `Display a list of all universal monsters available in the Terminal of Terror.`,
	Run: func(cmd *cobra.Command, args []string) {
		allMonsters := monsters.GetAllMonsters()
		fmt.Println("Universal Monsters:")
		fmt.Println("==================")
		for i, monster := range allMonsters {
			fmt.Printf("%d. %s - %s\n", i+1, monster.Name, monster.Description)
		}
		fmt.Printf("\nTotal: %d monsters\n", len(allMonsters))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
```

**Key Points:**
- `Run`: Function executed when command is called
- `cmd` and `args`: Access to command context and arguments
- `init()`: Registers command with root command
- Simple, straightforward output with fmt

### Step 2: Create the Random Command

Create `cmd/random.go`:

```go
package cmd

import (
	"fmt"

	"github.com/yourusername/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Get a random monster fact",
	Long:  `Display a random fact about a random universal monster.`,
	Run: func(cmd *cobra.Command, args []string) {
		monsterName, fact := monsters.GetRandomFact()
		fmt.Printf("🎃 Random Terror Fact 🎃\n\n")
		fmt.Printf("Monster: %s\n", monsterName)
		fmt.Printf("Fact: %s\n", fact)
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
}
```

**Test Your Commands:**
```bash
go run main.go list
go run main.go random
go run main.go --help
```

## Creating an Interactive UI with Bubbletea

Now for the exciting part - building an interactive terminal UI!

### Understanding the Bubbletea Model

The Elm Architecture in Bubbletea:

```
┌─────────────┐
│    Model    │  ←  Application State
└─────────────┘
       ↓
┌─────────────┐
│   Update    │  ←  Handle Events, Update State
└─────────────┘
       ↓
┌─────────────┐
│    View     │  ←  Render UI from State
└─────────────┘
```

### Step 1: Create the UI Model

Create `internal/ui/ui.go`:

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/terminal-of-terror/internal/monsters"
)

// Model holds the application state
type model struct {
	monsters        []monsters.Monster
	currentIndex    int
	quitting        bool
	width           int
	height          int
	showAllMonsters bool
}

// InitialModel creates the initial state
func InitialModel(showAll bool) model {
	allMonsters := monsters.GetAllMonsters()
	return model{
		monsters:        allMonsters,
		currentIndex:    0,
		showAllMonsters: showAll,
	}
}

// Init is called when the program starts
func (m model) Init() tea.Cmd {
	return nil
}
```

**State Management:**
- `monsters`: All monster data
- `currentIndex`: Which monster is currently displayed
- `quitting`: Whether the user wants to exit
- `width/height`: Terminal dimensions
- `showAllMonsters`: Display mode flag

### Step 2: Implement the Update Function

Add to `internal/ui/ui.go`:

```go
// Update handles events and updates the state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard input
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
		// Handle terminal resize
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}
```

**Event Handling:**
- **tea.KeyMsg**: Keyboard input events
- **tea.WindowSizeMsg**: Terminal resize events
- **Circular Navigation**: Using modulo for wraparound
- **Return Values**: Updated model and optional command to execute

### Step 3: Implement the View Function

Add to `internal/ui/ui.go`:

```go
// View renders the UI
func (m model) View() string {
	if m.quitting {
		return "Thanks for exploring the monsters! 👻\n"
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render("🎃 TERMINAL OF TERROR 🎃")
	b.WriteString(title + "\n\n")

	if m.showAllMonsters {
		// Show all monsters in a list
		for i, monster := range m.monsters {
			if i == m.currentIndex {
				b.WriteString("▶ ")
			} else {
				b.WriteString("  ")
			}
			b.WriteString(monsterNameStyle.Render(monster.Name) + "\n")
			b.WriteString("  " + descriptionStyle.Render(monster.Description) + "\n")
			b.WriteString("  " + metaStyle.Render(fmt.Sprintf("Origin: %s | First Appearance: %s", 
				monster.Origin, monster.FirstApp)) + "\n\n")
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
		b.WriteString(lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color("#FFA500")).Render("Terrifying Facts:") + "\n")
		for _, fact := range monster.Facts {
			b.WriteString(factStyle.Render("• "+fact) + "\n")
		}

		// Meta information
		b.WriteString("\n")
		b.WriteString(metaStyle.Render(fmt.Sprintf("Origin: %s", monster.Origin)) + "\n")
		b.WriteString(metaStyle.Render(fmt.Sprintf("First Appearance: %s", monster.FirstApp)) + "\n")

		// Navigation info
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(fmt.Sprintf("Monster %d of %d", 
			m.currentIndex+1, len(m.monsters))) + "\n")
	}

	// Help text
	b.WriteString(helpStyle.Render("Navigation: ← → or h l or p n | Quit: q or Ctrl+C") + "\n")

	return b.String()
}
```

**Rendering Concepts:**
- **strings.Builder**: Efficient string concatenation
- **Conditional Rendering**: Different views based on state
- **Styles**: Applied to different UI elements
- **String Return**: Complete UI as a single string

### Step 4: Add the RunUI Function

Add to `internal/ui/ui.go`:

```go
// RunUI starts the Bubbletea program
func RunUI(showAll bool) error {
	p := tea.NewProgram(InitialModel(showAll))
	_, err := p.Run()
	return err
}
```

**Program Lifecycle:**
1. Create program with initial model
2. Run starts the event loop
3. Update handles events
4. View renders UI
5. Repeat until tea.Quit is returned

## Styling with Lipgloss

### Understanding Lipgloss Styles

Add style definitions at the top of `internal/ui/ui.go`:

```go
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
```

**Style Properties:**
- **Foreground/Background**: Colors (hex or ANSI)
- **Bold/Italic**: Text formatting
- **Padding**: Space inside element
- **Margin**: Space outside element
- **Border**: Various border styles

**Color Reference:**
- `#FF0000`: Red
- `#FFD700`: Gold
- `#FF6347`: Tomato
- `#87CEEB`: Sky Blue
- `#90EE90`: Light Green
- `#808080`: Gray

### Applying Styles

Styles are applied using the `Render()` method:

```go
styledText := myStyle.Render("Hello, World!")
```

Styles can be chained:

```go
style := lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FF0000")).
	Padding(1)
```

## Adding Additional Commands

### Create the Monster Command

Create `cmd/monster.go`:

```go
package cmd

import (
	"github.com/yourusername/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var showAll bool

var monsterCmd = &cobra.Command{
	Use:   "monster",
	Short: "Explore universal monsters interactively",
	Long: `Display universal monsters with their descriptions, facts, and origins.
Use arrow keys or h/l to navigate between monsters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.RunUI(showAll)
	},
}

func init() {
	rootCmd.AddCommand(monsterCmd)
	monsterCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all monsters at once")
}
```

**Key Features:**
- **RunE**: Run function that can return errors
- **Flags**: `--all` or `-a` to change display mode
- **BoolVarP**: Boolean flag with long and short forms

## Building and Testing

### Build the Application

```bash
go build -o terminal-of-terror
```

This creates an executable binary named `terminal-of-terror`.

### Test All Commands

```bash
# Show help
./terminal-of-terror --help

# List all monsters
./terminal-of-terror list

# Get a random fact
./terminal-of-terror random

# Interactive explorer
./terminal-of-terror monster

# Show all monsters
./terminal-of-terror monster --all

# Check version
./terminal-of-terror --version
```

### Install Globally

```bash
go install
```

This installs the binary to `$GOPATH/bin`, making it available system-wide.

### Development Workflow

For rapid development:

```bash
# Run without building
go run main.go monster

# Build with verbose output
go build -v -o terminal-of-terror

# Run tests (if you add them)
go test ./...
```

## Next Steps and Enhancements

### Ideas for Expansion

1. **Add More Monsters**
   - Research and add more classic monsters
   - Include lesser-known creatures
   - Add custom ASCII art

2. **Additional Features**
   - Search functionality
   - Filter by origin or time period
   - Save favorite monsters
   - Export to different formats (JSON, markdown)

3. **Testing**
   - Add unit tests for monster functions
   - Test UI components
   - Integration tests for commands

4. **Configuration**
   - Config file support
   - Theme customization
   - Keybinding customization

5. **Advanced UI Features**
   - Mouse support
   - Pagination for long lists
   - Search mode
   - Animation effects

6. **Distribution**
   - Create releases on GitHub
   - Publish to package managers (Homebrew, etc.)
   - Docker container
   - Cross-platform builds

### Learning Resources

**Go Programming:**
- [Go Official Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go by Example](https://gobyexample.com/)

**Cobra:**
- [Cobra GitHub](https://github.com/spf13/cobra)
- [Cobra Documentation](https://cobra.dev/)

**Bubbletea:**
- [Bubbletea GitHub](https://github.com/charmbracelet/bubbletea)
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Bubbletea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)

**Lipgloss:**
- [Lipgloss GitHub](https://github.com/charmbracelet/lipgloss)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)

**Other Charm Tools:**
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Gum](https://github.com/charmbracelet/gum) - Shell scripts made easy

### Best Practices

1. **Code Organization**
   - Keep packages focused and cohesive
   - Use internal packages for implementation details
   - Separate concerns (data, UI, commands)

2. **Error Handling**
   - Always check and handle errors
   - Provide meaningful error messages
   - Use `RunE` instead of `Run` for commands that can fail
   - Add validation for edge cases (empty slices, nil pointers, etc.)
   - In production code, check for empty data before operations like `rand.Intn(len(slice))`

3. **Documentation**
   - Comment exported functions and types
   - Keep README updated
   - Use godoc conventions

4. **Dependencies**
   - Keep dependencies minimal
   - Use `go mod tidy` to clean up
   - Version pin critical dependencies

5. **User Experience**
   - Provide helpful error messages
   - Include comprehensive help text
   - Support common navigation patterns
   - Test on different terminal emulators

### Troubleshooting Common Issues

**Build Errors:**
```bash
# Update dependencies
go mod download
go mod tidy

# Clear cache
go clean -modcache
```

**Terminal Display Issues:**
- Some terminals may not support all colors
- ASCII art may not display correctly with certain fonts
- Test on multiple terminal emulators

**Performance:**
- Embed directive increases binary size
- Consider external files for very large datasets
- Profile with `go tool pprof` if needed

## Conclusion

You've now learned how to build a complete CLI application with an interactive terminal UI! The key concepts covered include:

- Setting up a Go project with proper structure
- Using Cobra for command-line interfaces
- Creating interactive UIs with Bubbletea
- Styling terminal output with Lipgloss
- Working with embedded files in Go
- Managing application state
- Event-driven programming

This foundation can be applied to many different types of CLI applications. Experiment, extend, and create your own terminal applications!

Happy coding! 🎃👻
