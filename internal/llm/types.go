package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Input struct {
	Messages    []Message
	MaxTokens   int
	Temperature *float32
	Model       string
	Schema      any
}

type Streamer interface {
	ChatStream(ctx context.Context, request Input, onData func(message string) error) error
}
