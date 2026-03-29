package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	configpkg "github.com/yiblet/hlp/internal/config"
	"github.com/yiblet/hlp/internal/xio"
)

type MainCmd struct {
	Cat        *CatCmd    `arg:"subcommand" help:"read from stdin and positional to apply some change"`
	Ask        *AskCmd    `arg:"subcommand" help:"ask for a question and print the answer"`
	Config     *ConfigCmd `arg:"subcommand" help:"configure hlp"`
	Chat       *ChatCmd   `arg:"subcommand" help:"chat with hlp using chat-file format"`
	ConfigName string     `arg:"-c,--config,env:HLP_CONFIG" help:"name of the configuration set"`
	Debug      bool       `arg:"-d,--debug" help:"enable debug mode"`
}

func (args *MainCmd) SetupConfig() (configpkg.Config, error) {
	cfg, err := configpkg.Read(args.ConfigName, args.Debug)
	if err != nil {
		return configpkg.Config{}, fmt.Errorf("failed fetching configs: %w", err)
	}
	return cfg, nil
}

func (args *MainCmd) Execute(ctx context.Context) error {
	config, err := args.SetupConfig()
	if err != nil {
		return err
	}

	switch {
	case args.Ask != nil:
		err = args.Ask.Execute(ctx, &config)
	case args.Config != nil:
		err = args.Config.Execute(ctx, &config)
	case args.Chat != nil:
		err = args.Chat.Execute(ctx, &config)
	case args.Cat != nil:
		err = args.Cat.Execute(ctx, &config)
	default:
		err = xio.WriteHelp(args, os.Stderr)
	}

	return err
}

func Run() error {
	var args MainCmd
	ctx := context.Background()
	arg.MustParse(&args)
	return args.Execute(ctx)
}
