# Contributing to Terminal of Terror

Thank you for your interest in contributing to Terminal of Terror! We welcome contributions from the community, whether that's a new monster, a whole pack of them, a fix or a better fact.

> 📺 *"Every monster in the vault was once a stranger at the door. Come on in."* - Count Cathode

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/terminal-of-terror.git
   cd terminal-of-terror
   ```
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.24 or higher
- Git

### Building the Project

```bash
go mod download
go build -o terminal-of-terror
```

### Running the Application

```bash
./terminal-of-terror
```

### Testing Your Changes

Run the automated tests:

```bash
go test ./...
```

CI runs these on Linux, macOS and Windows for every pull request, along with `gofmt`, `go vet`, a `go mod tidy` check and a smoke test of each command. Run `gofmt -w .` and `go vet ./...` before pushing to catch the same problems locally.

Then try the various commands:

```bash
./terminal-of-terror --help
./terminal-of-terror list
./terminal-of-terror random
./terminal-of-terror monster
./terminal-of-terror monster --all
./terminal-of-terror quiz
./terminal-of-terror guess
./terminal-of-terror mash
./terminal-of-terror crypt
./terminal-of-terror tonight
./terminal-of-terror countdown
./terminal-of-terror packs
```

Set `TERMINAL_OF_TERROR_HOME` to a scratch directory while testing, so your real progress isn't touched.

## Making Changes

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Write clear, descriptive commit messages
- Keep functions focused and concise

### Adding New Monsters

Monsters live in **packs** under `internal/monsters/packs/<pack-id>/`. The built-in packs are `universal` (Universal Classics) and `folklore` (World Folklore). Each pack has:

- `pack.json` — the pack's `id`, `name`, `description` and the display `order` of monster ids
- `<monster-id>.json` — one file per monster
- `<monster-id>.txt` — the monster's ASCII art (optional but encouraged)

A monster file looks like this (see `packs/universal/dracula.json` for a full example):

```json
{
  "id": "your-monster",
  "name": "Your Monster",
  "aliases": ["Nickname"],
  "emoji": "👹",
  "description": "A one-line description",
  "origin": "Where the legend comes from",
  "debut": {"medium": "novel", "title": "Book Title", "creator": "Author", "year": 1900},
  "film": {"title": "Film Title", "year": 1931, "releaseDate": "1931-01-01", "studio": "Universal",
           "director": "Director", "star": "Actor", "makeup": "Artist", "silent": false},
  "legend": "A paragraph on the real history or folklore behind the monster.",
  "myths": [{"claim": "A popular belief.", "true": false, "explanation": "What's really going on."}],
  "quotes": [{"text": "A line from a public-domain source.", "speaker": "Who", "source": "Where"}],
  "powers": ["Power one", "Power two"],
  "weaknesses": ["Weakness one", "Weakness two"],
  "stats": {"strength": 5, "speed": 5, "cunning": 5, "dread": 5},
  "legacy": ["Sequels, remakes and crossovers"],
  "facts": ["Fact 1", "Fact 2", "Fact 3", "Fact 4", "Fact 5"],
  "hostIntro": "What the horror host says when introducing this monster.",
  "theme": {"primary": "#RRGGBB", "accent": "#RRGGBB"}
}
```

Guidelines:

- **Accuracy matters.** This is a teaching tool, so double-check facts, dates and credits. Prefer "reportedly" over a confident guess.
- **Quotes must come from public-domain sources** such as 19th-century novels. Don't quote dialogue from copyrighted films; describe famous scenes in your own words instead.
- `film` is optional (folklore monsters may not have one). `debut` needs either a `year` or an `era`.
- Stats are 1–10 and just for fun (they power the Monster Mash).
- ASCII art must use single-width, left-to-right characters (no emoji, CJK or Hebrew/Arabic letters), or it will misalign and break the fog in Guess the Monster.
- Every fact, myth check, quote and film credit automatically becomes quiz material, so write facts that make good questions.
- Treat folklore from living cultures with respect: explain where a legend comes from and what it means to the people who tell it.
- Run `go test ./...`: the data tests check every built-in monster has the required fields.

### Community Packs

You can try a pack without touching the repository:

```bash
terminal-of-terror packs new my-pack
```

This creates a pack in your config directory (`$TERMINAL_OF_TERROR_HOME/packs` if that's set). Community monsters only need `name`, `description` and `facts`; everything else gets a sensible default. Broken files are skipped with a warning. When your pack is polished, it can be moved into `internal/monsters/packs/` and contributed as a built-in pack. Built-in monsters need every field.

### Adding New Features

1. Consider if the feature fits the project's scope
2. Open an issue to discuss major changes before implementing
3. Write clean, well-documented code
4. Update the README.md if adding new commands or features

### Updating Documentation

- Update README.md for user-facing changes
- Update this CONTRIBUTING.md for process changes
- Ensure all commands are documented, including in the README's command reference
- Re-record the demos when a change shows up in them (see below)

### Recording the Demos

The README's GIFs and screenshots are recorded from scripts in [`docs/demos/`](docs/demos): one [VHS](https://github.com/charmbracelet/vhs) `.tape` file per demo, plus `theme.json`, the colour theme `render.sh` fills into every tape. To re-record them:

```bash
docs/demos/render.sh              # everything
docs/demos/render.sh hero quiz    # just some
```

You'll need `go`, `vhs`, `ffmpeg` and **ttyd 1.7.7 or newer**. Older ttyd builds draw emoji one cell wide, which knocks every box border out of line; `TTYD=/path/to/ttyd` picks a specific binary. The script builds the app, uses a scratch home directory with sample crypt progress, and writes the results to `docs/assets/`. Look at the results before committing, and keep the hero GIF under about 2 MB so the README loads quickly.

Recordings are repeatable: `render.sh` sets `TERMINAL_OF_TERROR_SEED` (13 unless you set `DEMO_SEED`), which fixes every random choice (quiz questions, portraits, fights and the host's lines), so the keys a tape presses always get the same result. If you change the seed or the monster data, check that each tape still shows what it should: the guess tape, for example, expects the right answer to be option 1. `TERMINAL_OF_TERROR_SEED` is also handy for reproducing a bug report.

To add a demo, copy a similar tape, change its `Output` line and commands, and add its name to the `stills` list in `render.sh` if it should be a PNG rather than a GIF. Keep `Set Theme @theme.json` and `Set Framerate` in the tape: the script fills in the theme and reads the frame rate back when encoding.

## Commit Guidelines

### Commit Message Format

```
type: brief description

Detailed explanation of changes (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat: add new vampire monster with facts
fix: correct navigation in monster TUI
docs: update installation instructions
```

## Submitting Changes

1. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: your descriptive message"
   ```

2. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Open a Pull Request** on GitHub:
   - Provide a clear title and description
   - Reference any related issues
   - Explain what changes you made and why

## Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Ensure your code builds successfully
- Test your changes thoroughly
- Update documentation as needed
- Be responsive to feedback and review comments

## Code Review Process

1. A maintainer will review your PR
2. Address any requested changes
3. Once approved, a maintainer will merge your PR

## Reporting Issues

Found a bug or have a suggestion?

1. Check if the issue already exists
2. Open a new issue with:
   - Clear title and description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Your environment (OS, Go version)

## Questions?

Feel free to open an issue for questions or reach out to the maintainers.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Respect differing opinions and experiences

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to Terminal of Terror! 🎃
