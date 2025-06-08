package conversation

import (
	"fmt"
	"sync"
	"time"

	"github.com/alchemy-labs-co/riptide/internal/api"
	"github.com/alchemy-labs-co/riptide/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

// History manages the conversation history
type History struct {
	mu                  sync.RWMutex
	messages            []api.ConversationMessage
	config              *config.Config
	inputTokens         int
	outputTokens        int
	cachedTokens        int
	offPeakInputTokens  int
	offPeakOutputTokens int
	offPeakCachedTokens int
}

// NewHistory creates a new conversation history
func NewHistory(cfg *config.Config) *History {
	h := &History{
		messages: make([]api.ConversationMessage, 0),
		config:   cfg,
	}

	// Add system prompt
	h.AddSystemMessage(api.GetSystemPrompt())

	return h
}

// AddMessage adds a message to the conversation history
func (h *History) AddMessage(role, content string, toolCalls []api.ToolCall, toolCallID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	msg := api.ConversationMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	if toolCallID != "" {
		msg.ToolCallID = toolCallID
	}

	h.messages = append(h.messages, msg)
}

// AddUserMessage adds a user message to the history
func (h *History) AddUserMessage(content string) {
	h.AddMessage("user", content, nil, "")
}

// AddAssistantMessage adds an assistant message to the history
func (h *History) AddAssistantMessage(content string, toolCalls []api.ToolCall) {
	h.AddMessage("assistant", content, toolCalls, "")
}

// AddToolMessage adds a tool response message to the history
func (h *History) AddToolMessage(toolCallID, content string) {
	h.AddMessage("tool", content, nil, toolCallID)
}

// AddSystemMessage adds a system message (e.g., file content) to the history
func (h *History) AddSystemMessage(content string) {
	h.AddMessage("system", content, nil, "")
}

// GetMessages returns OpenAI-formatted messages for API calls
func (h *History) GetMessages() []openai.ChatCompletionMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Convert internal messages to OpenAI format
	result := make([]openai.ChatCompletionMessage, 0, len(h.messages))

	for _, msg := range h.messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Add tool calls if present
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				openaiMsg.ToolCalls = append(openaiMsg.ToolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
		}

		// Add tool call ID if present
		if msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		result = append(result, openaiMsg)
	}

	return result
}

// GetRawMessages returns the raw conversation messages
func (h *History) GetRawMessages() []api.ConversationMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to prevent external modification
	messages := make([]api.ConversationMessage, len(h.messages))
	copy(messages, h.messages)
	return messages
}

// Trim trims the conversation history to prevent token overflow
func (h *History) Trim() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Don't trim if conversation is still small
	if len(h.messages) <= 20 {
		return
	}

	// Separate system messages and other messages
	var systemMessages []api.ConversationMessage
	var otherMessages []api.ConversationMessage

	for _, msg := range h.messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	// Keep only the configured number of recent messages
	maxMessages := h.config.UI.MaxHistoryMessages
	if len(otherMessages) > maxMessages {
		otherMessages = otherMessages[len(otherMessages)-maxMessages:]
	}

	// Rebuild conversation history
	h.messages = make([]api.ConversationMessage, 0, len(systemMessages)+len(otherMessages))
	h.messages = append(h.messages, systemMessages...)
	h.messages = append(h.messages, otherMessages...)
}

// Clear clears the conversation history (except system prompt)
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Keep only the system prompt
	systemPrompt := h.messages[0]
	h.messages = []api.ConversationMessage{systemPrompt}

	// Reset token counters
	h.inputTokens = 0
	h.outputTokens = 0
	h.cachedTokens = 0
	h.offPeakInputTokens = 0
	h.offPeakOutputTokens = 0
	h.offPeakCachedTokens = 0
}

// FileAlreadyInContext checks if a file is already in the conversation context
func (h *History) FileAlreadyInContext(filePath string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	fileMarker := fmt.Sprintf("Content of file '%s'", filePath)
	for _, msg := range h.messages {
		if msg.Role == "system" &&
			len(msg.Content) > 0 &&
			contains(msg.Content, fileMarker) {
			return true
		}
	}
	return false
}

// GetConversationLength returns the number of messages in the history
func (h *History) GetConversationLength() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.messages)
}

// GetLastUserMessage returns the last user message if any
func (h *History) GetLastUserMessage() (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := len(h.messages) - 1; i >= 0; i-- {
		if h.messages[i].Role == "user" {
			return h.messages[i].Content, true
		}
	}
	return "", false
}

// GetLastAssistantMessage returns the last assistant message if any
func (h *History) GetLastAssistantMessage() (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := len(h.messages) - 1; i >= 0; i-- {
		if h.messages[i].Role == "assistant" {
			return h.messages[i].Content, true
		}
	}
	return "", false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

// findSubstring finds a substring in a string
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ConversationStats provides statistics about the conversation
type ConversationStats struct {
	TotalMessages       int
	UserMessages        int
	AssistantMessages   int
	SystemMessages      int
	ToolMessages        int
	InputTokens         int
	OutputTokens        int
	CachedTokens        int
	OffPeakInputTokens  int
	OffPeakOutputTokens int
	OffPeakCachedTokens int
	StartTime           time.Time
}

// GetStats returns statistics about the conversation
func (h *History) GetStats() ConversationStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := ConversationStats{
		TotalMessages:       len(h.messages),
		InputTokens:         h.inputTokens,
		OutputTokens:        h.outputTokens,
		CachedTokens:        h.cachedTokens,
		OffPeakInputTokens:  h.offPeakInputTokens,
		OffPeakOutputTokens: h.offPeakOutputTokens,
		OffPeakCachedTokens: h.offPeakCachedTokens,
		StartTime:           time.Now(), // This would need to be tracked separately
	}

	for _, msg := range h.messages {
		switch msg.Role {
		case "user":
			stats.UserMessages++
		case "assistant":
			stats.AssistantMessages++
		case "system":
			stats.SystemMessages++
		case "tool":
			stats.ToolMessages++
		}
	}

	return stats
}

// UpdateTokenUsage updates the token usage counters
func (h *History) UpdateTokenUsage(inputTokens, outputTokens, cachedTokens int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if current time is in off-peak hours (16:30-00:30 UTC)
	now := time.Now().UTC()
	hour := now.Hour()
	minute := now.Minute()
	isOffPeak := (hour == 16 && minute >= 30) || (hour > 16) || (hour == 0 && minute <= 30)

	// Update total counters
	h.inputTokens += inputTokens
	h.outputTokens += outputTokens
	h.cachedTokens += cachedTokens

	// Update off-peak counters if applicable
	if isOffPeak {
		h.offPeakInputTokens += inputTokens
		h.offPeakOutputTokens += outputTokens
		h.offPeakCachedTokens += cachedTokens
	}
}
