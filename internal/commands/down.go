package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/tmux"
)

const defaultDownTimeout = 10 * time.Second

// DownCmd implements the down command.
type DownCmd struct {
	flags  *Flags
	client tmux.Client
}

// NewDownCmd creates a new down command.
func NewDownCmd(flags *Flags, client tmux.Client) *DownCmd {
	return &DownCmd{flags: flags, client: client}
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

	log.Debug().Str("source", result.Source).Msg("resolved project config")

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

	var killed int

	for _, tab := range result.Project.Tabs {
		win, found := windowsByName[tab.Title]
		if !found {
			log.Warn().Str("window", tab.Title).Msg("window not found, skipping")
			continue
		}

		// Send C-c to interrupt any running process; ignore errors since the
		// process may have already exited.
		_ = cmd.client.SendKeys(ctx, win.ID, "C-c")

		cmd.waitForExit(ctx, win.ID, timeout)

		if err := cmd.client.KillWindow(ctx, win.ID); err != nil {
			return fmt.Errorf("killing window %q: %w", tab.Title, err)
		}

		killed++
	}

	log.Info().Int("killed", killed).Int("total", len(result.Project.Tabs)).Msg("windows torn down")

	return nil
}

// waitForExit polls the pane command until it returns to a shell or the timeout
// expires. Polling happens at 500ms intervals.
func (cmd *DownCmd) waitForExit(ctx context.Context, target string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warn().Str("target", target).Msg("timeout waiting for process to exit")
			return
		case <-ticker.C:
			paneCmd, err := cmd.client.ListPaneCommand(ctx, target)
			if err != nil {
				log.Warn().Err(err).Str("target", target).Msg("error checking pane command")
				return
			}

			if tmux.IsShell(paneCmd) {
				return
			}
		}
	}
}
