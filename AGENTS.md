# AGENTS.md

Terminal of Terror is a Go CLI that teaches people about classic Universal monsters, world folklore, cryptids,
literary monsters, classical mythology and the silent screen through an interactive explorer, games and nightly
rituals, all presented by a fictional late-night horror host, Count Cathode. It uses **Cobra** (commands),
**Bubbletea** (interactive UI) and **Lipgloss** (styling). It also generates a static website and carries a
31-night build-it-yourself course.

## Layout

```
cmd/                     Cobra commands: root.go (--pack, community packs), monster, list, random, quiz, guess,
                         mash, crypt, tonight, countdown, packs, progress.go (unlocks), output.go (--json)
internal/monsters/       Monster types, pack loader, search
internal/monsters/packs/ Built-in data: <pack>/pack.json, <id>.json, <id>.txt art
internal/ui/             Bubbletea screens and Lipgloss rendering
internal/host/           Count Cathode, the horror host
internal/quiz/ mash/     Quiz question generation, Monster Mash fight simulation
internal/crypt/ store/   Captures and badges; progress saved as JSON in the config dir
internal/calendar/       Moon phases, Halloween, anniversaries (offline)
site/                    Website generator (go run ./site), writes site/dist
course/                  Count Cathode's Night School, the 31-night course (also rendered on the website)
docs/assets/ docs/demos/ README GIFs and screenshots, and the VHS tapes that record them
.github/workflows/       ci.yml on every PR, release.yml on v* tags, pages.yml deploys the site, ss-*.yml slopstopper
```

## Rules

- **Quotes must come from public-domain sources** (e.g. the 19th-century novels). Never add dialogue from
  copyrighted films; describe famous scenes in your own words instead.
- Avoid emoji that need a variation selector (⚰️, 🎞️): terminals disagree on their width and they break box
  alignment. Prefer emoji that are wide by default (🦴, 🎬, 💀).
- Never push a `v*` tag without the maintainer's go-ahead: it publishes a release. Keep versions below v2.0.0
  unless the module path gains a `/v2` suffix.
- Prefer the standard library; add a dependency only for core functionality.
- Return errors from `RunE` in cobra commands, with messages a player can act on.
- Comment exported functions and types; keep `gofmt` formatting and Go naming.
- Keep keyboard navigation consistent (h/l, arrows, q) and colours from `internal/ui/styles.go` or each
  monster's `theme`.

## Build and test

```bash
go build -o terminal-of-terror
go test ./...                       # the data tests validate every monster
task ss:hygiene:test                # slopstopper's static checks (also the pre-push hook)
```

Set `TERMINAL_OF_TERROR_HOME` to a scratch directory when testing so real progress isn't touched, and
`TERMINAL_OF_TERROR_SEED` to a number to make every random choice repeatable. CI runs gofmt, go vet, a
`go mod tidy` check, the tests (with `-race` on Linux) on Linux, macOS and Windows, and a smoke test of every
command: add new commands to the smoke test.

## Common tasks

- **Adding a monster:** add `internal/monsters/packs/<pack>/<id>.json` (and `<id>.txt` art), add the id to
  `order` in `internal/monsters/packs/<pack>/pack.json`, include every field with 5+ accurate facts, then run `go test ./...`.
- **Adding a command:** new file in `cmd/`, a `cobra.Command` registered with `rootCmd` in `init()`, logic in
  `RunE`, then document it in README.md and add it to the CI smoke test.

Before you add or edit a monster or pack, read [CONTRIBUTING.md](CONTRIBUTING.md) for every monster field with a full example.
Before you cut a release, read [CONTRIBUTING.md](CONTRIBUTING.md) "Releasing" for the git-cliff changelog and tagging steps.
Before you change the explorer or any screen in `internal/ui`, read [docs/ui.md](docs/ui.md) for the files, tests and demo re-recording.
Before you change the website in `site/`, read [docs/website.md](docs/website.md) for the templates, link tests and landing-page copy.
Before you edit a lesson in `course/` or change code a lesson quotes, read [docs/course-authoring.md](docs/course-authoring.md) for the lesson rules.
For any task not covered above, read [docs/README.md](docs/README.md) for the routing table for every doc in this repo.

## Git

Conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`), branches `feature/…`, `fix/…`, `docs/…`.
