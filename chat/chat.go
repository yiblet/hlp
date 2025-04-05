package chat

import (
	"context"
	"errors"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type OpenAIStreamer struct {
	client openai.Client
}

// ChatStream implements Streamer.
func (o *OpenAIStreamer) ChatStream(ctx context.Context, request Input, onData func(message string) error) error {
	// Map Input messages to OpenAI message parameters
	messages := make([]openai.ChatCompletionMessageParamUnion, len(request.Messages))
	for i, msg := range request.Messages {
		switch msg.Role {
		case "user":
			messages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			messages[i] = openai.AssistantMessage(msg.Content)
		case "system":
			messages[i] = openai.SystemMessage(msg.Content)
		// Add other roles if needed (e.g., tool, function)
		default:
			// Potentially handle unknown roles or return an error
			messages[i] = openai.UserMessage(msg.Content) // Default to user for safety
		}
	}

	// Prepare the OpenAI request parameters
	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(request.Model),
		Messages: messages,
	}
	if request.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(request.MaxTokens))
	}
	if request.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*request.Temperature))
	}

	// Create the stream
	stream := o.client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close() // Ensure stream is closed

	// Process the stream
	for stream.Next() {
		chunk := stream.Current()

		// Check if choices are available and extract content
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			content := delta.Content

			if content != "" {
				// Call the onData callback with the content delta
				if err := onData(content); err != nil {
					// Handle callback error (e.g., stop streaming)
					// Close the stream explicitly here as we are exiting the function early
					return err
				}
			}
		}
	}

	// Check for errors after the stream is finished
	if err := stream.Err(); err != nil {
		// Check if the error is EOF, which is expected at the end of a stream
		if errors.Is(err, io.EOF) {
			return nil // Normal end of stream
		}
		return err // Return other stream errors
	}

	return nil // No errors encountered
}

func NewOpenAIStreamer(client openai.Client) *OpenAIStreamer {
	return &OpenAIStreamer{client: client}
}

// ensure that OpenAIStreamer implements the Streamer interface
var _ Streamer = (*OpenAIStreamer)(nil)
