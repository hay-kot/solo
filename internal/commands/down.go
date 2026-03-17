package commands

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// DownCmd implements the down command
type DownCmd struct {
	flags *Flags
}

// NewDownCmd creates a new down command
func NewDownCmd(flags *Flags) *DownCmd {
	return &DownCmd{flags: flags}
}

// Register adds the down command to the application
func (cmd *DownCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "down",
		Usage: "down command",
		Flags: []cli.Flag{
			// Add command-specific flags here
		},
		Action: cmd.run,
	})

	return app
}

func (cmd *DownCmd) run(ctx context.Context, c *cli.Command) error {
	log.Info().Msg("running down command")

	fmt.Println("Hello World!")

	return nil
}
