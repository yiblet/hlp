package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

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
	MaxTokens   int      `arg:"--tokens,-t" default:"0" help:"the maximum amount of tokens allowed in the output"`
	Temperature *float32 `arg:"--temp"`
	Bash        bool     `arg:"--bash" help:"output only valid bash"`
	Model       string   `arg:"--model,-m" help:"set openai model"`
	Attach      []string `arg:"--attach,-a,separate" help:"attach additional files at the end of the message. pass '-' to pass in stdin"`
	Once        bool     `arg:"--once,-o" help:"whether to just ask the model once"`
}

func (args *askCmd) buildContent(ctx context.Context) (string, error) {
	var sb strings.Builder
	for idx, q := range args.Question {
		if idx != 0 {
			sb.WriteRune(' ')
		}
		sb.WriteString(q)
	}

	if len(args.Question) > 0 &&
		!strings.HasSuffix(args.Question[len(args.Question)-1], "\n") {
		sb.WriteRune('\n')
	}

	for _, a := range args.Attach {
		sb.WriteRune('\n')

		var file *os.File
		if a == "-" {
			file = os.Stdin
		} else {
			var err error
			file, err = os.Open(a)
			if err != nil {
				return "", err
			}
		}
		defer file.Close()

		_, err := io.Copy(&sb, file)
		if err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

func (args *askCmd) messages(content string) []gpt3.ChatCompletionRequestMessage {
	if args.Bash {
		return []gpt3.ChatCompletionRequestMessage{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: content},
		}
	} else {
		return []gpt3.ChatCompletionRequestMessage{
			{Role: "system", Content: content},
		}
	}

}

func (args *askCmd) poll(input *bufio.Reader) (string, bool, error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	lineCh := make(chan string)
	defer close(lineCh)
	errCh := make(chan error)
	defer close(errCh)

	go func() {
		line, err := input.ReadString('\n')
		if err != nil {
			errCh <- err
		} else {
			lineCh <- line
		}
	}()

	select {
	case err := <-errCh:
		return "", false, err
	case <-sigCh:
		return "", false, nil
	case line := <-lineCh:
		line = strings.TrimSpace(line)
		if line == "" {
			return "", false, nil
		}
		signal.Stop(sigCh)
		close(sigCh)
		return line, true, nil
	}
}

func (args *askCmd) Execute(ctx context.Context, config *config) error {
	model := args.Model
	if model == "" {
		model = strings.TrimSpace(config.Model())
	}
	client := config.Client()

	lastMessage := ""
	content, err := args.buildContent(ctx)
	if err != nil {
		return fmt.Errorf("cannot build message: %w", err)
	}

	input := bufio.NewReader(os.Stdin)

	messages := args.messages(content)
	for {
		var response strings.Builder
		out := io.MultiWriter(os.Stdout, &response)

		err = aiStream(ctx, client, aiStreamInput{
			Messages:    messages,
			MaxTokens:   args.MaxTokens,
			Temperature: args.Temperature,
			Model:       model,
			Timeout:     2 * time.Minute,
		}, func(message string) error {
			if message != "" {
				lastMessage = message
			}
			_, err := fmt.Fprintf(out, "%s", message)
			return err
		})
		if err != nil {
			return err
		}
		if len(lastMessage) == 0 || lastMessage[len(lastMessage)-1] != '\n' {
			_, err := fmt.Fprintf(out, "\n")
			if err != nil {
				return err
			}
		}

		if args.Once {
			break
		}

		_, err := fmt.Printf("%shlp>%s ", colorGreen, colorReset)
		if err != nil {
			return err
		}

		line, cont, err := args.poll(input)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				return terminateSilently(err)
			}
			return err
		}

		if !cont {
			return nil
		}

		messages = append(
			messages,
			gpt3.ChatCompletionRequestMessage{Role: "assistant", Content: response.String()},
			gpt3.ChatCompletionRequestMessage{Role: "user", Content: line},
		)
	}

	return nil
}
