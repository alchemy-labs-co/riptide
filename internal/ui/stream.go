package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/alchemy-labs-co/riptide/internal/api"
)

// StreamManager handles the streaming integration between API and UI
type StreamManager struct {
	program *tea.Program
}

// NewStreamManager creates a new stream manager
func NewStreamManager(program *tea.Program) *StreamManager {
	return &StreamManager{
		program: program,
	}
}

// This file contains helper functions for streaming integration

// FollowUpMsg indicates we need a follow-up response after tool execution
type FollowUpMsg struct{}

// ExecuteToolsMsg contains tool calls to execute
type ExecuteToolsMsg struct {
	ToolCalls []api.ToolCall
}

// handleFollowUp processes the follow-up after tool execution
func (m Model) handleFollowUp() (tea.Model, tea.Cmd) {
	m.state = StateStreaming
	m.currentContent = ""
	m.isReasoning = false
	m.accumulatedContent = ""
	m.hasContent = false
	m.pendingToolCalls = nil

	// Create new context
	ctx, cancel := context.WithCancel(context.Background())
	m.streamCtx = ctx
	m.streamCancel = cancel

	// Add a system message indicating we're processing results
	m.addSystemMessage("Processing results...")

	// Get messages from history (includes tool responses)
	messages := m.history.GetMessages()

	// Create stream for follow-up
	eventChan, err := m.apiClient.CreateChatCompletionStream(ctx, messages)
	if err != nil {
		m.state = StateError
		m.addErrorMessage(fmt.Sprintf("Failed to create follow-up stream: %v", err))
		return m, nil
	}

	// Store the event channel
	m.streamEvents = eventChan

	return m, tea.Batch(
		m.nextStreamMsg(),
		m.spinner.Tick,
	)
}

// Attach attaches the stream manager to a model
func (sm *StreamManager) Attach(model *Model) {
	// This allows the model to send messages back to the UI
	model.program = sm.program
}

// Helper to properly format tool execution messages
func formatToolExecution(functionName string, enableEmoji bool) string {
	icon := GetIcon("lightning", enableEmoji)
	return fmt.Sprintf("%s Executing function: %s", icon, functionName)
}

// Helper to format tool results
func formatToolResult(success bool, message string, enableEmoji bool) string {
	if success {
		return FormatSuccess(message, enableEmoji)
	}
	return FormatError(message, enableEmoji)
}
