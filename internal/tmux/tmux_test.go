package tmux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWindows(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []Window
	}{
		{
			name:   "empty",
			input:  "",
			expect: nil,
		},
		{
			name:   "whitespace only",
			input:  "  \n\n  ",
			expect: nil,
		},
		{
			name:  "single window",
			input: "main @0\n",
			expect: []Window{
				{Name: "main", ID: "@0"},
			},
		},
		{
			name:  "multiple windows",
			input: "editor @0\nbuild @1\nlogs @2\n",
			expect: []Window{
				{Name: "editor", ID: "@0"},
				{Name: "build", ID: "@1"},
				{Name: "logs", ID: "@2"},
			},
		},
		{
			name:  "window name with spaces",
			input: "my window @3\n",
			expect: []Window{
				{Name: "my window", ID: "@3"},
			},
		},
		{
			name:   "no id field",
			input:  "onlyaname\n",
			expect: nil,
		},
		{
			name:  "trailing whitespace",
			input: "  main @0  \n",
			expect: []Window{
				{Name: "main", ID: "@0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseWindows(tt.input)
			if tt.expect == nil {
				require.Empty(t, got)
				return
			}
			assert.Equal(t, tt.expect, got)
		})
	}
}

func TestIsShell(t *testing.T) {
	shellNames := []string{"bash", "zsh", "sh", "fish", "dash", "ksh", "tcsh", "csh"}
	for _, name := range shellNames {
		t.Run(name, func(t *testing.T) {
			assert.True(t, IsShell(name))
		})
	}

	nonShells := []string{"vim", "nvim", "python", "node", "go", "tmux", "", "BASH", "Zsh"}
	for _, name := range nonShells {
		t.Run(name, func(t *testing.T) {
			assert.False(t, IsShell(name))
		})
	}
}

func TestExecClient_InTmux(t *testing.T) {
	c := NewExecClient("")

	t.Run("not in tmux", func(t *testing.T) {
		t.Setenv("TMUX", "")
		assert.False(t, c.InTmux())
	})

	t.Run("in tmux", func(t *testing.T) {
		t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
		assert.True(t, c.InTmux())
	})
}

func TestNewExecClient(t *testing.T) {
	t.Run("default bin", func(t *testing.T) {
		c := NewExecClient("")
		assert.Equal(t, "tmux", c.bin)
	})

	t.Run("custom bin", func(t *testing.T) {
		c := NewExecClient("/usr/local/bin/tmux")
		assert.Equal(t, "/usr/local/bin/tmux", c.bin)
	})
}
