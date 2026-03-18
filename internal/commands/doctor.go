package commands

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/paths"
	"github.com/hay-kot/solo/internal/ui"
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
	report := ui.NewReport("Solo Doctor", cmd.flags.NoColor)

	report.AddSection(cmd.checkDependencies())
	report.AddSection(cmd.checkEnvironment())
	report.AddSection(cmd.checkConfiguration())
	report.AddSection(cmd.checkPaths())

	report.Render(cmd.out)
	return nil
}

func (cmd *DoctorCmd) checkDependencies() ui.Section {
	sec := ui.Section{Title: "Dependencies"}

	path, err := exec.LookPath("tmux")
	if err != nil {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusWarn,
			Label:  "tmux",
			Detail: "not found in PATH",
		})
	} else {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusPass,
			Label:  "tmux",
			Detail: path,
		})
	}

	return sec
}

func (cmd *DoctorCmd) checkEnvironment() ui.Section {
	sec := ui.Section{Title: "Environment"}

	if os.Getenv("TMUX") != "" {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusPass,
			Label:  "tmux session",
		})
	} else {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusInfo,
			Label:  "tmux session",
			Detail: "not inside a tmux session",
		})
	}

	return sec
}

func (cmd *DoctorCmd) checkConfiguration() ui.Section {
	sec := ui.Section{Title: "Configuration"}

	// Global config
	cfgPath := filepath.Join(paths.ConfigDir(), "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusInfo,
			Label:  "global config",
			Detail: "not found",
		})
	} else if _, err := config.ReadFrom(cfgPath); err != nil {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusWarn,
			Label:  "global config",
			Detail: "parse error",
		})
	} else {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusPass,
			Label:  "global config",
			Detail: cfgPath,
		})
	}

	// Project config
	cwd, err := os.Getwd()
	if err != nil {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusWarn,
			Label:  "project config",
			Detail: "cannot determine working directory",
		})
		return sec
	}

	result, err := config.ResolveProject(cwd, cmd.flags.Config)
	if err != nil {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusInfo,
			Label:  "project config",
			Detail: "no match for current directory",
		})
	} else {
		sec.Checks = append(sec.Checks, ui.Check{
			Status: ui.StatusPass,
			Label:  "project config",
			Detail: formatSource(result.Source),
		})
	}

	return sec
}

func (cmd *DoctorCmd) checkPaths() ui.Section {
	sec := ui.Section{Title: "Paths"}

	cfgPath := filepath.Join(paths.ConfigDir(), "config.yaml")
	sec.Checks = append(sec.Checks, ui.Check{
		Status: ui.StatusInfo,
		Label:  "config",
		Detail: cfgPath,
	})

	logFile := cmd.flags.LogFile
	if logFile == "" {
		logFile = filepath.Join(paths.DataDir(), "solo.log")
	}
	sec.Checks = append(sec.Checks, ui.Check{
		Status: ui.StatusInfo,
		Label:  "log file",
		Detail: logFile,
	})

	return sec
}
