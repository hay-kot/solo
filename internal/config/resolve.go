package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResolveResult holds the resolved project configuration and its source.
type ResolveResult struct {
	Project Project
	Source  string // "local", "global-exact", "global-prefix"
}

// ResolveProject resolves project configuration for the given directory.
//
// Resolution order:
//  1. .solo.yml in dir (source: "local")
//  2. .solo.yaml in dir (source: "local")
//  3. Exact match in cfg.Projects (source: "global-exact")
//  4. Longest prefix match in cfg.Projects (source: "global-prefix")
func ResolveProject(dir string, cfg Config) (ResolveResult, error) {
	// Try local files first.
	for _, name := range []string{".solo.yml", ".solo.yaml"} {
		result, ok, err := tryLocalFile(filepath.Join(dir, name))
		if err != nil {
			return ResolveResult{}, err
		}

		if ok {
			return result, nil
		}
	}

	// Try global exact match.
	if proj, ok := cfg.Projects[dir]; ok {
		return ResolveResult{
			Project: proj,
			Source:  "global-exact",
		}, nil
	}

	// Try longest prefix match.
	var bestKey string

	for key := range cfg.Projects {
		expanded := expandHome(key)
		if !strings.HasPrefix(dir, expanded) {
			continue
		}

		if len(expanded) > len(bestKey) {
			bestKey = expanded
		}
	}

	if bestKey != "" {
		// Find the original key that produced this expanded path.
		for key := range cfg.Projects {
			if expandHome(key) == bestKey {
				return ResolveResult{
					Project: cfg.Projects[key],
					Source:  "global-prefix",
				}, nil
			}
		}
	}

	return ResolveResult{}, fmt.Errorf("no project configuration found for %s", dir)
}

func tryLocalFile(path string) (ResolveResult, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ResolveResult{}, false, nil
		}

		return ResolveResult{}, false, fmt.Errorf("reading %s: %w", path, err)
	}

	var proj Project
	if err := yaml.Unmarshal(data, &proj); err != nil {
		return ResolveResult{}, false, fmt.Errorf("parsing %s: %w", path, err)
	}

	return ResolveResult{
		Project: proj,
		Source:  "local",
	}, true, nil
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path[2:])
}
