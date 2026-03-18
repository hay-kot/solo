package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/hay-kot/solo/internal/paths"
)

// Tab represents a terminal tab configuration with a title and command.
type Tab struct {
	Title string `yaml:"title"`
	Cmd   string `yaml:"cmd"`
}

// Project defines a project configuration with tabs and a shutdown timeout.
type Project struct {
	Dir     string `yaml:"dir"`
	Tabs    []Tab  `yaml:"tabs"`
	Timeout int    `yaml:"timeout"` // seconds for down command, default 10
}

// Config holds the application configuration loaded from a YAML file.
type Config struct {
	LogLevel string             `yaml:"log_level"`
	LogFile  string             `yaml:"log_file"`
	Projects map[string]Project `yaml:"projects"`
}

// Default returns a Config with default values.
func Default() Config {
	return Config{
		LogLevel: "info",
	}
}

// Read loads config from the default XDG config path.
// Returns default config if the file does not exist.
func Read() (Config, error) {
	return ReadFrom(filepath.Join(paths.ConfigDir(), "config.yaml"))
}

// ReadFrom loads config from the given file path.
// Returns default config if the file does not exist.
func ReadFrom(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}
