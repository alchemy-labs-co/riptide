package functions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath returns a canonical, absolute version of the path with security checks
func NormalizePath(pathStr string) (string, error) {
	// Prevent empty paths
	if pathStr == "" {
		return "", fmt.Errorf("empty path provided")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(pathStr)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}

	// Clean the path
	cleanPath := filepath.Clean(absPath)

	// Check for directory traversal attempts
	if strings.Contains(pathStr, "..") {
		// Allow paths outside cwd but prevent obvious traversal attacks
		if strings.Contains(filepath.Clean(pathStr), "..") {
			return "", fmt.Errorf("invalid path: contains parent directory references")
		}
	}

	// Check for home directory references starting with ~
	if strings.HasPrefix(pathStr, "~") {
		return "", fmt.Errorf("home directory references (~) are not allowed")
	}

	return cleanPath, nil
}

// IsBinaryFile checks if a file is likely binary by peeking at its content
func IsBinaryFile(filePath string, peekSize int) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return true, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Read a chunk of the file
	buffer := make([]byte, peekSize)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return true, fmt.Errorf("reading file: %w", err)
	}

	// Check for null bytes in the sample
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true, nil
		}
	}

	// Additional checks for common binary signatures
	if n >= 4 {
		// Check for common binary file headers
		signature := buffer[:4]

		// ELF (Linux executables)
		if signature[0] == 0x7f && signature[1] == 'E' && signature[2] == 'L' && signature[3] == 'F' {
			return true, nil
		}

		// PE (Windows executables)
		if signature[0] == 'M' && signature[1] == 'Z' {
			return true, nil
		}

		// Mach-O (macOS executables)
		if (signature[0] == 0xfe && signature[1] == 0xed && signature[2] == 0xfa &&
			(signature[3] == 0xce || signature[3] == 0xcf)) ||
			(signature[0] == 0xce && signature[1] == 0xfa && signature[2] == 0xed && signature[3] == 0xfe) {
			return true, nil
		}

		// ZIP files
		if signature[0] == 'P' && signature[1] == 'K' && (signature[2] == 0x03 || signature[2] == 0x05) {
			return true, nil
		}
	}

	// Check for high proportion of non-printable characters
	nonPrintable := 0
	for i := 0; i < n; i++ {
		b := buffer[i]
		// Allow common whitespace characters
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintable++
		}
		if b > 126 && b < 160 {
			nonPrintable++
		}
	}

	// If more than 30% non-printable, consider it binary
	if float64(nonPrintable)/float64(n) > 0.3 {
		return true, nil
	}

	return false, nil
}

// ValidateFileSize checks if file size is within limits
func ValidateFileSize(filePath string, maxSizeMB int) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}

	maxSize := int64(maxSizeMB * 1024 * 1024)
	if fileInfo.Size() > maxSize {
		return fmt.Errorf("file size %d bytes exceeds %dMB limit", fileInfo.Size(), maxSizeMB)
	}

	return nil
}

// IsHiddenFile checks if a file or directory is hidden
func IsHiddenFile(name string) bool {
	return strings.HasPrefix(filepath.Base(name), ".")
}
