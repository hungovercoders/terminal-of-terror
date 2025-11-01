# Terminal of Terror 🎃

A terminal tool that terrifies you with universal monsters!

Terminal of Terror is a CLI application that brings classic universal monsters to your terminal. Explore famous monsters like Dracula, Frankenstein's Monster, The Wolf Man, and more! Learn fascinating and terrifying facts about these legendary creatures from horror history.

## Features

- 🧛 Explore 8 classic universal monsters
- 📚 Learn terrifying facts about each creature
- 🎨 Beautiful terminal UI powered by Bubbletea
- 🎲 Get random monster facts
- 📖 Interactive navigation between monsters
- ⚡ Fast and lightweight CLI tool

## Installation

### Prerequisites

- Go 1.19 or higher

### From Source

```bash
git clone https://github.com/hungovercoders/terminal-of-terror.git
cd terminal-of-terror
go build -o terminal-of-terror
```

### Install Globally

```bash
go install github.com/hungovercoders/terminal-of-terror@latest
```

## Usage

### Commands

#### Interactive Monster Explorer

Explore monsters interactively with a beautiful TUI:

```bash
terminal-of-terror monster
```

Show all monsters at once:

```bash
terminal-of-terror monster --all
# or
terminal-of-terror monster -a
```

**Navigation:**
- Use arrow keys (← →) or `h`/`l` or `p`/`n` to navigate between monsters
- Press `q` or `Ctrl+C` to quit

#### List All Monsters

Display a simple list of all available monsters:

```bash
terminal-of-terror list
```

#### Random Monster Fact

Get a random fact about a random monster:

```bash
terminal-of-terror random
```

#### Help

View help information:

```bash
terminal-of-terror --help
terminal-of-terror [command] --help
```

#### Version

Check the version:

```bash
terminal-of-terror --version
```

## Monsters Included

- 🧛 **Dracula** - The legendary vampire count from Transylvania
- 🧟 **Frankenstein's Monster** - The tragic creature created by Dr. Victor Frankenstein
- 🐺 **The Wolf Man** - A man cursed to transform into a werewolf
- 🏺 **The Mummy** - An ancient Egyptian priest brought back to life
- 🐟 **The Creature from the Black Lagoon** - An amphibious humanoid from the Amazon
- 👻 **The Invisible Man** - A scientist who discovers the secret of invisibility
- 🎭 **The Phantom of the Opera** - A disfigured musical genius haunting the Paris Opera House
- 🔔 **The Hunchback of Notre Dame** - The deformed bell-ringer of Notre Dame Cathedral

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for details about changes in each version.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Built With

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UIs

## Acknowledgments

- Inspired by the classic Universal Monsters films
- Monster facts compiled from various horror history sources
