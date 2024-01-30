package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexflint/go-arg"
)

type mainCmd struct {
	Ask    *askCmd    `arg:"subcommand"`
	Config *configCmd `arg:"subcommand"`
	Chat   *chatCmd   `arg:"subcommand"`
}

func (args *mainCmd) Execute(ctx context.Context) error {
	var err error
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("failed fetching configs: %w", err)
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

func main() {
	var args mainCmd
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()
	arg.MustParse(&args)

	if err := args.Execute(ctx); err != nil {
		log.Panic(err)
	}
}
