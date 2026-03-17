package ui

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpinner_Stop(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSpinner(&buf, false)

	s.Start("loading")
	time.Sleep(100 * time.Millisecond)
	s.Stop("done loading")

	out := buf.String()
	assert.Contains(t, out, "✓")
	assert.Contains(t, out, "done loading")
}

func TestSpinner_Warn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSpinner(&buf, false)

	s.Start("checking")
	time.Sleep(100 * time.Millisecond)
	s.Warn("something wrong")

	out := buf.String()
	assert.Contains(t, out, "⚠")
	assert.Contains(t, out, "something wrong")
}

func TestSpinner_NoColor(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSpinner(&buf, true)

	s.Start("working")
	time.Sleep(100 * time.Millisecond)
	s.Stop("finished")

	out := buf.String()
	assert.Contains(t, out, "* done:")
	assert.Contains(t, out, "finished")
	assert.NotContains(t, out, "\033[32m")
}

func TestSpinner_Update(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSpinner(&buf, true)

	s.Start("step 1")
	time.Sleep(100 * time.Millisecond)
	s.Update("step 2")
	time.Sleep(100 * time.Millisecond)
	s.Stop("all steps done")

	out := buf.String()
	assert.Contains(t, out, "step 1")
	assert.Contains(t, out, "step 2")
	assert.Contains(t, out, "all steps done")
}
