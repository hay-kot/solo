package commands

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/tmux"
)

func TestUpCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		client       *mockClient
		cfg          config.Config
		wantErr      string
		wantCreated  []string
		wantSentKeys []sendKeysCall
		wantSelected []string
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
				inTmux:     true,
				currentWin: "main",
			},
			cfg:     config.Config{},
			wantErr: "resolving project",
		},
		{
			name: "creates windows and sends commands",
			client: &mockClient{
				inTmux:     true,
				currentWin: "main",
			},
			cfg: cfgWithTabs(t,
				config.Tab{Title: "server", Cmd: "go run ."},
				config.Tab{Title: "logs", Cmd: "tail -f log"},
			),
			wantCreated: []string{"server", "logs"},
			wantSentKeys: []sendKeysCall{
				{Target: "@0", Keys: []string{"go run .", "Enter"}},
				{Target: "@1", Keys: []string{"tail -f log", "Enter"}},
			},
			wantSelected: []string{"main"},
		},
		{
			name: "skips existing windows",
			client: &mockClient{
				inTmux:     true,
				currentWin: "main",
				windows: []tmux.Window{
					{Name: "server", ID: "@1"},
				},
			},
			cfg: cfgWithTabs(t,
				config.Tab{Title: "server", Cmd: "go run ."},
				config.Tab{Title: "logs", Cmd: "tail -f log"},
			),
			wantCreated: []string{"logs"},
			wantSentKeys: []sendKeysCall{
				{Target: "@0", Keys: []string{"tail -f log", "Enter"}},
			},
			wantSelected: []string{"main"},
		},
		{
			name: "returns to original window",
			client: &mockClient{
				inTmux:     true,
				currentWin: "editor",
			},
			cfg: cfgWithTabs(t,
				config.Tab{Title: "build", Cmd: "make"},
			),
			wantCreated: []string{"build"},
			wantSentKeys: []sendKeysCall{
				{Target: "@0", Keys: []string{"make", "Enter"}},
			},
			wantSelected: []string{"editor"},
		},
		{
			name: "empty command skips SendKeys",
			client: &mockClient{
				inTmux:     true,
				currentWin: "main",
			},
			cfg: cfgWithTabs(t,
				config.Tab{Title: "shell", Cmd: ""},
			),
			wantCreated:  []string{"shell"},
			wantSelected: []string{"main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := &Flags{Config: tt.cfg}
			cmd := NewUpCmd(flags, tt.client)
			cmd.out = io.Discard

			err := cmd.run(context.Background(), nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCreated, tt.client.createdWins)

			if tt.wantSentKeys != nil {
				assert.Equal(t, tt.wantSentKeys, tt.client.sentKeys)
			} else {
				assert.Empty(t, tt.client.sentKeys)
			}

			assert.Equal(t, tt.wantSelected, tt.client.selectedWins)
		})
	}
}

// cfgWithTabs creates a Config with the given tabs mapped to the current working
// directory so ResolveProject finds them via global-exact match.
func cfgWithTabs(t *testing.T, tabs ...config.Tab) config.Config {
	t.Helper()

	// The up command calls os.Getwd(), which in tests returns the package directory.
	cwd, err := os.Getwd()
	require.NoError(t, err)

	return config.Config{
		Projects: map[string]config.Project{
			cwd: {Tabs: tabs},
		},
	}
}
