package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// testStream implements the Streamer interface for testing.
type testStream struct {
	t        *testing.T
	response string
	// We can add fields here later if we need to inspect the Input passed to ChatStream
}

// ChatStream simulates streaming the response string chunk by chunk (split by space).
func (ts *testStream) ChatStream(ctx context.Context, request Input, onData func(message string) error) error {
	// Simulate receiving chunks based on spaces
	chunks := strings.Split(ts.response, " ")
	for i, chunk := range chunks {
		message := chunk
		if i > 0 {
			message = " " + chunk // Add back the space separator for subsequent chunks
		}
		ts.t.Logf("Simulating stream chunk: %#v", message)
		if err := onData(message); err != nil {
			// If the callback returns an error, the stream should stop.
			return fmt.Errorf("onData callback failed: %w", err)
		}
		// Check context cancellation, though not strictly necessary for this simple mock
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

func testValidStream(t *testing.T) {
	t.Parallel()
	expectedResponse := "this is a valid stream response"
	streamer := &testStream{
		t:        t,
		response: expectedResponse,
	}

	var sb strings.Builder
	input := Input{ // Example input, adjust if needed for future tests
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

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
	streamer := &testStream{
		t:        t,
		response: expectedResponse,
	}
	callbackError := fmt.Errorf("intentional callback error")
	stopWord := "stop" // Stop after receiving this word

	var sb strings.Builder
	input := Input{} // Minimal input

	err := streamer.ChatStream(context.Background(), input, func(message string) error {
		t.Logf("Received stream chunk: %#v", message)
		_, writeErr := sb.WriteString(message)
		if writeErr != nil {
			return fmt.Errorf("failed to write to string builder: %w", writeErr)
		}
		// Return an error if the message contains the stop word
		if strings.Contains(message, stopWord) {
			return callbackError
		}
		return nil
	})

	if !errors.Is(err, callbackError) {
		t.Errorf("ChatStream did not return the expected callback error. Got: %v", err)
	}

	actualResponse := sb.String()
	expectedPartialResponse := "this stream will stop" // Should include the chunk that caused the error
	if !strings.HasPrefix(expectedResponse, actualResponse) {
		t.Errorf("Stream result should be a prefix of the full response when stopped early.\nExpected prefix: %#v\nActual:   %#v", expectedPartialResponse, actualResponse)
	}
	if actualResponse != expectedPartialResponse {
		t.Errorf("Stream result did not match expected partial response.\nExpected: %#v\nActual:   %#v", expectedPartialResponse, actualResponse)
	}
}
