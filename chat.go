package main

import (
	"context"
	"time"

	"github.com/PullRequestInc/go-gpt3"
)

type aiStreamInput struct {
	Messages    []gpt3.ChatCompletionRequestMessage
	MaxTokens   int
	Temperature *float32
	Model       string
	Timeout     time.Duration
}

type chatCompletionStreamer interface {
	// ChatCompletion creates a completion with the Chat completion endpoint which
	// is what powers the ChatGPT experience.
	ChatCompletionStream(ctx context.Context, request gpt3.ChatCompletionRequest, onData func(*gpt3.ChatCompletionStreamResponse) error) error
}

func aiStream(
	ctx context.Context,
	streamer chatCompletionStreamer,
	input aiStreamInput,
	handler func(message string) error,
) error {
	timeout := 2 * time.Minute
	if input.Timeout != 0 {
		timeout = input.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := streamer.ChatCompletionStream(ctx, gpt3.ChatCompletionRequest{
		Messages:    input.Messages,
		MaxTokens:   input.MaxTokens,
		Temperature: input.Temperature,
		Stream:      true,
		Model:       input.Model,
	}, func(cr *gpt3.ChatCompletionStreamResponse) error {
		message := cr.Choices[0].Delta.Content
		return handler(message)
	})
	return err
}
