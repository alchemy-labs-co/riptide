package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	PrimaryColor   = lipgloss.Color("#0066ff")
	SecondaryColor = lipgloss.Color("#3b82f6")
	AccentColor    = lipgloss.Color("#00d4ff")

	// Status colors
	SuccessColor = lipgloss.Color("#10b981")
	WarningColor = lipgloss.Color("#f59e0b")
	ErrorColor   = lipgloss.Color("#ef4444")

	// Text colors
	BrightCyan   = lipgloss.Color("#00ffff")
	DimTextColor = lipgloss.Color("#6b7280")
	WhiteColor   = lipgloss.Color("#ffffff")
	LinkColor    = lipgloss.Color("#2e7de9")

	// Background colors
	DarkBgColor  = lipgloss.Color("#1e3a8a")
	LightBgColor = lipgloss.Color("#e0f2fe")
)

// Styles for different UI elements
var (
	// Title and header styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			Padding(0, 2)

	SubheaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 2)

	WelcomePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SecondaryColor). // Use our light blue
				PaddingLeft(10).
				PaddingRight(10).
				PaddingTop(1).
				PaddingBottom(1)

	// Message styles
	UserPromptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor)

	AssistantLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SecondaryColor)

	ReasoningLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#60a5fa"))

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SuccessColor)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WarningColor)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ErrorColor)

	InfoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor)

	// File operation styles
	FilePathStyle = lipgloss.NewStyle().
			Foreground(BrightCyan)

	FileIconStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	// Content styles
	ContentStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	CodeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1f2937")).
			Foreground(WhiteColor).
			Padding(1).
			MarginTop(1).
			MarginBottom(1)

	InlineCodeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")).
			Foreground(lipgloss.Color("#f9fafb")).
			Padding(0, 1)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// Spinner styles
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(DimTextColor)

	// Table styles for diffs
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SecondaryColor).
				Background(DarkBgColor).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	DiffOldStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Faint(true)

	DiffNewStyle = lipgloss.NewStyle().
			Foreground(SuccessColor)
)

// GetIcon returns an icon with optional emoji support
func GetIcon(iconType string, enableEmoji bool) string {
	if !enableEmoji {
		switch iconType {
		case "success":
			return "[✓]"
		case "error":
			return "[✗]"
		case "warning":
			return "[!]"
		case "info":
			return "[i]"
		case "file":
			return "[F]"
		case "folder":
			return "[D]"
		case "arrow":
			return "→"
		case "thinking":
			return "[...]"
		case "bot":
			return "[AI]"
		case "user":
			return "[You]"
		case "whale":
			return "[DS]"
		case "moon":
			return "[OFF-PEAK]"
		default:
			return ""
		}
	}

	switch iconType {
	case "success":
		return "✓"
	case "error":
		return "✗"
	case "warning":
		return "⚠"
	case "info":
		return "ℹ"
	case "file":
		return "📄"
	case "folder":
		return "📁"
	case "arrow":
		return "→"
	case "thinking":
		return "💭"
	case "bot":
		return "🤖"
	case "user":
		return "🔵"
	case "whale":
		return "🐋"
	case "sparkle":
		return "✨"
	case "lightning":
		return "⚡"
	case "search":
		return "🔍"
	case "wave":
		return "👋"
	case "moon":
		return "🌙"
	default:
		return ""
	}
}

// FormatSuccess formats a success message
func FormatSuccess(msg string, enableEmoji bool) string {
	icon := GetIcon("success", enableEmoji)
	return SuccessStyle.Render(icon) + " " + msg
}

// FormatError formats an error message
func FormatError(msg string, enableEmoji bool) string {
	icon := GetIcon("error", enableEmoji)
	return ErrorStyle.Render(icon) + " " + msg
}

// FormatWarning formats a warning message
func FormatWarning(msg string, enableEmoji bool) string {
	icon := GetIcon("warning", enableEmoji)
	return WarningStyle.Render(icon) + " " + msg
}

// FormatInfo formats an info message
func FormatInfo(msg string, enableEmoji bool) string {
	icon := GetIcon("info", enableEmoji)
	return InfoStyle.Render(icon) + " " + msg
}

// FormatFilePath formats a file path with styling
func FormatFilePath(path string) string {
	return FilePathStyle.Render(path)
}
