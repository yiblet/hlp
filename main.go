package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/alexflint/go-arg"
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

func main() {
	var args struct {
		QuestionWords []string `arg:"positional"`
		OpenAIAPIKey  string   `arg:"env:OPENAI_API_KEY,required"`
		MaxTokens     int      `default:"512"`
		Temperature   float32  `default:"0.7"`
	}

	arg.MustParse(&args)

	ctx := context.Background()

	client := gpt3.NewClient(args.OpenAIAPIKey, gpt3.WithDefaultEngine(gpt3.TextDavinci003Engine))

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

	if err != nil {
		log.Panic(err)
	}
}
