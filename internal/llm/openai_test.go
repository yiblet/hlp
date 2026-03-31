package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStream struct {
	t        *testing.T
	response string
}

func (ts *testStream) ChatStream(ctx context.Context, request Input, onData func(message string) error) error {
	chunks := strings.Split(ts.response, " ")
	for i, chunk := range chunks {
		message := chunk
		if i > 0 {
			message = " " + chunk
		}
		ts.t.Logf("Simulating stream chunk: %#v", message)
		if err := onData(message); err != nil {
			return fmt.Errorf("onData callback failed: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return nil
}

func TestStreamer_ChatStream(t *testing.T) {
	t.Parallel()

	t.Run("valid stream", testValidStream)
	t.Run("callback error stops stream", testCallbackErrorStopsStream)
}

func TestCreateParamsIncludesStructuredOutputSchema(t *testing.T) {
	t.Parallel()

	params := createParams(Input{
		Model:    "gpt-4o-2024-08-06",
		Messages: []Message{{Role: "user", Content: "Hello"}},
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"answer": map[string]any{"type": "string"},
			},
			"required":             []string{"answer"},
			"additionalProperties": false,
		},
	})

	body, err := json.Marshal(params)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(body, &got))

	responseFormat, ok := got["response_format"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "json_schema", responseFormat["type"])

	jsonSchema, ok := responseFormat["json_schema"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "hlp_response", jsonSchema["name"])
	assert.Equal(t, true, jsonSchema["strict"])

	schema, ok := jsonSchema["schema"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", schema["type"])
	assert.Equal(t, false, schema["additionalProperties"])
	assert.Contains(t, schema, "properties")
}

func testValidStream(t *testing.T) {
	t.Parallel()
	expectedResponse := "this is a valid stream response"
	streamer := &testStream{t: t, response: expectedResponse}

	var sb strings.Builder
	input := Input{Model: "test-model", Messages: []Message{{Role: "user", Content: "Hello"}}}

	err := streamer.ChatStream(context.Background(), input, func(message string) error {
		t.Logf("Received stream chunk: %#v", message)
		_, err := sb.WriteString(message)
		if err != nil {
			return fmt.Errorf("failed to write to string builder: %w", err)
		}
		return nil
	})

	if err != nil {
		t.Errorf("ChatStream returned an unexpected error: %v", err)
	}

	actualResponse := sb.String()
	if expectedResponse != actualResponse {
		t.Errorf("Unexpected stream result.\nExpected: %#v\nActual:   %#v", expectedResponse, actualResponse)
	}
}

func testCallbackErrorStopsStream(t *testing.T) {
	t.Parallel()
	expectedResponse := "this stream will stop early"
	streamer := &testStream{t: t, response: expectedResponse}
	callbackError := fmt.Errorf("intentional callback error")
	stopWord := "stop"

	var sb strings.Builder
	input := Input{}

	err := streamer.ChatStream(context.Background(), input, func(message string) error {
		t.Logf("Received stream chunk: %#v", message)
		_, writeErr := sb.WriteString(message)
		if writeErr != nil {
			return fmt.Errorf("failed to write to string builder: %w", writeErr)
		}
		if strings.Contains(message, stopWord) {
			return callbackError
		}
		return nil
	})

	if !errors.Is(err, callbackError) {
		t.Errorf("ChatStream did not return the expected callback error. Got: %v", err)
	}

	actualResponse := sb.String()
	expectedPartialResponse := "this stream will stop"
	if !strings.HasPrefix(expectedResponse, actualResponse) {
		t.Errorf("Stream result should be a prefix of the full response when stopped early.\nExpected prefix: %#v\nActual:   %#v", expectedPartialResponse, actualResponse)
	}
	if actualResponse != expectedPartialResponse {
		t.Errorf("Stream result did not match expected partial response.\nExpected: %#v\nActual:   %#v", expectedPartialResponse, actualResponse)
	}
}
