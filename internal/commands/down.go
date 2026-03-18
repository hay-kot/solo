package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/tmux"
	"github.com/hay-kot/solo/internal/ui"
)

const defaultDownTimeout = 10 * time.Second

// DownCmd implements the down command.
type DownCmd struct {
	flags  *Flags
	client tmux.Client
	out    io.Writer
}

// NewDownCmd creates a new down command.
func NewDownCmd(flags *Flags, client tmux.Client) *DownCmd {
	return &DownCmd{flags: flags, client: client, out: os.Stderr}
}

// Register adds the down command to the application.
func (cmd *DownCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "down",
		Usage:  "Tear down tmux windows for the current project",
		Flags:  []cli.Flag{},
		Action: cmd.run,
	})

	return app
}

func (cmd *DownCmd) run(ctx context.Context, _ *cli.Command) error {
	if !cmd.client.InTmux() {
		return errors.New("not inside a tmux session")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	result, err := config.ResolveProject(cwd, cmd.flags.Config)
	if err != nil {
		return fmt.Errorf("resolving project: %w", err)
	}

	timeout := defaultDownTimeout
	if result.Project.Timeout > 0 {
		timeout = time.Duration(result.Project.Timeout) * time.Second
	}

	existing, err := cmd.client.ListWindows(ctx)
	if err != nil {
		return fmt.Errorf("listing windows: %w", err)
	}

	windowsByName := make(map[string]tmux.Window, len(existing))
	for _, w := range existing {
		windowsByName[w.Name] = w
	}

	spin := ui.NewSpinner(cmd.out, cmd.flags.NoColor)
	var killed int

	for i := len(result.Project.Tabs) - 1; i >= 0; i-- {
		tab := result.Project.Tabs[i]
		win, found := windowsByName[tab.Title]
		if !found {
			spin.Warn(fmt.Sprintf("Window %s not found, skipping", tab.Title))
			continue
		}

		spin.Start(fmt.Sprintf("Sending interrupt to %s...", tab.Title))

		// Send C-c to interrupt any running process; ignore errors since the
		// process may have already exited.
		_ = cmd.client.SendKeys(ctx, win.ID, "C-c")

		cmd.waitForExit(ctx, spin, win.ID, tab.Title, timeout)

		if err := cmd.client.KillWindow(ctx, win.ID); err != nil {
			return fmt.Errorf("killing window %q: %w", tab.Title, err)
		}

		killed++
	}

	_, _ = fmt.Fprintf(cmd.out, "Torn down %d/%d windows\n", killed, len(result.Project.Tabs))

	return nil
}

// waitForExit polls the pane command until it returns to a shell or the timeout
// expires. Polling happens at 500ms intervals.
func (cmd *DownCmd) waitForExit(ctx context.Context, spin *ui.Spinner, target, name string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			spin.Warn(fmt.Sprintf("%s timed out, killed", name))
			return
		case <-ticker.C:
			paneCmd, err := cmd.client.ListPaneCommand(ctx, target)
			if err != nil {
				spin.Warn(fmt.Sprintf("%s error checking pane", name))
				return
			}

			if paneCmd == "" || tmux.IsShell(paneCmd) {
				spin.Stop(fmt.Sprintf("%s stopped", name))
				return
			}

			spin.Update(fmt.Sprintf("Waiting for %s to exit...", name))
		}
	}
}
