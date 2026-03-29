package chatfile

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/yiblet/hlp/internal/llm"
)

var boundaryRegexp = regexp.MustCompile(`^---\s*(user|system|assistant)\s*$`)

func ParseChatFile(file io.Reader) ([]llm.Message, error) {
	scanner := bufio.NewScanner(file)
	messages := []llm.Message{}

	var currentRole string
	var currentMessage strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if matches := boundaryRegexp.FindStringSubmatch(strings.ToLower(line)); matches != nil {
			if currentRole != "" && currentMessage.Len() > 0 {
				messages = append(messages, llm.Message{Role: currentRole, Content: currentMessage.String()})
				currentMessage.Reset()
			}

			currentRole = matches[1]
			if err := ValidateRole(currentRole); err != nil {
				return nil, err
			}
			continue
		}

		if currentRole == "" && strings.TrimSpace(line) != "" {
			currentRole = "system"
		}
		if currentRole != "" {
			fmt.Fprintf(&currentMessage, "%s\n", line)
		}
	}

	if currentRole != "" && currentMessage.Len() > 0 {
		messages = append(messages, llm.Message{Role: currentRole, Content: currentMessage.String()})
	}

	return messages, nil
}

func ValidateRole(role string) error {
	if role != "system" && role != "assistant" && role != "user" {
		return &InvalidRoleError{role}
	}
	return nil
}
