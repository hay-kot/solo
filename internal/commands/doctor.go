package commands

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// DoctorCmd implements the doctor command
type DoctorCmd struct {
	flags *Flags
}

// NewDoctorCmd creates a new doctor command
func NewDoctorCmd(flags *Flags) *DoctorCmd {
	return &DoctorCmd{flags: flags}
}

// Register adds the doctor command to the application
func (cmd *DoctorCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "doctor",
		Usage: "doctor command",
		Flags: []cli.Flag{
			// Add command-specific flags here
		},
		Action: cmd.run,
	})

	return app
}

func (cmd *DoctorCmd) run(ctx context.Context, c *cli.Command) error {
	log.Info().Msg("running doctor command")

	fmt.Println("Hello World!")

	return nil
}
