// config provides the ConfigManager struct for handling file paths and configuration operations.
package config

import (
	"os"
	"path/filepath"
)

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
