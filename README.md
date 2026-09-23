# Terminal of Terror 🎃

A terminal tool that terrifies you with universal monsters!

Terminal of Terror is a CLI application that brings classic universal monsters to your terminal. Explore famous monsters like Dracula, Frankenstein's Monster, The Wolf Man, and more! Learn fascinating and terrifying facts about these legendary creatures from horror history.

## Features

- 🧛 Explore 8 classic universal monsters
- 📚 Learn terrifying facts about each creature, plus the real folklore and film history behind them
- 🔍 Myth vs. movie checks that separate what the legends said from what Hollywood invented
- 📺 A late-night Creature Feature hosted by Count Cathode, with a TV-static opening
- 🎨 Colours for each monster, and a black-and-white silent-film mode
- 🔎 Search across every monster's facts, films and legends
- 🎲 Get random monster facts
- 📖 Interactive navigation between monsters
- ⚡ Fast and lightweight CLI tool

## Installation

### Prerequisites

- Go 1.24 or higher

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

Tune in to the Channel 13 Creature Feature, hosted by **Count Cathode**:

```bash
terminal-of-terror monster
```

Each monster has its own colours and a page with sections you can flip through:

| Section | What you'll learn |
|---------|-------------------|
| **Facts** | Terrifying facts, origin and first appearance |
| **Legend** | The real folklore and history behind the monster |
| **The Film** | The classic film's director, star, makeup artist and release date |
| **Myth vs Movie** | True-or-false claims. Guess first, then press `r` to reveal the verdicts |
| **Quotes** | Lines from the original public-domain novels |
| **Stat Card** | Strength, speed, cunning and dread, plus powers and weaknesses |
| **Legacy** | Sequels, remakes and crossovers |

Silent-era monsters (the Phantom and the Hunchback) are shown in black and white with title cards and flickering film grain.

Jump straight to a monster by name. Partial names and nicknames work too:

```bash
terminal-of-terror monster dracula
terminal-of-terror monster wolfman
terminal-of-terror monster quasimodo
```

Open the gallery of every monster, or skip the opening:

```bash
terminal-of-terror monster --all      # or -a
terminal-of-terror monster --no-intro
```

**Keys:**

| Key | Action |
|-----|--------|
| `←` `→` / `h` `l` / `p` `n` | Previous / next monster |
| `tab` `shift+tab` / `1`-`7` | Switch section |
| `↑` `↓` / `j` `k`, `space`, `pgup`/`pgdn` | Scroll |
| `r` | Reveal Myth vs Movie verdicts |
| `/` | Search names, facts, films and legends |
| `g` / `esc` | Back to the gallery |
| `?` | Full help |
| `q` / `Ctrl+C` | Quit |

The explorer uses the full terminal and adapts to its size. On wide terminals the art sits beside the text.

#### List All Monsters

Display every monster, grouped by pack:

```bash
terminal-of-terror list
```

#### Random Monster Fact

Get a random fact about a random monster, with a word from your host:

```bash
terminal-of-terror random
```

#### JSON Output

`list`, `random` and `monster` accept `--json` for scripting:

```bash
terminal-of-terror list --json
terminal-of-terror random --json
terminal-of-terror monster dracula --json
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
