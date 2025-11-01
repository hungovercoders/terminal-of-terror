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

- Go 1.19 or higher
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

Test the various commands:

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

To add a new monster, edit `internal/monsters/monsters.go`:

1. Add a new `Monster` struct to the `monsters` slice
2. Include the following fields:
   - `Name`: The monster's name
   - `Description`: A brief description
   - `Origin`: Where the monster originates from
   - `FirstApp`: First appearance in literature/film
   - `Facts`: A slice of interesting facts (aim for 5-6 facts)

Example:
```go
{
    Name:        "Your Monster",
    Description: "A brief description",
    Origin:      "Origin source",
    FirstApp:    "First appearance (Year)",
    Facts: []string{
        "Fact 1",
        "Fact 2",
        "Fact 3",
    },
},
```

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
