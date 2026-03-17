package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/hay-kot/solo/internal/config"
)

// ConfigCmd implements the config command.
type ConfigCmd struct {
	flags *Flags

	// out is the writer for command output. Defaults to os.Stdout.
	out io.Writer
}

// NewConfigCmd creates a new config command.
func NewConfigCmd(flags *Flags) *ConfigCmd {
	return &ConfigCmd{flags: flags, out: os.Stdout}
}

// Register adds the config command to the application.
func (cmd *ConfigCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "config",
		Usage:  "Show resolved project configuration",
		Flags:  []cli.Flag{},
		Action: cmd.run,
	})

	return app
}

func (cmd *ConfigCmd) run(_ context.Context, _ *cli.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	result, err := config.ResolveProject(cwd, cmd.flags.Config)
	if err != nil {
		return fmt.Errorf("resolving project: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.out, "Source: %s\n", formatSource(result.Source))
	_, _ = fmt.Fprintf(cmd.out, "---\n")

	data, err := yaml.Marshal(result.Project)
	if err != nil {
		return fmt.Errorf("marshaling project config: %w", err)
	}

	_, _ = fmt.Fprint(cmd.out, string(data))

	return nil
}

func formatSource(source string) string {
	switch source {
	case "local":
		return "local"
	case "global-exact":
		return "global (exact match)"
	case "global-basename":
		return "global (basename match)"
	case "global-glob":
		return "global (glob match)"
	default:
		return source
	}
}
