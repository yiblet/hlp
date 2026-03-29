package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yiblet/hlp/internal/chatfile"
	configpkg "github.com/yiblet/hlp/internal/config"
	"github.com/yiblet/hlp/internal/llm"
	"github.com/yiblet/hlp/internal/term"
)

type ChatCmd struct {
	File        string   `arg:"required,positional" help:"the input chat file, if you pass - the command will read from stdin"`
	Write       *string  `arg:"positional" help:"the output chat file, if you pass - the output will be the same as input"`
	MaxTokens   int      `arg:"--tokens,-t" default:"0" help:"the maximum amount of tokens allowed in the output"`
	Temperature *float32 `-arg:"--temp"`
	Color       bool     `default:"false"`
	Model       string   `arg:"--model,-m" help:"set openai model"`
}

func (args *ChatCmd) appendChatFile(writer io.Writer, role, content string) error {
	if err := chatfile.ValidateRole(role); err != nil {
		return err
	}

	buf := bufio.NewWriter(writer)
	if _, err := fmt.Fprintf(buf, "--- %s\n%s\n", role, content); err != nil {
		return err
	}
	return buf.Flush()
}

func (args *ChatCmd) writeTo(input string, content string, writer io.Writer) error {
	output := bufio.NewWriter(writer)
	if _, err := output.WriteString(input); err != nil {
		return err
	}

	output.WriteRune('\n')
	if input[len(input)-1] != '\n' {
		output.WriteRune('\n')
	}
	if err := args.appendChatFile(output, "assistant", content); err != nil {
		return err
	}

	return output.Flush()
}

func (args *ChatCmd) write(input string, content string) error {
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

func (args *ChatCmd) outputWriter() (io.Writer, func() error) {
	if !args.Color {
		return os.Stdout, func() error { return nil }
	}
	return term.OutputWriter()
}

func (args *ChatCmd) readAll(reader io.Reader) error {
	var buf [4096]byte
	for {
		_, err := reader.Read(buf[:])
		if err != nil {
			return err
		}
	}
}

func (args *ChatCmd) Execute(ctx context.Context, config *configpkg.Config) error {
	model := args.Model
	if model == "" {
		model = strings.TrimSpace(config.Model())
	}

	client := config.Client()

	var (
		err  error
		file io.ReadCloser
	)
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

	messages, err := chatfile.ParseChatFile(reader)
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
	err = client.ChatStream(ctx, llm.Input{
		Messages:    messages,
		MaxTokens:   args.MaxTokens,
		Temperature: args.Temperature,
		Model:       model,
	}, func(message string) error {
		_, err := fmt.Fprint(writer, message)
		return err
	})
	if err != nil {
		return err
	}

	return args.write(inputContent.String(), outputContent.String())
}
