package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
)

const systemMessage = `
For the user's following questions, let's think step by step in bash comments to output to make sure
we output the correct bash command with comments. Make sure to ensure your output is always valid bash.

use the following example to understand the desired response style:
Question:
How do I recursively alter all files to the standard chmod permissions in a directory

Answer:
# To recursively alter all files to the standard chmod permissions in a directory, you can use the following command with comments:
# use the chmod command to change the file permissions recursively
chmod -R 644 /path/to/directory/
# -R option stands for recursive, which will apply the permissions to all files and subdirectories within the directory
# 644 is the standard permission for files, which means the owner has read and write access, and others have only read access
`

type askCmd struct {
	Question    []string `arg:"positional"`
	MaxTokens   int      `default:"0"`
	Temperature float32  `default:"0.7"`
	Bash        bool     `arg:"--bash" help:"output only valid bash"`
}

func (args *askCmd) messages() []gpt3.ChatCompletionRequestMessage {
	if args.Bash {
		return []gpt3.ChatCompletionRequestMessage{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: strings.Join(args.Question, " ")},
		}
	} else {
		return []gpt3.ChatCompletionRequestMessage{
			{Role: "system", Content: strings.Join(args.Question, " ")},
		}
	}

}

func (args *askCmd) Execute(ctx context.Context, config *config) error {
	model := strings.TrimSpace(config.Model())
	client := config.Client()

	lastMessage := ""
	err := client.ChatCompletionStream(ctx, gpt3.ChatCompletionRequest{
		Messages:    args.messages(),
		MaxTokens:   args.MaxTokens,
		Temperature: &args.Temperature,
		Stream:      true,
		Model:       model,
	}, func(cr *gpt3.ChatCompletionStreamResponse) error {
		message := cr.Choices[0].Delta.Content
		if message != "" {
			lastMessage = message
		}
		fmt.Printf("%s", cr.Choices[0].Delta.Content)
		return nil
	})
	if err != nil {
		return err
	}
	if len(lastMessage) == 0 || lastMessage[len(lastMessage)-1] != '\n' {
		fmt.Printf("\n")
	}
	return nil
}
