# AI Agent Optimization Guide

This document provides guidance for AI agents working on the Terminal of Terror project.

## Project Overview

Terminal of Terror is a Go CLI application that displays information about classic universal monsters. It uses:
- **Cobra** for CLI commands and structure
- **Bubbletea** for interactive terminal UI
- **Lipgloss** for terminal styling

## Project Structure

```
terminal-of-terror/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command and CLI setup
│   ├── monster.go         # Interactive monster explorer command
│   ├── list.go            # List all monsters command
│   └── random.go          # Random fact command
├── internal/              # Internal packages
│   ├── monsters/          # Monster data and logic
│   │   └── monsters.go    # Monster structs and data
│   └── ui/                # Bubbletea UI components
│       └── ui.go          # TUI implementation
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── go.sum                 # Go module checksums
```

## Key Conventions

### Code Style
- Use standard Go formatting (`gofmt`)
- Follow Go naming conventions
- Keep functions focused and single-purpose
- Use tabs for indentation (Go standard)

### Monster Data Structure
Monsters are defined in `internal/monsters/monsters.go` with:
- `Name`: Monster's name
- `Description`: Brief description
- `Origin`: Origin of the monster (folklore, literature, etc.)
- `FirstApp`: First appearance in media
- `Facts`: Slice of interesting facts (aim for 5-6 facts)

### Adding New Commands
1. Create a new file in `cmd/` directory
2. Define a cobra.Command
3. Register it with `rootCmd` in the `init()` function
4. Update README.md with command documentation

### UI Development
- Use Lipgloss for consistent styling
- Follow the color scheme established in `internal/ui/ui.go`
- Maintain keyboard navigation patterns (h/l, arrows, q)

## Building and Testing

### Build
```bash
go build -o terminal-of-terror
```

### Test Commands
```bash
./terminal-of-terror --help
./terminal-of-terror list
./terminal-of-terror random
./terminal-of-terror monster
./terminal-of-terror monster --all
```

### Dependencies
Install/update dependencies:
```bash
go mod download
go mod tidy
```

## Common Tasks

### Adding a New Monster
1. Edit `internal/monsters/monsters.go`
2. Add new `Monster` struct to the `monsters` slice
3. Include all required fields (Name, Description, Origin, FirstApp, Facts)
4. Ensure facts are accurate and interesting

### Adding a New Command
1. Create new file in `cmd/` directory (e.g., `cmd/newcmd.go`)
2. Define command struct with cobra.Command
3. Implement command logic in Run/RunE function
4. Register with rootCmd in init()
5. Update README.md documentation

### Modifying UI
1. Edit `internal/ui/ui.go`
2. Update model struct if needed
3. Modify Update() for new interactions
4. Adjust View() for display changes
5. Test interactivity thoroughly

## Documentation Standards

### README.md
- Keep installation instructions clear and up-to-date
- Document all commands with examples
- Include prerequisites
- Show expected output when helpful

### CONTRIBUTING.md
- Maintain clear contribution guidelines
- Update when processes change
- Include examples for common tasks

### Code Comments
- Comment exported functions and types
- Explain non-obvious logic
- Keep comments concise and relevant

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling

### When to Add Dependencies
- Only when necessary for core functionality
- Prefer standard library when possible
- Consider package size and maintenance status
- Update go.mod and go.sum appropriately

## Error Handling

- Use Go idiomatic error handling
- Return errors from RunE in cobra commands
- Provide helpful error messages to users
- Log errors appropriately

## Performance Considerations

- Monster data is loaded once at startup
- Keep UI updates minimal and efficient
- Avoid unnecessary allocations in tight loops
- Profile if performance issues arise

## Testing Approach

Currently, testing is manual:
1. Build the application
2. Test each command
3. Verify output format and content
4. Test interactive UI navigation
5. Ensure error cases are handled

## Git Workflow

### Commit Messages
Follow conventional commits:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation
- `refactor:` for code refactoring
- `chore:` for maintenance tasks

### Branch Naming
- `feature/description` for new features
- `fix/description` for bug fixes
- `docs/description` for documentation

## AI-Specific Tips

1. **Always build and test** after code changes
2. **Check for unused imports** - Go compiler is strict
3. **Maintain consistent styling** with existing code
4. **Update documentation** when adding features
5. **Test terminal UI changes** - they may not work as expected without testing
6. **Consider terminal width/height** when modifying UI
7. **Use go mod tidy** to clean up dependencies
8. **Add to .gitignore** any new build artifacts

## Future Considerations

Potential areas for expansion:
- Unit tests for monster data functions
- Integration tests for CLI commands
- Additional monster categories
- Filtering and search capabilities
- Configuration file support
- Additional output formats (JSON, etc.)
- Localization/internationalization

## Resources

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Effective Go](https://golang.org/doc/effective_go.html)
