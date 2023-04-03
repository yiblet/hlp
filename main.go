package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/alexflint/go-arg"
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
	Model       string   `default:"gpt-3.5-turbo"`
}

func (args *askCmd) Execute(ctx context.Context) error {
	config := authConfig{}

	model := strings.TrimSpace(args.Model)
	if model == "" {
		model = gpt3.TextDavinci003Engine
	}

	client, err := config.NewClient()
	if err != nil {
		return err
	}

	lastMessage := ""
	err = client.ChatCompletionStream(ctx, gpt3.ChatCompletionRequest{
		Messages: []gpt3.ChatCompletionRequestMessage{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: strings.Join(args.Question, " ")},
		},
		MaxTokens:   args.MaxTokens,
		Temperature: &args.Temperature,
		Stream:      true,
		Model:       model,
	}, func(cr *gpt3.ChatCompletionStreamResponse) {
		message := cr.Choices[0].Delta.Content
		if message != "" {
			lastMessage = message
		}
		fmt.Printf("%s", cr.Choices[0].Delta.Content)
	})
	if err != nil {
		return err
	}
	if len(lastMessage) == 0 || lastMessage[len(lastMessage)-1] != '\n' {
		fmt.Printf("\n")
	}
	return nil
}

type authCmd struct {
	OpenAIAPIKey string `arg:"env:OPENAI_API_KEY"`
}

func (c *authCmd) Execute(ctx context.Context) error {
	if c.OpenAIAPIKey == "" {
		fmt.Printf("The openai api key was not passed in via env or command line argument. Enter it in the following line:\n")
		var key string
		_, err := fmt.Scanf("%s\n", &key)
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("invalid openai key")
		}
		c.OpenAIAPIKey = key
	}

	cmd := authConfig{
		OpenAIAPIKey: c.OpenAIAPIKey,
	}

	fmt.Printf("saving openai api key...\n")
	err := cmd.Write()
	if err != nil {
		return err
	}

	fmt.Printf("api key stored in cofig\n")
	return nil
}

type chatCmd struct {
	File        string  `arg:"required,positional" help:"the input chat file, if you pass - the command will read from stdin"`
	Write       *string `arg:"positional" help:"the output chat file, if you pass - the output will be the same as input"`
	MaxTokens   int     `default:"0"`
	Temperature float32 `default:"0.7"`
	Model       string  `default:"gpt-3.5-turbo"`
	Color       bool    `default:"false"`
}

func (c *chatCmd) writeTo(
	input string,
	content string,
	writer io.Writer,
) error {
	// generate output for writing
	output := bufio.NewWriter(writer)
	if _, err := output.WriteString(input); err != nil {
		return err
	}

	output.WriteRune('\n')
	if input[len(input)-1] != '\n' { // add an extra line if needed
		output.WriteRune('\n')
	}
	if err := appendChatFile(output, "assistant", content); err != nil {
		return err
	}

	return output.Flush()
}

func (c *chatCmd) write(
	input string,
	content string,
) error {
	if c.Write == nil {
		return nil
	}

	outfile := *c.Write

	if outfile == "-" {
		if c.File == "-" {
			return fmt.Errorf("cannot output to stdin")
		}
		outfile = c.File
	}

	file, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer file.Close()
	if err != nil {
		return err
	}

	return c.writeTo(input, content, file)
}

func (c *chatCmd) outputWriter() (io.Writer, func() error) {
	var outputWriter io.Writer
	var close func() error
	if !c.Color {
		outputWriter = os.Stdout
		close = func() error { return nil }
	} else {
		outputWriter, close = getOutputWriter()
	}
	return outputWriter, close
}

func (c *chatCmd) Execute(ctx context.Context) error {
	config := authConfig{}
	client, err := config.NewClient()
	if err != nil {
		return err
	}

	var file io.ReadCloser
	if c.File != "-" {
		file, err = os.Open(c.File)
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		file = os.Stdin
	}

	var inputContent strings.Builder
	reader := io.TeeReader(file, &inputContent)

	// Read and parse the file
	messages, err := parseChatFile(reader)
	if err != nil {
		return err
	}
	if err := readAll(reader); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	var outputContent strings.Builder
	outputWriter, closeWriter := c.outputWriter()
	defer closeWriter()

	writer := io.MultiWriter(&outputContent, outputWriter)

	// Call ChatCompletionStream with the parsed messages
	err = client.ChatCompletionStream(ctx, gpt3.ChatCompletionRequest{
		Messages:    messages,
		Stream:      true,
		MaxTokens:   c.MaxTokens,
		Temperature: &c.Temperature,
		Model:       c.Model,
	}, func(cr *gpt3.ChatCompletionStreamResponse) {
		content := cr.Choices[0].Delta.Content
		fmt.Fprint(writer, content)
	})

	if err != nil {
		return err
	}

	return c.write(inputContent.String(), outputContent.String())
}

func readAll(reader io.Reader) error {
	var buf [4096]byte
	for {
		_, err := reader.Read(buf[:])
		if err != nil {
			return err
		}
	}
}

// parseChatFile parses a chat file with the following format:
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
func parseChatFile(file io.Reader) ([]gpt3.ChatCompletionRequestMessage, error) {
	scanner := bufio.NewScanner(file)
	messages := []gpt3.ChatCompletionRequestMessage{}

	var currentRole string
	var currentMessage strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "---") {
			if currentRole != "" && currentMessage.Len() > 0 {
				messages = append(messages, gpt3.ChatCompletionRequestMessage{
					Role:    currentRole,
					Content: currentMessage.String(),
				})
				currentMessage.Reset()
			}
			currentRole = strings.TrimSpace(strings.TrimPrefix(line, "---"))

			if err := validateRole(currentRole); err != nil {
				return nil, err
			}
			continue
		}

		if currentRole != "" {
			currentMessage.WriteString(line)
			currentMessage.WriteString("\n")
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

func validateRole(role string) error {
	if role != "system" && role != "assistant" && role != "user" {
		return fmt.Errorf("invalid role: %s", role)
	}
	return nil
}

// appendChatFile appends a new chat response to the specified file.
// The role parameter must be one of "system", "assistant", or "user".
// The content parameter should contain the text of the chat response.
// The file will be created if it doesn't exist, and the new chat response
// will be appended to the existing content in the file.
//
// The format of the appended chat response will be:
//
// --- role
// content
//
// The function returns an error if the role is invalid or if there's an issue
// while opening or writing to the file.
func appendChatFile(writer io.Writer, role, content string) error {
	if err := validateRole(role); err != nil {
		return err
	}

	buf := bufio.NewWriter(writer)
	if _, err := fmt.Fprintf(buf, "--- %s\n%s\n", role, content); err != nil {
		return err
	}
	return buf.Flush()
}

type mainCmd struct {
	Ask  *askCmd  `arg:"subcommand"`
	Auth *authCmd `arg:"subcommand"`
	Chat *chatCmd `arg:"subcommand"`
}

func main() {
	var args mainCmd
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()
	arg.MustParse(&args)

	var err error
	switch {
	case args.Ask != nil:
		err = args.Ask.Execute(ctx)
	case args.Auth != nil:
		err = args.Auth.Execute(ctx)
	case args.Chat != nil:
		err = args.Chat.Execute(ctx)
	default:
		err = fmt.Errorf("invalid command: run with --help")
	}

	if err != nil {
		log.Panic(err)
	}
}
