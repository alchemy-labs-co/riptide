# Deep Code Development Notes

This document contains important lessons learned, architectural decisions, and implementation quirks discovered during development.

## Architecture Overview

### Core Components

1. **TUI Framework**: Built with Bubble Tea (bubbletea) for terminal UI
2. **Styling**: Uses lipgloss for terminal styling
3. **API Client**: Wraps go-openai library for DeepSeek API compatibility
4. **State Management**: Model-View-Update pattern with centralized state

### Key Files

- `main.go` - Entry point, initializes TUI program
- `internal/ui/model.go` - Core state management and update logic
- `internal/ui/render.go` - All rendering logic for UI components
- `internal/ui/styles.go` - Centralized styling definitions
- `internal/api/client.go` - DeepSeek API client implementation
- `internal/conversation/history.go` - Token tracking and conversation management

## Critical Implementation Lessons

### 1. Bubble Tea Model Copying Issues

**Problem**: Bubble Tea passes models by value, which can cause panics with certain types.

```go
// BAD - causes panic: "strings: illegal use of non-zero Builder copied by value"
type Model struct {
    currentContent strings.Builder
}

// GOOD - use string instead
type Model struct {
    currentContent string
}
```

**Lesson**: Avoid types with copy-check mechanisms in Bubble Tea models. Use simple types or pointers.

### 2. Context Cancellation in Streaming

**Problem**: Improper context cancellation caused immediate stream termination.

```go
// BAD - cancels immediately!
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()

// GOOD - cancel only when streaming completes
go func() {
    defer cancel()
    // ... streaming logic
}()
```

### 3. Terminal UI Rendering Quirks

#### Double-Line Input Box Issue
- **Problem**: Using `textinput.View()` directly can cause double-line rendering
- **Solution**: Manually construct input content with cursor positioning
- **Key**: Set `ti.Placeholder = ""` to prevent conflicts

#### Cursor Visibility
```go
// Make cursor visible as white block
ti.TextStyle = lipgloss.NewStyle().Foreground(WhiteColor)
ti.CursorStyle = lipgloss.NewStyle().
    Background(WhiteColor).
    Foreground(lipgloss.Color("#000000"))
```

#### Manual Cursor Rendering
```go
// Get cursor position
cursorPos := m.textInput.Position()

// Render cursor manually for better control
cursor := lipgloss.NewStyle().
    Background(WhiteColor).
    Foreground(lipgloss.Color("#000000")).
    Render(" ")
```

### 4. Autocomplete Implementation

#### Design Decisions
1. **No borders on dropdown**: Cleaner terminal appearance
2. **Text highlighting over background**: Better for dark terminals
3. **Enter/Tab behavior**: Fills command + space without submitting
4. **Dynamic filtering**: Only show commands matching typed prefix
5. **Escape cancels**: Standard UX pattern

#### Color Scheme
- Selected: `AccentColor` (#00d4ff - DeepSeek light blue) + Bold
- Unselected: `DimTextColor` (#6b7280)
- Descriptions: `#9CA3AF` (optimized for dark terminals)

#### Key Code Pattern
```go
// Track all matches, not just first
m.autocompleteMatches = []Command{}
for _, cmd := range availableCommands {
    if strings.HasPrefix(cmd.Name, lowerInput) && cmd.Name != lowerInput {
        m.autocompleteMatches = append(m.autocompleteMatches, cmd)
    }
}
```

### 5. Message Rendering and Scrolling

#### Smart Auto-Scrolling
```go
func (m *Model) updateViewport() {
    atBottom := m.viewport.AtBottom()
    nearBottom := m.viewport.YOffset >= (m.viewport.TotalLineCount() - m.viewport.Height - 5)
    
    m.viewport.SetContent(content)
    
    // Only scroll if already at/near bottom
    if atBottom || nearBottom {
        m.viewport.GotoBottom()
    }
}
```

**Key Insight**: Preserve user's scroll position unless they're at the bottom.

### 6. Markdown Rendering

**Implementation**: Regex-based parsing with Lip Gloss styling
```go
// Bold: **text** → Bold style
boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
text = boldRegex.ReplaceAllStringFunc(text, func(match string) string {
    content := strings.Trim(match, "*_")
    return lipgloss.NewStyle().Bold(true).Render(content)
})
```

**Note**: No external markdown library needed - Lip Gloss is sufficient.

### 7. Logging Interference

**Critical**: Never use `log.Printf` in TUI code - it corrupts the terminal display!

```go
// BAD - corrupts terminal
log.Printf("Debug info")

// GOOD - use debug messages or error display
m.addErrorMessage(fmt.Sprintf("Debug: %v", info))
```

### 8. State Management Patterns

#### Seeking Indicator Pattern
```go
// Add when starting
m.addSeekingIndicator()

// Remove when content arrives
m.removeSeekingIndicator()
```

#### Message Type System
- `"user"` - User input with blue triangle
- `"assistant-label"` - White dot indicator
- `"reasoning-label"` - Blue dot indicator
- `"content"` - Actual response content
- `"reasoning"` - Reasoning content (different style)
- `"seeking"` - Animated loading indicator
- `"system"` - Help/status messages
- `"error"` - Error messages

### 8.1 Config Menu Implementation

#### Design Pattern
- Full-screen overlay with centered content
- Arrow key navigation with visual indicators
- Immediate visual feedback on selection
- Tree-style confirmation messages

#### State Management
```go
// Config menu state fields
configMenuActive   bool
configMenuIndex    int
configOptions      []ConfigOption
configMenuChanged  bool
originalConfig     config.Config  // For comparison
```

#### Confirmation Message Pattern
```go
// Show command execution
m.addUserMessage("/config")

// Show results with tree branch
m.addSystemMessage("⎿  Disabled emoji")

// Multiple changes
m.addSystemMessage("⎿  Enabled emoji\n   Changed model to deepseek-chat")
```

#### Configuration Options Structure
- Predefined values only (no free text input)
- Store original config for comparison
- Apply changes only on save (not escape)
- Show "No changes made" when appropriate

### 9. API Integration Details

#### Token Usage Tracking
- Track regular vs off-peak tokens separately
- Off-peak hours: 16:30-00:30 UTC (75% discount)
- Update costs in real-time based on UTC time

#### Streaming Event Flow
1. User input → Show seeking indicator
2. First content → Remove seeking, show assistant label
3. Reasoning → Blue dot + reasoning style
4. Content → Regular content style
5. Tool calls → Execute after stream
6. Done → Update token usage

### 10. Input Handling Best Practices

#### Command Design
- Avoid redundancy: `/quit` not `/exit` and `/quit`
- Auto-space after command completion
- Clear command list with descriptions

#### Key Bindings
- **Tab/Enter**: Accept autocomplete
- **↑/↓**: Navigate autocomplete or scroll
- **Esc**: Cancel autocomplete
- **Ctrl+C**: Cancel stream or quit
- **PgUp/PgDn**: Scroll conversation

#### Config Menu (`/config`)
- **↑/↓**: Navigate options
- **Enter/Tab/Space**: Cycle through values
- **q or Ctrl+S**: Save changes and exit
- **Esc**: Cancel without saving

## Common Pitfalls and Solutions

### 1. Import Management
```go
// Always import lipgloss when using styles
import "github.com/charmbracelet/lipgloss"
```

### 2. File Modifications
- Always read a file before writing
- Check for external modifications
- Use `MultiEdit` for multiple changes

### 3. Color Consistency
- Use predefined constants from `styles.go`
- Test in both light and dark terminals
- Avoid hardcoded colors

### 4. Performance Considerations
- Minimize viewport updates
- Use targeted re-renders
- Batch tool executions

## Debugging Strategies

1. **Temporary Status Messages**: Use `ProcessCompleteMsg`
2. **State Tracking**: Add debug fields temporarily
3. **Build Frequently**: `go build -o deep-code main.go`
4. **Test Edge Cases**: Empty input, long text, rapid commands

## Code Style Guidelines

1. **No Comments**: Unless specifically requested
2. **Error Display**: Show in UI, don't panic
3. **Concise Responses**: Match terminal UI style
4. **Feature Parity**: Maintain with Python version

## Testing Checklist

- [ ] Autocomplete with 0, 1, multiple matches
- [ ] Scrolling preservation during updates
- [ ] Markdown rendering combinations
- [ ] Stream interruption (Ctrl+C)
- [ ] Double-line input box issue
- [ ] Cursor visibility in all states
- [ ] Off-peak hour transitions
- [ ] Command execution flow
- [ ] Config menu navigation and value cycling
- [ ] Config save/cancel confirmation messages
- [ ] Config changes persistence to config.json

## Future Enhancement Ideas

1. **Config Menu**: ✅ Implemented with full keyboard navigation
2. **Search**: Ctrl+F within conversation
3. **Export**: Save conversations
4. **Themes**: Light/dark mode support
5. **Multi-model**: Switch between DeepSeek models

## Important Constants

- Viewport height: `msg.Height - 8`
- Input width: `m.width - 2`
- Scroll threshold: 5 lines from bottom
- Off-peak discount: 75% (0.25 multiplier)
- Max file size: 1MB default

## Security Considerations

1. Path validation and normalization
2. File size limits enforcement
3. Binary file detection
4. No API key exposure in logs

---

*Last Updated: After implementing autocomplete, markdown rendering, cursor visibility, and fixing all rendering issues*