//go:build integration

package integration

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const commandTimeout = 30 * time.Second

// Harness wraps solo binary invocations with isolated environment per test.
type Harness struct {
	t          *testing.T
	configDir  string
	dataDir    string
	homeDir    string
	configPath string
}

// NewHarness creates a new test harness with isolated temp directories.
func NewHarness(t *testing.T) *Harness {
	t.Helper()

	configDir := t.TempDir()
	dataDir := t.TempDir()
	homeDir := t.TempDir()

	return &Harness{
		t:          t,
		configDir:  configDir,
		dataDir:    dataDir,
		homeDir:    homeDir,
		configPath: testdataConfig(),
	}
}

// Run executes solo with the given arguments and returns combined output.
func (h *Harness) Run(args ...string) (string, error) {
	h.t.Helper()
	cmd := h.command(args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunInDir executes solo with a specific working directory.
func (h *Harness) RunInDir(dir string, args ...string) (string, error) {
	h.t.Helper()
	cmd := h.command(args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunStdout executes solo and returns only stdout (ignoring stderr).
func (h *Harness) RunStdout(args ...string) (string, error) {
	h.t.Helper()
	cmd := h.command(args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			h.t.Logf("stderr from solo %v: %s", args, exitErr.Stderr)
		}
	}
	return string(out), err
}

// WithConfig writes a YAML config file to the harness config dir.
func (h *Harness) WithConfig(yaml string) *Harness {
	h.t.Helper()
	configPath := filepath.Join(h.configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0o644); err != nil {
		h.t.Fatalf("writing test config: %v", err)
	}
	h.configPath = configPath
	return h
}

// DataDir returns the isolated data directory path.
func (h *Harness) DataDir() string { return h.dataDir }

// HomeDir returns the isolated home directory path.
func (h *Harness) HomeDir() string { return h.homeDir }

func (h *Harness) command(args ...string) *exec.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	h.t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, soloBin, args...)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"TMPDIR=" + os.Getenv("TMPDIR"),
		"TERM=" + os.Getenv("TERM"),
		"HOME=" + h.homeDir,
		"XDG_CONFIG_HOME=" + h.configDir,
		"XDG_DATA_HOME=" + h.dataDir,
		"NO_COLOR=1",
	}
	if h.configPath != "" {
		cmd.Env = append(cmd.Env, "CONFIG_FILE="+h.configPath)
	}
	// Propagate tmux socket isolation if set
	if tmuxDir := os.Getenv("TMUX_TMPDIR"); tmuxDir != "" {
		cmd.Env = append(cmd.Env, "TMUX_TMPDIR="+tmuxDir)
	}
	return cmd
}

// testdataConfig resolves the path to the test config fixture relative to this source file.
func testdataConfig() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot resolve test source file path")
	}
	return filepath.Join(filepath.Dir(file), "testdata", "config.yaml")
}
