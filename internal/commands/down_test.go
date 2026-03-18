package commands

import (
	"context"
	"io"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/tmux"
)

func TestDownCmd(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name       string
		client     *mockClient
		cfg        config.Config
		wantErr    string
		wantKilled []string
		wantKeys   []sendKeysCall
	}{
		{
			name:    "not in tmux returns error",
			client:  &mockClient{inTmux: false},
			cfg:     config.Config{},
			wantErr: "not inside a tmux session",
		},
		{
			name: "no project config returns error",
			client: &mockClient{
				inTmux: true,
			},
			cfg:     config.Config{},
			wantErr: "resolving project",
		},
		{
			name: "tears down matching windows",
			client: &mockClient{
				inTmux: true,
				windows: []tmux.Window{
					{Name: "server", ID: "@1"},
					{Name: "logs", ID: "@2"},
				},
				listPaneCommandFn: func(_ string) (string, error) {
					return "zsh", nil
				},
			},
			cfg: config.Config{
				Projects: map[string]config.Project{
					cwd: {
						Tabs: []config.Tab{
							{Title: "server", Cmd: "go run ."},
							{Title: "logs", Cmd: "tail -f log"},
						},
					},
				},
			},
			wantKeys: []sendKeysCall{
				{Target: "@2", Keys: []string{"C-c"}},
				{Target: "@1", Keys: []string{"C-c"}},
			},
			wantKilled: []string{"@2", "@1"},
		},
		{
			name: "skips non-matching windows",
			client: &mockClient{
				inTmux: true,
				windows: []tmux.Window{
					{Name: "server", ID: "@1"},
				},
				listPaneCommandFn: func(_ string) (string, error) {
					return "bash", nil
				},
			},
			cfg: config.Config{
				Projects: map[string]config.Project{
					cwd: {
						Tabs: []config.Tab{
							{Title: "server", Cmd: "go run ."},
							{Title: "missing", Cmd: "echo hi"},
						},
					},
				},
			},
			wantKeys: []sendKeysCall{
				{Target: "@1", Keys: []string{"C-c"}},
			},
			wantKilled: []string{"@1"},
		},
		{
			name: "handles empty pane command as exited",
			client: &mockClient{
				inTmux: true,
				windows: []tmux.Window{
					{Name: "worker", ID: "@4"},
				},
				listPaneCommandFn: func(_ string) (string, error) {
					return "", nil
				},
			},
			cfg: config.Config{
				Projects: map[string]config.Project{
					cwd: {
						Tabs: []config.Tab{
							{Title: "worker", Cmd: "python worker.py"},
						},
					},
				},
			},
			wantKeys: []sendKeysCall{
				{Target: "@4", Keys: []string{"C-c"}},
			},
			wantKilled: []string{"@4"},
		},
		{
			name: "handles immediate shell exit",
			client: &mockClient{
				inTmux: true,
				windows: []tmux.Window{
					{Name: "build", ID: "@3"},
				},
				listPaneCommandFn: func(_ string) (string, error) {
					return "fish", nil
				},
			},
			cfg: config.Config{
				Projects: map[string]config.Project{
					cwd: {
						Tabs: []config.Tab{
							{Title: "build", Cmd: "make"},
						},
					},
				},
			},
			wantKeys: []sendKeysCall{
				{Target: "@3", Keys: []string{"C-c"}},
			},
			wantKilled: []string{"@3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := &Flags{Config: tt.cfg}
			cmd := NewDownCmd(flags, tt.client)
			cmd.out = io.Discard

			err := cmd.run(context.Background(), nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantKilled, tt.client.killedWins)

			if tt.wantKeys != nil {
				assert.Equal(t, tt.wantKeys, tt.client.sentKeys)
			} else {
				assert.Empty(t, tt.client.sentKeys)
			}
		})
	}
}

func TestDownCmd_Timeout(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Process never returns to shell — ListPaneCommand always returns a
	// non-shell command. The down command should still kill the window after
	// the timeout expires.
	client := &mockClient{
		inTmux: true,
		windows: []tmux.Window{
			{Name: "stuck", ID: "@5"},
		},
		listPaneCommandFn: func(_ string) (string, error) {
			return "node", nil
		},
	}

	cfg := config.Config{
		Projects: map[string]config.Project{
			cwd: {
				Tabs: []config.Tab{
					{Title: "stuck", Cmd: "node server.js"},
				},
				Timeout: 1, // 1 second to keep test fast
			},
		},
	}

	flags := &Flags{Config: cfg}
	cmd := NewDownCmd(flags, client)
	cmd.out = io.Discard

	err = cmd.run(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"@5"}, client.killedWins)
	assert.Equal(t, []sendKeysCall{{Target: "@5", Keys: []string{"C-c"}}}, client.sentKeys)
}

func TestDownCmd_ProcessExitsAfterPolling(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Simulate a process that exits after 2 polls.
	var calls atomic.Int32

	client := &mockClient{
		inTmux: true,
		windows: []tmux.Window{
			{Name: "worker", ID: "@7"},
		},
		listPaneCommandFn: func(_ string) (string, error) {
			n := calls.Add(1)
			if n >= 3 {
				return "zsh", nil
			}
			return "python", nil
		},
	}

	cfg := config.Config{
		Projects: map[string]config.Project{
			cwd: {
				Tabs: []config.Tab{
					{Title: "worker", Cmd: "python worker.py"},
				},
				Timeout: 10,
			},
		},
	}

	flags := &Flags{Config: cfg}
	cmd := NewDownCmd(flags, client)
	cmd.out = io.Discard

	err = cmd.run(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"@7"}, client.killedWins)
	assert.GreaterOrEqual(t, int(calls.Load()), 3)
}
