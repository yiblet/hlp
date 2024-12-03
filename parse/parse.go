package parse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
)

var boundaryRegexp = regexp.MustCompile(`^---\s*(user|system|assistant)\s*$`)

// ParseChatFile parses a chat file with the following format:
//
// --- Role1
// Content for Role1
// (additional lines of content for Role1 if necessary)
// --- Role2
// Content for Role2
// (additional lines of content for Role2 if necessary)
// ---
//
// The file should have alternating roles and content, separated by a line containing "---".
// Each role and its corresponding content must be separated by a newline.
//
// valid roles are "system", "assistant", and "user". System can only appear
// as the first role in the chat log.
func ParseChatFile(file io.Reader) ([]gpt3.ChatCompletionRequestMessage, error) {
	scanner := bufio.NewScanner(file)
	messages := []gpt3.ChatCompletionRequestMessage{}

	var currentRole string
	var currentMessage strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if matches := boundaryRegexp.FindStringSubmatch(strings.ToLower(line)); matches != nil {
			if currentRole != "" && currentMessage.Len() > 0 {
				messages = append(messages, gpt3.ChatCompletionRequestMessage{
					Role:    currentRole,
					Content: currentMessage.String(),
				})
				currentMessage.Reset()
			}

			currentRole = matches[1]
			if err := ValidateRole(currentRole); err != nil {
				return nil, err
			}
			continue
		}

		// if there is no role, but there is some sort of content assume
		// that it's the user talking.
		if currentRole == "" && strings.TrimSpace(line) != "" {
			currentRole = "system"
		}
		if currentRole != "" {
			fmt.Fprintf(&currentMessage, "%s\n", line)
		}
	}

	if currentRole != "" && currentMessage.Len() > 0 {
		messages = append(messages, gpt3.ChatCompletionRequestMessage{
			Role:    currentRole,
			Content: currentMessage.String(),
		})
	}

	return messages, nil
}

func ValidateRole(role string) error {
	if role != "system" && role != "assistant" && role != "user" {
		return &InvalidRoleError{role}
	}
	return nil
}
