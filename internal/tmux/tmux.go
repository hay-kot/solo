package tmux

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Window represents a tmux window with its name and ID.
type Window struct {
	Name string
	ID   string // e.g. "@1"
}

// Client defines the interface for interacting with tmux.
type Client interface {
	InTmux() bool
	NewWindow(ctx context.Context, name string) error
	SendKeys(ctx context.Context, target string, keys ...string) error
	ListWindows(ctx context.Context) ([]Window, error)
	ListPaneCommand(ctx context.Context, target string) (string, error)
	KillWindow(ctx context.Context, target string) error
	SelectWindow(ctx context.Context, target string) error
	CurrentWindow(ctx context.Context) (string, error)
}

// ExecClient implements Client by shelling out to the tmux binary.
type ExecClient struct {
	bin string
}

// NewExecClient returns a new ExecClient that uses the given tmux binary path.
// If bin is empty, "tmux" is used.
func NewExecClient(bin string) *ExecClient {
	if bin == "" {
		bin = "tmux"
	}
	return &ExecClient{bin: bin}
}

// InTmux reports whether the current process is running inside a tmux session.
func (c *ExecClient) InTmux() bool {
	return os.Getenv("TMUX") != ""
}

// NewWindow creates a new tmux window with the given name.
func (c *ExecClient) NewWindow(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, c.bin, "new-window", "-n", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("new-window %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SendKeys sends keystrokes to the target tmux window/pane. Each key is passed
// as a separate argument to tmux send-keys. The caller must explicitly include
// "Enter" when a carriage return is needed.
func (c *ExecClient) SendKeys(ctx context.Context, target string, keys ...string) error {
	args := []string{"send-keys", "-t", target}
	args = append(args, keys...)
	cmd := exec.CommandContext(ctx, c.bin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("send-keys %q: %s: %w", target, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// ListWindows returns all windows in the current tmux session.
func (c *ExecClient) ListWindows(ctx context.Context) ([]Window, error) {
	cmd := exec.CommandContext(ctx, c.bin, "list-windows", "-F", "#{window_name} #{window_id}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list-windows: %w", err)
	}
	return parseWindows(string(out)), nil
}

// ListPaneCommand returns the current command running in the target window's pane.
func (c *ExecClient) ListPaneCommand(ctx context.Context, target string) (string, error) {
	cmd := exec.CommandContext(ctx, c.bin, "list-panes", "-t", target, "-F", "#{pane_current_command}")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("list-panes %q: %w", target, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// KillWindow destroys the target tmux window.
func (c *ExecClient) KillWindow(ctx context.Context, target string) error {
	cmd := exec.CommandContext(ctx, c.bin, "kill-window", "-t", target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kill-window %q: %s: %w", target, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SelectWindow focuses the target tmux window.
func (c *ExecClient) SelectWindow(ctx context.Context, target string) error {
	cmd := exec.CommandContext(ctx, c.bin, "select-window", "-t", target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("select-window %q: %s: %w", target, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// CurrentWindow returns the name of the currently active tmux window.
func (c *ExecClient) CurrentWindow(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, c.bin, "display-message", "-p", "#{window_name}")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("display-message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// parseWindows parses the output of `tmux list-windows -F '#{window_name} #{window_id}'`
// into a slice of Window values.
func parseWindows(output string) []Window {
	var windows []Window
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// The last space-separated field is the window ID (e.g. @1).
		// Everything before it is the window name.
		idx := strings.LastIndex(line, " ")
		if idx < 0 {
			continue
		}
		windows = append(windows, Window{
			Name: line[:idx],
			ID:   line[idx+1:],
		})
	}
	return windows
}

// shells is the set of common shell binary names.
var shells = map[string]struct{}{
	"bash": {},
	"zsh":  {},
	"sh":   {},
	"fish": {},
	"dash": {},
	"ksh":  {},
	"tcsh": {},
	"csh":  {},
}

// IsShell reports whether cmd is a known shell command name.
func IsShell(cmd string) bool {
	_, ok := shells[cmd]
	return ok
}
