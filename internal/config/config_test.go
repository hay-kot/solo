package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFrom(t *testing.T) {
	t.Parallel()

	t.Run("default config when file missing", func(t *testing.T) {
		t.Parallel()
		cfg, err := ReadFrom("/nonexistent/path/config.yaml")
		require.NoError(t, err)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Empty(t, cfg.LogFile)
		assert.Nil(t, cfg.Projects)
	})

	t.Run("parses full config with projects", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")

		writeFile(t, path, `
log_level: debug
log_file: /tmp/solo.log
projects:
  /home/user/code/myproject:
    tabs:
      - title: server
        cmd: go run .
      - title: docker
        cmd: docker compose up
    timeout: 5
  /home/user/code:
    tabs:
      - title: shell
        cmd: ""
`)

		cfg, err := ReadFrom(path)
		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, "/tmp/solo.log", cfg.LogFile)
		require.Len(t, cfg.Projects, 2)

		proj := cfg.Projects["/home/user/code/myproject"]
		assert.Len(t, proj.Tabs, 2)
		assert.Equal(t, "server", proj.Tabs[0].Title)
		assert.Equal(t, "go run .", proj.Tabs[0].Cmd)
		assert.Equal(t, 5, proj.Timeout)

		proj2 := cfg.Projects["/home/user/code"]
		assert.Len(t, proj2.Tabs, 1)
		assert.Equal(t, "shell", proj2.Tabs[0].Title)
		assert.Empty(t, proj2.Tabs[0].Cmd)
		assert.Equal(t, 0, proj2.Timeout)
	})

	t.Run("parses config without projects", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")

		writeFile(t, path, `
log_level: warn
`)

		cfg, err := ReadFrom(path)
		require.NoError(t, err)
		assert.Equal(t, "warn", cfg.LogLevel)
		assert.Nil(t, cfg.Projects)
	})
}
