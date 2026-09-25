# 📺 Count Cathode's Night School

### Build Terminal of Terror from nothing, in 31 nights

> *"Good evening, students of the strange. You've watched the Creature Feature. Tonight, you'll learn how the picture gets made. Thirty-one nights, one lesson each, and by Halloween you'll have built the whole thing yourself. Pencils sharp, garlic close."*

This is a course in writing a real Go command-line application, taught by building **Terminal of Terror** from an empty folder to a released program. Every night adds one idea and leaves you with code that runs. Along the way you'll meet the standard library, [Cobra](https://github.com/spf13/cobra) for commands, [Lip Gloss](https://github.com/charmbracelet/lipgloss) for style and [Bubble Tea](https://github.com/charmbracelet/bubbletea) for interactive screens, and you'll learn how to test, ship and release the lot.

**Who it's for:** anyone who can open a terminal and has written a little code in any language. No Go experience is needed; Go is taught as we go. If you already know Go, skip to Night 11, where the interactive terminal starts.

**How it works:** each night takes 30 to 60 minutes. Nights 1 to 14 are *build-along*: type the code, run it, see it work. From Night 15 the app gets big, so those nights teach the *ideas* with the key code, and point at the finished files in this repository for the rest. Every night ends with an exercise ("Before dawn") and a Go fact ("Terrifying fact").

Use the same module name as this repository, `github.com/hungovercoders/terminal-of-terror`, so the code in the lessons matches yours line for line. You can rename it on the last night.

## 📅 The programme

### Week 1 · The Vault — foundations

| Night | Lesson | You'll learn |
|-------|--------|--------------|
| [1](night-01.md) | It's Alive! | Installing Go, modules, `main`, `go run` and `go build` |
| [2](night-02.md) | The Skeleton Crew | Cobra: a root command, `--help`, `--version` |
| [3](night-03.md) | Meet the Monsters | Structs, slices, loops, and a `list` command |
| [4](night-04.md) | A Random Terror | `math/rand`, seeds, functions that return two values |
| [5](night-05.md) | Data in the Coffin | JSON, struct tags, `//go:embed` |
| [6](night-06.md) | Packs of Monsters | One file per monster, `embed.FS`, maps and sorting |
| [7](night-07.md) | Check Under the Bed | `go test`, table tests, testing with a fake filesystem |

### Week 2 · The Séance — finding and styling

| Night | Lesson | You'll learn |
|-------|--------|--------------|
| [8](night-08.md) | Speak Its Name | Fuzzy lookup, aliases, command arguments, helpful errors |
| [9](night-09.md) | A Splash of Blood | Lip Gloss styles, colours and borders; terminal width |
| [10](night-10.md) | Every Monster in Its Colours | Themes per monster, a reusable fact card |
| [11](night-11.md) | The Séance Begins | Bubble Tea: Model, Update, View; your first explorer |
| [12](night-12.md) | Fitting the Screen | Window size, wrapping, side-by-side layout, scrolling |
| [13](night-13.md) | Friday the 13th: Turn the Page | Tabs and sections, number keys, header and footer |
| [14](night-14.md) | Whispers in the Dark | Search mode: text input, matching, snippets |

### Week 3 · The Double Feature — atmosphere and games

| Night | Lesson | You'll learn |
|-------|--------|--------------|
| [15](night-15.md) | Your Host, Count Cathode | A host package, timers with `tea.Tick`, TV static, typewriter text |
| [16](night-16.md) | Silent Picture | Palettes, title cards, film grain, animating only when needed |
| [17](night-17.md) | The Gallery | A list view with a cursor and preview, keeping the selection visible |
| [18](night-18.md) | Writing the Questions | Generating quiz questions from data, hiding the answer, distractors |
| [19](night-19.md) | The Midnight Quiz | A second Bubble Tea program, feedback, ranks, wiring a command |
| [20](night-20.md) | Who Lurks in the Fog | Grids of runes, random reveal order, stages and points |
| [21](night-21.md) | What the Crypt Remembers | Saving progress: config dirs, JSON, atomic writes, `t.Setenv` |

### Week 4 · The Finale — rituals, packs and release

| Night | Lesson | You'll learn |
|-------|--------|--------------|
| [22](night-22.md) | Captures and Badges | Outcomes, unlock rules, small predicate helpers |
| [23](night-23.md) | Monster Mash | Dice, best of three, narration templates |
| [24](night-24.md) | By the Light of the Moon | Moon-phase maths, countdowns, pure functions of time |
| [25](night-25.md) | Nightly Rituals | Daily seeds, tickets, block-letter digits |
| [26](night-26.md) | Speaking to Machines | `--json`, persistent flags, reproducible runs |
| [27](night-27.md) | Open the Doors | Community packs from disk, validation, defaults, scaffolding |
| [28](night-28.md) | Haunting the Tests | Testing Bubble Tea models, size sweeps, driving a real terminal |
| [29](night-29.md) | The Night Watch | CI with GitHub Actions on three operating systems |
| [30](night-30.md) | Lights, Camera | Recording demo GIFs, the README |
| [31](night-31.md) | Halloween: Sign Off | Versions from build info, GoReleaser, tagging a release |

## Before Night 1

You'll need:

- A terminal. On macOS or Linux, the one you have. On Windows, [Windows Terminal](https://aka.ms/terminal) with PowerShell works well.
- A text editor. [VS Code](https://code.visualstudio.com) with the Go extension is a fine choice.
- [Go 1.24 or newer](https://go.dev/dl/). Night 1 walks you through installing it.
- [Git](https://git-scm.com), so you can save your progress and look at the finished code.

The finished program is in this repository. Whenever a lesson says "the full file in the repo", it means the file at the same path here, for example [`cmd/list.go`](../../cmd/list.go). Reading ahead is allowed. Copying is encouraged. Understanding is the point.

📺 *"Channel 13 will now begin its broadcast. Night 1 is [this way](night-01.md)."*
