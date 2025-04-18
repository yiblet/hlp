package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"
)

type mainCmd struct {
	Ask        *askCmd    `arg:"subcommand"`
	Config     *configCmd `arg:"subcommand"`
	Chat       *chatCmd   `arg:"subcommand"`
	ConfigName string     `arg:"-c,--config,env:HLP_CONFIG" help:"name of the configuration set"`
	Debug      bool       `arg:"-d,--debug" help:"enable debug mode"`
}

func (args *mainCmd) SetupConfig() (config, error) {
	var err error
	cfg, err := ReadConfig(args.ConfigName, args.Debug)

	if err != nil {
		return config{}, fmt.Errorf("failed fetching configs: %w", err)
	}
	return cfg, nil
}

func (args *mainCmd) Execute(ctx context.Context) error {
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
	default:
		err = writeHelp(args, os.Stderr)
	}

	return err
}

func run() error {
	var args mainCmd
	ctx := context.Background()
	arg.MustParse(&args)
	if err := args.Execute(ctx); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		var errtype *terminateSilentlyError
		if errors.As(err, &errtype) {
			os.Exit(0)
		}
		log.Printf("error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
