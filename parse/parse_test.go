package parse_test

import (
	"strings"
	"testing"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/yiblet/hlp/parse"
)

func TestParseChatFile(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    string
		expected []gpt3.ChatCompletionRequestMessage
		err      bool
	}{
		{
			name:  "Valid chat",
			input: "--- system\nSystem message\n--- user\nUser message\n",
			expected: []gpt3.ChatCompletionRequestMessage{
				{Role: "system", Content: "System message\n"},
				{Role: "user", Content: "User message\n"},
			},
			err: false,
		},
		{
			name:  "Valid chat",
			input: "--- user\nUser message\n",
			expected: []gpt3.ChatCompletionRequestMessage{
				{Role: "user", Content: "User message\n"},
			},
			err: false,
		},
		{
			name:  "Empty input",
			input: "",
			err:   false,
		},
		// Add more test cases as needed.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			output, err := parse.ParseChatFile(reader)
			if (err != nil) != tc.err {
				t.Fatalf("Expected error: %v, got: %v", tc.err, err)
			}

			if !tc.err && !compareMessages(output, tc.expected) {
				t.Fatalf("Expected %v, got %v", tc.expected, output)
			}
		})
	}
}

// Utility function to compare two slices of gpt3.ChatCompletionRequestMessage
func compareMessages(a, b []gpt3.ChatCompletionRequestMessage) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v.Role != b[i].Role || v.Content != b[i].Content {
			return false
		}
	}

	return true
}
