package commands

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// ConfigCmd implements the config command
type ConfigCmd struct {
	flags *Flags
}

// NewConfigCmd creates a new config command
func NewConfigCmd(flags *Flags) *ConfigCmd {
	return &ConfigCmd{flags: flags}
}

// Register adds the config command to the application
func (cmd *ConfigCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "config",
		Usage: "config command",
		Flags: []cli.Flag{
			// Add command-specific flags here
		},
		Action: cmd.run,
	})

	return app
}

func (cmd *ConfigCmd) run(ctx context.Context, c *cli.Command) error {
	log.Info().Msg("running config command")

	fmt.Println("Hello World!")

	return nil
}
