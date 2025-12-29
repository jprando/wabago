// config handles the creation, listing, and restoration of configuration backups.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

// CreateNamedBackup creates a backup with a specific name and returns the paths
func (m *ConfigManager) CreateNamedBackup(name string) (string, string, error) {
	if err := os.MkdirAll(m.BackupDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	var configBackupPath, styleBackupPath string

	// Backup config
	if m.ConfigExists() {
		content, err := os.ReadFile(m.ConfigPath)
		if err == nil {
			configBackupPath = filepath.Join(m.BackupDir, fmt.Sprintf("config_%s", name))
			if err := os.WriteFile(configBackupPath, content, 0644); err != nil {
				return "", "", fmt.Errorf("failed to write config backup: %w", err)
			}
		}
	}

	// Backup style
	if m.StyleExists() {
		content, err := os.ReadFile(m.StylePath)
		if err == nil {
			styleBackupPath = filepath.Join(m.BackupDir, fmt.Sprintf("style_%s.css", name))
			if err := os.WriteFile(styleBackupPath, content, 0644); err != nil {
				return "", "", fmt.Errorf("failed to write style backup: %w", err)
			}
		}
	}

	return configBackupPath, styleBackupPath, nil
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
