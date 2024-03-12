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

func aiStream(
	ctx context.Context,
	client gpt3.Client,
	input aiStreamInput,
	handler func(message string) error,
) error {
	timeout := 2 * time.Minute
	if input.Timeout != 0 {
		timeout = input.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := client.ChatCompletionStream(ctx, gpt3.ChatCompletionRequest{
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
