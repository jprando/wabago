// ui defines the core application model, state management, and the NewApp constructor.
package ui

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/jprando/wabago/internal/config"
)

// View represents the current view/screen
type View int

const (
	ViewMain View = iota
	ViewBarSettings
	ViewModulesLeft
	ViewModulesCenter
	ViewModulesRight
	ViewModuleEditor
	ViewModuleAdd
	ViewStyleEditor
	ViewBackups
	ViewHelp
	ViewRestoreConfirm
	ViewModal
	ViewModuleCatalog
	ViewDiff
)

// ModalType distinguishes between selection and text input modals
type ModalType int
...
	// Config
	configManager *config.ConfigManager
	config        *config.WaybarConfig
	styleContent  string
	hasChanges    bool
	diffContent   string

	// Components
	keyMap     KeyMap

	// Module editing
	editingModule       string
	moduleProperties    []config.PropertyDefinition
	modulePropertyIndex int

	// Module adding
	addModuleCategory int
	addModuleIndex    int

	// Module Catalog
	catalogCategory int
	catalogIndex    int

	// Notifications
	notification    string
	notificationType string

	// Status
	configPath string
	stylePath  string
	
	// Layout
	contentHeight int

	// Startup Backup
	startupConfigBackup string
	startupStyleBackup  string
	
	// Modal
	modal ModalState
}

// NewApp creates a new application instance
func NewApp() *App {
	cm := config.NewConfigManager()

	app := &App{
		configManager: cm,
		keyMap:        DefaultKeyMap(),
		help:          help.New(),
		view:          ViewMain,
		configPath:    cm.ConfigPath,
		stylePath:     cm.StylePath,
	}

	// Validate config before creating startup backup
	// We only want to backup if the current config is at least syntactically valid
	if _, err := cm.LoadConfig(); err == nil {
		// Create startup backup with timestamp
		timestamp := time.Now().Format("20060102_150405")
		cfgBackup, styleBackup, err := cm.CreateNamedBackup("startup_" + timestamp)
		if err == nil {
			app.startupConfigBackup = cfgBackup
			app.startupStyleBackup = styleBackup
		}
	} else {
		// If config is invalid on startup, we don't have a safe "startup" state to restore to.
		// app.loadConfig() called below will report the error to the user.
		app.notification = "Warning: Startup config is invalid. Restore disabled."
		app.notificationType = "warning"
	}

	app.loadConfig()

	// If loadConfig didn't report an error, and we have a backup, inform the user
	if app.notification == "" && app.startupConfigBackup != "" {
		app.notification = "Startup backup created successfully"
		app.notificationType = "info"
	}

	return app
}

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Icon        string
	View        View
}

// Cleanup removes startup backups
func (a *App) Cleanup() {
	if a.startupConfigBackup != "" {
		os.Remove(a.startupConfigBackup)
	}
	if a.startupStyleBackup != "" {
		os.Remove(a.startupStyleBackup)
	}
}