# Contributing to Terminal of Terror

Thank you for your interest in contributing to Terminal of Terror! We welcome contributions from the community.

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

Then try the various commands:

```bash
./terminal-of-terror --help
./terminal-of-terror list
./terminal-of-terror random
./terminal-of-terror monster
./terminal-of-terror monster --all
```

## Making Changes

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Write clear, descriptive commit messages
- Keep functions focused and concise

### Adding New Monsters

Monsters live in **packs** under `internal/monsters/packs/<pack-id>/`. Each pack has:

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
- Run `go test ./...`: the data tests check every monster has the required fields.

### Adding New Features

1. Consider if the feature fits the project's scope
2. Open an issue to discuss major changes before implementing
3. Write clean, well-documented code
4. Update the README.md if adding new commands or features

### Updating Documentation

- Update README.md for user-facing changes
- Update this CONTRIBUTING.md for process changes
- Ensure all commands are documented

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
