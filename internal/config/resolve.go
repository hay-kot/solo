package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ResolveResult holds the resolved project configuration and its source.
type ResolveResult struct {
	Project Project
	Source  string // "local", "global-exact", "global-basename", "global-glob"
}

// ResolveProject resolves project configuration for the given directory.
//
// Resolution order:
//  1. .solo.yml in dir (source: "local")
//  2. .solo.yaml in dir (source: "local")
//  3. Exact match in cfg.Projects on full path (source: "global-exact")
//  4. Exact match in cfg.Projects on basename (source: "global-basename")
//  5. Glob match in cfg.Projects on basename, longest pattern wins (source: "global-glob")
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

	// Try global exact match on full path.
	if proj, ok := cfg.Projects[dir]; ok {
		return ResolveResult{
			Project: proj,
			Source:  "global-exact",
		}, nil
	}

	basename := filepath.Base(dir)

	// Try global exact match on basename.
	if proj, ok := cfg.Projects[basename]; ok {
		return ResolveResult{
			Project: proj,
			Source:  "global-basename",
		}, nil
	}

	// Try glob match on basename, longest pattern wins.
	var bestKey string

	for key := range cfg.Projects {
		matched, err := filepath.Match(key, basename)
		if err != nil {
			continue
		}

		if matched && len(key) > len(bestKey) {
			bestKey = key
		}
	}

	if bestKey != "" {
		return ResolveResult{
			Project: cfg.Projects[bestKey],
			Source:  "global-glob",
		}, nil
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
