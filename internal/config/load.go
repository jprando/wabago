// config implements functionality for loading and parsing Waybar configuration and style files.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
	// Include can be array or single string
	configDir := filepath.Dir(m.ConfigPath)

	var includePaths []string
	if includes, ok := raw["include"].([]interface{}); ok {
		for _, inc := range includes {
			if incPath, ok := inc.(string); ok {
				includePaths = append(includePaths, incPath)
			}
		}
	} else if include, ok := raw["include"].(string); ok {
		// Single include path
		includePaths = append(includePaths, include)
	}

	for _, incPath := range includePaths {
		// Expand ~ to home directory first
		if strings.HasPrefix(incPath, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				incPath = filepath.Join(home, incPath[2:])
			}
		}
		// Handle relative paths
		if !filepath.IsAbs(incPath) {
			incPath = filepath.Join(configDir, incPath)
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
