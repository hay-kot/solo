package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/paths"
)

const llmTextTemplate = `# solo - tmux session manager

solo manages tmux windows for development projects. It reads a configuration
file to know which windows to open and what commands to run in each.

## Global Config File

Location: %s

This is a YAML file. If it does not exist, create it (and any parent directories).

## Config File Format

` + "```yaml" + `
log_level: info        # optional: debug, info, warn, error
log_file: ""           # optional: path to log file

projects:
  # Key is the project name. It matches against the directory basename first,
  # then the full path, then glob patterns (longest match wins).
  my-project:
    tabs:
      - title: editor  # tmux window title
        cmd: nvim       # command to run (optional — omit for a bare shell)
      - title: server
        cmd: make run
      - title: shell   # window with no command, just a shell
    timeout: 10        # seconds to wait on 'solo down' before force-kill (default: 10)

  # Glob patterns are supported — longest match wins
  "api-*":
    tabs:
      - title: server
        cmd: go run .
      - title: logs
        cmd: tail -f app.log
` + "```" + `

## Adding an Entry for the Current Project

The current project directory is: %s
The directory basename is:        %s

To add this project, insert a new entry under the 'projects' key using the
basename as the key (or the full path for an exact match):

` + "```yaml" + `
projects:
  %s:
    tabs:
      - title: editor
        cmd: nvim
      - title: shell
` + "```" + `

## Project Resolution Order

When solo runs in a directory, it looks for config in this order:
1. .solo.yml or .solo.yaml in the current directory (local config)
2. Exact full-path match in the global config
3. Exact basename match in the global config
4. Glob pattern match on basename (longest pattern wins)
5. No config found — error

## Key Commands

  solo up      Create tmux windows for the current project
  solo down    Gracefully shut down and close those windows
  solo config  Show the resolved config for the current directory
  solo doctor  Check dependencies, environment, and config health

## Local Config (Alternative to Global)

Instead of editing the global config, you can create .solo.yml in the project
root with the same structure as a single project entry:

` + "```yaml" + `
tabs:
  - title: editor
    cmd: nvim
  - title: server
    cmd: make run
timeout: 10
` + "```" + `
`

// LLMTextCmd implements the llmtext command.
type LLMTextCmd struct {
	flags *Flags
	out   io.Writer
}

// NewLLMTextCmd creates a new llmtext command.
func NewLLMTextCmd(flags *Flags) *LLMTextCmd {
	return &LLMTextCmd{flags: flags, out: os.Stdout}
}

// Register adds the llmtext command to the application.
func (cmd *LLMTextCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "llmtext",
		Usage:  "Print an LLM-friendly description of solo and how to configure the current project",
		Action: cmd.run,
	})

	return app
}

func (cmd *LLMTextCmd) run(_ context.Context, _ *cli.Command) error {
	configPath := filepath.Join(paths.ConfigDir(), "config.yaml")

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	basename := filepath.Base(cwd)

	_, err = fmt.Fprintf(cmd.out, llmTextTemplate, configPath, cwd, basename, basename)
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
