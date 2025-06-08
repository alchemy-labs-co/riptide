package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/deep-code/deep-code/internal/config"
	"github.com/deep-code/deep-code/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nPlease ensure:\n")
		fmt.Fprintf(os.Stderr, "1. DEEPSEEK_API_KEY environment variable is set\n")
		fmt.Fprintf(os.Stderr, "2. config.json exists (optional)\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  export DEEPSEEK_API_KEY=your_api_key_here\n")
		fmt.Fprintf(os.Stderr, "  ./deep-code\n")
		os.Exit(1)
	}

	// Create the model
	model, err := ui.NewModel(cfg)
	if err != nil {
		log.Fatal("Error creating model:", err)
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Set up the program reference for streaming
	model.SetProgram(p)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
}

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Deep Code (Go implementation)\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Handle help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Deep Code - AI-powered coding assistant")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  deep-code [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -h, --help     Show this help message")
		fmt.Println("  -v, --version  Show version information")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  DEEPSEEK_API_KEY       Your DeepSeek API key (required)")
		fmt.Println("  DEEPSEEK_CONFIG_PATH   Path to config.json (optional)")
		fmt.Println()
		fmt.Println("Configuration:")
		fmt.Println("  Create a config.json file to customize settings")
		fmt.Println("  See config.json.example for available options")
		os.Exit(0)
	}
}
