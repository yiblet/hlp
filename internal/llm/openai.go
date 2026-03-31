package llm

import (
	"context"
	"errors"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type OpenAIStreamer struct {
	client        openai.Client
	disableStream bool
}

func (o *OpenAIStreamer) ChatStream(ctx context.Context, request Input, onData func(string) error) error {
	if o.disableStream {
		return o.chatWithoutStream(ctx, request, onData)
	}
	return o.chatWithStream(ctx, request, onData)
}

func (o *OpenAIStreamer) chatWithStream(ctx context.Context, request Input, onData func(message string) error) error {
	params := createParams(request)
	stream := o.client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			content := delta.Content
			if content != "" {
				if err := onData(content); err != nil {
					return err
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}

func (o *OpenAIStreamer) chatWithoutStream(ctx context.Context, request Input, onData func(string) error) error {
	params := createParams(request)
	res, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return err
	}
	return onData(res.Choices[0].Message.Content)
}

func NewOpenAIStreamer(client openai.Client) *OpenAIStreamer {
	return &OpenAIStreamer{client: client}
}

var _ Streamer = (*OpenAIStreamer)(nil)

func createParams(request Input) openai.ChatCompletionNewParams {
	messages := make([]openai.ChatCompletionMessageParamUnion, len(request.Messages))
	for i, msg := range request.Messages {
		switch msg.Role {
		case "user":
			messages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			messages[i] = openai.AssistantMessage(msg.Content)
		case "system":
			messages[i] = openai.SystemMessage(msg.Content)
		default:
			messages[i] = openai.UserMessage(msg.Content)
		}
	}
	params := openai.ChatCompletionNewParams{
		Model:    request.Model,
		Messages: messages,
	}
	if request.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(request.MaxTokens))
	}
	if request.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*request.Temperature))
	}
	if request.Schema != nil {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "hlp_response",
					Schema: request.Schema,
					Strict: param.NewOpt(true),
				},
			},
		}
	}
	return params
}
