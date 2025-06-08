package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deep-code/deep-code/internal/api"
	"github.com/deep-code/deep-code/internal/config"
	"github.com/deep-code/deep-code/internal/conversation"
	"github.com/deep-code/deep-code/internal/functions"
)

// Command represents a slash command with its description
type Command struct {
	Name        string
	Description string
	Usage       string
}

// Available slash commands
var availableCommands = []Command{
	{Name: "/add", Description: "Add file or directory to context", Usage: "/add <path>"},
	{Name: "/clear", Description: "Clear conversation history", Usage: "/clear"},
	{Name: "/config", Description: "Configure settings", Usage: "/config"},
	{Name: "/help", Description: "Show help information", Usage: "/help"},
	{Name: "/status", Description: "Show current configuration and pricing", Usage: "/status"},
	{Name: "/quit", Description: "Quit the application", Usage: "/quit"},
}

// State represents the current state of the application
type State int

const (
	StateReady State = iota
	StateProcessing
	StateStreaming
	StateWaitingForInput
	StateError
	StateQuitting
)

// Model represents the Bubble Tea model
type Model struct {
	// Core components
	config    *config.Config
	apiClient *api.Client
	fileOps   *functions.FileOperations
	scanner   *functions.DirectoryScanner
	history   *conversation.History

	// UI components
	viewport  viewport.Model
	textInput textinput.Model
	spinner   spinner.Model

	// State
	state          State
	messages       []Message
	currentContent string

	// Display settings
	width       int
	height      int
	showWelcome bool

	// Autocomplete state
	autocompleteActive        bool
	autocompleteSuggestion    string
	autocompleteCommand       *Command
	autocompleteMatches       []Command
	autocompleteSelectedIndex int

	// Config menu state
	configMenuActive  bool
	configMenuIndex   int
	configOptions     []ConfigOption
	configMenuChanged bool
	originalConfig    config.Config

	// Streaming state
	streamCtx          context.Context
	streamCancel       context.CancelFunc
	isReasoning        bool
	pendingToolCalls   []api.ToolCall
	streamEvents       <-chan api.StreamEvent
	accumulatedContent string
	hasContent         bool

	// Program reference for sending messages
	program *tea.Program
}

// Message represents a message in the conversation
type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
	IsError   bool
}

// StreamMsg is sent when streaming content is received
type StreamMsg struct {
	Event api.StreamEvent
}

// StreamCompleteMsg is sent when streaming is complete
type StreamCompleteMsg struct {
	Error error
}

// ProcessCompleteMsg is sent when processing is complete
type ProcessCompleteMsg struct {
	Result string
	Error  error
}

// NewModel creates a new Bubble Tea model
func NewModel(cfg *config.Config) (*Model, error) {
	// Create API client
	apiClient := api.NewClient(cfg)

	// Create file operations handler
	fileOps := functions.NewFileOperations(cfg)

	// Create directory scanner
	scanner := functions.NewDirectoryScanner(cfg)

	// Create conversation history
	history := conversation.NewHistory(cfg)

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "" // Disable placeholder to prevent double-line issue
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().Foreground(WhiteColor)
	ti.Cursor.Style = lipgloss.NewStyle().Background(WhiteColor).Foreground(lipgloss.Color("#000000"))

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	// Create viewport
	vp := viewport.New(80, 20)

	return &Model{
		config:      cfg,
		apiClient:   apiClient,
		fileOps:     fileOps,
		scanner:     scanner,
		history:     history,
		viewport:    vp,
		textInput:   ti,
		spinner:     s,
		state:       StateReady,
		messages:    make([]Message, 0),
		showWelcome: true,
	}, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		m.textInput.Width = msg.Width - 4
		return m, nil

	case StreamMsg:
		return m.handleStreamEvent(msg.Event)

	case StreamCompleteMsg:
		m.state = StateReady
		if msg.Error != nil {
			m.addErrorMessage(fmt.Sprintf("Stream error: %v", msg.Error))
		}
		m.updateViewport()
		return m, nil

	case ProcessCompleteMsg:
		m.state = StateReady
		if msg.Error != nil {
			m.addErrorMessage(fmt.Sprintf("Process error: %v", msg.Error))
		} else if msg.Result != "" {
			m.addSystemMessage(msg.Result)
		}
		m.updateViewport()
		return m, nil

	case FollowUpMsg:
		return m.handleFollowUp()

	case ExecuteToolsMsg:
		return m.handleExecuteTools(msg.ToolCalls)

	case spinner.TickMsg:
		if m.state == StateProcessing || m.state == StateStreaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Update viewport for non-key messages
	if _, ok := msg.(tea.KeyMsg); !ok {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Show config menu if active
	if m.configMenuActive {
		return m.renderConfigMenu()
	}

	var content strings.Builder

	// Show welcome screen on first run
	if m.showWelcome {
		content.WriteString(m.renderWelcome())
		content.WriteString("\n\n")
	}

	// Render messages
	content.WriteString(m.renderMessages())

	// Update viewport content
	m.viewport.SetContent(content.String())

	// Build the final view
	var view strings.Builder

	// Header (simplified, no status)
	view.WriteString(m.renderHeader())
	view.WriteString("\n\n")

	// Viewport
	view.WriteString(m.viewport.View())
	view.WriteString("\n")

	// Status line
	view.WriteString(m.renderStatusLine())
	view.WriteString("\n")

	// Input with status
	view.WriteString(m.renderInput())

	return view.String()
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle config menu if active
	if m.configMenuActive {
		return m.handleConfigMenuKeyPress(msg)
	}

	// Handle special keys first
	switch msg.Type {
	case tea.KeyCtrlC:
		// Ctrl+C pressed
		if m.streamCancel != nil {
			// Canceling stream
			m.streamCancel()
		}
		m.state = StateQuitting
		return m, tea.Quit

	case tea.KeyCtrlD:
		if m.state == StateReady {
			m.state = StateQuitting
			return m, tea.Quit
		}

	case tea.KeyEnter:
		if m.state == StateReady {
			// If autocomplete is active, fill the command instead of submitting
			if m.autocompleteActive && m.autocompleteSuggestion != "" {
				m.textInput.SetValue(m.autocompleteSuggestion + " ")
				m.textInput.SetCursor(len(m.autocompleteSuggestion) + 1)
				m.updateAutocomplete()
				return m, nil
			}

			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}

			// Check for commands
			if strings.HasPrefix(input, "/") {
				return m.handleCommand(input)
			}

			// Check for exit commands
			if input == "exit" || input == "quit" {
				m.state = StateQuitting
				return m, tea.Quit
			}

			// Process regular input
			m.showWelcome = false
			m.addUserMessage(input)
			m.textInput.SetValue("")

			// Force scroll to bottom for new user messages
			content := m.renderMessages()
			m.viewport.SetContent(content)
			m.viewport.GotoBottom()

			// Start processing
			return m.startConversation(input)
		}

	case tea.KeyPgUp:
		m.viewport.LineUp(5)
		return m, nil

	case tea.KeyPgDown:
		m.viewport.LineDown(5)
		return m, nil

	case tea.KeyTab:
		// Accept autocomplete suggestion
		if m.state == StateReady && m.autocompleteActive && m.autocompleteSuggestion != "" {
			m.textInput.SetValue(m.autocompleteSuggestion + " ")
			m.textInput.SetCursor(len(m.autocompleteSuggestion) + 1)
			m.updateAutocomplete()
			return m, nil
		}

	case tea.KeyUp:
		// Navigate up in autocomplete list
		if m.state == StateReady && m.autocompleteActive && len(m.autocompleteMatches) > 0 {
			m.autocompleteSelectedIndex--
			if m.autocompleteSelectedIndex < 0 {
				m.autocompleteSelectedIndex = len(m.autocompleteMatches) - 1
			}
			m.autocompleteSuggestion = m.autocompleteMatches[m.autocompleteSelectedIndex].Name
			m.autocompleteCommand = &m.autocompleteMatches[m.autocompleteSelectedIndex]
			return m, nil
		}
		// Otherwise, scroll viewport up
		m.viewport.LineUp(1)
		return m, nil

	case tea.KeyDown:
		// Navigate down in autocomplete list
		if m.state == StateReady && m.autocompleteActive && len(m.autocompleteMatches) > 0 {
			m.autocompleteSelectedIndex++
			if m.autocompleteSelectedIndex >= len(m.autocompleteMatches) {
				m.autocompleteSelectedIndex = 0
			}
			m.autocompleteSuggestion = m.autocompleteMatches[m.autocompleteSelectedIndex].Name
			m.autocompleteCommand = &m.autocompleteMatches[m.autocompleteSelectedIndex]
			return m, nil
		}
		// Otherwise, scroll viewport down
		m.viewport.LineDown(1)
		return m, nil

	case tea.KeyEsc:
		// Cancel autocomplete
		if m.state == StateReady && m.autocompleteActive {
			m.autocompleteActive = false
			m.autocompleteSuggestion = ""
			m.autocompleteCommand = nil
			m.autocompleteMatches = nil
			m.autocompleteSelectedIndex = 0
			return m, nil
		}
	}

	// For all other keys, update the text input if we're in ready state
	if m.state == StateReady {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		// Update autocomplete suggestions after text change
		m.updateAutocomplete()
		return m, cmd
	}

	return m, nil
}

// handleCommand handles slash commands
func (m Model) handleCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.SplitN(input, " ", 2)
	command := strings.ToLower(parts[0])

	switch command {
	case "/add":
		if len(parts) < 2 {
			m.addErrorMessage("Usage: /add <file or directory path>")
			m.updateViewport()
			return m, nil
		}
		m.textInput.SetValue("")
		return m.handleAddCommand(parts[1])

	case "/clear":
		m.messages = []Message{}
		m.history.Clear()
		m.showWelcome = true
		m.textInput.SetValue("")
		m.updateViewport()
		return m, nil

	case "/help":
		m.addSystemMessage(m.getHelpText())
		m.textInput.SetValue("")
		m.updateViewport()
		return m, nil

	case "/status":
		m.addSystemMessage(m.getStatusText())
		m.textInput.SetValue("")
		m.updateViewport()
		return m, nil

	case "/config":
		// Enter config menu
		m.configMenuActive = true
		m.state = StateConfigMenu
		m.configMenuIndex = 0
		m.configMenuChanged = false
		m.originalConfig = *m.config // Save original config for comparison
		m.initializeConfigOptions()
		m.textInput.SetValue("")
		return m, nil

	case "/quit":
		m.state = StateQuitting
		return m, tea.Quit

	default:
		m.addErrorMessage(fmt.Sprintf("Unknown command: %s", command))
		m.textInput.SetValue("")
		m.updateViewport()
		return m, nil
	}
}

// startConversation starts a new conversation with the API
func (m Model) startConversation(input string) (tea.Model, tea.Cmd) {
	// Starting conversation
	m.state = StateStreaming
	m.history.AddUserMessage(input)
	m.currentContent = ""
	m.isReasoning = false
	m.pendingToolCalls = nil
	m.accumulatedContent = ""
	m.hasContent = false

	// Add seeking indicator
	m.addSeekingIndicator()

	// Create context for streaming
	ctx, cancel := context.WithCancel(context.Background())
	m.streamCtx = ctx
	m.streamCancel = cancel
	// Stream context created

	// Get messages from history
	messages := m.history.GetMessages()
	// Retrieved messages from history

	// Create stream
	// Creating stream
	eventChan, err := m.apiClient.CreateChatCompletionStream(ctx, messages)
	if err != nil {
		// Failed to create stream
		m.state = StateError
		m.addErrorMessage(fmt.Sprintf("Failed to create stream: %v", err))
		return m, nil
	}
	// Stream created

	// Store the event channel
	m.streamEvents = eventChan

	return m, tea.Batch(
		m.nextStreamMsg(),
		m.spinner.Tick,
	)
}

// nextStreamMsg waits for the next message from the stream
func (m Model) nextStreamMsg() tea.Cmd {
	return func() tea.Msg {
		if m.streamEvents == nil {
			// No stream events
			return StreamCompleteMsg{}
		}

		select {
		case <-m.streamCtx.Done():
			// Context canceled
			return StreamCompleteMsg{Error: m.streamCtx.Err()}
		case event, ok := <-m.streamEvents:
			if !ok {
				// Channel closed, stream complete
				// Stream complete
				return StreamCompleteMsg{}
			}
			// Event received
			return StreamMsg{Event: event}
		}
	}
}

// handleStreamEvent handles streaming events
func (m Model) handleStreamEvent(event api.StreamEvent) (tea.Model, tea.Cmd) {
	switch event.Type {
	case api.EventTypeReasoning:
		if !m.isReasoning {
			m.isReasoning = true
			m.removeSeekingIndicator()
			m.addReasoningLabel()
		}
		m.currentContent += event.ReasoningContent
		m.updateCurrentMessage()

	case api.EventTypeContent:
		if m.isReasoning {
			m.isReasoning = false
			m.finalizeCurrentMessage()
			m.addAssistantLabel()
		} else if !m.hasContent {
			// First content, remove seeking indicator
			m.removeSeekingIndicator()
			m.addAssistantLabel()
		}
		m.currentContent += event.Content
		m.accumulatedContent += event.Content
		m.hasContent = true
		m.updateCurrentMessage()

	case api.EventTypeToolCall:
		m.pendingToolCalls = event.ToolCalls
		if len(m.currentContent) > 0 {
			m.finalizeCurrentMessage()
		}
		// Tool calls will be executed when stream completes

	case api.EventTypeDone:
		m.finalizeCurrentMessage()
		// Store in history
		if m.hasContent || len(m.pendingToolCalls) > 0 {
			m.history.AddAssistantMessage(m.accumulatedContent, m.pendingToolCalls)
		}
		// Update token usage if available
		if event.Usage != nil {
			m.history.UpdateTokenUsage(event.Usage.InputTokens, event.Usage.OutputTokens, event.Usage.CachedTokens)
		}
		// Check if we need to execute tools
		if len(m.pendingToolCalls) > 0 {
			return m, func() tea.Msg {
				return ExecuteToolsMsg{ToolCalls: m.pendingToolCalls}
			}
		}
		return m, func() tea.Msg {
			return StreamCompleteMsg{}
		}

	case api.EventTypeError:
		return m, func() tea.Msg {
			return StreamCompleteMsg{Error: event.Error}
		}
	}

	// Continue reading from stream
	return m, m.nextStreamMsg()
}

// handleExecuteTools executes the tool calls
func (m Model) handleExecuteTools(toolCalls []api.ToolCall) (tea.Model, tea.Cmd) {
	m.state = StateProcessing

	return m, func() tea.Msg {
		// Execute each tool call
		for _, toolCall := range toolCalls {
			// Show function being executed
			functionName := fmt.Sprintf("→ Executing: %s", toolCall.Function.Name)
			if m.program != nil {
				m.program.Send(ProcessCompleteMsg{Result: functionName})
			}

			// Execute the function
			result, err := m.fileOps.ExecuteFunction(toolCall)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Add tool response to history
			m.history.AddToolMessage(toolCall.ID, result)

			// Show result
			if m.program != nil {
				m.program.Send(ProcessCompleteMsg{Result: result})
			}
		}

		// After executing tools, we need a follow-up response
		return FollowUpMsg{}
	}
}

// Message rendering helpers
func (m *Model) addUserMessage(content string) {
	m.messages = append(m.messages, Message{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (m *Model) addAssistantLabel() {
	m.messages = append(m.messages, Message{
		Role:      "assistant-label",
		Content:   "",
		Timestamp: time.Now(),
	})
}

func (m *Model) addReasoningLabel() {
	m.messages = append(m.messages, Message{
		Role:      "reasoning-label",
		Content:   "",
		Timestamp: time.Now(),
	})
}

func (m *Model) addSystemMessage(content string) {
	m.messages = append(m.messages, Message{
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (m *Model) addErrorMessage(content string) {
	m.messages = append(m.messages, Message{
		Role:      "error",
		Content:   content,
		Timestamp: time.Now(),
		IsError:   true,
	})
}

func (m *Model) addSeekingIndicator() {
	m.messages = append(m.messages, Message{
		Role:      "seeking",
		Content:   "",
		Timestamp: time.Now(),
	})
}

func (m *Model) removeSeekingIndicator() {
	// Remove the last seeking message if present
	if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "seeking" {
		m.messages = m.messages[:len(m.messages)-1]
	}
}

func (m *Model) updateCurrentMessage() {
	if len(m.currentContent) == 0 {
		return
	}

	// Determine the message role based on current state
	messageRole := "content"
	if m.isReasoning {
		messageRole = "reasoning"
	}

	// Find the last message with the same role and update it
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == messageRole {
			m.messages[i].Content = m.currentContent
			m.updateViewport()
			return
		}
	}

	// No matching message found, create a new one
	m.messages = append(m.messages, Message{
		Role:      messageRole,
		Content:   m.currentContent,
		Timestamp: time.Now(),
	})
	m.updateViewport()
}

func (m *Model) finalizeCurrentMessage() {
	if len(m.currentContent) > 0 {
		m.updateCurrentMessage()
		m.currentContent = ""
	}
}

func (m *Model) updateViewport() {
	// Only auto-scroll if we're already at or near the bottom
	// This preserves the user's scroll position if they've scrolled up
	atBottom := m.viewport.AtBottom()
	nearBottom := m.viewport.YOffset >= (m.viewport.TotalLineCount() - m.viewport.Height - 5)

	// Update the content
	content := m.renderMessages()
	m.viewport.SetContent(content)

	// Only scroll to bottom if we were already there or near there
	if atBottom || nearBottom {
		m.viewport.GotoBottom()
	}
}

// SetProgram sets the tea.Program reference for streaming
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// updateAutocomplete updates the autocomplete suggestion based on current input
func (m *Model) updateAutocomplete() {
	currentValue := m.textInput.Value()

	// Reset autocomplete if input is empty or doesn't start with /
	if currentValue == "" || !strings.HasPrefix(currentValue, "/") {
		m.autocompleteActive = false
		m.autocompleteSuggestion = ""
		m.autocompleteCommand = nil
		m.autocompleteMatches = nil
		m.autocompleteSelectedIndex = 0
		return
	}

	// Find all matching commands
	lowerInput := strings.ToLower(currentValue)
	m.autocompleteMatches = []Command{}

	for _, cmd := range availableCommands {
		if strings.HasPrefix(cmd.Name, lowerInput) && cmd.Name != lowerInput {
			m.autocompleteMatches = append(m.autocompleteMatches, cmd)
		}
	}

	// Update state based on matches
	if len(m.autocompleteMatches) > 0 {
		m.autocompleteActive = true
		// Ensure selected index is valid
		if m.autocompleteSelectedIndex >= len(m.autocompleteMatches) {
			m.autocompleteSelectedIndex = 0
		}
		m.autocompleteSuggestion = m.autocompleteMatches[m.autocompleteSelectedIndex].Name
		m.autocompleteCommand = &m.autocompleteMatches[m.autocompleteSelectedIndex]
	} else {
		// No matches found
		m.autocompleteActive = false
		m.autocompleteSuggestion = ""
		m.autocompleteCommand = nil
		m.autocompleteSelectedIndex = 0
	}
}
