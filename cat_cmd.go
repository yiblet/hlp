package main

import "context"

type catCmd struct {
	Question    []string `arg:"positional"`
	MaxTokens   int      `arg:"--tokens,-t" default:"0" help:"the maximum amount of tokens allowed in the output"`
	Temperature *float32 `arg:"--temp"`
	Model       string   `arg:"--model,-m" help:"set openai model"`
}

func (cat *catCmd) Execute(ctx context.Context, config *config) error {
	ask := askCmd{
		Question:    cat.Question,
		MaxTokens:   cat.MaxTokens,
		Temperature: cat.Temperature,
		Model:       cat.Model,
		Once:        true,
		Attach:      []string{"-"},
	}
	return ask.Execute(ctx, config)
}
