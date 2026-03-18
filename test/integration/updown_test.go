//go:build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tmuxRun executes a tmux command and returns its trimmed output.
func tmuxRun(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("tmux", args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "tmux %v failed: %s", args, out)
	return strings.TrimSpace(string(out))
}

// tmuxWindowNames returns the names of all windows in the given session.
func tmuxWindowNames(t *testing.T, session string) []string {
	t.Helper()
	raw := tmuxRun(t, "list-windows", "-t", session, "-F", "#{window_name}")
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

func TestUpDown(t *testing.T) {
	session := fmt.Sprintf("solo-test-%d", time.Now().UnixNano())

	// Start an isolated tmux session
	tmuxRun(t, "new-session", "-d", "-s", session, "-x", "200", "-y", "50")
	t.Cleanup(func() {
		_ = exec.Command("tmux", "kill-session", "-t", session).Run()
	})

	workDir := t.TempDir()

	configYAML := fmt.Sprintf(`projects:
  %s:
    timeout: 2
    tabs:
      - title: alpha
        cmd: "echo hello"
      - title: bravo
        cmd: "echo world"
`, workDir)

	h := NewHarness(t).WithConfig(configYAML)

	// Run solo up inside the tmux session
	soloUpCmd := fmt.Sprintf("%s --config %s up", soloBin, h.configPath)
	tmuxRun(t, "send-keys", "-t", session, fmt.Sprintf("cd %s && %s", workDir, soloUpCmd), "Enter")

	// Wait for solo up to complete
	require.Eventually(t, func() bool {
		names := tmuxWindowNames(t, session)
		return containsAll(names, "alpha", "bravo")
	}, 10*time.Second, 250*time.Millisecond, "windows alpha and bravo should be created")

	windows := tmuxWindowNames(t, session)
	assert.Contains(t, windows, "alpha")
	assert.Contains(t, windows, "bravo")

	// Run solo down inside the tmux session
	soloDownCmd := fmt.Sprintf("%s --config %s down", soloBin, h.configPath)
	tmuxRun(t, "send-keys", "-t", session, soloDownCmd, "Enter")

	// Wait for solo down to complete
	require.Eventually(t, func() bool {
		names := tmuxWindowNames(t, session)
		return !containsAny(names, "alpha", "bravo")
	}, 15*time.Second, 250*time.Millisecond, "windows alpha and bravo should be torn down")

	windows = tmuxWindowNames(t, session)
	assert.NotContains(t, windows, "alpha")
	assert.NotContains(t, windows, "bravo")
}

func TestUpSkipsExisting(t *testing.T) {
	session := fmt.Sprintf("solo-skip-%d", time.Now().UnixNano())

	tmuxRun(t, "new-session", "-d", "-s", session, "-x", "200", "-y", "50")
	t.Cleanup(func() {
		_ = exec.Command("tmux", "kill-session", "-t", session).Run()
	})

	// Pre-create "alpha" window
	tmuxRun(t, "new-window", "-t", session, "-n", "alpha")

	workDir := t.TempDir()

	configYAML := fmt.Sprintf(`projects:
  %s:
    tabs:
      - title: alpha
        cmd: "echo exists"
      - title: bravo
        cmd: "echo new"
`, workDir)

	h := NewHarness(t).WithConfig(configYAML)

	soloUpCmd := fmt.Sprintf("%s --config %s up", soloBin, h.configPath)
	tmuxRun(t, "send-keys", "-t", session, fmt.Sprintf("cd %s && %s", workDir, soloUpCmd), "Enter")

	require.Eventually(t, func() bool {
		names := tmuxWindowNames(t, session)
		return containsAll(names, "alpha", "bravo")
	}, 10*time.Second, 250*time.Millisecond)

	// Verify only one "alpha" window exists (wasn't duplicated)
	windows := tmuxWindowNames(t, session)
	count := 0
	for _, name := range windows {
		if name == "alpha" {
			count++
		}
	}
	assert.Equal(t, 1, count, "alpha should appear exactly once (not recreated)")
}

func containsAll(slice []string, items ...string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	for _, item := range items {
		if _, ok := set[item]; !ok {
			return false
		}
	}
	return true
}

func containsAny(slice []string, items ...string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	for _, item := range items {
		if _, ok := set[item]; ok {
			return true
		}
	}
	return false
}
