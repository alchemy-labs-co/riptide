# Deep Code Go Implementation
BINARY_NAME=deep-code
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Directories
CMD_DIR=.
INTERNAL_DIR=internal
DIST_DIR=dist

.PHONY: all build clean test coverage fmt vet lint run install help

# Default target
all: clean fmt vet test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: ./$(BINARY_NAME)"

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	
	@echo "All builds complete in $(DIST_DIR)/"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(DIST_DIR)
	@$(GOCMD) clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install from https://golangci-lint.run/"; exit 1; }
	@golangci-lint run

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete"

# Create example config
config-example:
	@echo "Creating example configuration..."
	@cp config.json config.json.example
	@echo "Created config.json.example"

# Development setup
setup: deps
	@echo "Setting up development environment..."
	@cp .env.example .env 2>/dev/null || echo "No .env.example found"
	@echo "Setup complete. Don't forget to:"
	@echo "  1. Set DEEPSEEK_API_KEY in .env"
	@echo "  2. Adjust config.json as needed"

# Help
help:
	@echo "Deep Code - Go Implementation"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build the binary"
	@echo "  make build-all   - Build for multiple platforms"
	@echo "  make run         - Build and run the application"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make coverage    - Run tests with coverage report"
	@echo "  make fmt         - Format code"
	@echo "  make vet         - Run go vet"
	@echo "  make lint        - Run linter (requires golangci-lint)"
	@echo "  make deps        - Download dependencies"
	@echo "  make install     - Install binary to GOPATH/bin"
	@echo "  make setup       - Setup development environment"
	@echo "  make help        - Show this help message"