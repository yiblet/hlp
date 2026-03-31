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
	Cat        *CatCmd    `arg:"subcommand" help:"read stdin, append it to the prompt, and answer once"`
	Ask        *AskCmd    `arg:"subcommand" help:"ask a question and print the answer"`
	Config     *ConfigCmd `arg:"subcommand" help:"read and update stored configuration values"`
	Chat       *ChatCmd   `arg:"subcommand" help:"run a chat-file conversation and optionally write the reply back"`
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
