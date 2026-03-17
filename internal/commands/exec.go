package commands

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// ExecCmd implements the exec command
type ExecCmd struct {
	flags *Flags
}

// NewExecCmd creates a new exec command
func NewExecCmd(flags *Flags) *ExecCmd {
	return &ExecCmd{flags: flags}
}

// Register adds the exec command to the application
func (cmd *ExecCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "exec",
		Usage: "exec command",
		Flags: []cli.Flag{
			// Add command-specific flags here
		},
		Action: cmd.run,
	})

	return app
}

func (cmd *ExecCmd) run(ctx context.Context, c *cli.Command) error {
	log.Info().Msg("running exec command")

	fmt.Println("Hello World!")

	return nil
}
