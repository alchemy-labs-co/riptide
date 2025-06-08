package functions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alchemy-labs-co/riptide/internal/api"
	"github.com/alchemy-labs-co/riptide/internal/config"
)

// FileOperations handles all file-related operations
type FileOperations struct {
	config *config.Config
}

// NewFileOperations creates a new FileOperations instance
func NewFileOperations(cfg *config.Config) *FileOperations {
	return &FileOperations{
		config: cfg,
	}
}

// ExecuteFunction executes a function call and returns the result
func (f *FileOperations) ExecuteFunction(toolCall api.ToolCall) (string, error) {
	var args api.FileOperationArgs
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("parsing arguments: %w", err)
	}

	switch toolCall.Function.Name {
	case "read_file":
		return f.readFile(args.FilePath)
	case "read_multiple_files":
		return f.readMultipleFiles(args.FilePaths)
	case "create_file":
		return f.createFile(args.FilePath, args.Content)
	case "create_multiple_files":
		return f.createMultipleFiles(args.Files)
	case "edit_file":
		return f.editFile(args.FilePath, args.OriginalSnippet, args.NewSnippet)
	default:
		return "", fmt.Errorf("unknown function: %s", toolCall.Function.Name)
	}
}

// readFile reads the content of a single file
func (f *FileOperations) readFile(filePath string) (string, error) {
	normalizedPath, err := NormalizePath(filePath)
	if err != nil {
		return "", fmt.Errorf("normalizing path: %w", err)
	}

	content, err := os.ReadFile(normalizedPath)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return fmt.Sprintf("Content of file '%s':\n\n%s", normalizedPath, string(content)), nil
}

// readMultipleFiles reads the content of multiple files
func (f *FileOperations) readMultipleFiles(filePaths []string) (string, error) {
	var results []string
	separator := strings.Repeat("=", 50)

	for _, filePath := range filePaths {
		normalizedPath, err := NormalizePath(filePath)
		if err != nil {
			results = append(results, fmt.Sprintf("Error reading '%s': %v", filePath, err))
			continue
		}

		content, err := os.ReadFile(normalizedPath)
		if err != nil {
			results = append(results, fmt.Sprintf("Error reading '%s': %v", filePath, err))
			continue
		}

		results = append(results, fmt.Sprintf("Content of file '%s':\n\n%s", normalizedPath, string(content)))
	}

	return "\n\n" + separator + "\n\n" + strings.Join(results, "\n\n"+separator+"\n\n"), nil
}

// createFile creates or overwrites a file
func (f *FileOperations) createFile(filePath, content string) (string, error) {
	normalizedPath, err := NormalizePath(filePath)
	if err != nil {
		return "", fmt.Errorf("normalizing path: %w", err)
	}

	// Validate file size
	maxSize := f.config.FileOperations.MaxFileSizeMB * 1024 * 1024
	if len(content) > maxSize {
		return "", fmt.Errorf("file content exceeds %dMB size limit", f.config.FileOperations.MaxFileSizeMB)
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(normalizedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating parent directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(normalizedPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return fmt.Sprintf("Successfully created file '%s'", normalizedPath), nil
}

// createMultipleFiles creates multiple files at once
func (f *FileOperations) createMultipleFiles(files []api.FileToCreate) (string, error) {
	var createdFiles []string

	for _, file := range files {
		if _, err := f.createFile(file.Path, file.Content); err != nil {
			return "", fmt.Errorf("creating file '%s': %w", file.Path, err)
		}
		createdFiles = append(createdFiles, file.Path)
	}

	return fmt.Sprintf("Successfully created %d files: %s", len(createdFiles), strings.Join(createdFiles, ", ")), nil
}

// editFile edits a file by replacing a snippet
func (f *FileOperations) editFile(filePath, originalSnippet, newSnippet string) (string, error) {
	normalizedPath, err := NormalizePath(filePath)
	if err != nil {
		return "", fmt.Errorf("normalizing path: %w", err)
	}

	// Read the current content
	content, err := os.ReadFile(normalizedPath)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	contentStr := string(content)

	// Check occurrences
	occurrences := strings.Count(contentStr, originalSnippet)
	if occurrences == 0 {
		return "", fmt.Errorf("original snippet not found in file")
	}
	if occurrences > 1 {
		return "", fmt.Errorf("ambiguous edit: %d matches found for the snippet", occurrences)
	}

	// Replace the snippet
	updatedContent := strings.Replace(contentStr, originalSnippet, newSnippet, 1)

	// Write the updated content
	if err := os.WriteFile(normalizedPath, []byte(updatedContent), 0644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return fmt.Sprintf("Successfully edited file '%s'", normalizedPath), nil
}

// ReadFileForContext reads a file and returns it formatted for conversation context
func (f *FileOperations) ReadFileForContext(filePath string) (string, error) {
	normalizedPath, err := NormalizePath(filePath)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(normalizedPath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Content of file '%s':\n\n%s", normalizedPath, string(content)), nil
}
