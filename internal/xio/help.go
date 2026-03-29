package xio

import (
	"io"

	"github.com/alexflint/go-arg"
)

func WriteHelp(config any, output io.Writer) error {
	parser, err := arg.NewParser(arg.Config{}, config)
	if err != nil {
		return err
	}
	parser.WriteHelp(output)
	return nil
}
