# Pomodoro CLI Timer

A simple, interactive, keyboard-driven Pomodoro timer for the terminal.

<p align="center">
  <img src="assets/Initial%20Screen.png" alt="Setup Screen" width="40%">
  <!--&nbsp; &nbsp;-->
  <img src="assets/timer.png" alt="Timer Running" width="45%">
</p>

## Features

- Clean, colorful terminal interface with ASCII timer display
- Interactive setup menu and quick-start via command-line arguments
- Pause, resume, skip, and adjust time on the fly
- Desktop notifications when sessions end (Windows, macOS, Linux)
- Sound alerts for session changes
- Fully customizable work/break durations and number of sessions

## Installation

### Option 1: Go Install

If you have Go installed, you can install it directly:

```bash
go install github.com/nirabyte/pomo@latest
```

### Option 2: Build from Source

Clone the repo and build the binary.

```bash
git clone https://github.com/niraletter/pomodoro-cli.git
cd pomo
go build -o pomo main.go
```

### Option 3: Cross-platform Builds

Use the provided Makefile to build for all platforms:

```bash
make build-all    # Build for all platforms
make build-linux  # Build Linux variants only
make build-macos  # Build macOS variants only
make build-windows # Build Windows variants only
make clean        # Remove all built binaries
make help         # Show available commands
```

_Builds are optimized for size using `-ldflags="-s -w"` and `-trimpath` flags._

## Usage

### 1. Interactive Mode

Simply run the tool without arguments to enter the setup screen:

```bash
pomo
```

### 2. Quick Start (CLI Arguments)

Skip the setup and start the timer immediately.

**Syntax:** `pomo [work] [break] [sessions]`

```bash
# Start 25m work (defaults: 5m break, 4 sessions)
pomo 25m

# Start 50m work, 10m break
pomo 50m 10m

# Start 45m work, 15m break, 6 sessions
pomo 45m 15m 6
```

## Presets & Aliases

Create shell aliases for common configurations:

```bash
# Standard 25-minute sessions
alias pomo25='pomo 25m 5m 4'

# Long 50-minute focus sessions
alias pomo50='pomo 50m 10m 3'

# Quick 15-minute sprints
alias pomo15='pomo 15m 3m 6'

# Custom ultra-focus mode
alias pomofocus='pomo 90m 15m 2'
```

Or add to your shell config (`.bashrc`, `.zshrc`, etc.):

```bash
# Pomodoro presets
pomo25() { pomo 25m 5m 4; }
pomo50() { pomo 50m 10m 3; }
pomo15() { pomo 15m 3m 6; }
```

Then source your config file:

```bash
# For Bash
source ~/.bashrc

# For Zsh
source ~/.zshrc

```

## Key Bindings

Bind keyboard shortcuts in your window manager:

### Hyprland (`~/.config/hypr/hyprland.conf`)

```bash
# Pomodoro shortcuts
bind = $mainMod, P, exec, pomo 25m 5m 4
bind = $mainMod SHIFT, P, exec, pomo 50m 10m 3
bind = $mainMod CTRL, P, exec, pomo 15m 3m 6
```

### i3/Sway (`~/.config/i3/config` or `~/.config/sway/config`)

```bash
# Pomodoro shortcuts
bindsym $mod+p exec pomo 25m 5m 4
bindsym $mod+Shift+p exec pomo 50m 10m 3
bindsym $mod+Ctrl+p exec pomo 15m 3m 6
```

## Controls

### Setup Screen

| Key           | Action                        |
| :------------ | :---------------------------- |
| `TAB`         | Switch inputs                 |
| `Mouse wheel` | Change values (focused input) |
| `ENTER`       | Start Timer                   |
| `a`           | Toggle Autobreak              |
| `q`           | Quit                          |

### Timer Screen

| Key           | Action                   |
| :------------ | :----------------------- |
| `SPACE`       | Pause / Resume           |
| `s`           | **Skip** current session |
| `↑` / `↓`     | +/- 1 minute             |
| `Mouse wheel` | +/- 1 minute             |
| `q`           | Quit                     |

### Built With

- **[Go](https://go.dev/)**
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI Framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — Styling
- **[Beeep](https://github.com/gen2brain/beeep)** — Notifications

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
