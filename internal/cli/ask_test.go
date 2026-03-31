package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskCmdSchema(t *testing.T) {
	t.Parallel()

	t.Run("accepts object schema", func(t *testing.T) {
		t.Parallel()

		args := AskCmd{Schema: `{"type":"object","properties":{"answer":{"type":"string"}}}`}

		schema, err := args.schema()
		require.NoError(t, err)
		assert.IsType(t, map[string]any{}, schema)
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		t.Parallel()

		args := AskCmd{Schema: `{`}

		_, err := args.schema()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON schema")
	})

	t.Run("rejects non object schema", func(t *testing.T) {
		t.Parallel()

		args := AskCmd{Schema: `[]`}

		_, err := args.schema()
		require.Error(t, err)
		assert.EqualError(t, err, "invalid JSON schema: expected top-level object")
	})
}
