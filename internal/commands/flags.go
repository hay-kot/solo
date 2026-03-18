package commands

import "github.com/hay-kot/solo/internal/config"

// Flags holds global flags shared across all commands
type Flags struct {
	LogLevel   string
	NoColor    bool
	LogFile    string
	ConfigFile string
	Config     config.Config
}
