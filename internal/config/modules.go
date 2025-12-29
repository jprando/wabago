package config

// ModuleCategory represents a category of waybar modules
type ModuleCategory struct {
	Name        string
	Description string
	Modules     []ModuleDefinition
}

// ModuleDefinition defines a waybar module and its properties
type ModuleDefinition struct {
	Name        string
	Description string
	Properties  []PropertyDefinition
}

// PropertyDefinition defines a property for a module
type PropertyDefinition struct {
	Name        string
	Type        string // "string", "integer", "boolean", "array", "object"
	Default     interface{}
	Description string
	Options     []string // For enum-like properties
	Required    bool
}

// CommonProperties returns properties common to all modules
func CommonProperties() []PropertyDefinition {
	return []PropertyDefinition{
		{Name: "format", Type: "string", Description: "Display format template with Pango markup support"},
		{Name: "format-icons", Type: "array", Description: "Array of icons for different states"},
		{Name: "rotate", Type: "integer", Default: 0, Description: "Rotation in degrees: 0, 90, 180, 270"},
		{Name: "max-length", Type: "integer", Description: "Maximum length of the module text"},
		{Name: "min-length", Type: "integer", Description: "Minimum length of the module text"},
		{Name: "align", Type: "number", Description: "Alignment of text (0=left, 0.5=center, 1=right)"},
		{Name: "on-click", Type: "string", Description: "Command to execute on left click"},
		{Name: "on-click-middle", Type: "string", Description: "Command to execute on middle click"},
		{Name: "on-click-right", Type: "string", Description: "Command to execute on right click"},
		{Name: "on-double-click", Type: "string", Description: "Command to execute on double click"},
		{Name: "on-triple-click", Type: "string", Description: "Command to execute on triple click"},
		{Name: "on-scroll-up", Type: "string", Description: "Command to execute on scroll up"},
		{Name: "on-scroll-down", Type: "string", Description: "Command to execute on scroll down"},
		{Name: "on-update", Type: "string", Description: "Command to execute when module updates"},
		{Name: "tooltip", Type: "boolean", Default: true, Description: "Enable/disable tooltip"},
		{Name: "tooltip-format", Type: "string", Description: "Tooltip format template"},
	}
}

// GetModuleCategories returns all available waybar module categories
func GetModuleCategories() []ModuleCategory {
	return []ModuleCategory{
		{
			Name:        "System",
			Description: "System monitoring modules",
			Modules: []ModuleDefinition{
				{
					Name:        "battery",
					Description: "Display battery status and percentage",
					Properties: []PropertyDefinition{
						{Name: "bat", Type: "string", Description: "Battery device name (e.g., BAT0)"},
						{Name: "adapter", Type: "string", Description: "Power adapter name"},
						{Name: "full-at", Type: "integer", Default: 100, Description: "Consider full at percentage"},
						{Name: "design-capacity", Type: "boolean", Default: false, Description: "Use design capacity"},
						{Name: "interval", Type: "integer", Default: 60, Description: "Update interval in seconds"},
						{Name: "states", Type: "object", Description: "Battery level states for styling"},
						{Name: "format-charging", Type: "string", Description: "Format when charging"},
						{Name: "format-discharging", Type: "string", Description: "Format when discharging"},
						{Name: "format-full", Type: "string", Description: "Format when full"},
						{Name: "format-plugged", Type: "string", Description: "Format when plugged"},
						{Name: "format-time", Type: "string", Description: "Time remaining format"},
					},
				},
				{
					Name:        "cpu",
					Description: "Display CPU usage",
					Properties: []PropertyDefinition{
						{Name: "interval", Type: "integer", Default: 10, Description: "Update interval in seconds"},
						{Name: "format", Type: "string", Default: "{usage}%", Description: "Display format"},
						{Name: "states", Type: "object", Description: "CPU usage states for styling"},
					},
				},
				{
					Name:        "memory",
					Description: "Display memory usage",
					Properties: []PropertyDefinition{
						{Name: "interval", Type: "integer", Default: 30, Description: "Update interval in seconds"},
						{Name: "format", Type: "string", Default: "{}%", Description: "Display format"},
						{Name: "states", Type: "object", Description: "Memory usage states for styling"},
					},
				},
				{
					Name:        "disk",
					Description: "Display disk usage",
					Properties: []PropertyDefinition{
						{Name: "path", Type: "string", Default: "/", Description: "Mount path to monitor"},
						{Name: "interval", Type: "integer", Default: 30, Description: "Update interval in seconds"},
						{Name: "format", Type: "string", Default: "{percentage_used}%", Description: "Display format"},
						{Name: "unit", Type: "string", Default: "GB", Description: "Unit for size display", Options: []string{"kB", "kiB", "MB", "MiB", "GB", "GiB", "TB", "TiB"}},
					},
				},
				{
					Name:        "temperature",
					Description: "Display temperature sensor data",
					Properties: []PropertyDefinition{
						{Name: "hwmon-path", Type: "string", Description: "Path to hwmon temperature file"},
						{Name: "hwmon-path-abs", Type: "string", Description: "Absolute path to hwmon device"},
						{Name: "input-filename", Type: "string", Description: "Temperature input file name"},
						{Name: "thermal-zone", Type: "integer", Description: "Thermal zone number"},
						{Name: "critical-threshold", Type: "integer", Default: 80, Description: "Critical temperature threshold"},
						{Name: "warning-threshold", Type: "integer", Default: 70, Description: "Warning temperature threshold"},
						{Name: "interval", Type: "integer", Default: 10, Description: "Update interval in seconds"},
						{Name: "format", Type: "string", Default: "{temperatureC}°C", Description: "Display format"},
						{Name: "format-critical", Type: "string", Description: "Format when critical"},
					},
				},
				{
					Name:        "load",
					Description: "Display system load average",
					Properties: []PropertyDefinition{
						{Name: "interval", Type: "integer", Default: 10, Description: "Update interval in seconds"},
						{Name: "format", Type: "string", Default: "{load1}", Description: "Display format"},
					},
				},
			},
		},
		{
			Name:        "Audio/Media",
			Description: "Audio and media control modules",
			Modules: []ModuleDefinition{
				{
					Name:        "pulseaudio",
					Description: "PulseAudio volume control",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{volume}%", Description: "Display format"},
						{Name: "format-muted", Type: "string", Description: "Format when muted"},
						{Name: "format-bluetooth", Type: "string", Description: "Format for bluetooth devices"},
						{Name: "format-source", Type: "string", Description: "Format for source (microphone)"},
						{Name: "format-source-muted", Type: "string", Description: "Format when source muted"},
						{Name: "scroll-step", Type: "number", Default: 1, Description: "Volume step on scroll"},
						{Name: "max-volume", Type: "integer", Default: 100, Description: "Maximum volume percentage"},
						{Name: "ignored-sinks", Type: "array", Description: "List of sinks to ignore"},
					},
				},
				{
					Name:        "wireplumber",
					Description: "WirePlumber/PipeWire volume control",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{volume}%", Description: "Display format"},
						{Name: "format-muted", Type: "string", Description: "Format when muted"},
						{Name: "scroll-step", Type: "number", Default: 1, Description: "Volume step on scroll"},
						{Name: "max-volume", Type: "integer", Default: 100, Description: "Maximum volume percentage"},
					},
				},
				{
					Name:        "mpris",
					Description: "MPRIS media player control",
					Properties: []PropertyDefinition{
						{Name: "player", Type: "string", Description: "Specific player to track"},
						{Name: "ignored-players", Type: "array", Description: "Players to ignore"},
						{Name: "format", Type: "string", Default: "{player_icon} {dynamic}", Description: "Display format"},
						{Name: "format-paused", Type: "string", Description: "Format when paused"},
						{Name: "format-stopped", Type: "string", Description: "Format when stopped"},
						{Name: "artist-len", Type: "integer", Description: "Max artist name length"},
						{Name: "album-len", Type: "integer", Description: "Max album name length"},
						{Name: "title-len", Type: "integer", Description: "Max title length"},
						{Name: "dynamic-len", Type: "integer", Description: "Max dynamic text length"},
						{Name: "dynamic-order", Type: "array", Description: "Order of dynamic elements"},
						{Name: "interval", Type: "integer", Default: 1, Description: "Update interval in seconds"},
					},
				},
			},
		},
		{
			Name:        "Workspaces",
			Description: "Workspace and window management modules",
			Modules: []ModuleDefinition{
				{
					Name:        "hyprland/workspaces",
					Description: "Hyprland workspace indicator",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{name}", Description: "Display format"},
						{Name: "format-icons", Type: "object", Description: "Icons for workspace states"},
						{Name: "show-special", Type: "boolean", Default: false, Description: "Show special workspaces"},
						{Name: "all-outputs", Type: "boolean", Default: false, Description: "Show all workspaces on all outputs"},
						{Name: "active-only", Type: "boolean", Default: false, Description: "Show only active workspace"},
						{Name: "sort-by", Type: "string", Default: "default", Description: "Sorting method", Options: []string{"default", "number", "name", "id"}},
						{Name: "persistent-workspaces", Type: "object", Description: "Workspaces to always show"},
					},
				},
				{
					Name:        "hyprland/window",
					Description: "Hyprland active window title",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{title}", Description: "Display format"},
						{Name: "max-length", Type: "integer", Description: "Maximum title length"},
						{Name: "separate-outputs", Type: "boolean", Default: false, Description: "Separate per output"},
						{Name: "rewrite", Type: "object", Description: "Title rewrite rules"},
					},
				},
				{
					Name:        "hyprland/submap",
					Description: "Hyprland submap indicator",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{}", Description: "Display format"},
						{Name: "always-on", Type: "boolean", Default: false, Description: "Always show module"},
					},
				},
				{
					Name:        "hyprland/language",
					Description: "Hyprland keyboard layout",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{}", Description: "Display format"},
						{Name: "format-short", Type: "string", Description: "Short format (first 2 chars)"},
						{Name: "keyboard-name", Type: "string", Description: "Specific keyboard to track"},
					},
				},
				{
					Name:        "sway/workspaces",
					Description: "Sway workspace indicator",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{name}", Description: "Display format"},
						{Name: "format-icons", Type: "object", Description: "Icons for workspace states"},
						{Name: "all-outputs", Type: "boolean", Default: false, Description: "Show all workspaces on all outputs"},
						{Name: "disable-scroll", Type: "boolean", Default: false, Description: "Disable scroll to switch"},
						{Name: "disable-click", Type: "boolean", Default: false, Description: "Disable click to switch"},
						{Name: "persistent-workspaces", Type: "object", Description: "Workspaces to always show"},
					},
				},
				{
					Name:        "sway/window",
					Description: "Sway active window title",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{title}", Description: "Display format"},
						{Name: "max-length", Type: "integer", Description: "Maximum title length"},
						{Name: "tooltip", Type: "boolean", Default: true, Description: "Enable tooltip"},
					},
				},
				{
					Name:        "sway/mode",
					Description: "Sway mode indicator",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{}", Description: "Display format"},
					},
				},
				{
					Name:        "wlr/taskbar",
					Description: "Task bar for wlroots compositors",
					Properties: []PropertyDefinition{
						{Name: "all-outputs", Type: "boolean", Default: false, Description: "Show all windows on all outputs"},
						{Name: "format", Type: "string", Default: "{icon}", Description: "Display format"},
						{Name: "icon-size", Type: "integer", Default: 16, Description: "Icon size in pixels"},
						{Name: "icon-theme", Type: "string", Description: "Icon theme name"},
						{Name: "tooltip-format", Type: "string", Description: "Tooltip format"},
						{Name: "active-first", Type: "boolean", Default: false, Description: "Show active window first"},
						{Name: "sort-by-app-id", Type: "boolean", Default: false, Description: "Sort by application ID"},
					},
				},
			},
		},
		{
			Name:        "Network",
			Description: "Network and connectivity modules",
			Modules: []ModuleDefinition{
				{
					Name:        "network",
					Description: "Network status and speed",
					Properties: []PropertyDefinition{
						{Name: "interface", Type: "string", Description: "Specific interface to monitor"},
						{Name: "interval", Type: "integer", Default: 60, Description: "Update interval in seconds"},
						{Name: "format", Type: "string", Description: "Display format"},
						{Name: "format-wifi", Type: "string", Description: "Format for WiFi connections"},
						{Name: "format-ethernet", Type: "string", Description: "Format for Ethernet connections"},
						{Name: "format-linked", Type: "string", Description: "Format when linked but no IP"},
						{Name: "format-disconnected", Type: "string", Description: "Format when disconnected"},
						{Name: "format-disabled", Type: "string", Description: "Format when disabled"},
						{Name: "tooltip-format", Type: "string", Description: "Tooltip format"},
						{Name: "tooltip-format-wifi", Type: "string", Description: "Tooltip for WiFi"},
						{Name: "tooltip-format-ethernet", Type: "string", Description: "Tooltip for Ethernet"},
						{Name: "tooltip-format-disconnected", Type: "string", Description: "Tooltip when disconnected"},
						{Name: "max-length", Type: "integer", Description: "Maximum text length"},
					},
				},
				{
					Name:        "bluetooth",
					Description: "Bluetooth status",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{status}", Description: "Display format"},
						{Name: "format-connected", Type: "string", Description: "Format when connected"},
						{Name: "format-connected-battery", Type: "string", Description: "Format with battery info"},
						{Name: "format-disabled", Type: "string", Description: "Format when disabled"},
						{Name: "format-off", Type: "string", Description: "Format when off"},
						{Name: "format-on", Type: "string", Description: "Format when on"},
						{Name: "format-no-controller", Type: "string", Description: "Format when no controller"},
						{Name: "controller-alias", Type: "string", Description: "Controller to monitor"},
					},
				},
			},
		},
		{
			Name:        "Display",
			Description: "Display and visual modules",
			Modules: []ModuleDefinition{
				{
					Name:        "clock",
					Description: "Date and time display",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{:%H:%M}", Description: "Time format (strftime)"},
						{Name: "format-alt", Type: "string", Description: "Alternative format on click"},
						{Name: "timezone", Type: "string", Description: "Timezone (e.g., America/New_York)"},
						{Name: "timezones", Type: "array", Description: "List of timezones to cycle"},
						{Name: "interval", Type: "integer", Default: 60, Description: "Update interval in seconds"},
						{Name: "locale", Type: "string", Description: "Locale for formatting"},
						{Name: "calendar", Type: "object", Description: "Calendar popup settings"},
						{Name: "actions", Type: "object", Description: "Calendar action bindings"},
					},
				},
				{
					Name:        "backlight",
					Description: "Screen backlight control",
					Properties: []PropertyDefinition{
						{Name: "device", Type: "string", Description: "Backlight device name"},
						{Name: "format", Type: "string", Default: "{percent}%", Description: "Display format"},
						{Name: "format-icons", Type: "array", Description: "Icons for brightness levels"},
						{Name: "scroll-step", Type: "number", Default: 1, Description: "Brightness step on scroll"},
						{Name: "reverse-scrolling", Type: "boolean", Default: false, Description: "Reverse scroll direction"},
						{Name: "smooth-scrolling-threshold", Type: "number", Description: "Smooth scrolling threshold"},
					},
				},
				{
					Name:        "tray",
					Description: "System tray",
					Properties: []PropertyDefinition{
						{Name: "icon-size", Type: "integer", Default: 16, Description: "Icon size in pixels"},
						{Name: "show-passive-items", Type: "boolean", Default: false, Description: "Show passive items"},
						{Name: "spacing", Type: "integer", Default: 4, Description: "Spacing between icons"},
						{Name: "reverse-direction", Type: "boolean", Default: false, Description: "Reverse icon order"},
					},
				},
				{
					Name:        "idle-inhibitor",
					Description: "Idle inhibitor toggle",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{icon}", Description: "Display format"},
						{Name: "format-icons", Type: "object", Description: "Icons for states (activated/deactivated)"},
						{Name: "timeout", Type: "number", Description: "Auto-deactivate timeout in minutes"},
						{Name: "start-activated", Type: "boolean", Default: false, Description: "Start in activated state"},
					},
				},
				{
					Name:        "image",
					Description: "Display an image",
					Properties: []PropertyDefinition{
						{Name: "path", Type: "string", Description: "Path to image file"},
						{Name: "size", Type: "integer", Default: 16, Description: "Image size in pixels"},
						{Name: "interval", Type: "integer", Default: 0, Description: "Refresh interval in seconds"},
					},
				},
			},
		},
		{
			Name:        "Input",
			Description: "Input device modules",
			Modules: []ModuleDefinition{
				{
					Name:        "keyboard-state",
					Description: "Keyboard lock states (Caps, Num, Scroll)",
					Properties: []PropertyDefinition{
						{Name: "capslock", Type: "boolean", Default: false, Description: "Show Caps Lock state"},
						{Name: "numlock", Type: "boolean", Default: false, Description: "Show Num Lock state"},
						{Name: "scrolllock", Type: "boolean", Default: false, Description: "Show Scroll Lock state"},
						{Name: "format", Type: "object", Description: "Format for each lock key"},
						{Name: "format-icons", Type: "object", Description: "Icons for lock states"},
						{Name: "device-path", Type: "string", Description: "Input device path"},
					},
				},
				{
					Name:        "sway/language",
					Description: "Sway keyboard layout",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{}", Description: "Display format"},
						{Name: "format-short", Type: "string", Description: "Short format (2 chars)"},
					},
				},
			},
		},
		{
			Name:        "Custom",
			Description: "Custom and user-defined modules",
			Modules: []ModuleDefinition{
				{
					Name:        "custom/",
					Description: "Custom script module (append name after slash)",
					Properties: []PropertyDefinition{
						{Name: "exec", Type: "string", Description: "Command to execute"},
						{Name: "exec-if", Type: "string", Description: "Condition command for visibility"},
						{Name: "exec-on-event", Type: "boolean", Default: true, Description: "Re-exec on click/scroll"},
						{Name: "return-type", Type: "string", Description: "Output format: json or empty", Options: []string{"", "json"}},
						{Name: "interval", Type: "integer", Description: "Update interval in seconds"},
						{Name: "restart-interval", Type: "integer", Description: "Restart interval for continuous"},
						{Name: "signal", Type: "integer", Description: "Signal number for updates"},
						{Name: "format", Type: "string", Default: "{}", Description: "Display format"},
						{Name: "format-icons", Type: "array", Description: "Icons for alt values"},
						{Name: "escape", Type: "boolean", Default: false, Description: "Escape output markup"},
					},
				},
				{
					Name:        "group/",
					Description: "Group multiple modules (append name after slash)",
					Properties: []PropertyDefinition{
						{Name: "modules", Type: "array", Description: "List of modules in group", Required: true},
						{Name: "orientation", Type: "string", Default: "inherit", Description: "Group orientation", Options: []string{"inherit", "horizontal", "vertical", "orthogonal"}},
						{Name: "drawer", Type: "object", Description: "Drawer animation settings"},
					},
				},
			},
		},
		{
			Name:        "Other",
			Description: "Miscellaneous modules",
			Modules: []ModuleDefinition{
				{
					Name:        "user",
					Description: "Display current user info",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{user}", Description: "Display format"},
						{Name: "interval", Type: "integer", Default: 60, Description: "Update interval in seconds"},
						{Name: "icon", Type: "boolean", Default: true, Description: "Show user avatar"},
						{Name: "icon-size", Type: "integer", Default: 24, Description: "Avatar size in pixels"},
					},
				},
				{
					Name:        "gamemode",
					Description: "Feral GameMode indicator",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{glyph}", Description: "Display format"},
						{Name: "format-alt", Type: "string", Description: "Alternative format"},
						{Name: "glyph", Type: "string", Default: "󰊴", Description: "Icon glyph"},
						{Name: "hide-not-running", Type: "boolean", Default: true, Description: "Hide when not running"},
						{Name: "use-icon", Type: "boolean", Default: true, Description: "Use icon or text"},
						{Name: "icon-name", Type: "string", Default: "input-gaming-symbolic", Description: "Icon name"},
						{Name: "icon-size", Type: "integer", Default: 20, Description: "Icon size in pixels"},
						{Name: "icon-spacing", Type: "integer", Default: 4, Description: "Icon spacing"},
					},
				},
				{
					Name:        "privacy",
					Description: "Privacy indicators (camera, microphone, screenshare)",
					Properties: []PropertyDefinition{
						{Name: "icon-size", Type: "integer", Default: 20, Description: "Icon size in pixels"},
						{Name: "icon-spacing", Type: "integer", Default: 4, Description: "Icon spacing"},
						{Name: "transition-duration", Type: "integer", Default: 250, Description: "Animation duration (ms)"},
						{Name: "modules", Type: "array", Description: "Privacy modules to show"},
					},
				},
				{
					Name:        "systemd-failed-units",
					Description: "Show failed systemd units",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{nr_failed} failed", Description: "Display format"},
						{Name: "format-ok", Type: "string", Default: "", Description: "Format when no failures"},
						{Name: "user", Type: "boolean", Default: true, Description: "Include user units"},
						{Name: "system", Type: "boolean", Default: true, Description: "Include system units"},
						{Name: "hide-on-ok", Type: "boolean", Default: true, Description: "Hide when no failures"},
					},
				},
				{
					Name:        "power-profiles-daemon",
					Description: "Power profile indicator/control",
					Properties: []PropertyDefinition{
						{Name: "format", Type: "string", Default: "{icon}", Description: "Display format"},
						{Name: "format-icons", Type: "object", Description: "Icons for each profile"},
						{Name: "tooltip-format", Type: "string", Description: "Tooltip format"},
					},
				},
				{
					Name:        "upower",
					Description: "UPower battery/device info",
					Properties: []PropertyDefinition{
						{Name: "native-path", Type: "string", Description: "Device native path"},
						{Name: "model", Type: "string", Description: "Device model filter"},
						{Name: "icon-size", Type: "integer", Default: 20, Description: "Icon size in pixels"},
						{Name: "format", Type: "string", Description: "Display format"},
						{Name: "format-alt", Type: "string", Description: "Alternative format"},
						{Name: "hide-if-empty", Type: "boolean", Default: true, Description: "Hide when no device"},
						{Name: "show-icon", Type: "boolean", Default: true, Description: "Show device icon"},
					},
				},
				{
					Name:        "cava",
					Description: "Audio visualizer",
					Properties: []PropertyDefinition{
						{Name: "framerate", Type: "integer", Default: 30, Description: "Frames per second"},
						{Name: "bars", Type: "integer", Default: 12, Description: "Number of bars"},
						{Name: "method", Type: "string", Default: "pulse", Description: "Audio method", Options: []string{"pulse", "pipewire", "alsa", "fifo", "sndio", "oss"}},
						{Name: "source", Type: "string", Default: "auto", Description: "Audio source"},
						{Name: "bar_delimiter", Type: "integer", Default: 0, Description: "Delimiter character code"},
						{Name: "format-icons", Type: "array", Description: "Bar level icons"},
						{Name: "input_delay", Type: "integer", Default: 2, Description: "Input delay frames"},
						{Name: "sleep_timer", Type: "integer", Default: 5, Description: "Sleep after silence (seconds)"},
						{Name: "hide_on_silence", Type: "boolean", Default: false, Description: "Hide when silent"},
						{Name: "sensitivity", Type: "integer", Default: 100, Description: "Visualization sensitivity"},
						{Name: "lower_cutoff_freq", Type: "integer", Default: 50, Description: "Lower frequency cutoff (Hz)"},
						{Name: "higher_cutoff_freq", Type: "integer", Default: 10000, Description: "Higher frequency cutoff (Hz)"},
						{Name: "stereo", Type: "boolean", Default: true, Description: "Enable stereo mode"},
						{Name: "monstercat", Type: "boolean", Default: false, Description: "Monstercat smoothing"},
						{Name: "waves", Type: "boolean", Default: false, Description: "Wave mode"},
						{Name: "noise_reduction", Type: "number", Default: 0.77, Description: "Noise reduction level"},
					},
				},
			},
		},
	}
}

// GetAllModuleNames returns all available module names
func GetAllModuleNames() []string {
	var names []string
	for _, category := range GetModuleCategories() {
		for _, module := range category.Modules {
			names = append(names, module.Name)
		}
	}
	return names
}

// GetModuleDefinition returns the definition for a specific module
func GetModuleDefinition(name string) *ModuleDefinition {
	// Handle custom/ and group/ modules
	baseName := name
	if len(name) > 7 && name[:7] == "custom/" {
		baseName = "custom/"
	} else if len(name) > 6 && name[:6] == "group/" {
		baseName = "group/"
	}

	// Handle instance notation (module#instance)
	for i, c := range baseName {
		if c == '#' {
			baseName = baseName[:i]
			break
		}
	}

	for _, category := range GetModuleCategories() {
		for _, module := range category.Modules {
			if module.Name == baseName {
				return &module
			}
		}
	}
	return nil
}
