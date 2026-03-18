package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		cfg         Config
		wantSource  string
		wantTabs    int
		wantTimeout int
		wantErr     bool
	}{
		{
			name: "local .solo.yml found",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, ".solo.yml"), `
tabs:
  - title: server
    cmd: go run .
  - title: docker
    cmd: docker compose up
timeout: 5
`)
				return dir
			},
			cfg:        Config{},
			wantSource: "local",
			wantTabs:   2,
		},
		{
			name: "local .solo.yaml found",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, ".solo.yaml"), `
tabs:
  - title: tests
    cmd: go test ./...
`)
				return dir
			},
			cfg:        Config{},
			wantSource: "local",
			wantTabs:   1,
		},
		{
			name: ".solo.yml takes precedence over .solo.yaml",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, ".solo.yml"), `
tabs:
  - title: yml-tab
    cmd: echo yml
`)
				writeFile(t, filepath.Join(dir, ".solo.yaml"), `
tabs:
  - title: yaml-tab1
    cmd: echo yaml1
  - title: yaml-tab2
    cmd: echo yaml2
`)
				return dir
			},
			cfg:        Config{},
			wantSource: "local",
			wantTabs:   1, // from .solo.yml, not .solo.yaml
		},
		{
			name: "global exact match",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			cfg: Config{
				Projects: map[string]Project{
					"PLACEHOLDER": {
						Tabs: []Tab{
							{Title: "dev", Cmd: "make dev"},
							{Title: "logs", Cmd: "tail -f log"},
						},
					},
				},
			},
			wantSource: "global-exact",
			wantTabs:   2,
		},
		{
			name: "basename exact match",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "recipinned")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"recipinned": {
						Tabs: []Tab{
							{Title: "dev", Cmd: "make dev"},
						},
					},
				},
			},
			wantSource: "global-basename",
			wantTabs:   1,
		},
		{
			name: "basename glob match",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "recipinned-v2")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"recipinned*": {
						Tabs: []Tab{
							{Title: "dev", Cmd: "make dev"},
							{Title: "logs", Cmd: "tail -f log"},
						},
					},
				},
			},
			wantSource: "global-glob",
			wantTabs:   2,
		},
		{
			name: "longest glob wins",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "recipinned-v2")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"rec*": {
						Tabs: []Tab{
							{Title: "short", Cmd: "echo short"},
						},
					},
					"recipinned*": {
						Tabs: []Tab{
							{Title: "long", Cmd: "echo long"},
							{Title: "long2", Cmd: "echo long2"},
						},
					},
				},
			},
			wantSource: "global-glob",
			wantTabs:   2, // longest pattern "recipinned*" wins
		},
		{
			name: "glob with question mark",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "app1")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"app?": {
						Tabs: []Tab{
							{Title: "dev", Cmd: "make dev"},
						},
					},
				},
			},
			wantSource: "global-glob",
			wantTabs:   1,
		},
		{
			name: "glob question mark does not match long suffix",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "application")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"app?": {
						Tabs: []Tab{
							{Title: "dev", Cmd: "make dev"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "full path exact takes precedence over basename",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "myproject")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"PLACEHOLDER": {
						Tabs: []Tab{
							{Title: "exact", Cmd: "echo exact"},
							{Title: "exact2", Cmd: "echo exact2"},
						},
					},
					"myproject": {
						Tabs: []Tab{
							{Title: "basename", Cmd: "echo basename"},
						},
					},
				},
			},
			wantSource: "global-exact",
			wantTabs:   2,
		},
		{
			name: "local takes precedence over basename",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				sub := filepath.Join(dir, "myproject")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				writeFile(t, filepath.Join(sub, ".solo.yml"), `
tabs:
  - title: local
    cmd: echo local
`)
				return sub
			},
			cfg: Config{
				Projects: map[string]Project{
					"myproject": {
						Tabs: []Tab{
							{Title: "basename", Cmd: "echo basename"},
							{Title: "basename2", Cmd: "echo basename2"},
						},
					},
				},
			},
			wantSource: "local",
			wantTabs:   1,
		},
		{
			name: "no match returns error",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			cfg:     Config{},
			wantErr: true,
		},
		{
			name: "local file with timeout",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, ".solo.yml"), `
tabs:
  - title: app
    cmd: ./run.sh
timeout: 30
`)
				return dir
			},
			cfg:         Config{},
			wantSource:  "local",
			wantTabs:    1,
			wantTimeout: 30,
		},
		{
			name: "local file takes precedence over global exact",
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, ".solo.yml"), `
tabs:
  - title: local
    cmd: echo local
`)
				return dir
			},
			cfg: Config{
				Projects: map[string]Project{
					"PLACEHOLDER": {
						Tabs: []Tab{
							{Title: "global", Cmd: "echo global"},
							{Title: "global2", Cmd: "echo global2"},
						},
					},
				},
			},
			wantSource: "local",
			wantTabs:   1, // local wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := tt.setupDir(t)

			// Replace PLACEHOLDER keys with actual temp dir paths.
			cfg := tt.cfg
			if cfg.Projects != nil {
				resolved := make(map[string]Project, len(cfg.Projects))
				for key, proj := range cfg.Projects {
					switch key {
					case "PLACEHOLDER":
						resolved[dir] = proj
					default:
						resolved[key] = proj
					}
				}
				cfg.Projects = resolved
			}

			result, err := ResolveProject(dir, cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSource, result.Source)
			assert.Len(t, result.Project.Tabs, tt.wantTabs)

			if tt.wantTimeout > 0 {
				assert.Equal(t, tt.wantTimeout, result.Project.Timeout)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
