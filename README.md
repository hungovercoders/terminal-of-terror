# Terminal of Terror 🎃

A terminal tool that terrifies you with universal monsters!

Terminal of Terror is a CLI application that brings classic universal monsters to your terminal. Explore famous monsters like Dracula, Frankenstein's Monster, The Wolf Man, and more! Learn fascinating and terrifying facts about these legendary creatures from horror history.

## Features

- 🧛 Explore 19 monsters: 12 Universal Classics and 7 from world folklore
- 📚 Learn terrifying facts about each creature, plus the real folklore and film history behind them
- 🔍 Myth vs. movie checks that separate what the legends said from what Hollywood invented
- 📺 A late-night Creature Feature hosted by Count Cathode, with a TV-static opening
- 🎨 Colours for each monster, and a black-and-white silent-film mode
- 🔎 Search across every monster's facts, films and legends
- 🧠 The Midnight Quiz, Guess the Monster and the Monster Mash
- ⚰️ Capture monsters for your crypt and earn badges
- 🎲 Get random monster facts, or a Fact of the Night for your shell's startup
- 🌕 Full-moon warnings, Friday the 13th surprises, film anniversaries and a Halloween countdown
- 🎟️ A nightly double-feature ticket
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

`list`, `random`, `monster`, `quiz` and `crypt` accept `--json` for scripting:

```bash
terminal-of-terror list --json
terminal-of-terror random --json
terminal-of-terror monster dracula --json
```

### Games

#### The Midnight Quiz

Multiple-choice questions built from the monster data: facts, film credits, quotes from the original novels, true-or-false myth checks, and more. Every answer comes with an explanation, so you learn as you play.

```bash
terminal-of-terror quiz                 # 10 questions
terminal-of-terror quiz --questions 5   # or -n 5
terminal-of-terror quiz dracula         # only questions about Dracula
```

Finish with a rank from **Ghoul-in-Training** up to **Master of Horror**.

#### Guess the Monster

A monster's portrait is hidden in fog. Guess who it is, or press `space` to clear more fog and get clues. The earlier you guess, the more points you score.

```bash
terminal-of-terror guess
terminal-of-terror guess --rounds 8
```

#### The Crypt

Answer 3 questions about a monster correctly, in quizzes or Guess the Monster, to **capture** it. Your crypt shows your captured monsters, your progress towards the rest, your badges and your records:

```bash
terminal-of-terror crypt
terminal-of-terror crypt reset   # start again
```

Badges include **Flawless Fiend** (a perfect quiz), **Eagle Eye** (a guess before the fog lifts), **Silent Era Scholar**, **Monster Kid** and **Master of the Crypt**.

Progress is saved as JSON in your config directory (e.g. `~/.config/terminal-of-terror/progress.json` on Linux). Set `TERMINAL_OF_TERROR_HOME` to keep it somewhere else.

#### Monster Mash

Pit two monsters against each other in a three-round bout decided by their stat cards and a roll of the dice:

```bash
terminal-of-terror mash                     # two random contenders
terminal-of-terror mash dracula "wolf man"
terminal-of-terror mash --fast              # skip the dramatic pauses
```

### Nightly Rituals

#### Fact of the Night

`random --daily` gives everyone the same fact all day, and a new one tomorrow. Add it to your shell's startup file for a spooky greeting in every new terminal:

```bash
terminal-of-terror random --daily
echo 'terminal-of-terror random --daily' >> ~/.bashrc   # or ~/.zshrc
terminal-of-terror random --date 2026-10-31             # peek at another night
```

The card also mentions anything special about the date: full moons, Friday the 13th, classic film anniversaries (like Frankenstein's release on 21 November 1931) and the Halloween countdown.

#### Halloween Countdown

```bash
terminal-of-terror countdown
```

Shows the nights until Halloween, tonight's moon phase and the next film anniversary. During October it's the **31 Nights of Fright**, with a different monster featured each night.

#### Tonight's Double Feature

```bash
terminal-of-terror tonight
terminal-of-terror tonight --shuffle   # a different bill
```

Prints a retro ticket for tonight's Channel 13 double bill, with showtimes, directors and stars. The line-up changes every night.

Moon phases, anniversaries and countdowns are all worked out offline. Playing games on a full moon, Friday the 13th or Halloween earns special badges.

### More

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

### Universal Classics (`--pack universal`)

- 🧛 **Dracula** - The legendary vampire count from Transylvania
- 🧟 **Frankenstein's Monster** - The tragic creature created by Dr. Victor Frankenstein
- 🐺 **The Wolf Man** - A man cursed to transform into a werewolf
- 🏺 **The Mummy** - An ancient Egyptian priest brought back to life
- 🐟 **The Creature from the Black Lagoon** - An amphibious humanoid from the Amazon
- 👻 **The Invisible Man** - A scientist who discovers the secret of invisibility
- 🎭 **The Phantom of the Opera** - A disfigured musical genius haunting the Paris Opera House
- 🔔 **The Hunchback of Notre Dame** - The deformed bell-ringer of Notre Dame Cathedral
- 👰 **The Bride of Frankenstein** - The Monster's electrifying mate
- 🦇 **Dracula's Daughter** - A melancholy countess desperate to escape her father's curse
- 🌿 **The Werewolf of London** - Hollywood's first mainstream werewolf, six years before the Wolf Man
- 👽 **The Metaluna Mutant** - A big-brained worker from a dying alien world

### World Folklore (`--pack folklore`)

- 🗿 **The Golem** - A clay giant from Jewish folklore, brought to life to protect its people
- 😱 **The Banshee** - The Irish fairy woman whose wail foretells a death
- 🧙 **Baba Yaga** - The Slavic forest witch with a hut on chicken legs
- 🥒 **The Kappa** - A polite but dangerous Japanese river imp
- 🏮 **The Jiangshi** - The Chinese hopping corpse that drains the breath of life
- 💧 **La Llorona** - The weeping woman of Mexican and Latin American legend
- 😈 **Krampus** - The horned Alpine demon who punishes naughty children

## Monster Packs

Monsters come in packs. Use `--pack` with any command to choose which ones to use:

```bash
terminal-of-terror packs                        # list every pack
terminal-of-terror quiz --pack folklore         # a folklore-only quiz
terminal-of-terror monster --pack universal,folklore
```

### Community Packs

Make your own pack of monsters, with no Go code required:

```bash
terminal-of-terror packs new cryptids
```

This creates `~/.config/terminal-of-terror/packs/cryptids/` (or the equivalent config directory on your OS) with an example monster to edit. Every pack in that folder is loaded automatically, and your monsters appear in the explorer, quiz, guessing game, Mash and everything else. Only `name`, `description` and `facts` are required. See [CONTRIBUTING.md](CONTRIBUTING.md) for every field, and consider contributing a great pack back to the project!

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

- Inspired by the classic Universal Monsters films and the late-night horror hosts who showed them
- Monster facts compiled from horror history and folklore sources
- Quotes come only from public-domain texts, such as the original 19th-century novels
- Count Cathode and Channel 13 are fictional
