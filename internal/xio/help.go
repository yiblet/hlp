package xio

import (
	"io"
	"strings"

	"github.com/alexflint/go-arg"
)

type helpEpiloger interface {
	HelpEpilogue() string
}

func WriteHelp(config any, output io.Writer) error {
	parser, err := arg.NewParser(arg.Config{}, config)
	if err != nil {
		return err
	}
	if err := writeHelp(parser, config, output); err != nil {
		return err
	}
	return nil
}

func WriteParserHelp(parser *arg.Parser, output io.Writer) error {
	return writeHelp(parser, parser.Subcommand(), output)
}

func writeHelp(parser *arg.Parser, config any, output io.Writer) error {
	parser.WriteHelp(output)
	if config == nil {
		return nil
	}

	if epiloger, ok := config.(helpEpiloger); ok {
		epilogue := strings.TrimSpace(epiloger.HelpEpilogue())
		if epilogue != "" {
			if _, err := io.WriteString(output, "\n\n"+epilogue+"\n"); err != nil {
				return err
			}
		}
	}
	return nil
}
