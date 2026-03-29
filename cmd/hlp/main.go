package main

import (
	"errors"
	"log"
	"os"

	"github.com/yiblet/hlp/internal/cli"
	"github.com/yiblet/hlp/internal/xerr"
)

func main() {
	if err := cli.Run(); err != nil {
		var errtype *xerr.SilentTerminationError
		if errors.As(err, &errtype) {
			os.Exit(0)
		}
		log.Printf("error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
