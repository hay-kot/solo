# solo

A tmux workspace manager. Define per-project tabs and commands in a config file, then spin them up or tear them down with a single command.

## Installation

```bash
go install github.com/hay-kot/solo@latest
```

## Usage

### Commands

| Command  | Description                                  |
| -------- | -------------------------------------------- |
| `up`     | Create tmux windows for the current project  |
| `down`   | Tear down tmux windows for the current project |
| `config` | Show resolved project configuration          |
| `doctor` | Check environment and configuration          |

### Quick Start

1. Create a `.solo.yaml` in your project directory:

```yaml
tabs:
  - title: editor
    cmd: nvim
  - title: server
    cmd: go run ./cmd/server
  - title: shell
```

2. Inside a tmux session, run:

```bash
solo up     # creates the windows
solo down   # tears them down
```

### Configuration

Solo resolves project configuration in the following order:

1. `.solo.yml` or `.solo.yaml` in the current directory
2. Global config match by full path
3. Global config match by directory basename
4. Global config glob match on basename (longest pattern wins)

The global config file lives at `$XDG_CONFIG_HOME/solo/config.yaml` (defaults to `~/.config/solo/config.yaml`):

```yaml
log_level: info

projects:
  # exact directory name match
  myproject:
    tabs:
      - title: editor
        cmd: nvim
      - title: server
        cmd: make run

  # glob match
  "api-*":
    tabs:
      - title: server
        cmd: go run .
      - title: logs
        cmd: tail -f /var/log/app.log
```

### Global Flags

| Flag          | Env Var            | Description                  |
| ------------- | ------------------ | ---------------------------- |
| `--log-level` | `SOLO_LOG_LEVEL`   | Log level (default: `info`)  |
| `--no-color`  | `SOLO_NO_COLOR`    | Disable colored output       |
| `--log-file`  | `SOLO_LOG_FILE`    | Path to log file             |
| `--config`    | `SOLO_CONFIG_FILE` | Path to config file          |
