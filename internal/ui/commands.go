package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/alchemy-labs-co/riptide/internal/functions"
)

// handleAddCommand handles the /add command to add files or directories to context
func (m Model) handleAddCommand(path string) (tea.Model, tea.Cmd) {
	path = strings.TrimSpace(path)
	if path == "" {
		m.addErrorMessage("No path provided")
		m.updateViewport()
		return m, nil
	}

	m.state = StateProcessing
	enableEmoji := m.config.UI.EnableEmoji

	return m, func() tea.Msg {
		// Normalize the path
		normalizedPath, err := functions.NormalizePath(path)
		if err != nil {
			return ProcessCompleteMsg{
				Error: fmt.Errorf("invalid path: %w", err),
			}
		}

		// Check if it's a file or directory
		fileInfo, err := os.Stat(normalizedPath)
		if err != nil {
			return ProcessCompleteMsg{
				Error: fmt.Errorf("accessing path: %w", err),
			}
		}

		if fileInfo.IsDir() {
			// Handle directory
			return m.addDirectoryToContext(normalizedPath, enableEmoji)
		} else {
			// Handle single file
			return m.addFileToContext(normalizedPath, enableEmoji)
		}
	}
}

// addFileToContext adds a single file to the conversation context
func (m Model) addFileToContext(filePath string, enableEmoji bool) tea.Msg {
	// Check if file is already in context
	if m.history.FileAlreadyInContext(filePath) {
		return ProcessCompleteMsg{
			Result: FormatInfo(fmt.Sprintf("File '%s' is already in the conversation context", filePath), enableEmoji),
		}
	}

	// Read the file
	content, err := m.fileOps.ReadFileForContext(filePath)
	if err != nil {
		return ProcessCompleteMsg{
			Error: fmt.Errorf("reading file: %w", err),
		}
	}

	// Add to history
	m.history.AddSystemMessage(content)

	return ProcessCompleteMsg{
		Result: FormatSuccess(fmt.Sprintf("Added file '%s' to conversation", FormatFilePath(filePath)), enableEmoji),
	}
}

// addDirectoryToContext adds all eligible files from a directory to context
func (m Model) addDirectoryToContext(dirPath string, enableEmoji bool) tea.Msg {
	// Show scanning status
	// Scanning directory...

	// Scan the directory
	result, err := m.scanner.ScanDirectory(dirPath)
	if err != nil {
		return ProcessCompleteMsg{
			Error: fmt.Errorf("scanning directory: %w", err),
		}
	}

	// Read all files
	if len(result.AddedFiles) > 0 {
		fileContents, err := m.scanner.ReadFiles(result.AddedFiles)
		if err != nil {
			return ProcessCompleteMsg{
				Error: fmt.Errorf("reading files: %w", err),
			}
		}

		// Add each file to conversation history
		addedCount := 0
		for filePath, content := range fileContents {
			if !m.history.FileAlreadyInContext(filePath) {
				m.history.AddSystemMessage(fmt.Sprintf("Content of file '%s':\n\n%s", filePath, content))
				addedCount++
			}
		}

		// Format result message
		var resultMsg strings.Builder
		resultMsg.WriteString(FormatSuccess(
			fmt.Sprintf("Added folder '%s' to conversation", FormatFilePath(dirPath)),
			enableEmoji,
		))

		if addedCount > 0 {
			resultMsg.WriteString(fmt.Sprintf("\n\n%s Added files: (%d)\n",
				GetIcon("folder", enableEmoji), addedCount))

			// Show up to 10 files
			displayCount := len(result.AddedFiles)
			if displayCount > 10 {
				displayCount = 10
			}

			for i := 0; i < displayCount; i++ {
				resultMsg.WriteString(fmt.Sprintf("  %s %s\n",
					GetIcon("file", enableEmoji),
					FormatFilePath(result.AddedFiles[i]),
				))
			}

			if len(result.AddedFiles) > 10 {
				resultMsg.WriteString(fmt.Sprintf("  ... and %d more\n", len(result.AddedFiles)-10))
			}
		}

		if len(result.SkippedFiles) > 0 {
			resultMsg.WriteString(fmt.Sprintf("\n%s Skipped files: (%d)\n",
				GetIcon("warning", enableEmoji), len(result.SkippedFiles)))

			// Show first few skipped files
			displayCount := len(result.SkippedFiles)
			if displayCount > 5 {
				displayCount = 5
			}

			for i := 0; i < displayCount; i++ {
				resultMsg.WriteString(fmt.Sprintf("  %s %s\n",
					GetIcon("warning", enableEmoji),
					result.SkippedFiles[i],
				))
			}

			if len(result.SkippedFiles) > 5 {
				resultMsg.WriteString(fmt.Sprintf("  ... and %d more\n", len(result.SkippedFiles)-5))
			}
		}

		if len(result.Errors) > 0 {
			resultMsg.WriteString(fmt.Sprintf("\n%s Errors: (%d)\n",
				GetIcon("error", enableEmoji), len(result.Errors)))
			for _, err := range result.Errors {
				resultMsg.WriteString(fmt.Sprintf("  %s %v\n",
					GetIcon("error", enableEmoji), err))
			}
		}

		return ProcessCompleteMsg{
			Result: resultMsg.String(),
		}
	}

	return ProcessCompleteMsg{
		Result: FormatWarning("No eligible files found in the directory", enableEmoji),
	}
}

// parseCommand parses a command and returns the command name and arguments
func parseCommand(input string) (string, []string) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return "", nil
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	return command, args
}

// isExitCommand checks if the input is an exit command
func isExitCommand(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	return lower == "exit" || lower == "quit"
}
