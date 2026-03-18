package commands

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hay-kot/solo/internal/config"
)

func TestDoctorCmd_TmuxInstalled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := &DoctorCmd{flags: &Flags{NoColor: true}, out: &buf}

	sec := cmd.checkDependencies()

	assert.Equal(t, "Dependencies", sec.Title)
	require.Len(t, sec.Checks, 1)
	assert.Equal(t, "tmux", sec.Checks[0].Label)

	if _, err := exec.LookPath("tmux"); err == nil {
		assert.Contains(t, sec.Checks[0].Detail, "/tmux")
	} else {
		assert.Equal(t, "not found in PATH", sec.Checks[0].Detail)
	}
}

func TestDoctorCmd_TmuxSession(t *testing.T) {
	tests := []struct {
		name       string
		envVal     string
		wantDetail string
	}{
		{
			name:       "inside tmux session",
			envVal:     "/tmp/tmux-1000/default,12345,0",
			wantDetail: "",
		},
		{
			name:       "not inside tmux session",
			envVal:     "",
			wantDetail: "not inside a tmux session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TMUX", tt.envVal)

			cmd := &DoctorCmd{flags: &Flags{NoColor: true}, out: &bytes.Buffer{}}
			sec := cmd.checkEnvironment()

			require.Len(t, sec.Checks, 1)
			assert.Equal(t, "tmux session", sec.Checks[0].Label)
			assert.Equal(t, tt.wantDetail, sec.Checks[0].Detail)
		})
	}
}

func TestDoctorCmd_ProjectConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        config.Config
		wantDetail string
	}{
		{
			name:       "project found",
			cfg:        cfgWithTabs(t, config.Tab{Title: "test", Cmd: "echo hi"}),
			wantDetail: "global (exact match)",
		},
		{
			name:       "no project",
			cfg:        config.Config{},
			wantDetail: "no match for current directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := &DoctorCmd{
				flags: &Flags{Config: tt.cfg, NoColor: true},
				out:   &bytes.Buffer{},
			}

			sec := cmd.checkConfiguration()

			// Find the project config check
			var found bool
			for _, c := range sec.Checks {
				if c.Label == "project config" {
					assert.Equal(t, tt.wantDetail, c.Detail)
					found = true
				}
			}
			assert.True(t, found, "expected project config check")
		})
	}
}

func TestDoctorCmd_GlobalConfig(t *testing.T) {
	t.Run("config exists and valid", func(t *testing.T) {
		dir := t.TempDir()
		soloDir := filepath.Join(dir, "solo")
		require.NoError(t, os.MkdirAll(soloDir, 0o755))

		cfgPath := filepath.Join(soloDir, "config.yaml")
		err := os.WriteFile(cfgPath, []byte("log_level: debug\n"), 0o644)
		require.NoError(t, err)

		t.Setenv("XDG_CONFIG_HOME", dir)

		cmd := &DoctorCmd{flags: &Flags{NoColor: true}, out: &bytes.Buffer{}}
		sec := cmd.checkConfiguration()

		var found bool
		for _, c := range sec.Checks {
			if c.Label == "global config" {
				assert.Contains(t, c.Detail, cfgPath)
				found = true
			}
		}
		assert.True(t, found, "expected global config check")
	})

	t.Run("config not found", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cmd := &DoctorCmd{flags: &Flags{NoColor: true}, out: &bytes.Buffer{}}
		sec := cmd.checkConfiguration()

		var found bool
		for _, c := range sec.Checks {
			if c.Label == "global config" {
				assert.Equal(t, "not found", c.Detail)
				found = true
			}
		}
		assert.True(t, found, "expected global config check")
	})
}

func TestDoctorCmd_Run(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := &DoctorCmd{flags: &Flags{NoColor: true}, out: &buf}

	err := cmd.run(context.Background(), nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Solo Doctor")
	assert.Contains(t, output, "Dependencies")
	assert.Contains(t, output, "tmux")
	assert.Contains(t, output, "passed")
}
