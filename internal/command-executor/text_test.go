package executor_test

import (
	"testing"

	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	assert "gotest.tools/v3/assert"
)

func TestStripPrefix(t *testing.T) {
	assert.Equal(t, executor.StripPrefix("!", "ps")("!ps test"), "test")
	assert.Equal(t, executor.StripPrefix("!", "ps")("!ps       test"), "test")
	assert.Equal(t, executor.StripPrefix("!", "ps")("!pstest"), "est")
}

func TestHasPrefix(t *testing.T) {
	assert.Equal(t, executor.HasCommandPrefix("!", "ps", "!ps test"), true)
	assert.Equal(t, executor.HasCommandPrefix("!", "ps", "!ps      test"), true)
	assert.Equal(t, executor.HasCommandPrefix("!", "ps", "!pstest"), false)

	assert.Equal(t, executor.HasCommandPrefix("!", "ps", "!pslist"), false)
}
