package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yiblet/hlp/chat"
	"github.com/yiblet/hlp/parse"
)

type chatCmd struct {
	File        string   `arg:"required,positional" help:"the input chat file, if you pass - the command will read from stdin"`
	Write       *string  `arg:"positional" help:"the output chat file, if you pass - the output will be the same as input"`
	MaxTokens   int      `arg:"--tokens,-t" default:"0" help:"the maximum amount of tokens allowed in the output"`
	Temperature *float32 `-arg:"--temp"`
	Color       bool     `default:"false"`
	Model       string   `arg:"--model,-m" help:"set openai model"`
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
func (args *chatCmd) appendChatFile(writer io.Writer, role, content string) error {
	if err := parse.ValidateRole(role); err != nil {
		return err
	}

	buf := bufio.NewWriter(writer)
	if _, err := fmt.Fprintf(buf, "--- %s\n%s\n", role, content); err != nil {
		return err
	}
	return buf.Flush()
}

func (args *chatCmd) writeTo(
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
	if err := args.appendChatFile(output, "assistant", content); err != nil {
		return err
	}

	return output.Flush()
}

func (args *chatCmd) write(
	input string,
	content string,
) error {
	if args.Write == nil {
		return nil
	}

	outfile := *args.Write

	if outfile == "-" {
		if args.File == "-" {
			return fmt.Errorf("cannot output to stdin")
		}
		outfile = args.File
	}

	file, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer file.Close()
	return args.writeTo(input, content, file)
}

func (args *chatCmd) outputWriter() (io.Writer, func() error) {
	var outputWriter io.Writer
	var close func() error
	if !args.Color {
		outputWriter = os.Stdout
		close = func() error { return nil }
	} else {
		outputWriter, close = getOutputWriter()
	}
	return outputWriter, close
}

func (args *chatCmd) readAll(reader io.Reader) error {
	var buf [4096]byte
	for {
		_, err := reader.Read(buf[:])
		if err != nil {
			return err
		}
	}
}

func (args *chatCmd) Execute(ctx context.Context, config *config) error {
	model := args.Model
	if model == "" {
		model = strings.TrimSpace(config.Model())
	}

	var err error
	client := config.Client()

	var file io.ReadCloser
	if args.File != "-" {
		file, err = os.Open(args.File)
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
	if err := args.readAll(reader); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	var outputContent strings.Builder
	outputWriter, closeWriter := args.outputWriter()
	defer closeWriter()

	writer := io.MultiWriter(&outputContent, outputWriter)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()
	// Call ChatCompletionStream with the parsed messages
	err = client.ChatStream(ctx, chat.Input{
		Messages:    messages,
		MaxTokens:   args.MaxTokens,
		Temperature: args.Temperature,
		Model:       model,
	}, func(message string) error {
		fmt.Fprint(writer, message)
		return nil
	})
	if err != nil {
		return err
	}

	return args.write(inputContent.String(), outputContent.String())
}
