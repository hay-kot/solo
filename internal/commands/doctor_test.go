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

	cmd := &DoctorCmd{
		flags: &Flags{},
		out:   &buf,
	}

	cmd.checkTmuxInstalled()

	output := buf.String()

	// tmux may or may not be installed in the test environment.
	assert.Contains(t, output, "tmux")

	if _, err := exec.LookPath("tmux"); err == nil {
		assert.Contains(t, output, "[pass] tmux installed:")
	} else {
		assert.Contains(t, output, "[warn] tmux not found in PATH")
	}
}

func TestDoctorCmd_TmuxSession(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   string
	}{
		{
			name:   "inside tmux session",
			envVal: "/tmp/tmux-1000/default,12345,0",
			want:   "[pass] inside tmux session",
		},
		{
			name:   "not inside tmux session",
			envVal: "",
			want:   "[info] not inside tmux session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TMUX", tt.envVal)

			var buf bytes.Buffer

			cmd := &DoctorCmd{
				flags: &Flags{},
				out:   &buf,
			}

			cmd.checkTmuxSession()

			assert.Contains(t, buf.String(), tt.want)
		})
	}
}

func TestDoctorCmd_ProjectConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.Config
		want string
	}{
		{
			name: "project found",
			cfg:  cfgWithTabs(t, config.Tab{Title: "test", Cmd: "echo hi"}),
			want: "[pass] project config:",
		},
		{
			name: "no project",
			cfg:  config.Config{},
			want: "[info] no project config found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			cmd := &DoctorCmd{
				flags: &Flags{Config: tt.cfg},
				out:   &buf,
			}

			cmd.checkProjectConfig()

			assert.Contains(t, buf.String(), tt.want)
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

		var buf bytes.Buffer

		cmd := &DoctorCmd{
			flags: &Flags{},
			out:   &buf,
		}

		cmd.checkGlobalConfig()

		assert.Contains(t, buf.String(), "[pass] global config:")
	})

	t.Run("config not found", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		var buf bytes.Buffer

		cmd := &DoctorCmd{
			flags: &Flags{},
			out:   &buf,
		}

		cmd.checkGlobalConfig()

		assert.Contains(t, buf.String(), "[info] global config not found")
	})
}

func TestDoctorCmd_Run(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := &DoctorCmd{
		flags: &Flags{},
		out:   &buf,
	}

	err := cmd.run(context.Background(), nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "tmux")
}
