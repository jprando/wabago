# Wabago - Waybar Configuration TUI Editor

A beautiful Terminal User Interface (TUI) application for editing Waybar configurations, built with Go using the Charmbracelet libraries.

## Features

- **Bar Settings Editor**: Configure position, height, width, spacing, margins, and more
- **Module Management**: Add, remove, reorder, and edit modules for left, center, and right sections
- **Full Module Support**: All Waybar modules with their properties pre-configured
- **CSS Style Viewer**: View your Waybar CSS styles
- **Backup System**: Automatic backups before saving, with restore functionality
- **Keyboard-Driven UI**: Vim-style navigation (hjkl) and intuitive keybindings
- **Beautiful UI**: Nord-inspired color scheme with smooth navigation

## Supported Modules

### System
- Battery, CPU, Memory, Disk, Temperature, Load

### Audio/Media
- PulseAudio, WirePlumber, MPRIS

### Workspaces
- Hyprland (workspaces, window, submap, language)
- Sway (workspaces, window, mode, language)
- wlr/taskbar

### Network
- Network, Bluetooth

### Display
- Clock, Backlight, Tray, Idle Inhibitor, Image

### Input
- Keyboard State, Language

### Custom
- Custom script modules (custom/*)
- Group modules (group/*)

### Other
- User, Gamemode, Privacy, Systemd Failed Units, Power Profiles Daemon, UPower, Cava

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/jprando/wabago.git
cd wabago

# Build
go build -o wabago ./cmd/wabago

# Install (optional)
sudo mv wabago /usr/local/bin/
```

### Requirements

- Go 1.21 or later
- A terminal with UTF-8 and true color support

## Usage

```bash
# Run the application
./wabago

# Or if installed globally
wabago
```

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `h` | Move left / Previous category |
| `→` / `l` | Move right / Next category |
| `Enter` | Select / Confirm |
| `Esc` | Go back / Cancel |
| `Tab` | Next field |

### Actions
| Key | Action |
|-----|--------|
| `a` | Add new module |
| `d` / `Delete` | Delete selected module |
| `m` | Toggle move mode (reorder modules) |
| `Ctrl+S` | Save configuration |
| `r` | Reload configuration from disk |
| `b` | Create backup |

### General
| Key | Action |
|-----|--------|
| `?` | Show/hide help |
| `q` / `Ctrl+C` | Quit application |

## Configuration Path

Wabago looks for Waybar configuration in the following locations (in order):

1. `$XDG_CONFIG_HOME/waybar/config` (or `config.jsonc`)
2. `~/.config/waybar/config` (or `config.jsonc`)

Backups are stored in `~/.config/waybar/backups/`

## Screenshots

The TUI features a clean, Nord-inspired interface with:
- Main menu for quick navigation
- Module lists with descriptions
- Property editors with placeholders
- Status bar with save state indicator
- Keyboard shortcut help

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Related

- [Waybar](https://github.com/Alexays/Waybar) - The status bar this tool configures
- [Waybar Wiki](https://github.com/Alexays/Waybar/wiki) - Official Waybar documentation
