# Deep Code (Go)

A powerful AI coding assistant powered by DeepSeek's reasoning models, featuring a beautiful terminal UI built with Bubble Tea.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- 🚀 **Chain-of-Thought Reasoning** - Watch the AI think through problems step-by-step
- 📝 **File Operations** - Read, create, and edit files directly through function calls
- 🎨 **Beautiful TUI** - Built with Charm's Bubble Tea framework
- 📁 **Smart Context Management** - Add files and directories to conversation context
- 🔄 **Streaming Responses** - Real-time streaming of AI responses
- 🛡️ **Security Features** - Path validation and file size limits
- 💰 **Token Usage Tracking** - Real-time cost estimation based on DeepSeek pricing
- 🎯 **Extensible Architecture** - Well-structured codebase for easy modifications

## Installation

### Prerequisites

- Go 1.22 or higher
- A DeepSeek API key (get one at [platform.deepseek.com](https://platform.deepseek.com))

### From Source

```bash
git clone https://github.com/yourusername/deep-code.git
cd deep-code
make build
```

### Using Go Install

```bash
go install github.com/yourusername/deep-code@latest
```

## Configuration

### Environment Variables

Set your DeepSeek API key:

```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

### Configuration File

Create a `config.json` file (optional):

```json
{
  "api": {
    "base_url": "https://api.deepseek.com/v1",
    "model": "deepseek-reasoner",
    "max_completion_tokens": 8192,
    "timeout_seconds": 300
  },
  "ui": {
    "enable_emoji": true,
    "theme": "default",
    "max_history_messages": 15
  },
  "file_operations": {
    "max_file_size": 1048576,
    "allowed_extensions": [".go", ".py", ".js", ".ts", ".json", ".md", ".txt"]
  },
  "scanner": {
    "exclude_patterns": [
      "node_modules",
      ".git",
      "__pycache__",
      "*.pyc",
      ".env"
    ]
  }
}
```

## Usage

### Basic Usage

```bash
./deep-code
```

### Commands

- `/add <path>` - Add a file or directory to the conversation context
- `/clear` - Clear the conversation history
- `/help` - Show help information
- `exit` or `quit` - Exit the application
- `Ctrl+C` - Force quit
- `PgUp/PgDown` - Scroll conversation history

### Example Workflow

1. Start the application:
   ```bash
   ./deep-code
   ```

2. Add files to context:
   ```
   /add main.go
   /add src/
   ```

3. Ask questions or request changes:
   ```
   Can you help me optimize the database queries in this code?
   ```

4. The AI can directly read and modify files:
   ```
   Please add error handling to the fetchUser function
   ```

## Function Capabilities

Deep Code can perform the following file operations:

- **read_file** - Read a single file's content
- **read_multiple_files** - Read multiple files at once
- **create_file** - Create new files or overwrite existing ones
- **create_multiple_files** - Create multiple files in one operation
- **edit_file** - Make precise edits using find-and-replace

## Architecture

```
deep-code/
├── main.go                 # Entry point
├── internal/
│   ├── api/               # DeepSeek API client
│   ├── config/            # Configuration management
│   ├── conversation/      # Conversation history
│   ├── functions/         # File operations
│   └── ui/                # Terminal UI components
├── config.json            # Configuration file
└── Makefile              # Build automation
```

## Performance

- Streaming responses for real-time interaction
- Efficient file scanning with pattern matching
- Concurrent file operations
- Minimal memory footprint
- Real-time token usage and cost tracking

## Security

- Path traversal protection
- File size limits
- Configurable file extension filtering
- No execution of arbitrary commands
- Binary file detection and exclusion

## Feature Parity with Python Version

This Go implementation maintains **complete feature parity** with the original Python version:

### ✅ Core Features
- [x] DeepSeek API integration with streaming
- [x] Chain-of-thought reasoning display
- [x] All 5 file operation functions
- [x] `/add` command for context management
- [x] `/clear` command
- [x] `/help` command
- [x] Conversation history management
- [x] File security and validation

### ✅ UI Features
- [x] Real-time streaming display
- [x] Colored output and formatting
- [x] Emoji support (configurable)
- [x] Scrollable conversation view
- [x] Status indicators (Ready/Seeking/Processing)
- [x] Token and cost tracking

### ✅ Additional Features in Go Version
- [x] **Single binary distribution** - No Python/dependencies needed
- [x] **Better performance** - Compiled language advantages
- [x] **Type safety** - Compile-time error checking
- [x] **Concurrent operations** - Leverages Go's goroutines
- [x] **Modern TUI** - Reactive UI with Bubble Tea

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

## Troubleshooting

### Common Issues

1. **API Key Not Found**
   ```bash
   export DEEPSEEK_API_KEY=your_api_key_here
   ```

2. **Connection Errors**
   - Check your internet connection
   - Verify API endpoint in config.json
   - Ensure your API key is valid

3. **UI Display Issues**
   - Ensure terminal supports Unicode
   - Try resizing terminal window
   - Disable emoji in config if needed

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [DeepSeek](https://deepseek.com) for the amazing reasoning models
- [Charm](https://charm.sh) for the beautiful TUI libraries
- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) for the OpenAI client
- Original [deepseek-engineer](https://github.com/ruvnet/deepseek-engineer) Python implementation

## Related Projects

- [deepseek-engineer](https://github.com/ruvnet/deepseek-engineer) - Original Python implementation
- [continue.dev](https://continue.dev) - VS Code AI assistant
- [aider](https://github.com/paul-gauthier/aider) - AI pair programming tool