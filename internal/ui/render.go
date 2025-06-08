package ui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/deep-code/deep-code/internal/conversation"
)

// renderWelcome renders the welcome screen
func (m Model) renderWelcome() string {
	enableEmoji := m.config.UI.EnableEmoji

	// Get current time info
	now := time.Now()
	localTime := now.Format("3:04 PM")
	_, tzOffset := now.Zone()
	tzOffsetHours := tzOffset / 3600
	tzSign := "+"
	if tzOffsetHours < 0 {
		tzSign = ""
	}

	utcTime := now.UTC()
	utcHour := utcTime.Hour()
	utcMinute := utcTime.Minute()
	isOffPeak := (utcHour == 16 && utcMinute >= 30) || (utcHour > 16) || (utcHour == 0 && utcMinute <= 30)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "Unknown"
	}

	// Create header with welcome message
	headerContent := fmt.Sprintf("%s Welcome to %s!",
		GetIcon("whale", enableEmoji),
		lipgloss.NewStyle().Bold(true).Render("Deep Code"),
	)

	// Create subtitle
	subtitle := lipgloss.NewStyle().
		Foreground(DimTextColor).
		Italic(true).
		Render("/help for help, /status for your current setup")

	// Create cwd line
	cwdLine := fmt.Sprintf("cwd: %s", lipgloss.NewStyle().Foreground(DimTextColor).Render(cwd))

	// Create time and pricing info
	var timeAndPricing string
	if isOffPeak {
		timeAndPricing = fmt.Sprintf("Local time: %s (UTC%s%d) • %s Off-peak pricing active (75%% off)",
			localTime,
			tzSign,
			tzOffsetHours,
			GetIcon("moon", enableEmoji),
		)
	} else {
		if utcHour < 16 || (utcHour == 16 && utcMinute < 30) {
			// Calculate hours until off-peak
			hoursUntil := 16 - utcHour
			if utcMinute > 30 {
				hoursUntil--
			}
			minutesUntil := 30 - utcMinute
			if minutesUntil < 0 {
				minutesUntil += 60
			}
			timeAndPricing = fmt.Sprintf("Local time: %s (UTC%s%d) • Off-peak in %dh %dm",
				localTime,
				tzSign,
				tzOffsetHours,
				hoursUntil,
				minutesUntil,
			)
		} else {
			timeAndPricing = fmt.Sprintf("Local time: %s (UTC%s%d)",
				localTime,
				tzSign,
				tzOffsetHours,
			)
		}
	}

	// Create the welcome panel content
	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		headerContent,
		"",
		subtitle,
		"",
		cwdLine,
		"",
		lipgloss.NewStyle().Foreground(DimTextColor).Render(timeAndPricing),
	)

	// Apply panel styling - don't set width, let it auto-fit
	panel := WelcomePanelStyle.Render(panelContent)

	return panel
}

// renderHeader renders the header
func (m Model) renderHeader() string {
	enableEmoji := m.config.UI.EnableEmoji
	title := fmt.Sprintf("%s Deep Code", GetIcon("whale", enableEmoji))
	return TitleStyle.Render(title)
}

// renderMessages renders all messages
func (m Model) renderMessages() string {
	var content strings.Builder

	// Log out the message type for debugging
	fmt.Println("Message type:", m.messages[0].Role)
	fmt.Println("Message Metadata:", m.messages[0])

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			// Blue triangle for user messages
			blueTriangle := lipgloss.NewStyle().Foreground(SecondaryColor).Render("▶")
			content.WriteString(fmt.Sprintf("\n%s %s %s\n",
				blueTriangle,
				msg.Content,
				HelpStyle.Render(msg.Timestamp.Format("15:04:05")),
			))

		case "assistant-label":
			// White dot for output tokens
			whiteDot := lipgloss.NewStyle().Foreground(WhiteColor).Render("●")
			content.WriteString(fmt.Sprintf("\n%s %s\n",
				whiteDot,
				AssistantLabelStyle.Render("Assistant>"),
			))

		case "reasoning-label":
			// Blue dot for reasoning tokens
			blueDot := lipgloss.NewStyle().Foreground(SecondaryColor).Render("●")
			content.WriteString(fmt.Sprintf("\n%s %s\n",
				blueDot,
				ReasoningLabelStyle.Render("Reasoning:"),
			))

		case "content":
			// Apply markdown rendering to content
			renderedContent := renderMarkdown(msg.Content)
			// Apply padding to each line
			lines := strings.Split(renderedContent, "\n")
			for _, line := range lines {
				content.WriteString(ContentStyle.Render(line))
				content.WriteString("\n")
			}

		case "reasoning":
			// Show reasoning content with different styling
			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				content.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("#60a5fa")).
					PaddingLeft(2).
					Render(line))
				content.WriteString("\n")
			}

		case "system":
			content.WriteString(fmt.Sprintf("\n%s\n", InfoStyle.Render(msg.Content)))

		case "error":
			content.WriteString(fmt.Sprintf("\n%s\n", ErrorStyle.Render(msg.Content)))

		case "seeking":
			// Show animated seeking indicator
			enableEmoji := m.config.UI.EnableEmoji
			whale := GetIcon("whale", enableEmoji)
			content.WriteString(fmt.Sprintf("\n%s %s %s\n",
				whale,
				m.spinner.View(),
				InfoStyle.Render("Seeking..."),
			))
		}
	}

	return content.String()
}

// renderStatusLine renders the status line
func (m Model) renderStatusLine() string {
	stats := m.history.GetStats()

	// DeepSeek pricing per million tokens (regular hours)
	inputTokensPriceCached := 0.14 / 1_000_000 // $0.14 per 1M tokens
	inputTokensPrice := 0.55 / 1_000_000       // $0.55 per 1M tokens
	outputTokensPrice := 2.19 / 1_000_000      // $2.19 per 1M tokens

	// Off-peak discount (75% off during 16:30-00:30 UTC)
	// Off-peak prices: Cached Input: $0.035, Input: $0.135, Output: $0.550 per 1M tokens
	offPeakDiscount := 0.25 // 75% off means paying 25% of original price

	// Calculate regular hour tokens (total - off-peak)
	regularInputTokens := stats.InputTokens - stats.OffPeakInputTokens
	regularOutputTokens := stats.OutputTokens - stats.OffPeakOutputTokens
	regularCachedTokens := stats.CachedTokens - stats.OffPeakCachedTokens

	// Calculate costs for regular hours
	regularInputCost := float64(regularInputTokens) * inputTokensPrice
	regularOutputCost := float64(regularOutputTokens) * outputTokensPrice
	regularCachedCost := float64(regularCachedTokens) * inputTokensPriceCached

	// Calculate costs for off-peak hours
	offPeakInputCost := float64(stats.OffPeakInputTokens) * inputTokensPrice * offPeakDiscount
	offPeakOutputCost := float64(stats.OffPeakOutputTokens) * outputTokensPrice * offPeakDiscount
	offPeakCachedCost := float64(stats.OffPeakCachedTokens) * inputTokensPriceCached * offPeakDiscount

	// Total cost
	totalCost := regularInputCost + regularOutputCost + regularCachedCost +
		offPeakInputCost + offPeakOutputCost + offPeakCachedCost

	// Check if we're currently in off-peak hours
	now := time.Now().UTC()
	hour := now.Hour()
	minute := now.Minute()
	isOffPeak := (hour == 16 && minute >= 30) || (hour > 16) || (hour == 0 && minute <= 30)

	// Format cost string with off-peak indicator
	costString := fmt.Sprintf("$%.4f", totalCost)
	if isOffPeak {
		costString += " 🌙" // Moon emoji to indicate currently in off-peak
	}

	left := HelpStyle.Render(fmt.Sprintf(
		"Messages: %d | Input: %d | Output: %d | Cached: %d | Cost: %s",
		stats.TotalMessages,
		stats.InputTokens,
		stats.OutputTokens,
		stats.CachedTokens,
		costString,
	))

	right := HelpStyle.Render(fmt.Sprintf(
		"Model: %s",
		m.config.API.Model,
	))

	statusLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(m.width/2).Render(left),
		lipgloss.NewStyle().Width(m.width/2).Align(lipgloss.Right).Render(right),
	)

	return statusLine
}

// renderInput renders the input area
func (m Model) renderInput() string {
	// Use a blue triangle instead of "You"
	prompt := lipgloss.NewStyle().Foreground(SecondaryColor).Render("▶ ")

	if m.state != StateReady {
		inputContent := prompt + HelpStyle.Render("(waiting...)")
		// Create input box with full width
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(0, 1).
			Width(m.width - 2).
			Render(inputContent)
	}

	// Get current input and build content
	currentValue := m.textInput.Value()
	var inputContent string
	cursorPos := m.textInput.Position()

	// Add autocomplete suggestion if active
	if m.autocompleteActive && m.autocompleteSuggestion != "" {
		// Show the suggestion as dimmed text after the current input
		suggestionPart := m.autocompleteSuggestion[len(currentValue):]
		dimmedSuggestion := HelpStyle.Render(suggestionPart)

		// Build with cursor
		if cursorPos < len(currentValue) {
			// Cursor in middle of text
			before := currentValue[:cursorPos]
			after := currentValue[cursorPos:]
			cursor := lipgloss.NewStyle().Background(WhiteColor).Foreground(lipgloss.Color("#000000")).Render(" ")
			inputContent = prompt + before + cursor + after + dimmedSuggestion
		} else {
			// Cursor at end
			cursor := lipgloss.NewStyle().Background(WhiteColor).Foreground(lipgloss.Color("#000000")).Render(" ")
			inputContent = prompt + currentValue + cursor + dimmedSuggestion
		}
	} else if currentValue == "" {
		// Show placeholder when empty with cursor
		cursor := lipgloss.NewStyle().Background(WhiteColor).Foreground(lipgloss.Color("#000000")).Render(" ")
		inputContent = prompt + cursor + HelpStyle.Render("Type your message...")
	} else {
		// Show typed content with cursor
		if cursorPos < len(currentValue) {
			// Cursor in middle of text
			before := currentValue[:cursorPos]
			after := currentValue[cursorPos:]
			cursor := lipgloss.NewStyle().Background(WhiteColor).Foreground(lipgloss.Color("#000000")).Render(" ")
			inputContent = prompt + before + cursor + after
		} else {
			// Cursor at end
			cursor := lipgloss.NewStyle().Background(WhiteColor).Foreground(lipgloss.Color("#000000")).Render(" ")
			inputContent = prompt + currentValue + cursor
		}
	}

	// Create input box with full width
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(0, 1).
		Width(m.width - 2).
		Render(inputContent)

	// Add autocomplete dropdown if active
	if m.autocompleteActive && len(m.autocompleteMatches) > 0 {
		dropdown := m.renderAutocompleteDropdown()
		hintLine := HelpStyle.Render("  Enter/Tab to complete • ↑↓ to navigate • Esc to cancel")
		return inputBox + "\n" + dropdown + "\n" + hintLine
	}

	// Add status indicator on the right
	var statusText string
	switch m.state {
	case StateStreaming:
		statusText = m.spinner.View() + " " + InfoStyle.Render("Seeking...")
	case StateProcessing:
		statusText = m.spinner.View() + " " + InfoStyle.Render("Processing...")
	case StateError:
		statusText = ErrorStyle.Render("Error occurred")
	case StateReady:
		statusText = SuccessStyle.Render("Ready")
	}

	// Place status on the right below the input box
	if statusText != "" {
		statusLine := lipgloss.NewStyle().
			Width(m.width - 2).
			Align(lipgloss.Right).
			MarginTop(1).
			Render(statusText)
		return inputBox + "\n" + statusLine
	}

	return inputBox
}

// getHelpText returns the help text
func (m Model) getHelpText() string {
	enableEmoji := m.config.UI.EnableEmoji

	return fmt.Sprintf(`%s Deep Code Help

%s Commands:
  /add <path>     - Add file or directory to conversation context
  /clear          - Clear conversation history
  /config         - Configure settings
  /help           - Show this help message
  /status         - Show current configuration and pricing info
  exit/quit       - Exit the application
  Ctrl+C          - Force quit
  Ctrl+D          - Quit (when ready)
  PgUp/PgDown     - Scroll conversation

%s File Operations:
  The AI can automatically:
  • Read single or multiple files
  • Create new files or overwrite existing ones
  • Edit files by replacing specific snippets
  • Scan directories with smart filtering

%s Tips:
  • Just mention file names naturally - the AI will read them
  • Ask for code changes and the AI will implement them
  • The AI shows its reasoning process before responding
  • File operations are executed automatically when needed`,
		GetIcon("info", enableEmoji),
		GetIcon("arrow", enableEmoji),
		GetIcon("file", enableEmoji),
		GetIcon("sparkle", enableEmoji),
	)
}

// getStatusText returns the status information
func (m Model) getStatusText() string {
	enableEmoji := m.config.UI.EnableEmoji

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "Unknown"
	}

	// Get current time info
	now := time.Now()
	localTime := now.Format("3:04 PM")
	_, tzOffset := now.Zone()
	tzOffsetHours := tzOffset / 3600
	tzSign := "+"
	if tzOffsetHours < 0 {
		tzSign = ""
	}

	utcTime := now.UTC()
	utcHour := utcTime.Hour()
	utcMinute := utcTime.Minute()
	isOffPeak := (utcHour == 16 && utcMinute >= 30) || (utcHour > 16) || (utcHour == 0 && utcMinute <= 30)

	// Calculate off-peak hours in local time
	offPeakStartUTC := time.Date(now.Year(), now.Month(), now.Day(), 16, 30, 0, 0, time.UTC)
	offPeakEndUTC := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 30, 0, 0, time.UTC)
	offPeakStartLocal := offPeakStartUTC.In(now.Location()).Format("3:04 PM")
	offPeakEndLocal := offPeakEndUTC.In(now.Location()).Format("3:04 PM")

	// Get stats
	stats := m.history.GetStats()
	totalCost := m.calculateTotalCost(stats)

	// Create pricing status line
	var pricingStatusLine string
	if isOffPeak {
		pricingStatusLine = fmt.Sprintf("└ Status: %s Off-peak pricing ACTIVE (75%% off)", GetIcon("moon", enableEmoji))
	} else {
		pricingStatusLine = "└ Status: Regular pricing"
	}

	// Build status text in structured format
	// Use lipgloss styles for consistent formatting
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(SecondaryColor)

	statusText := fmt.Sprintf(`%s

%s
└ %s

%s
└ Current: %s (UTC%s%d)
└ Off-peak: %s - %s daily (75%% off)
%s

%s
└ Default: %s with 64K context window

%s
└ Messages: %d
└ Tokens: %d input, %d output, %d cached
└ Cost: $%.4f`,
		"Deep Code Status v1.0.0",
		headerStyle.Render("Working Directory"),
		cwd,
		headerStyle.Render("Time • "+localTime),
		localTime,
		tzSign,
		tzOffsetHours,
		offPeakStartLocal,
		offPeakEndLocal,
		pricingStatusLine,
		headerStyle.Render("Model • /model"),
		m.config.API.Model,
		headerStyle.Render("Session • /clear"),
		stats.TotalMessages,
		stats.InputTokens,
		stats.OutputTokens,
		stats.CachedTokens,
		totalCost,
	)

	return statusText + "\n\nPress Enter to continue..."
}

// calculateTotalCost calculates the total cost from stats
func (m Model) calculateTotalCost(stats conversation.ConversationStats) float64 {
	// DeepSeek pricing per million tokens (regular hours)
	inputTokensPriceCached := 0.14 / 1_000_000
	inputTokensPrice := 0.55 / 1_000_000
	outputTokensPrice := 2.19 / 1_000_000

	// Off-peak discount (75% off)
	offPeakDiscount := 0.25

	// Calculate regular hour tokens (total - off-peak)
	regularInputTokens := stats.InputTokens - stats.OffPeakInputTokens
	regularOutputTokens := stats.OutputTokens - stats.OffPeakOutputTokens
	regularCachedTokens := stats.CachedTokens - stats.OffPeakCachedTokens

	// Calculate costs for regular hours
	regularInputCost := float64(regularInputTokens) * inputTokensPrice
	regularOutputCost := float64(regularOutputTokens) * outputTokensPrice
	regularCachedCost := float64(regularCachedTokens) * inputTokensPriceCached

	// Calculate costs for off-peak hours
	offPeakInputCost := float64(stats.OffPeakInputTokens) * inputTokensPrice * offPeakDiscount
	offPeakOutputCost := float64(stats.OffPeakOutputTokens) * outputTokensPrice * offPeakDiscount
	offPeakCachedCost := float64(stats.OffPeakCachedTokens) * inputTokensPriceCached * offPeakDiscount

	// Total cost
	return regularInputCost + regularOutputCost + regularCachedCost +
		offPeakInputCost + offPeakOutputCost + offPeakCachedCost
}

// renderDiffTable renders a table showing file edits
func renderDiffTable(edits []DiffEdit, enableEmoji bool) string {
	if len(edits) == 0 {
		return ""
	}

	var rows []string

	// Header
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		TableHeaderStyle.Width(30).Render("File Path"),
		TableHeaderStyle.Width(40).Render("Original"),
		TableHeaderStyle.Width(40).Render("New"),
	)
	rows = append(rows, header)

	// Rows
	for _, edit := range edits {
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			TableCellStyle.Width(30).Render(FilePathStyle.Render(edit.Path)),
			TableCellStyle.Width(40).Render(DiffOldStyle.Render(truncate(edit.Original, 35))),
			TableCellStyle.Width(40).Render(DiffNewStyle.Render(truncate(edit.New, 35))),
		)
		rows = append(rows, row)
	}

	title := fmt.Sprintf("%s Proposed Edits", GetIcon("file", enableEmoji))

	return PanelStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render(title),
			strings.Join(rows, "\n"),
		),
	)
}

// DiffEdit represents a file edit for display
type DiffEdit struct {
	Path     string
	Original string
	New      string
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%.1fm", d.Minutes())
	default:
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// renderMarkdown applies basic markdown formatting to text
func renderMarkdown(text string) string {
	// Bold text: **text** or __text__
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	text = boldRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "*_")
		return lipgloss.NewStyle().Bold(true).Render(content)
	})

	// Inline code: `code`
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	text = inlineCodeRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "`")
		return InlineCodeStyle.Render(content)
	})

	// Headers: # Header (at start of line)
	headerRegex := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	text = headerRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := strings.SplitN(match, " ", 2)
		if len(parts) < 2 {
			return match
		}
		content := parts[1]
		return lipgloss.NewStyle().Bold(true).Underline(true).Render(content)
	})

	// Lists: - item or * item (at start of line)
	listItemRegex := regexp.MustCompile(`(?m)^(\s*)([-*])\s+(.+)$`)
	text = listItemRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := regexp.MustCompile(`^(\s*)([-*])\s+(.+)$`).FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		indent := parts[1]
		content := parts[3]

		// Use a bullet point
		return indent + lipgloss.NewStyle().Foreground(AccentColor).Render("•") + " " + content
	})

	return text
}

// truncate truncates a string to the specified length, adding "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// renderAutocompleteDropdown renders the autocomplete suggestions as a dropdown list
func (m Model) renderAutocompleteDropdown() string {
	if !m.autocompleteActive || len(m.autocompleteMatches) == 0 {
		return ""
	}

	// Create list items with selection indicator
	var items []string
	maxNameLen := 0

	// Find the maximum command name length for alignment
	for _, cmd := range m.autocompleteMatches {
		if len(cmd.Name) > maxNameLen {
			maxNameLen = len(cmd.Name)
		}
	}

	// Build each item in the list
	for i, cmd := range m.autocompleteMatches {
		var item string

		// Add selection indicator
		if i == m.autocompleteSelectedIndex {
			// Selected item with highlighted text (no background)
			commandStyle := lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				Padding(0, 1)

			descStyle := lipgloss.NewStyle().
				Foreground(WhiteColor).
				Padding(0, 1)

			item = fmt.Sprintf("%s  %s",
				commandStyle.Width(maxNameLen+2).Render(cmd.Name),
				descStyle.Render(cmd.Description),
			)
		} else {
			// Normal item
			commandStyle := lipgloss.NewStyle().
				Foreground(DimTextColor).
				Padding(0, 1)

			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Padding(0, 1)

			item = fmt.Sprintf("%s  %s",
				commandStyle.Width(maxNameLen+2).Render(cmd.Name),
				descStyle.Render(cmd.Description),
			)
		}

		items = append(items, item)
	}

	// Join items and create dropdown box
	dropdownContent := strings.Join(items, "\n")

	// Create dropdown without border
	dropdown := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2).
		Render(dropdownContent)

	return dropdown
}
