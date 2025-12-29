// config provides functionality for comparing configuration changes.
package config

import (
	"fmt"
	"os"
	"os/exec"
)

// GetDiff returns the unified diff between the saved config/style and the current in-memory state
func (m *ConfigManager) GetDiff(currentConfig *WaybarConfig, currentStyle string) (string, error) {
	var diffOutput string

	// 1. Config Diff
	if m.ConfigExists() {
		// Generate JSON for current in-memory config
		newConfigBytes, err := m.MarshalConfig(currentConfig)
		if err != nil {
			return "", fmt.Errorf("failed to marshal current config: %w", err)
		}

		// Create temp file for new config
		tmpConfig, err := os.CreateTemp("", "wabago_config_*.json")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpConfig.Name())
		defer tmpConfig.Close()

		if _, err := tmpConfig.Write(newConfigBytes); err != nil {
			return "", fmt.Errorf("failed to write temp config: %w", err)
		}

		// Run diff
		cmd := exec.Command("diff", "-u", "--label", "Saved Config", "--label", "Current Config", m.ConfigPath, tmpConfig.Name())
		out, err := cmd.CombinedOutput()
		// diff returns exit code 1 if files differ, 0 if same, 2 if error
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					diffOutput += string(out) + "\n"
				} else {
					return "", fmt.Errorf("diff command failed: %s", out)
				}
			} else {
				return "", fmt.Errorf("failed to run diff: %w", err)
			}
		} else {
			// No differences or error? check output
			if len(out) > 0 {
				diffOutput += string(out) + "\n"
			}
		}
	} else {
		diffOutput += "No saved configuration found to compare.\n"
	}

	// 2. Style Diff
	if m.StyleExists() {
		// Create temp file for new style
		tmpStyle, err := os.CreateTemp("", "wabago_style_*.css")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpStyle.Name())
		defer tmpStyle.Close()

		if _, err := tmpStyle.WriteString(currentStyle); err != nil {
			return "", fmt.Errorf("failed to write temp style: %w", err)
		}

		// Run diff
		cmd := exec.Command("diff", "-u", "--label", "Saved Style", "--label", "Current Style", m.StylePath, tmpStyle.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					diffOutput += string(out) + "\n"
				} else {
					return "", fmt.Errorf("diff command failed: %s", out)
				}
			} else {
				return "", fmt.Errorf("failed to run diff: %w", err)
			}
		} else {
			if len(out) > 0 {
				diffOutput += string(out) + "\n"
			}
		}
	}

	if diffOutput == "" {
		return "No changes detected.", nil
	}

	return diffOutput, nil
}
