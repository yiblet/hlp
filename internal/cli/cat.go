package cli

import (
	"context"

	configpkg "github.com/yiblet/hlp/internal/config"
)

type CatCmd struct {
	Question    []string `arg:"positional" help:"question or prompt text"`
	MaxTokens   int      `arg:"--tokens,-t" default:"0" help:"the maximum amount of tokens allowed in the output"`
	Temperature *float32 `arg:"--temp" help:"sampling temperature for the model response"`
	Model       string   `arg:"--model,-m" help:"set openai model"`
}

func (cat *CatCmd) Execute(ctx context.Context, config *configpkg.Config) error {
	ask := AskCmd{
		Question:    cat.Question,
		MaxTokens:   cat.MaxTokens,
		Temperature: cat.Temperature,
		Model:       cat.Model,
		Once:        true,
		Attach:      []string{"-"},
	}
	return ask.Execute(ctx, config)
}
