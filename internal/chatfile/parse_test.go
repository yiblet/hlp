package chatfile_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yiblet/hlp/internal/chatfile"
	"github.com/yiblet/hlp/internal/llm"
)

func TestParseChatFile(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []llm.Message
		err      bool
	}{
		{
			name:     "Valid chat",
			input:    "--- system\nSystem message\n--- user\nUser message\n",
			expected: []llm.Message{{Role: "system", Content: "System message\n"}, {Role: "user", Content: "User message\n"}},
			err:      false,
		},
		{
			name:     "Valid chat",
			input:    "--- user\nUser message\n",
			expected: []llm.Message{{Role: "user", Content: "User message\n"}},
			err:      false,
		},
		{
			name:     "Empty input",
			input:    "",
			err:      false,
			expected: []llm.Message{},
		},
		{
			name:     "test case",
			err:      false,
			input:    "--- user\n\n# Summary of Recent Commits Organized by Project",
			expected: []llm.Message{{Role: "user", Content: "\n# Summary of Recent Commits Organized by Project\n"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			output, err := chatfile.ParseChatFile(reader)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, output)
			}
		})
	}
}
