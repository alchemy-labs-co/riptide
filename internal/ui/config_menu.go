package ui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigOption represents a configuration option with possible values
type ConfigOption struct {
	Name           string
	Description    string
	CurrentValue   string
	PossibleValues []string
	ConfigKey      string // Key in config struct
	ConfigSection  string // Section in config (api, ui, file_operations)
}

// Add config menu state to State enum
const (
	StateConfigMenu State = iota + 10 // Start from 10 to avoid conflicts
)

// ConfigMenuKeyPress handles key presses in config menu
func (m Model) handleConfigMenuKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit config menu without saving
		m.configMenuActive = false
		m.state = StateReady
		// Show that config was opened but no changes made
		m.addUserMessage("/config")
		m.addSystemMessage("⎿  No changes made")
		m.updateViewport()
		return m, nil

	case tea.KeyUp:
		if m.configMenuIndex > 0 {
			m.configMenuIndex--
		} else {
			m.configMenuIndex = len(m.configOptions) - 1
		}
		return m, nil

	case tea.KeyDown:
		if m.configMenuIndex < len(m.configOptions)-1 {
			m.configMenuIndex++
		} else {
			m.configMenuIndex = 0
		}
		return m, nil

	case tea.KeySpace, tea.KeyEnter, tea.KeyTab:
		// Cycle through values for current option
		if m.configMenuIndex < len(m.configOptions) {
			opt := &m.configOptions[m.configMenuIndex]
			currentIdx := -1
			for i, v := range opt.PossibleValues {
				if v == opt.CurrentValue {
					currentIdx = i
					break
				}
			}

			// Move to next value
			nextIdx := (currentIdx + 1) % len(opt.PossibleValues)
			opt.CurrentValue = opt.PossibleValues[nextIdx]
			m.configMenuChanged = true
		}
		return m, nil

	case tea.KeyCtrlS:
		// Save and exit
		return m.saveConfigAndExit()
	}

	// Check for 'q' to quit
	if msg.String() == "q" {
		return m.saveConfigAndExit()
	}

	return m, nil
}

// saveConfigAndExit saves the configuration and exits the menu
func (m Model) saveConfigAndExit() (tea.Model, tea.Cmd) {
	m.configMenuActive = false
	m.state = StateReady

	// Always show the /config command was run
	m.addUserMessage("/config")

	if m.configMenuChanged {
		// Apply changes to config
		for _, opt := range m.configOptions {
			m.applyConfigOption(opt)
		}

		// Save config to file
		if err := m.config.Save("config.json"); err != nil {
			m.addErrorMessage(fmt.Sprintf("Failed to save config: %v", err))
		} else {
			// Get the changes summary
			summary := m.getConfigChangesSummary()
			if summary != "" {
				// Show the changes
				m.addSystemMessage(summary)
			} else {
				// No actual changes even though menu was changed
				m.addSystemMessage("⎿  No changes made")
			}
		}
	} else {
		// No changes were made
		m.addSystemMessage("⎿  No changes made")
	}

	m.updateViewport()
	return m, nil
}

// applyConfigOption applies a config option to the actual config struct
func (m *Model) applyConfigOption(opt ConfigOption) {
	switch opt.ConfigSection {
	case "api":
		switch opt.ConfigKey {
		case "model":
			m.config.API.Model = opt.CurrentValue
		case "max_completion_tokens":
			if val, err := strconv.Atoi(opt.CurrentValue); err == nil {
				m.config.API.MaxCompletionTokens = val
			}
		case "timeout_seconds":
			if val, err := strconv.Atoi(opt.CurrentValue); err == nil {
				m.config.API.TimeoutSeconds = val
			}
		}
	case "ui":
		switch opt.ConfigKey {
		case "theme":
			m.config.UI.Theme = opt.CurrentValue
		case "enable_emoji":
			m.config.UI.EnableEmoji = opt.CurrentValue == "true"
		case "max_history_messages":
			if val, err := strconv.Atoi(opt.CurrentValue); err == nil {
				m.config.UI.MaxHistoryMessages = val
			}
		}
	case "file_operations":
		switch opt.ConfigKey {
		case "max_file_size_mb":
			if val, err := strconv.Atoi(opt.CurrentValue); err == nil {
				m.config.FileOperations.MaxFileSizeMB = val
			}
		}
	}
}

// getConfigChangesSummary returns a summary of configuration changes
func (m Model) getConfigChangesSummary() string {
	var changes []string

	for _, opt := range m.configOptions {
		// Check if value changed from original
		originalValue := m.getOriginalConfigValue(opt)
		if originalValue != opt.CurrentValue {
			// Format specific messages based on the option
			switch opt.ConfigKey {
			case "enable_emoji":
				if opt.CurrentValue == "true" {
					changes = append(changes, "Enabled emoji")
				} else {
					changes = append(changes, "Disabled emoji")
				}
			case "theme":
				changes = append(changes, fmt.Sprintf("Changed theme to %s", opt.CurrentValue))
			case "model":
				changes = append(changes, fmt.Sprintf("Changed model to %s", opt.CurrentValue))
			case "max_history_messages":
				changes = append(changes, fmt.Sprintf("Set max history to %s messages", opt.CurrentValue))
			case "max_completion_tokens":
				changes = append(changes, fmt.Sprintf("Set max tokens to %s", opt.CurrentValue))
			case "timeout_seconds":
				changes = append(changes, fmt.Sprintf("Set timeout to %s seconds", opt.CurrentValue))
			case "max_file_size_mb":
				changes = append(changes, fmt.Sprintf("Set max file size to %s MB", opt.CurrentValue))
			default:
				changes = append(changes, fmt.Sprintf("%s: %s → %s", opt.Name, originalValue, opt.CurrentValue))
			}
		}
	}

	if len(changes) == 0 {
		return ""
	}

	// Format with tree branch
	if len(changes) == 1 {
		return "⎿  " + changes[0]
	}

	result := ""
	for i, change := range changes {
		if i == 0 {
			result += "⎿  " + change
		} else {
			result += "\n   " + change
		}
	}

	return result
}

// getOriginalConfigValue gets the original value from config before changes
func (m Model) getOriginalConfigValue(opt ConfigOption) string {
	switch opt.ConfigSection {
	case "api":
		switch opt.ConfigKey {
		case "model":
			return m.originalConfig.API.Model
		case "max_completion_tokens":
			return strconv.Itoa(m.originalConfig.API.MaxCompletionTokens)
		case "timeout_seconds":
			return strconv.Itoa(m.originalConfig.API.TimeoutSeconds)
		}
	case "ui":
		switch opt.ConfigKey {
		case "theme":
			return m.originalConfig.UI.Theme
		case "enable_emoji":
			return strconv.FormatBool(m.originalConfig.UI.EnableEmoji)
		case "max_history_messages":
			return strconv.Itoa(m.originalConfig.UI.MaxHistoryMessages)
		}
	case "file_operations":
		switch opt.ConfigKey {
		case "max_file_size_mb":
			return strconv.Itoa(m.originalConfig.FileOperations.MaxFileSizeMB)
		}
	}
	return ""
}

// initializeConfigOptions initializes the config menu options
func (m *Model) initializeConfigOptions() {
	m.configOptions = []ConfigOption{
		{
			Name:           "Model",
			Description:    "DeepSeek model to use",
			CurrentValue:   m.config.API.Model,
			PossibleValues: []string{"deepseek-reasoner", "deepseek-chat"},
			ConfigKey:      "model",
			ConfigSection:  "api",
		},
		{
			Name:           "Theme",
			Description:    "UI theme",
			CurrentValue:   m.config.UI.Theme,
			PossibleValues: []string{"default", "dark", "light"},
			ConfigKey:      "theme",
			ConfigSection:  "ui",
		},
		{
			Name:           "Enable Emoji",
			Description:    "Show emoji in UI",
			CurrentValue:   strconv.FormatBool(m.config.UI.EnableEmoji),
			PossibleValues: []string{"true", "false"},
			ConfigKey:      "enable_emoji",
			ConfigSection:  "ui",
		},
		{
			Name:           "Max History Messages",
			Description:    "Maximum messages to keep in history",
			CurrentValue:   strconv.Itoa(m.config.UI.MaxHistoryMessages),
			PossibleValues: []string{"10", "15", "20", "30", "50"},
			ConfigKey:      "max_history_messages",
			ConfigSection:  "ui",
		},
		{
			Name:           "Max Completion Tokens",
			Description:    "Maximum tokens for completion",
			CurrentValue:   strconv.Itoa(m.config.API.MaxCompletionTokens),
			PossibleValues: []string{"32000", "64000", "128000"},
			ConfigKey:      "max_completion_tokens",
			ConfigSection:  "api",
		},
		{
			Name:           "Timeout (seconds)",
			Description:    "API timeout in seconds",
			CurrentValue:   strconv.Itoa(m.config.API.TimeoutSeconds),
			PossibleValues: []string{"120", "300", "600"},
			ConfigKey:      "timeout_seconds",
			ConfigSection:  "api",
		},
		{
			Name:           "Max File Size (MB)",
			Description:    "Maximum file size to read",
			CurrentValue:   strconv.Itoa(m.config.FileOperations.MaxFileSizeMB),
			PossibleValues: []string{"1", "5", "10", "20"},
			ConfigKey:      "max_file_size_mb",
			ConfigSection:  "file_operations",
		},
	}
}

// renderConfigMenu renders the configuration menu
func (m Model) renderConfigMenu() string {
	// Create the menu box
	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 4)

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(SecondaryColor).
		MarginBottom(1)

	title := titleStyle.Render("Settings")
	subtitle := HelpStyle.Render("Configure Riptide preferences")

	// Build menu content
	var content string
	content += title + "\n"
	content += subtitle + "\n\n"

	// Render each option
	for i, opt := range m.configOptions {
		var line string

		// Selection indicator
		if i == m.configMenuIndex {
			line += lipgloss.NewStyle().Foreground(AccentColor).Render("▶ ")
		} else {
			line += "  "
		}

		// Option name (left aligned)
		nameStyle := lipgloss.NewStyle().Width(25)
		if i == m.configMenuIndex {
			nameStyle = nameStyle.Bold(true).Foreground(AccentColor)
		}
		line += nameStyle.Render(opt.Name)

		// Current value (right aligned)
		valueStyle := lipgloss.NewStyle().
			Width(20).
			Align(lipgloss.Right)
		if i == m.configMenuIndex {
			valueStyle = valueStyle.Foreground(WhiteColor)
		} else {
			valueStyle = valueStyle.Foreground(DimTextColor)
		}
		line += valueStyle.Render(opt.CurrentValue)

		content += line + "\n"
	}

	// Footer with instructions
	footer := "\n\n" + HelpStyle.Render("↑/↓ to select • Enter/Tab/Space to change • q or Ctrl+S to save • Esc to cancel")

	return menuStyle.Render(content + footer)
}
