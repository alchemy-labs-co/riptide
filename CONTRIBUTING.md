# Contributing to Riptide

Thank you for your interest in contributing to Riptide! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

- Be welcoming and inclusive
- Be respectful of differing viewpoints
- Accept constructive criticism gracefully
- Focus on what's best for the community

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/alchemy-labs-co/riptide.git
   cd riptide
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/alchemy-labs-co/riptide.git
   ```

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Make
- Git
- A DeepSeek API key for testing

### Initial Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Set your DeepSeek API key**:
   ```bash
   export DEEPSEEK_API_KEY=your_api_key_here
   # Or create a config.json file with your settings
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Build the project**:
   ```bash
   make build
   ```

### Development Tools

We recommend installing these tools for a better development experience:

```bash
# Install golangci-lint for linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install air for live reload during development
go install github.com/cosmtrek/air@latest

# Install gofumpt for stricter formatting
go install mvdan.cc/gofumpt@latest
```

### Important Development Notes

- **Never use `strings.Builder` in Bubble Tea models** - causes panics due to copy-by-value
- **Avoid `log.Printf` in TUI code** - corrupts terminal display
- **Use predefined styles from `styles.go`** - ensures consistency
- **Read `docs/DEVELOPMENT_NOTES.md`** - contains critical implementation lessons

## How to Contribute

### Types of Contributions

1. **Bug Fixes** - Fix issues reported in GitHub Issues
2. **Features** - Add new functionality
3. **Documentation** - Improve README, code comments, or docs
4. **Tests** - Add missing tests or improve test coverage
5. **Performance** - Optimize code for better performance
6. **UI/UX** - Enhance the terminal user interface

### Finding Issues

- Check the [Issues](https://github.com/alchemy-labs-co/riptide/issues) page
- Look for issues labeled `good first issue` or `help wanted`
- Ask in discussions if you want to work on something specific

### Creating Issues

Before creating an issue:
1. Search existing issues to avoid duplicates
2. Use clear, descriptive titles
3. Provide detailed information:
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - System information (OS, Go version)
   - Relevant logs or screenshots

## Pull Request Process

### Before You Start

1. **Discuss major changes** - Open an issue first for significant changes
2. **Keep PRs focused** - One feature/fix per PR
3. **Update from upstream** regularly:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### Creating a Pull Request

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write clear, concise commit messages
   - Follow the code style guidelines
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**:
   ```bash
   make test
   make lint
   ```

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create the PR**:
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changes you made and why
   - Include screenshots for UI changes

### PR Review Process

- All PRs require at least one review
- Address reviewer feedback promptly
- Keep discussions focused and constructive
- Squash commits if requested

## Code Style

### Go Code Style

We follow standard Go conventions with some additional guidelines:

1. **Format code** with `gofumpt`:
   ```bash
   make fmt
   ```

2. **Lint code** with `golangci-lint`:
   ```bash
   make lint
   ```

3. **Naming conventions**:
   - Use meaningful variable names
   - Prefer clarity over brevity
   - Follow Go naming conventions

4. **Error handling**:
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to read file %s: %w", path, err)
   }
   
   // Avoid
   if err != nil {
       return err
   }
   ```

5. **Comments**:
   - Add godoc comments for all exported functions
   - Use clear, concise language
   - Explain "why" not just "what"

### Project Structure

```
riptide/
├── main.go                # Entry point - initializes TUI program
├── internal/
│   ├── api/              # DeepSeek API integration
│   │   ├── client.go     # OpenAI-compatible client wrapper
│   │   └── types.go      # API request/response types
│   ├── config/           # Configuration handling
│   │   └── config.go     # Config loading, validation, defaults
│   ├── conversation/     # Conversation management
│   │   └── history.go    # Message history and token tracking
│   ├── functions/        # File operation functions
│   │   ├── file_ops.go   # Read/write/edit operations
│   │   ├── scanner.go    # Directory scanning with patterns
│   │   └── security.go   # Path validation and limits
│   └── ui/               # Terminal UI (Bubble Tea)
│       ├── model.go      # MVC model and state management
│       ├── render.go     # Rendering logic for all components
│       ├── styles.go     # Centralized lipgloss styles
│       ├── stream.go     # Stream processing and updates
│       ├── commands.go   # Command definitions and autocomplete
│       └── config_menu.go # Interactive configuration menu
├── docs/
│   ├── DEVELOPMENT_NOTES.md # Implementation lessons and patterns
│   └── TASKS.md            # Feature/bug tracking
└── .riptide/              # Project-specific settings
```

Keep code organized in the appropriate packages. Refer to `docs/DEVELOPMENT_NOTES.md` for implementation patterns and lessons learned.

## Key Implementation Patterns

### UI State Management

The project uses Bubble Tea's Model-View-Update (MVU) pattern:

```go
// Model holds all application state
type Model struct {
    messages       []Message
    viewport       viewport.Model
    textInput      textinput.Model
    // ... other state fields
}

// Update handles all state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)

// View renders the UI based on current state
func (m Model) View() string
```

### Command Autocomplete

The autocomplete system provides smart suggestions:
- Shows dropdown with matching commands
- Highlights selected option with DeepSeek blue (#00d4ff)
- Tab/Enter fills command with space
- Escape cancels autocomplete

### Configuration Menu

Interactive settings adjustment (`/config`):
- Full-screen overlay with centered content
- Arrow key navigation
- Enter/Tab/Space cycles through values
- Shows confirmation of changes

### Streaming Responses

Handle DeepSeek API streaming with proper state management:
1. Show "Seeking..." indicator
2. Process reasoning and content separately
3. Update token usage in real-time
4. Support stream cancellation with Ctrl+C

## Testing

### Writing Tests

1. **Test files** should be named `*_test.go`
2. **Test functions** should start with `Test`
3. **Use table-driven tests** for multiple scenarios:
   ```go
   func TestFunction(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           expected string
       }{
           {"empty input", "", ""},
           {"normal input", "hello", "HELLO"},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got := Function(tt.input)
               if got != tt.expected {
                   t.Errorf("got %v, want %v", got, tt.expected)
               }
           })
       }
   }
   ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test ./internal/api

# Run tests with verbose output
go test -v ./...
```

## Documentation

### Code Documentation

- Add godoc comments for all exported types and functions
- Include examples in comments where helpful
- Keep comments up-to-date with code changes

### README Updates

When adding features or changing behavior:
1. Update the main README.md
2. Add examples if applicable
3. Update configuration documentation

### Commit Messages

Follow conventional commit format:
```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Build process or auxiliary tool changes

Example:
```
feat(ui): add syntax highlighting for code blocks

Added syntax highlighting support for code blocks in the conversation view
using the chroma library. Supports automatic language detection.

Closes #123
```

## Community

### Getting Help

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues and discussions first

### Stay Updated

- Watch the repository for updates
- Subscribe to release notifications
- Join community discussions

## License

By contributing to Riptide, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Riptide! Your efforts help make this tool better for everyone. 🚀