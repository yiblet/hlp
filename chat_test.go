package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	gpt3 "github.com/PullRequestInc/go-gpt3"
)

type testStream struct {
	t        *testing.T
	response string
	request  gpt3.ChatCompletionRequest
}

// ChatCompletionStream implements streamClient.
func (t *testStream) ChatCompletionStream(ctx context.Context, request gpt3.ChatCompletionRequest, onData func(*gpt3.ChatCompletionStreamResponse) error) error {
	convert := func(s string) *gpt3.ChatCompletionStreamResponse {
		return &gpt3.ChatCompletionStreamResponse{
			Choices: []gpt3.ChatCompletionStreamResponseChoice{
				{
					Delta: gpt3.ChatCompletionResponseMessage{
						Content: s,
					},
				},
			},
		}
	}

	t.request = request
	for i, elem := range strings.Split(t.response, " ") {
		t.t.Logf("inserting: %#v", elem)
		elem := elem
		if i == 0 {
			if err := onData(convert(elem)); err != nil {
				return err
			}
		} else {
			if err := onData(convert(fmt.Sprintf(" %s", elem))); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAIStream(t *testing.T) {
	t.Parallel()

	t.Run("valid stream", func(t *testing.T) {
		stream := &testStream{response: "valid stream", t: t}
		var sb strings.Builder
		err := aiStream(context.Background(), stream, aiStreamInput{}, func(message string) error {
			t.Logf("retrieved: %#v", message)
			_, err := sb.WriteString(message)
			return err
		})

		if err != nil {
			t.Errorf("stream should not error: %v", err)
		}

		expected := stream.response
		test := sb.String()
		if expected != test {
			t.Errorf("invalid stream response. Expected %#v got %#v", expected, test)
		}
	})

	t.Run("valid stream", func(t *testing.T) {
		stream := &testStream{response: "valid stream", t: t}
		var sb strings.Builder
		aiStream(context.Background(), stream, aiStreamInput{}, func(message string) error {
			t.Logf("retrieved: %#v", message)
			_, err := sb.WriteString(message)
			return err
		})

		expected := stream.response
		test := sb.String()
		if expected != test {
			t.Errorf("invalid stream response. Expected %#v got %#v", expected, test)
		}
	})

}
