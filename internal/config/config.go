package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	API            APIConfig            `json:"api"`
	UI             UIConfig             `json:"ui"`
	FileOperations FileOperationsConfig `json:"file_operations"`
	APIKey         string               `json:"-"` // Not stored in JSON, loaded from env
}

// APIConfig contains API-related settings
type APIConfig struct {
	BaseURL             string `json:"base_url"`
	Model               string `json:"model"`
	MaxCompletionTokens int    `json:"max_completion_tokens"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
}

// UIConfig contains UI-related settings
type UIConfig struct {
	Theme              string `json:"theme"`
	EnableEmoji        bool   `json:"enable_emoji"`
	MaxHistoryMessages int    `json:"max_history_messages"`
}

// FileOperationsConfig contains file operation settings
type FileOperationsConfig struct {
	MaxFileSizeMB   int `json:"max_file_size_mb"`
	MaxFilesPerScan int `json:"max_files_per_scan"`
	BinaryPeekSize  int `json:"binary_peek_size"`
}

// Load loads configuration from config.json and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get config file path
	configPath := os.Getenv("DEEPSEEK_CONFIG_PATH")
	if configPath == "" {
		// Default to config.json in current directory
		configPath = "config.json"
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If config doesn't exist, use defaults
		if os.IsNotExist(err) {
			return loadDefaults()
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Load API key from environment
	cfg.APIKey = os.Getenv("DEEPSEEK_API_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable not set")
	}

	return &cfg, nil
}

// loadDefaults returns a configuration with default values
func loadDefaults() (*Config, error) {
	cfg := &Config{
		API: APIConfig{
			BaseURL:             "https://api.deepseek.com/v1",
			Model:               "deepseek-reasoner",
			MaxCompletionTokens: 64000,
			TimeoutSeconds:      300,
		},
		UI: UIConfig{
			Theme:              "default",
			EnableEmoji:        true,
			MaxHistoryMessages: 15,
		},
		FileOperations: FileOperationsConfig{
			MaxFileSizeMB:   5,
			MaxFilesPerScan: 1000,
			BinaryPeekSize:  1024,
		},
	}

	// Load API key from environment
	cfg.APIKey = os.Getenv("DEEPSEEK_API_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable not set")
	}

	return cfg, nil
}

// GetExcludedFiles returns the list of files to exclude from scanning
func GetExcludedFiles() map[string]bool {
	return map[string]bool{
		// Python specific
		".DS_Store": true, "Thumbs.db": true, ".gitignore": true, ".python-version": true,
		"uv.lock": true, ".uv": true, "uvenv": true, ".uvenv": true, ".venv": true, "venv": true,
		"__pycache__": true, ".pytest_cache": true, ".coverage": true, ".mypy_cache": true,
		// Node.js / Web specific
		"node_modules": true, "package-lock.json": true, "yarn.lock": true, "pnpm-lock.yaml": true,
		".next": true, ".nuxt": true, "dist": true, "build": true, ".cache": true, ".parcel-cache": true,
		".turbo": true, ".vercel": true, ".output": true, ".contentlayer": true,
		// Build outputs
		"out": true, "coverage": true, ".nyc_output": true, "storybook-static": true,
		// Environment and config
		".env": true, ".env.local": true, ".env.development": true, ".env.production": true,
		// Misc
		".git": true, ".svn": true, ".hg": true, "CVS": true,
		// Go specific
		"vendor": true, "go.sum": true,
	}
}

// GetExcludedExtensions returns the list of file extensions to exclude
func GetExcludedExtensions() map[string]bool {
	return map[string]bool{
		// Binary and media files
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true, ".svg": true, ".webp": true, ".avif": true,
		".mp4": true, ".webm": true, ".mov": true, ".mp3": true, ".wav": true, ".ogg": true,
		".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true,
		// Documents
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		// Python specific
		".pyc": true, ".pyo": true, ".pyd": true, ".egg": true, ".whl": true,
		// UV specific
		".uv": true, ".uvenv": true,
		// Database and logs
		".db": true, ".sqlite": true, ".sqlite3": true, ".log": true,
		// IDE specific
		".idea": true, ".vscode": true,
		// Web specific
		".map": true, ".chunk.js": true, ".chunk.css": true,
		".min.js": true, ".min.css": true, ".bundle.js": true, ".bundle.css": true,
		// Cache and temp files
		".cache": true, ".tmp": true, ".temp": true,
		// Font files
		".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	}
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
