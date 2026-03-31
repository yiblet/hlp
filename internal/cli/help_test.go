package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/alexflint/go-arg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yiblet/hlp/internal/xio"
)

func TestWriteHelpRootIncludesImprovedCommandDescriptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := xio.WriteHelp(&MainCmd{}, &buf)
	require.NoError(t, err)

	help := buf.String()
	assert.Contains(t, help, "read stdin, append it to the prompt, and answer once")
	assert.Contains(t, help, "read and update stored configuration values")
	assert.Contains(t, help, "run a chat-file conversation and optionally write the reply back")
}

func TestWriteHelpAskIncludesPromptAndTemperatureHelp(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := xio.WriteHelp(&AskCmd{}, &buf)
	require.NoError(t, err)

	help := buf.String()
	assert.Contains(t, help, "question or prompt text")
	assert.Contains(t, help, "sampling temperature for the model response")
	assert.Contains(t, help, "constrain the output to the provided JSON schema")
	assert.Contains(t, help, "append file contents to the prompt")
	assert.Contains(t, help, "send a single request and exit instead of entering follow-up mode")
	assert.Contains(t, help, "Schema examples:")
	assert.Contains(t, help, "Simple object:")
	assert.Contains(t, help, "Object with array:")
	assert.Contains(t, help, "Enum field:")
}

func TestWriteParserHelpIncludesAskSchemaExamplesForSubcommandHelp(t *testing.T) {
	t.Parallel()

	cmd := &MainCmd{}
	parser, err := arg.NewParser(arg.Config{}, cmd)
	require.NoError(t, err)

	err = parser.Parse([]string{"ask", "--help"})
	require.True(t, errors.Is(err, arg.ErrHelp))

	var buf bytes.Buffer
	require.NoError(t, xio.WriteParserHelp(parser, &buf))

	help := buf.String()
	assert.Contains(t, help, "Global options:")
	assert.Contains(t, help, "Schema examples:")
	assert.Contains(t, help, "Simple object:")
	assert.Contains(t, help, "Enum field:")
}

func TestWriteHelpConfigIncludesSubcommandDescriptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := xio.WriteHelp(&ConfigCmd{}, &buf)
	require.NoError(t, err)

	help := buf.String()
	assert.Contains(t, help, "set a stored configuration value")
	assert.Contains(t, help, "print a stored configuration value")
	assert.Contains(t, help, "print the configuration directory path")
}

func TestWriteHelpConfigGetIncludesFieldDescriptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := xio.WriteHelp(&ConfigGetCmd{}, &buf)
	require.NoError(t, err)

	help := buf.String()
	assert.Contains(t, help, "print the default model")
	assert.Contains(t, help, "print the stored OpenAI API key")
	assert.Contains(t, help, "print the stored OpenAI API endpoint")
}

func TestChatParserAcceptsTempFlag(t *testing.T) {
	t.Parallel()

	cmd := &ChatCmd{}
	parser, err := arg.NewParser(arg.Config{}, cmd)
	require.NoError(t, err)

	err = parser.Parse([]string{"--temp", "0.7", "input.chat"})
	require.NoError(t, err)
	require.NotNil(t, cmd.Temperature)
	assert.InDelta(t, 0.7, *cmd.Temperature, 0.0001)
	assert.Equal(t, "input.chat", cmd.File)
}

func TestWriteHelpChatUsesTempFlagAndColorDescription(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := xio.WriteHelp(&ChatCmd{}, &buf)
	require.NoError(t, err)

	help := buf.String()
	assert.Contains(t, help, "--temp TEMP")
	assert.Contains(t, help, "sampling temperature for the model response")
	assert.Contains(t, help, "render streamed output through bat when available")
	assert.NotContains(t, help, "--temperature")
}
