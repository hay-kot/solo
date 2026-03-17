package commands

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/tmux"
)

// UpCmd implements the up command.
type UpCmd struct {
	flags  *Flags
	client tmux.Client
}

// NewUpCmd creates a new up command.
func NewUpCmd(flags *Flags, client tmux.Client) *UpCmd {
	return &UpCmd{flags: flags, client: client}
}

// Register adds the up command to the application.
func (cmd *UpCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "up",
		Usage:  "Create tmux windows for the current project",
		Flags:  []cli.Flag{},
		Action: cmd.run,
	})

	return app
}

func (cmd *UpCmd) run(ctx context.Context, _ *cli.Command) error {
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

	originalWindow, err := cmd.client.CurrentWindow(ctx)
	if err != nil {
		return fmt.Errorf("getting current window: %w", err)
	}

	existing, err := cmd.client.ListWindows(ctx)
	if err != nil {
		return fmt.Errorf("listing windows: %w", err)
	}

	existingNames := make(map[string]struct{}, len(existing))
	for _, w := range existing {
		existingNames[w.Name] = struct{}{}
	}

	var created int

	for _, tab := range result.Project.Tabs {
		if _, exists := existingNames[tab.Title]; exists {
			log.Warn().Str("window", tab.Title).Msg("window already exists, skipping")
			continue
		}

		winID, err := cmd.client.NewWindow(ctx, tab.Title)
		if err != nil {
			return fmt.Errorf("creating window %q: %w", tab.Title, err)
		}

		if tab.Cmd != "" {
			if err := cmd.client.SendKeys(ctx, winID, tab.Cmd, "Enter"); err != nil {
				return fmt.Errorf("sending keys to %q: %w", tab.Title, err)
			}
		}

		created++
	}

	if err := cmd.client.SelectWindow(ctx, originalWindow); err != nil {
		return fmt.Errorf("selecting original window: %w", err)
	}

	log.Info().Int("created", created).Int("total", len(result.Project.Tabs)).Msg("windows created")

	return nil
}
