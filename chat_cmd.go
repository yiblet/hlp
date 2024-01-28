package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/yiblet/hlp/parse"
)

type chatCmd struct {
	File        string  `arg:"required,positional" help:"the input chat file, if you pass - the command will read from stdin"`
	Write       *string `arg:"positional" help:"the output chat file, if you pass - the output will be the same as input"`
	MaxTokens   int     `default:"0"`
	Temperature float32 `default:"0.7"`
	Model       string  `default:"gpt-3.5-turbo"`
	Color       bool    `default:"false"`
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
func (c *chatCmd) appendChatFile(writer io.Writer, role, content string) error {
	if err := parse.ValidateRole(role); err != nil {
		return err
	}

	buf := bufio.NewWriter(writer)
	if _, err := fmt.Fprintf(buf, "--- %s\n%s\n", role, content); err != nil {
		return err
	}
	return buf.Flush()
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
	if err := c.appendChatFile(output, "assistant", content); err != nil {
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

func (c *chatCmd) readAll(reader io.Reader) error {
	var buf [4096]byte
	for {
		_, err := reader.Read(buf[:])
		if err != nil {
			return err
		}
	}
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
	messages, err := parse.ParseChatFile(reader)
	if err != nil {
		return err
	}
	if err := c.readAll(reader); err != nil && !errors.Is(err, io.EOF) {
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
