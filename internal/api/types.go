package api

import (
	"encoding/json"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// ToolDefinition represents a function tool definition
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition defines a callable function
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a function call request from the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the function name and arguments
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// FileOperationArgs represents arguments for file operations
type FileOperationArgs struct {
	FilePath        string         `json:"file_path,omitempty"`
	FilePaths       []string       `json:"file_paths,omitempty"`
	Content         string         `json:"content,omitempty"`
	OriginalSnippet string         `json:"original_snippet,omitempty"`
	NewSnippet      string         `json:"new_snippet,omitempty"`
	Files           []FileToCreate `json:"files,omitempty"`
}

// FileToCreate represents a file to be created
type FileToCreate struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// StreamEvent represents an event during streaming
type StreamEvent struct {
	Type             EventType
	Content          string
	ReasoningContent string
	ToolCalls        []ToolCall
	Error            error
	Usage            *TokenUsage
}

// TokenUsage represents token usage information
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	CachedTokens int
}

// EventType represents the type of streaming event
type EventType int

const (
	EventTypeContent EventType = iota
	EventTypeReasoning
	EventTypeToolCall
	EventTypeError
	EventTypeDone
)

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	Role             string     `json:"role"`
	Content          string     `json:"content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	Timestamp        time.Time  `json:"timestamp"`
}

// GetTools returns all available tool definitions
func GetTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "Read the content of a single file from the filesystem",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"file_path": {
							"type": "string",
							"description": "The path to the file to read (relative or absolute)"
						}
					},
					"required": ["file_path"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_multiple_files",
				Description: "Read the content of multiple files from the filesystem",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"file_paths": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Array of file paths to read (relative or absolute)"
						}
					},
					"required": ["file_paths"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_file",
				Description: "Create a new file or overwrite an existing file with the provided content",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"file_path": {
							"type": "string",
							"description": "The path where the file should be created"
						},
						"content": {
							"type": "string",
							"description": "The content to write to the file"
						}
					},
					"required": ["file_path", "content"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_multiple_files",
				Description: "Create multiple files at once",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"files": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"path": {"type": "string"},
									"content": {"type": "string"}
								},
								"required": ["path", "content"]
							},
							"description": "Array of files to create with their paths and content"
						}
					},
					"required": ["files"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "edit_file",
				Description: "Edit an existing file by replacing a specific snippet with new content",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"file_path": {
							"type": "string",
							"description": "The path to the file to edit"
						},
						"original_snippet": {
							"type": "string",
							"description": "The exact text snippet to find and replace"
						},
						"new_snippet": {
							"type": "string",
							"description": "The new text to replace the original snippet with"
						}
					},
					"required": ["file_path", "original_snippet", "new_snippet"]
				}`),
			},
		},
	}
}

// GetSystemPrompt returns the system prompt for Riptide
func GetSystemPrompt() string {
	return `You are an elite software engineer called Riptide with decades of experience across all programming domains.
Your expertise spans system design, algorithms, testing, and best practices.
You provide thoughtful, well-structured solutions while explaining your reasoning.

Core capabilities:
1. Code Analysis & Discussion
   - Analyze code with expert-level insight
   - Explain complex concepts clearly
   - Suggest optimizations and best practices
   - Debug issues with precision

2. File Operations (via function calls):
   - read_file: Read a single file's content
   - read_multiple_files: Read multiple files at once
   - create_file: Create or overwrite a single file
   - create_multiple_files: Create multiple files at once
   - edit_file: Make precise edits to existing files using snippet replacement

Guidelines:
1. Provide natural, conversational responses explaining your reasoning
2. Use function calls when you need to read or modify files
3. For file operations:
   - Always read files first before editing them to understand the context
   - Use precise snippet matching for edits
   - Explain what changes you're making and why
   - Consider the impact of changes on the overall codebase
4. Follow language-specific best practices
5. Suggest tests or validation steps when appropriate
6. Be thorough in your analysis and recommendations

IMPORTANT: In your thinking process, if you realize that something requires a tool call, cut your thinking short and proceed directly to the tool call. Don't overthink - act efficiently when file operations are needed.

Remember: You're a senior engineer - be thoughtful, precise, and explain your reasoning clearly.`
}
