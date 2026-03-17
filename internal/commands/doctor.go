package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/paths"
)

// DoctorCmd implements the doctor command.
type DoctorCmd struct {
	flags *Flags

	// out is the writer for command output. Defaults to os.Stdout.
	out io.Writer
}

// NewDoctorCmd creates a new doctor command.
func NewDoctorCmd(flags *Flags) *DoctorCmd {
	return &DoctorCmd{flags: flags, out: os.Stdout}
}

// Register adds the doctor command to the application.
func (cmd *DoctorCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "doctor",
		Usage:  "Check environment and configuration",
		Flags:  []cli.Flag{},
		Action: cmd.run,
	})

	return app
}

func (cmd *DoctorCmd) run(_ context.Context, _ *cli.Command) error {
	cmd.checkTmuxInstalled()
	cmd.checkTmuxSession()
	cmd.checkGlobalConfig()
	cmd.checkProjectConfig()

	return nil
}

func (cmd *DoctorCmd) writef(format string, args ...any) {
	_, _ = fmt.Fprintf(cmd.out, format, args...)
}

func (cmd *DoctorCmd) checkTmuxInstalled() {
	path, err := exec.LookPath("tmux")
	if err != nil {
		cmd.writef("[warn] tmux not found in PATH\n")
		return
	}

	cmd.writef("[pass] tmux installed: %s\n", path)
}

func (cmd *DoctorCmd) checkTmuxSession() {
	if os.Getenv("TMUX") != "" {
		cmd.writef("[pass] inside tmux session\n")
		return
	}

	cmd.writef("[info] not inside tmux session\n")
}

func (cmd *DoctorCmd) checkGlobalConfig() {
	cfgPath := filepath.Join(paths.ConfigDir(), "config.yaml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cmd.writef("[info] global config not found: %s\n", cfgPath)
		return
	}

	if _, err := config.ReadFrom(cfgPath); err != nil {
		cmd.writef("[warn] global config parse error: %v\n", err)
		return
	}

	cmd.writef("[pass] global config: %s\n", cfgPath)
}

func (cmd *DoctorCmd) checkProjectConfig() {
	cwd, err := os.Getwd()
	if err != nil {
		cmd.writef("[warn] cannot determine working directory: %v\n", err)
		return
	}

	result, err := config.ResolveProject(cwd, cmd.flags.Config)
	if err != nil {
		cmd.writef("[info] no project config found for current directory\n")
		return
	}

	cmd.writef("[pass] project config: %s\n", formatSource(result.Source))
}
