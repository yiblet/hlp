package parse_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yiblet/hlp/chat"
	"github.com/yiblet/hlp/parse"
)

func TestParseChatFile(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    string
		expected []chat.Message
		err      bool
	}{
		{
			name:  "Valid chat",
			input: "--- system\nSystem message\n--- user\nUser message\n",
			expected: []chat.Message{
				{Role: "system", Content: "System message\n"},
				{Role: "user", Content: "User message\n"},
			},
			err: false,
		},
		{
			name:  "Valid chat",
			input: "--- user\nUser message\n",
			expected: []chat.Message{
				{Role: "user", Content: "User message\n"},
			},
			err: false,
		},
		{
			name:     "Empty input",
			input:    "",
			err:      false,
			expected: []chat.Message{},
		},
		{
			name:  "test case",
			err:   false,
			input: "--- user\n\n# Summary of Recent Commits Organized by Project",
			expected: []chat.Message{
				{Role: "user", Content: "\n# Summary of Recent Commits Organized by Project\n"},
			},
		},
		// Add more test cases as needed.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			output, err := parse.ParseChatFile(reader)

			// Use testify assertions
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, output)
			}
		})
	}
}
