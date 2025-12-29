package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// WaybarConfig represents the main waybar configuration
type WaybarConfig struct {
	// Bar settings
	Layer          string   `json:"layer,omitempty"`
	Position       string   `json:"position,omitempty"`
	Height         int      `json:"height,omitempty"`
	Width          int      `json:"width,omitempty"`
	Spacing        int      `json:"spacing,omitempty"`
	Margin         string   `json:"margin,omitempty"`
	MarginTop      int      `json:"margin-top,omitempty"`
	MarginBottom   int      `json:"margin-bottom,omitempty"`
	MarginLeft     int      `json:"margin-left,omitempty"`
	MarginRight    int      `json:"margin-right,omitempty"`
	Exclusive      *bool    `json:"exclusive,omitempty"`
	FixedCenter    *bool    `json:"fixed-center,omitempty"`
	Passthrough    *bool    `json:"passthrough,omitempty"`
	GTKLayerShell  *bool    `json:"gtk-layer-shell,omitempty"`
	IPC            *bool    `json:"ipc,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	Output         []string `json:"output,omitempty"`
	Name           string   `json:"name,omitempty"`
	ReloadOnChange *bool    `json:"reload_style_on_change,omitempty"`

	// Module lists
	ModulesLeft   []string `json:"modules-left,omitempty"`
	ModulesCenter []string `json:"modules-center,omitempty"`
	ModulesRight  []string `json:"modules-right,omitempty"`

	// Module configurations (dynamic)
	Modules map[string]interface{} `json:"-"`

	// Raw data for preserving unknown fields
	Raw map[string]interface{} `json:"-"`
}

// ConfigManager handles waybar configuration files
type ConfigManager struct {
	ConfigPath string
	StylePath  string
	BackupDir  string
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "waybar")

	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configDir = filepath.Join(xdgConfig, "waybar")
	}

	return &ConfigManager{
		ConfigPath: filepath.Join(configDir, "config"),
		StylePath:  filepath.Join(configDir, "style.css"),
		BackupDir:  filepath.Join(configDir, "backups"),
	}
}

// GetConfigDir returns the waybar config directory
func (m *ConfigManager) GetConfigDir() string {
	return filepath.Dir(m.ConfigPath)
}

// ConfigExists checks if the config file exists
func (m *ConfigManager) ConfigExists() bool {
	// Check for config.jsonc first, then config
	if _, err := os.Stat(m.ConfigPath + ".jsonc"); err == nil {
		m.ConfigPath = m.ConfigPath + ".jsonc"
		return true
	}
	if _, err := os.Stat(m.ConfigPath); err == nil {
		return true
	}
	return false
}

// StyleExists checks if the style file exists
func (m *ConfigManager) StyleExists() bool {
	_, err := os.Stat(m.StylePath)
	return err == nil
}

// stripJSONCComments removes comments from JSONC content
// This handles strings correctly to avoid removing // inside quoted strings
func stripJSONCComments(content string) string {
	var result strings.Builder
	inString := false
	inSingleLineComment := false
	inMultiLineComment := false
	escaped := false

	runes := []rune(content)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		// Handle escape sequences in strings
		if inString && !escaped && c == '\\' {
			escaped = true
			result.WriteRune(c)
			continue
		}

		if escaped {
			escaped = false
			result.WriteRune(c)
			continue
		}

		// Handle string boundaries
		if c == '"' && !inSingleLineComment && !inMultiLineComment {
			inString = !inString
			result.WriteRune(c)
			continue
		}

		// Skip if in string
		if inString {
			result.WriteRune(c)
			continue
		}

		// Handle single-line comment start
		if !inMultiLineComment && c == '/' && next == '/' {
			inSingleLineComment = true
			i++ // skip next char
			continue
		}

		// Handle single-line comment end
		if inSingleLineComment && c == '\n' {
			inSingleLineComment = false
			result.WriteRune(c)
			continue
		}

		// Handle multi-line comment start
		if !inSingleLineComment && c == '/' && next == '*' {
			inMultiLineComment = true
			i++ // skip next char
			continue
		}

		// Handle multi-line comment end
		if inMultiLineComment && c == '*' && next == '/' {
			inMultiLineComment = false
			i++ // skip next char
			continue
		}

		// Write character if not in comment
		if !inSingleLineComment && !inMultiLineComment {
			result.WriteRune(c)
		}
	}

	// Remove trailing commas before } or ]
	cleaned := result.String()
	trailingCommaRe := regexp.MustCompile(`,\s*([}\]])`)
	cleaned = trailingCommaRe.ReplaceAllString(cleaned, "$1")

	return cleaned
}

// LoadConfig loads the waybar configuration
func (m *ConfigManager) LoadConfig() (*WaybarConfig, error) {
	if !m.ConfigExists() {
		return nil, fmt.Errorf("config file not found at %s", m.ConfigPath)
	}

	content, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Strip JSONC comments
	jsonContent := stripJSONCComments(string(content))

	// Try to parse - config can be object or array (for multiple bars)
	var raw map[string]interface{}

	// First try as object
	if err := json.Unmarshal([]byte(jsonContent), &raw); err != nil {
		// Try as array (multiple bars config)
		var rawArray []map[string]interface{}
		if err2 := json.Unmarshal([]byte(jsonContent), &rawArray); err2 != nil {
			return nil, fmt.Errorf("failed to parse config JSON: %w (also tried as array: %v)", err, err2)
		}
		// Use first bar config
		if len(rawArray) > 0 {
			raw = rawArray[0]
		} else {
			return nil, fmt.Errorf("config array is empty")
		}
	}

	// Process includes - Waybar supports including external config files
	if includes, ok := raw["include"].([]interface{}); ok {
		configDir := filepath.Dir(m.ConfigPath)
		for _, inc := range includes {
			if incPath, ok := inc.(string); ok {
				// Handle relative paths
				if !filepath.IsAbs(incPath) {
					incPath = filepath.Join(configDir, incPath)
				}
				// Expand ~ to home directory
				if strings.HasPrefix(incPath, "~") {
					if home, err := os.UserHomeDir(); err == nil {
						incPath = filepath.Join(home, incPath[1:])
					}
				}
				// Load and merge included file
				if incContent, err := os.ReadFile(incPath); err == nil {
					incJSON := stripJSONCComments(string(incContent))
					var incRaw map[string]interface{}
					if err := json.Unmarshal([]byte(incJSON), &incRaw); err == nil {
						// Merge included config into raw (included values override)
						for k, v := range incRaw {
							raw[k] = v
						}
					}
				}
			}
		}
	}

	// Then parse into struct
	config := &WaybarConfig{
		Raw:     raw,
		Modules: make(map[string]interface{}),
	}

	// Extract known fields
	if v, ok := raw["layer"].(string); ok {
		config.Layer = v
	}
	if v, ok := raw["position"].(string); ok {
		config.Position = v
	}
	if v, ok := raw["height"].(float64); ok {
		config.Height = int(v)
	}
	if v, ok := raw["width"].(float64); ok {
		config.Width = int(v)
	}
	if v, ok := raw["spacing"].(float64); ok {
		config.Spacing = int(v)
	}
	if v, ok := raw["margin"].(string); ok {
		config.Margin = v
	}
	if v, ok := raw["margin-top"].(float64); ok {
		config.MarginTop = int(v)
	}
	if v, ok := raw["margin-bottom"].(float64); ok {
		config.MarginBottom = int(v)
	}
	if v, ok := raw["margin-left"].(float64); ok {
		config.MarginLeft = int(v)
	}
	if v, ok := raw["margin-right"].(float64); ok {
		config.MarginRight = int(v)
	}
	if v, ok := raw["exclusive"].(bool); ok {
		config.Exclusive = &v
	}
	if v, ok := raw["fixed-center"].(bool); ok {
		config.FixedCenter = &v
	}
	if v, ok := raw["passthrough"].(bool); ok {
		config.Passthrough = &v
	}
	if v, ok := raw["gtk-layer-shell"].(bool); ok {
		config.GTKLayerShell = &v
	}
	if v, ok := raw["ipc"].(bool); ok {
		config.IPC = &v
	}
	if v, ok := raw["mode"].(string); ok {
		config.Mode = v
	}
	if v, ok := raw["name"].(string); ok {
		config.Name = v
	}
	if v, ok := raw["reload_style_on_change"].(bool); ok {
		config.ReloadOnChange = &v
	}

	// Handle output (can be string or array)
	if v, ok := raw["output"].(string); ok {
		config.Output = []string{v}
	} else if v, ok := raw["output"].([]interface{}); ok {
		for _, o := range v {
			if s, ok := o.(string); ok {
				config.Output = append(config.Output, s)
			}
		}
	}

	// Extract module lists
	if v, ok := raw["modules-left"].([]interface{}); ok {
		for _, m := range v {
			if s, ok := m.(string); ok {
				config.ModulesLeft = append(config.ModulesLeft, s)
			}
		}
	}
	if v, ok := raw["modules-center"].([]interface{}); ok {
		for _, m := range v {
			if s, ok := m.(string); ok {
				config.ModulesCenter = append(config.ModulesCenter, s)
			}
		}
	}
	if v, ok := raw["modules-right"].([]interface{}); ok {
		for _, m := range v {
			if s, ok := m.(string); ok {
				config.ModulesRight = append(config.ModulesRight, s)
			}
		}
	}

	// Extract module configurations
	knownKeys := map[string]bool{
		"layer": true, "position": true, "height": true, "width": true,
		"spacing": true, "margin": true, "margin-top": true, "margin-bottom": true,
		"margin-left": true, "margin-right": true, "exclusive": true,
		"fixed-center": true, "passthrough": true, "gtk-layer-shell": true,
		"ipc": true, "mode": true, "output": true, "name": true,
		"reload_style_on_change": true, "modules-left": true,
		"modules-center": true, "modules-right": true, "include": true,
	}

	for key, value := range raw {
		if !knownKeys[key] {
			config.Modules[key] = value
		}
	}

	return config, nil
}

// LoadStyle loads the waybar CSS style
func (m *ConfigManager) LoadStyle() (string, error) {
	if !m.StyleExists() {
		return "", nil
	}

	content, err := os.ReadFile(m.StylePath)
	if err != nil {
		return "", fmt.Errorf("failed to read style: %w", err)
	}

	return string(content), nil
}

// CreateBackup creates a backup of the current config and style
func (m *ConfigManager) CreateBackup() error {
	if err := os.MkdirAll(m.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Backup config
	if m.ConfigExists() {
		content, err := os.ReadFile(m.ConfigPath)
		if err == nil {
			backupPath := filepath.Join(m.BackupDir, fmt.Sprintf("config_%s", timestamp))
			os.WriteFile(backupPath, content, 0644)
		}
	}

	// Backup style
	if m.StyleExists() {
		content, err := os.ReadFile(m.StylePath)
		if err == nil {
			backupPath := filepath.Join(m.BackupDir, fmt.Sprintf("style_%s.css", timestamp))
			os.WriteFile(backupPath, content, 0644)
		}
	}

	return nil
}

// SaveConfig saves the waybar configuration
func (m *ConfigManager) SaveConfig(config *WaybarConfig) error {
	// Build the output map
	output := make(map[string]interface{})

	// Add bar settings
	if config.Layer != "" {
		output["layer"] = config.Layer
	}
	if config.Position != "" {
		output["position"] = config.Position
	}
	if config.Height > 0 {
		output["height"] = config.Height
	}
	if config.Width > 0 {
		output["width"] = config.Width
	}
	if config.Spacing > 0 {
		output["spacing"] = config.Spacing
	}
	if config.Margin != "" {
		output["margin"] = config.Margin
	}
	if config.MarginTop != 0 {
		output["margin-top"] = config.MarginTop
	}
	if config.MarginBottom != 0 {
		output["margin-bottom"] = config.MarginBottom
	}
	if config.MarginLeft != 0 {
		output["margin-left"] = config.MarginLeft
	}
	if config.MarginRight != 0 {
		output["margin-right"] = config.MarginRight
	}
	if config.Exclusive != nil {
		output["exclusive"] = *config.Exclusive
	}
	if config.FixedCenter != nil {
		output["fixed-center"] = *config.FixedCenter
	}
	if config.Passthrough != nil {
		output["passthrough"] = *config.Passthrough
	}
	if config.GTKLayerShell != nil {
		output["gtk-layer-shell"] = *config.GTKLayerShell
	}
	if config.IPC != nil {
		output["ipc"] = *config.IPC
	}
	if config.Mode != "" {
		output["mode"] = config.Mode
	}
	if config.Name != "" {
		output["name"] = config.Name
	}
	if config.ReloadOnChange != nil {
		output["reload_style_on_change"] = *config.ReloadOnChange
	}
	if len(config.Output) == 1 {
		output["output"] = config.Output[0]
	} else if len(config.Output) > 1 {
		output["output"] = config.Output
	}

	// Add module lists
	if len(config.ModulesLeft) > 0 {
		output["modules-left"] = config.ModulesLeft
	}
	if len(config.ModulesCenter) > 0 {
		output["modules-center"] = config.ModulesCenter
	}
	if len(config.ModulesRight) > 0 {
		output["modules-right"] = config.ModulesRight
	}

	// Add module configurations
	for key, value := range config.Modules {
		output[key] = value
	}

	// Format with indentation
	jsonData, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(m.ConfigPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(m.ConfigPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SaveStyle saves the waybar CSS style
func (m *ConfigManager) SaveStyle(content string) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(m.StylePath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.StylePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write style: %w", err)
	}

	return nil
}

// GetBackups returns a list of available backups
func (m *ConfigManager) GetBackups() ([]string, error) {
	entries, err := os.ReadDir(m.BackupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}

// RestoreBackup restores a specific backup
func (m *ConfigManager) RestoreBackup(filename string) error {
	backupPath := filepath.Join(m.BackupDir, filename)
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if strings.HasPrefix(filename, "config_") {
		return os.WriteFile(m.ConfigPath, content, 0644)
	} else if strings.HasPrefix(filename, "style_") {
		return os.WriteFile(m.StylePath, content, 0644)
	}

	return fmt.Errorf("unknown backup type: %s", filename)
}

// GetAllModules returns all modules used in the config
func (c *WaybarConfig) GetAllModules() []string {
	seen := make(map[string]bool)
	var modules []string

	for _, m := range c.ModulesLeft {
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}
	for _, m := range c.ModulesCenter {
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}
	for _, m := range c.ModulesRight {
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}

	return modules
}

// GetModuleConfig returns the configuration for a specific module
func (c *WaybarConfig) GetModuleConfig(name string) map[string]interface{} {
	if config, ok := c.Modules[name].(map[string]interface{}); ok {
		return config
	}
	return make(map[string]interface{})
}

// SetModuleConfig sets the configuration for a specific module
func (c *WaybarConfig) SetModuleConfig(name string, config map[string]interface{}) {
	c.Modules[name] = config
}
