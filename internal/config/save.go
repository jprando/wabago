// config implements functionality for persisting Waybar configuration and style changes to disk.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MarshalConfig converts the WaybarConfig to JSON bytes
func (m *ConfigManager) MarshalConfig(config *WaybarConfig) ([]byte, error) {
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
	return json.MarshalIndent(output, "", "    ")
}

// SaveConfig saves the waybar configuration
func (m *ConfigManager) SaveConfig(config *WaybarConfig) error {
	jsonData, err := m.MarshalConfig(config)
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
