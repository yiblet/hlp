package chat

import (
	"context"
)

// Message is a message to use as the context for the chat completion API
type Message struct {
	// Role is the role is the role of the the message. Can be "system", "user", or "assistant"
	Role string `json:"role"`

	// Content is the content of the message
	Content string `json:"content"`
}

type Input struct {
	Messages    []Message
	MaxTokens   int
	Temperature *float32
	Model       string
}

type Streamer interface {
	// ChatCompletion creates a completion with the Chat completion endpoint which
	// is what powers the ChatGPT experience.
	ChatStream(ctx context.Context, request Input, onData func(message string) error) error
}
