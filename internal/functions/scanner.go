package functions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alchemy-labs-co/riptide/internal/config"
)

// ScanResult represents the result of scanning a directory
type ScanResult struct {
	AddedFiles   []string
	SkippedFiles []string
	TotalScanned int
	Errors       []error
}

// DirectoryScanner handles directory scanning with exclusions
type DirectoryScanner struct {
	config             *config.Config
	excludedFiles      map[string]bool
	excludedExtensions map[string]bool
}

// NewDirectoryScanner creates a new DirectoryScanner instance
func NewDirectoryScanner(cfg *config.Config) *DirectoryScanner {
	return &DirectoryScanner{
		config:             cfg,
		excludedFiles:      config.GetExcludedFiles(),
		excludedExtensions: config.GetExcludedExtensions(),
	}
}

// ScanDirectory scans a directory and returns the results
func (s *DirectoryScanner) ScanDirectory(dirPath string) (*ScanResult, error) {
	normalizedPath, err := NormalizePath(dirPath)
	if err != nil {
		return nil, fmt.Errorf("normalizing directory path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("accessing directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", normalizedPath)
	}

	result := &ScanResult{
		AddedFiles:   make([]string, 0),
		SkippedFiles: make([]string, 0),
		Errors:       make([]error, 0),
	}

	// Walk the directory
	err = filepath.Walk(normalizedPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("accessing %s: %w", path, err))
			return nil // Continue walking
		}

		// Skip if we've reached the file limit
		if result.TotalScanned >= s.config.FileOperations.MaxFilesPerScan {
			return filepath.SkipAll
		}

		// Handle directories
		if info.IsDir() {
			// Skip hidden directories
			if IsHiddenFile(info.Name()) && path != normalizedPath {
				result.SkippedFiles = append(result.SkippedFiles, path+" (hidden directory)")
				return filepath.SkipDir
			}

			// Skip excluded directories
			if s.excludedFiles[info.Name()] {
				result.SkippedFiles = append(result.SkippedFiles, path+" (excluded directory)")
				return filepath.SkipDir
			}

			return nil // Continue into directory
		}

		// Handle files
		result.TotalScanned++

		// Skip hidden files
		if IsHiddenFile(info.Name()) {
			result.SkippedFiles = append(result.SkippedFiles, path+" (hidden file)")
			return nil
		}

		// Skip excluded files
		if s.excludedFiles[info.Name()] {
			result.SkippedFiles = append(result.SkippedFiles, path+" (excluded file)")
			return nil
		}

		// Skip by extension
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if s.excludedExtensions[ext] {
			result.SkippedFiles = append(result.SkippedFiles, path+" (excluded extension)")
			return nil
		}

		// Skip files that are too large
		maxSize := int64(s.config.FileOperations.MaxFileSizeMB * 1024 * 1024)
		if info.Size() > maxSize {
			result.SkippedFiles = append(result.SkippedFiles,
				fmt.Sprintf("%s (exceeds %dMB limit)", path, s.config.FileOperations.MaxFileSizeMB))
			return nil
		}

		// Skip binary files
		isBinary, err := IsBinaryFile(path, s.config.FileOperations.BinaryPeekSize)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("checking if %s is binary: %w", path, err))
			result.SkippedFiles = append(result.SkippedFiles, path+" (error checking file type)")
			return nil
		}
		if isBinary {
			result.SkippedFiles = append(result.SkippedFiles, path+" (binary file)")
			return nil
		}

		// File passed all checks
		result.AddedFiles = append(result.AddedFiles, path)

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return result, nil
}

// ReadFiles reads the content of multiple files and returns them as a map
func (s *DirectoryScanner) ReadFiles(filePaths []string) (map[string]string, error) {
	contents := make(map[string]string)

	for _, filePath := range filePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", filePath, err)
		}
		contents[filePath] = string(content)
	}

	return contents, nil
}

// FormatScanResult formats the scan result for display
func FormatScanResult(result *ScanResult, dirPath string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Scanned directory: %s\n", dirPath))
	sb.WriteString(fmt.Sprintf("Total files processed: %d\n", result.TotalScanned))

	if len(result.AddedFiles) > 0 {
		sb.WriteString(fmt.Sprintf("\nAdded files (%d):\n", len(result.AddedFiles)))
		for _, file := range result.AddedFiles {
			sb.WriteString(fmt.Sprintf("  ✓ %s\n", file))
		}
	}

	if len(result.SkippedFiles) > 0 {
		sb.WriteString(fmt.Sprintf("\nSkipped files (%d):\n", len(result.SkippedFiles)))
		// Show only first 10 to avoid clutter
		displayCount := len(result.SkippedFiles)
		if displayCount > 10 {
			displayCount = 10
		}
		for i := 0; i < displayCount; i++ {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", result.SkippedFiles[i]))
		}
		if len(result.SkippedFiles) > 10 {
			sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(result.SkippedFiles)-10))
		}
	}

	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\nErrors encountered (%d):\n", len(result.Errors)))
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  ✗ %v\n", err))
		}
	}

	return sb.String()
}
