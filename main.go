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
	Ask  *askCmd  `arg:"subcommand"`
	Auth *authCmd `arg:"subcommand"`
	Chat *chatCmd `arg:"subcommand"`
}

func main() {
	var args mainCmd
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()
	arg.MustParse(&args)

	var err error
	switch {
	case args.Ask != nil:
		err = args.Ask.Execute(ctx)
	case args.Auth != nil:
		err = args.Auth.Execute(ctx)
	case args.Chat != nil:
		err = args.Chat.Execute(ctx)
	default:
		fmt.Printf("invalid command: run with --help for more info\n")
		os.Exit(1)
		return
	}

	if err != nil {
		log.Panic(err)
	}
}
