package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/alexflint/go-arg"
	"github.com/kirsle/configdir"
)

var template = strings.ReplaceAll(`
For the following questions, let's think step by step in bash comments to output to make sure 
we output the correct bash command.

Question:
How do I go to my home directory
Answer:
\\\bash
# use the cd command to change directories
cd ~
# this will take you to the home directory
\\\
Question: 
%s
Answer:
\\\bash
`, "\\", "`")

type askCmd struct {
	QuestionWords []string `arg:"positional"`
	MaxTokens     int      `default:"512"`
	Temperature   float32  `default:"0.7"`
}

func (args *askCmd) Execute(ctx context.Context) error {
	config := authConfig{}
	if err := config.Read(); err != nil {
		return err
	}

	client := gpt3.NewClient(config.OpenAIAPIKey, gpt3.WithDefaultEngine(gpt3.TextDavinci003Engine))

	err := client.CompletionStream(ctx, gpt3.CompletionRequest{
		Prompt:      []string{fmt.Sprintf(template, strings.Join(args.QuestionWords, " "))},
		MaxTokens:   &args.MaxTokens,
		Temperature: &args.Temperature,
		Stop:        []string{"```"},
		Stream:      true,
	},
		func(cr *gpt3.CompletionResponse) {
			fmt.Printf("%s", cr.Choices[0].Text)
		})

	return err
}

type authConfig struct {
	OpenAIAPIKey string `json:"openai_api_key"`
}

func (c *authConfig) Read() error {
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	configPath := configdir.LocalConfig("hlp")
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, "configuration.json")
	if _, err = os.Stat(configFile); err != nil {
		return err
	}

	// Load the existing file.
	fh, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	decoder := json.NewDecoder(fh)
	return decoder.Decode(c)
}

func (c *authConfig) Write() error {
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	configPath := configdir.LocalConfig("hlp")
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, "configuration.json")
	fh, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	encoder := json.NewEncoder(fh)
	return encoder.Encode(c)
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

type mainCmd struct {
	Ask  *askCmd  `arg:"subcommand"`
	Auth *authCmd `arg:"subcommand"`
}

func main() {
	var args mainCmd
	ctx := context.Background()
	arg.MustParse(&args)

	var err error
	switch {
	case args.Ask != nil:
		err = args.Ask.Execute(ctx)
	case args.Auth != nil:
		err = args.Auth.Execute(ctx)
	default:
		err = fmt.Errorf("invalid command")
	}

	if err != nil {
		log.Panic(err)
	}
}
