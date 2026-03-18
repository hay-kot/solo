package commands

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hay-kot/solo/internal/config"
)

func TestConfigCmd_ResolvesAndPrintsYAML(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	cfg := config.Config{
		Projects: map[string]config.Project{
			cwd: {
				Tabs: []config.Tab{
					{Title: "server", Cmd: "go run ."},
					{Title: "docker", Cmd: "docker compose up"},
				},
				Timeout: 5,
			},
		},
	}

	var buf bytes.Buffer

	cmd := &ConfigCmd{
		flags: &Flags{Config: cfg},
		out:   &buf,
	}

	err = cmd.run(context.Background(), nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Source: global (exact match)")
	assert.Contains(t, output, "---")
	assert.Contains(t, output, "title: server")
	assert.Contains(t, output, "cmd: go run .")
	assert.Contains(t, output, "timeout: 5")
}

func TestConfigCmd_NoProjectReturnsError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := &ConfigCmd{
		flags: &Flags{Config: config.Config{}},
		out:   &buf,
	}

	err := cmd.run(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolving project")
}

func TestConfigCmd_LocalConfig(t *testing.T) {
	dir := t.TempDir()

	soloConfig := `tabs:
  - title: dev
    cmd: npm start
timeout: 3
`

	err := os.WriteFile(dir+"/.solo.yml", []byte(soloConfig), 0o644)
	require.NoError(t, err)

	t.Chdir(dir)

	var buf bytes.Buffer

	cmd := &ConfigCmd{
		flags: &Flags{Config: config.Config{}},
		out:   &buf,
	}

	err = cmd.run(context.Background(), nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Source: local")
	assert.Contains(t, output, "title: dev")
	assert.Contains(t, output, "cmd: npm start")
	assert.Contains(t, output, "timeout: 3")
}

func TestFormatSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "local", source: "local", want: "local"},
		{name: "global exact", source: "global-exact", want: "global (exact match)"},
		{name: "global basename", source: "global-basename", want: "global (basename match)"},
		{name: "global glob", source: "global-glob", want: "global (glob match)"},
		{name: "unknown", source: "custom", want: "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, formatSource(tt.source))
		})
	}
}
