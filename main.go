package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/hay-kot/solo/internal/commands"
	"github.com/hay-kot/solo/internal/config"
	"github.com/hay-kot/solo/internal/paths"
	"github.com/hay-kot/solo/internal/tmux"
)

var (
	// Build information. Populated at build-time via -ldflags flag.
	version = "dev"
	commit  = "HEAD"
	date    = "now"
)

func build() string {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					commit = s.Value
				case "vcs.time":
					date = s.Value
				}
			}
		}
	}

	short := commit
	if len(commit) > 7 {
		short = commit[:7]
	}

	return fmt.Sprintf("%s (%s) %s", version, short, date)
}

// soloEnv returns a cli.EnvVars source with the SOLO_ prefix and any
// additional unprefixed aliases (e.g. NO_COLOR).
func soloEnv(name string, aliases ...string) cli.ValueSourceChain {
	vars := make([]string, 0, 1+len(aliases))
	vars = append(vars, "SOLO_"+name)
	vars = append(vars, aliases...)
	return cli.EnvVars(vars...)
}

func setupLogger(level string, logFile string, noColor bool) error {
	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	var output io.Writer = zerolog.ConsoleWriter{Out: os.Stderr, NoColor: noColor}

	if logFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		// Write to both console and file
		output = io.MultiWriter(
			zerolog.ConsoleWriter{Out: os.Stderr, NoColor: noColor},
			zerolog.ConsoleWriter{Out: file, NoColor: true},
		)
	}

	log.Logger = log.Output(output).Level(parsedLevel)

	return nil
}

func main() {
	os.Exit(run())
}

func run() int {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flags := &commands.Flags{}

	app := &cli.Command{
		Name:                  "solo",
		Usage:                 `A set of utilities for my preferred development`,
		Version:               build(),
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "log level (debug, info, warn, error, fatal, panic)",
				Sources:     soloEnv("LOG_LEVEL"),
				Value:       "info",
				Destination: &flags.LogLevel,
			},
			&cli.BoolFlag{
				Name:        "no-color",
				Usage:       "disable colored output",
				Sources:     soloEnv("NO_COLOR", "NO_COLOR"),
				Destination: &flags.NoColor,
			},
			&cli.StringFlag{
				Name:        "log-file",
				Usage:       "path to log file (optional)",
				Sources:     soloEnv("LOG_FILE"),
				Destination: &flags.LogFile,
			},
			&cli.StringFlag{
				Name:        "config",
				Usage:       "path to config file",
				Sources:     soloEnv("CONFIG_FILE"),
				Destination: &flags.ConfigFile,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			cfg, err := func() (config.Config, error) {
				if flags.ConfigFile != "" {
					return config.ReadFrom(flags.ConfigFile)
				}
				return config.Read()
			}()
			if err != nil {
				return ctx, fmt.Errorf("loading config: %w", err)
			}

			flags.Config = cfg

			if flags.LogLevel == "info" && cfg.LogLevel != "" {
				flags.LogLevel = cfg.LogLevel
			}
			if flags.LogFile == "" && cfg.LogFile != "" {
				flags.LogFile = cfg.LogFile
			}
			logFile := flags.LogFile
			if logFile == "" {
				logFile = filepath.Join(paths.DataDir(), "solo.log")
			}

			if err := setupLogger(flags.LogLevel, logFile, flags.NoColor); err != nil {
				return ctx, err
			}

			return ctx, nil
		},
	}
	app = commands.NewUpCmd(flags, tmux.NewExecClient("")).Register(app)
	app = commands.NewDownCmd(flags, tmux.NewExecClient("")).Register(app)
	app = commands.NewConfigCmd(flags).Register(app)
	app = commands.NewDoctorCmd(flags).Register(app)
	// +scaffold:command:register
	app = commands.NewLLMTextCmd(flags).Register(app)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, os.Args); err != nil {
		colorRed := "\033[38;2;215;95;107m"
		colorGray := "\033[38;2;163;163;163m"
		colorReset := "\033[0m"
		if flags.NoColor {
			colorRed = ""
			colorGray = ""
			colorReset = ""
		}
		fmt.Fprintf(os.Stderr, "\n%s╭ Error%s\n%s│%s %s%s%s\n%s╵%s\n",
			colorRed, colorReset,
			colorRed, colorReset, colorGray, err.Error(), colorReset,
			colorRed, colorReset,
		)
		return 1
	}

	return 0
}
