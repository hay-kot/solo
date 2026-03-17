//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmokeHelp(t *testing.T) {
	h := NewHarness(t)
	out, err := h.Run("help")
	require.NoError(t, err, "solo help should succeed: %s", out)
	assert.Contains(t, out, "solo", "help output should mention solo")
}

func TestSmokeVersion(t *testing.T) {
	h := NewHarness(t)
	out, err := h.Run("--version")
	require.NoError(t, err, "solo --version should succeed: %s", out)
	assert.NotEmpty(t, out, "version output should not be empty")
	t.Logf("version: %s", out)
}
