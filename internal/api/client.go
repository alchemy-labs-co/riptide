package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/deep-code/deep-code/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

// Client wraps the OpenAI client for DeepSeek API access
type Client struct {
	client *openai.Client
	config *config.Config
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	openaiConfig := openai.DefaultConfig(cfg.APIKey)
	openaiConfig.BaseURL = cfg.API.BaseURL

	// Client created

	return &Client{
		client: openai.NewClientWithConfig(openaiConfig),
		config: cfg,
	}
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *Client) CreateChatCompletionStream(ctx context.Context, messages []openai.ChatCompletionMessage) (<-chan StreamEvent, error) {
	// Removed log to prevent UI interference

	// Create timeout context if configured
	var cancel context.CancelFunc
	if c.config.API.TimeoutSeconds > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.config.API.TimeoutSeconds)*time.Second)
		// Timeout context created
		// Don't defer cancel here - it will be called in the goroutine when streaming completes
	}

	// Create the request
	req := openai.ChatCompletionRequest{
		Model:    c.config.API.Model,
		Messages: messages,
		Tools:    GetTools(),
		Stream:   true,
		// MaxTokens is the standard field (not MaxCompletionTokens)
		MaxTokens: c.config.API.MaxCompletionTokens,
	}

	// Create the stream
	// Creating stream
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		// Stream creation failed
		if cancel != nil {
			cancel()
		}
		return nil, fmt.Errorf("creating chat completion stream: %w", err)
	}
	// Stream created successfully

	// Create event channel
	eventChan := make(chan StreamEvent, 100)

	// Start goroutine to process stream
	go func() {
		// Starting stream processing
		defer close(eventChan)
		defer stream.Close()
		// Cancel the timeout context when done
		if cancel != nil {
			defer cancel()
		}

		var currentContent string
		var toolCalls []ToolCall

		for {
			response, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// Stream finished
					// Stream finished
					eventChan <- StreamEvent{Type: EventTypeDone}
					return
				}
				// Stream error
				// Stream error occurred
				eventChan <- StreamEvent{
					Type:  EventTypeError,
					Error: fmt.Errorf("stream error: %w", err),
				}
				return
			}

			// Process the response
			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				delta := choice.Delta

				// Handle reasoning content if available
				// Note: The standard go-openai library doesn't have ReasoningContent field
				// For now, we'll skip reasoning display until we extend the library
				// TODO: Fork go-openai to add DeepSeek-specific fields

				// Handle regular content
				if delta.Content != "" {
					currentContent += delta.Content
					eventChan <- StreamEvent{
						Type:    EventTypeContent,
						Content: delta.Content,
					}
				}

				// Handle tool calls
				if len(delta.ToolCalls) > 0 {
					// Process tool call deltas
					for _, toolCallDelta := range delta.ToolCalls {
						if toolCallDelta.Index == nil {
							continue
						}

						index := *toolCallDelta.Index
						// Ensure we have enough tool calls
						for len(toolCalls) <= index {
							toolCalls = append(toolCalls, ToolCall{
								Type:     "function",
								Function: FunctionCall{},
							})
						}

						// Update tool call
						if toolCallDelta.ID != "" {
							toolCalls[index].ID = toolCallDelta.ID
						}
						// Function is not a pointer in go-openai, so we check the fields directly
						if toolCallDelta.Function.Name != "" {
							toolCalls[index].Function.Name += toolCallDelta.Function.Name
						}
						if toolCallDelta.Function.Arguments != "" {
							toolCalls[index].Function.Arguments += toolCallDelta.Function.Arguments
						}
					}
				}

				// Check if we have complete tool calls
				if choice.FinishReason == openai.FinishReasonToolCalls && len(toolCalls) > 0 {
					eventChan <- StreamEvent{
						Type:      EventTypeToolCall,
						ToolCalls: toolCalls,
					}
				}
			}

			// Check for usage information (typically sent at the end of stream)
			if response.Usage != nil {
				usage := &TokenUsage{
					InputTokens:  response.Usage.PromptTokens,
					OutputTokens: response.Usage.CompletionTokens,
					// Note: DeepSeek's cached tokens might be in a custom field
					// For now, we'll need to check if the API provides this
					CachedTokens: 0,
				}
				eventChan <- StreamEvent{
					Type:  EventTypeDone,
					Usage: usage,
				}
				return
			}
		}
	}()

	return eventChan, nil
}

// CreateChatCompletion creates a non-streaming chat completion (for follow-ups)
func (c *Client) CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error) {
	// Create timeout context if configured
	if c.config.API.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.config.API.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// Create the request
	req := openai.ChatCompletionRequest{
		Model:     c.config.API.Model,
		Messages:  messages,
		Tools:     GetTools(),
		MaxTokens: c.config.API.MaxCompletionTokens,
	}

	// Make the request
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating chat completion: %w", err)
	}

	return &resp, nil
}
