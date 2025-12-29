package ui

import (
	"fmt"
	"strings"

	"github.com/jprando/wabago/internal/config"
	"github.com/jprando/wabago/internal/ui/styles"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

// KeyMap defines the key bindings
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Back     key.Binding
	Tab      key.Binding
	Save     key.Binding
	Quit     key.Binding
	Help     key.Binding
	Delete   key.Binding
	Add      key.Binding
	Move     key.Binding
	Reload   key.Binding
	Backup   key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d", "delete"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Move: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move"),
		),
		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload"),
		),
		Backup: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "backup"),
		),
	}
}

// ShortHelp returns the short help keys
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns the full help keys
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Tab},
		{k.Save, k.Add, k.Delete},
		{k.Reload, k.Backup, k.Quit},
	}
}

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Icon        string
	View        View
}

// App is the main application model
type App struct {
	// Dimensions
	width  int
	height int

	// State
	view         View
	previousView View
	menuIndex    int
	listIndex    int
	fieldIndex   int
	moving       bool

	// Config
	configManager *config.ConfigManager
	config        *config.WaybarConfig
	styleContent  string
	hasChanges    bool

	// Components
	keyMap     KeyMap
	help       help.Model
	textInputs []textinput.Model

	// Module editing
	editingModule       string
	moduleProperties    []string
	modulePropertyIndex int

	// Module adding
	addModuleCategory int
	addModuleIndex    int

	// Notifications
	notification    string
	notificationType string

	// Status
	configPath string
	stylePath  string
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

	app.loadConfig()
	return app
}

func (a *App) loadConfig() {
	cfg, err := a.configManager.LoadConfig()
	if err != nil {
		a.config = &config.WaybarConfig{
			Position:      "top",
			Layer:         "top",
			Height:        30,
			Spacing:       4,
			ModulesLeft:   []string{},
			ModulesCenter: []string{},
			ModulesRight:  []string{},
			Modules:       make(map[string]interface{}),
		}
		a.notification = fmt.Sprintf("Config not found, using defaults: %v", err)
		a.notificationType = "warning"
	} else {
		a.config = cfg
		a.notification = "Configuration loaded successfully"
		a.notificationType = "success"
	}

	style, err := a.configManager.LoadStyle()
	if err != nil {
		a.styleContent = defaultStyle()
	} else {
		a.styleContent = style
	}
}

func defaultStyle() string {
	return `* {
    font-family: "JetBrains Mono", "Font Awesome 6 Free", monospace;
    font-size: 13px;
}

window#waybar {
    background-color: rgba(46, 52, 64, 0.9);
    color: #eceff4;
    border-bottom: 2px solid #88c0d0;
}

#workspaces button {
    padding: 0 5px;
    color: #d8dee9;
    border-radius: 5px;
    margin: 2px;
}

#workspaces button.active {
    background-color: #88c0d0;
    color: #2e3440;
}

#clock, #battery, #cpu, #memory, #network, #pulseaudio, #tray {
    padding: 0 10px;
    margin: 2px 4px;
    border-radius: 5px;
}
`
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width
		return a, nil

	case tea.KeyMsg:
		// Clear notification on any key press
		if a.notification != "" {
			a.notification = ""
		}

		// Global keys
		switch {
		case key.Matches(msg, a.keyMap.Quit):
			if a.view == ViewMain {
				return a, tea.Quit
			}
			a.view = ViewMain
			a.listIndex = 0
			a.fieldIndex = 0
			return a, nil

		case key.Matches(msg, a.keyMap.Help):
			if a.view == ViewHelp {
				a.view = a.previousView
			} else {
				a.previousView = a.view
				a.view = ViewHelp
			}
			return a, nil

		case key.Matches(msg, a.keyMap.Save):
			return a, a.saveConfig()

		case key.Matches(msg, a.keyMap.Reload):
			a.loadConfig()
			a.notification = "Configuration reloaded"
			a.notificationType = "info"
			return a, nil

		case key.Matches(msg, a.keyMap.Backup):
			if err := a.configManager.CreateBackup(); err != nil {
				a.notification = fmt.Sprintf("Backup failed: %v", err)
				a.notificationType = "error"
			} else {
				a.notification = "Backup created successfully"
				a.notificationType = "success"
			}
			return a, nil
		}

		// View-specific handling
		switch a.view {
		case ViewMain:
			return a.updateMainView(msg)
		case ViewBarSettings:
			return a.updateBarSettingsView(msg)
		case ViewModulesLeft, ViewModulesCenter, ViewModulesRight:
			return a.updateModulesListView(msg)
		case ViewModuleEditor:
			return a.updateModuleEditorView(msg)
		case ViewModuleAdd:
			return a.updateModuleAddView(msg)
		case ViewStyleEditor:
			return a.updateStyleEditorView(msg)
		case ViewBackups:
			return a.updateBackupsView(msg)
		case ViewHelp:
			return a.updateHelpView(msg)
		}
	}

	// Update text inputs if any
	var cmd tea.Cmd
	for i := range a.textInputs {
		a.textInputs[i], cmd = a.textInputs[i].Update(msg)
	}
	return a, cmd
}

func (a *App) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := a.configManager.CreateBackup(); err != nil {
			// Log backup error but continue
		}

		if err := a.configManager.SaveConfig(a.config); err != nil {
			a.notification = fmt.Sprintf("Failed to save config: %v", err)
			a.notificationType = "error"
		} else {
			a.notification = "Configuration saved!"
			a.notificationType = "success"
			a.hasChanges = false
		}

		if err := a.configManager.SaveStyle(a.styleContent); err != nil {
			a.notification = fmt.Sprintf("Failed to save style: %v", err)
			a.notificationType = "error"
		}

		return nil
	}
}

func (a *App) updateMainView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuItems := a.getMainMenuItems()

	switch {
	case key.Matches(msg, a.keyMap.Up):
		if a.menuIndex > 0 {
			a.menuIndex--
		}
	case key.Matches(msg, a.keyMap.Down):
		if a.menuIndex < len(menuItems)-1 {
			a.menuIndex++
		}
	case key.Matches(msg, a.keyMap.Enter):
		if a.menuIndex < len(menuItems) {
			a.view = menuItems[a.menuIndex].View
			a.listIndex = 0
			a.fieldIndex = 0
			a.initViewInputs()
		}
	}
	return a, nil
}

func (a *App) updateBarSettingsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.applyBarSettings()
		a.view = ViewMain
	case key.Matches(msg, a.keyMap.Up):
		if a.fieldIndex > 0 {
			a.textInputs[a.fieldIndex].Blur()
			a.fieldIndex--
			a.textInputs[a.fieldIndex].Focus()
		}
	case key.Matches(msg, a.keyMap.Down), key.Matches(msg, a.keyMap.Tab):
		if a.fieldIndex < len(a.textInputs)-1 {
			a.textInputs[a.fieldIndex].Blur()
			a.fieldIndex++
			a.textInputs[a.fieldIndex].Focus()
		}
	default:
		var cmd tea.Cmd
		a.textInputs[a.fieldIndex], cmd = a.textInputs[a.fieldIndex].Update(msg)
		a.hasChanges = true
		return a, cmd
	}
	return a, nil
}

func (a *App) updateModulesListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	modules := a.getCurrentModulesList()

	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.moving = false
		a.view = ViewMain
	case key.Matches(msg, a.keyMap.Up):
		if a.moving && a.listIndex > 0 {
			a.swapModules(a.listIndex, a.listIndex-1)
			a.listIndex--
			a.hasChanges = true
		} else if a.listIndex > 0 {
			a.listIndex--
		}
	case key.Matches(msg, a.keyMap.Down):
		if a.moving && a.listIndex < len(modules)-1 {
			a.swapModules(a.listIndex, a.listIndex+1)
			a.listIndex++
			a.hasChanges = true
		} else if a.listIndex < len(modules)-1 {
			a.listIndex++
		}
	case key.Matches(msg, a.keyMap.Enter):
		if len(modules) > 0 && a.listIndex < len(modules) {
			a.editingModule = modules[a.listIndex]
			a.view = ViewModuleEditor
			a.initModuleEditorInputs()
		}
	case key.Matches(msg, a.keyMap.Add):
		a.view = ViewModuleAdd
		a.addModuleCategory = 0
		a.addModuleIndex = 0
	case key.Matches(msg, a.keyMap.Delete):
		if len(modules) > 0 && a.listIndex < len(modules) {
			a.deleteModule(a.listIndex)
			if a.listIndex >= len(a.getCurrentModulesList()) && a.listIndex > 0 {
				a.listIndex--
			}
			a.hasChanges = true
		}
	case key.Matches(msg, a.keyMap.Move):
		if len(modules) > 0 {
			a.moving = !a.moving
		}
	}
	return a, nil
}

func (a *App) updateModuleEditorView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.applyModuleSettings()
		a.view = a.previousView
		if a.view == ViewMain {
			// Determine correct modules view
			for _, m := range a.config.ModulesLeft {
				if m == a.editingModule {
					a.view = ViewModulesLeft
					break
				}
			}
			for _, m := range a.config.ModulesCenter {
				if m == a.editingModule {
					a.view = ViewModulesCenter
					break
				}
			}
			for _, m := range a.config.ModulesRight {
				if m == a.editingModule {
					a.view = ViewModulesRight
					break
				}
			}
		}
	case key.Matches(msg, a.keyMap.Up):
		if a.fieldIndex > 0 {
			a.textInputs[a.fieldIndex].Blur()
			a.fieldIndex--
			a.textInputs[a.fieldIndex].Focus()
		}
	case key.Matches(msg, a.keyMap.Down), key.Matches(msg, a.keyMap.Tab):
		if a.fieldIndex < len(a.textInputs)-1 {
			a.textInputs[a.fieldIndex].Blur()
			a.fieldIndex++
			a.textInputs[a.fieldIndex].Focus()
		}
	default:
		var cmd tea.Cmd
		a.textInputs[a.fieldIndex], cmd = a.textInputs[a.fieldIndex].Update(msg)
		a.hasChanges = true
		return a, cmd
	}
	return a, nil
}

func (a *App) updateModuleAddView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	categories := config.GetModuleCategories()

	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = a.previousView
	case key.Matches(msg, a.keyMap.Left):
		if a.addModuleCategory > 0 {
			a.addModuleCategory--
			a.addModuleIndex = 0
		}
	case key.Matches(msg, a.keyMap.Right):
		if a.addModuleCategory < len(categories)-1 {
			a.addModuleCategory++
			a.addModuleIndex = 0
		}
	case key.Matches(msg, a.keyMap.Up):
		if a.addModuleIndex > 0 {
			a.addModuleIndex--
		}
	case key.Matches(msg, a.keyMap.Down):
		if a.addModuleCategory < len(categories) {
			mods := categories[a.addModuleCategory].Modules
			if a.addModuleIndex < len(mods)-1 {
				a.addModuleIndex++
			}
		}
	case key.Matches(msg, a.keyMap.Enter):
		if a.addModuleCategory < len(categories) {
			mods := categories[a.addModuleCategory].Modules
			if a.addModuleIndex < len(mods) {
				moduleName := mods[a.addModuleIndex].Name
				// Handle custom/ and group/ modules
				if strings.HasSuffix(moduleName, "/") {
					moduleName = moduleName + "name"
				}
				a.addModule(moduleName)
				a.hasChanges = true
				a.view = a.previousView
			}
		}
	}
	return a, nil
}

func (a *App) updateStyleEditorView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = ViewMain
	}
	// For now, style editor is read-only in this simplified version
	// A full implementation would use a textarea component
	return a, nil
}

func (a *App) updateBackupsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	backups, _ := a.configManager.GetBackups()

	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = ViewMain
	case key.Matches(msg, a.keyMap.Up):
		if a.listIndex > 0 {
			a.listIndex--
		}
	case key.Matches(msg, a.keyMap.Down):
		if a.listIndex < len(backups)-1 {
			a.listIndex++
		}
	case key.Matches(msg, a.keyMap.Enter):
		if len(backups) > 0 && a.listIndex < len(backups) {
			if err := a.configManager.RestoreBackup(backups[a.listIndex]); err != nil {
				a.notification = fmt.Sprintf("Restore failed: %v", err)
				a.notificationType = "error"
			} else {
				a.loadConfig()
				a.notification = "Backup restored successfully"
				a.notificationType = "success"
			}
		}
	}
	return a, nil
}

func (a *App) updateHelpView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keyMap.Back), key.Matches(msg, a.keyMap.Enter):
		a.view = a.previousView
	}
	return a, nil
}

func (a *App) getMainMenuItems() []MenuItem {
	return []MenuItem{
		{Title: "Bar Settings", Description: "Configure bar position, size, and behavior", Icon: styles.IconGear, View: ViewBarSettings},
		{Title: "Modules Left", Description: "Manage left-aligned modules", Icon: styles.IconArrowLeft, View: ViewModulesLeft},
		{Title: "Modules Center", Description: "Manage center-aligned modules", Icon: styles.IconDot, View: ViewModulesCenter},
		{Title: "Modules Right", Description: "Manage right-aligned modules", Icon: styles.IconArrowRight, View: ViewModulesRight},
		{Title: "Style Editor", Description: "Edit CSS styling", Icon: styles.IconPalette, View: ViewStyleEditor},
		{Title: "Backups", Description: "Manage configuration backups", Icon: styles.IconFolder, View: ViewBackups},
	}
}

func (a *App) getCurrentModulesList() []string {
	switch a.view {
	case ViewModulesLeft:
		return a.config.ModulesLeft
	case ViewModulesCenter:
		return a.config.ModulesCenter
	case ViewModulesRight:
		return a.config.ModulesRight
	}
	return nil
}

func (a *App) swapModules(i, j int) {
	switch a.view {
	case ViewModulesLeft:
		a.config.ModulesLeft[i], a.config.ModulesLeft[j] = a.config.ModulesLeft[j], a.config.ModulesLeft[i]
	case ViewModulesCenter:
		a.config.ModulesCenter[i], a.config.ModulesCenter[j] = a.config.ModulesCenter[j], a.config.ModulesCenter[i]
	case ViewModulesRight:
		a.config.ModulesRight[i], a.config.ModulesRight[j] = a.config.ModulesRight[j], a.config.ModulesRight[i]
	}
}

func (a *App) deleteModule(index int) {
	switch a.view {
	case ViewModulesLeft:
		a.config.ModulesLeft = append(a.config.ModulesLeft[:index], a.config.ModulesLeft[index+1:]...)
	case ViewModulesCenter:
		a.config.ModulesCenter = append(a.config.ModulesCenter[:index], a.config.ModulesCenter[index+1:]...)
	case ViewModulesRight:
		a.config.ModulesRight = append(a.config.ModulesRight[:index], a.config.ModulesRight[index+1:]...)
	}
}

func (a *App) addModule(name string) {
	a.previousView = a.view
	switch a.view {
	case ViewModulesLeft:
		a.config.ModulesLeft = append(a.config.ModulesLeft, name)
	case ViewModulesCenter:
		a.config.ModulesCenter = append(a.config.ModulesCenter, name)
	case ViewModulesRight:
		a.config.ModulesRight = append(a.config.ModulesRight, name)
	}
}

func (a *App) initViewInputs() {
	a.previousView = a.view
	switch a.view {
	case ViewBarSettings:
		a.initBarSettingsInputs()
	}
}

func (a *App) initBarSettingsInputs() {
	a.textInputs = make([]textinput.Model, 8)

	fields := []struct {
		placeholder string
		value       string
	}{
		{"position (top/bottom/left/right)", a.config.Position},
		{"layer (top/bottom)", a.config.Layer},
		{"height", fmt.Sprintf("%d", a.config.Height)},
		{"width (0 = auto)", fmt.Sprintf("%d", a.config.Width)},
		{"spacing", fmt.Sprintf("%d", a.config.Spacing)},
		{"margin", a.config.Margin},
		{"mode (dock/hide/invisible/overlay)", a.config.Mode},
		{"name", a.config.Name},
	}

	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.placeholder
		ti.SetValue(f.value)
		ti.CharLimit = 100
		ti.Width = 40
		if i == 0 {
			ti.Focus()
		}
		a.textInputs[i] = ti
	}
	a.fieldIndex = 0
}

func (a *App) applyBarSettings() {
	if len(a.textInputs) < 8 {
		return
	}

	a.config.Position = a.textInputs[0].Value()
	a.config.Layer = a.textInputs[1].Value()
	fmt.Sscanf(a.textInputs[2].Value(), "%d", &a.config.Height)
	fmt.Sscanf(a.textInputs[3].Value(), "%d", &a.config.Width)
	fmt.Sscanf(a.textInputs[4].Value(), "%d", &a.config.Spacing)
	a.config.Margin = a.textInputs[5].Value()
	a.config.Mode = a.textInputs[6].Value()
	a.config.Name = a.textInputs[7].Value()
}

func (a *App) initModuleEditorInputs() {
	moduleDef := config.GetModuleDefinition(a.editingModule)
	moduleConfig := a.config.GetModuleConfig(a.editingModule)

	// Combine common and module-specific properties
	allProps := config.CommonProperties()
	if moduleDef != nil {
		allProps = append(allProps, moduleDef.Properties...)
	}

	a.moduleProperties = make([]string, len(allProps))
	a.textInputs = make([]textinput.Model, len(allProps))

	for i, prop := range allProps {
		a.moduleProperties[i] = prop.Name

		ti := textinput.New()
		ti.Placeholder = prop.Description
		ti.CharLimit = 500
		ti.Width = 50

		// Set current value if exists
		if val, ok := moduleConfig[prop.Name]; ok {
			switch v := val.(type) {
			case string:
				ti.SetValue(v)
			case float64:
				ti.SetValue(fmt.Sprintf("%v", v))
			case bool:
				ti.SetValue(fmt.Sprintf("%v", v))
			default:
				// For complex types, show JSON
				ti.SetValue(fmt.Sprintf("%v", v))
			}
		}

		if i == 0 {
			ti.Focus()
		}
		a.textInputs[i] = ti
	}
	a.fieldIndex = 0
}

func (a *App) applyModuleSettings() {
	if len(a.moduleProperties) == 0 {
		return
	}

	moduleConfig := make(map[string]interface{})

	for i, propName := range a.moduleProperties {
		if i < len(a.textInputs) {
			value := a.textInputs[i].Value()
			if value != "" {
				// Try to parse as number or bool
				var numVal float64
				var boolVal bool
				if _, err := fmt.Sscanf(value, "%f", &numVal); err == nil {
					moduleConfig[propName] = numVal
				} else if _, err := fmt.Sscanf(value, "%t", &boolVal); err == nil {
					moduleConfig[propName] = boolVal
				} else {
					moduleConfig[propName] = value
				}
			}
		}
	}

	a.config.SetModuleConfig(a.editingModule, moduleConfig)
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var content string

	switch a.view {
	case ViewMain:
		content = a.renderMainView()
	case ViewBarSettings:
		content = a.renderBarSettingsView()
	case ViewModulesLeft, ViewModulesCenter, ViewModulesRight:
		content = a.renderModulesListView()
	case ViewModuleEditor:
		content = a.renderModuleEditorView()
	case ViewModuleAdd:
		content = a.renderModuleAddView()
	case ViewStyleEditor:
		content = a.renderStyleEditorView()
	case ViewBackups:
		content = a.renderBackupsView()
	case ViewHelp:
		content = a.renderHelpView()
	}

	return a.renderLayout(content)
}

func (a *App) renderLayout(content string) string {
	// Header
	header := a.renderHeader()

	// Footer with status and help
	footer := a.renderFooter()

	// Calculate content height
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := a.height - headerHeight - footerHeight - 2

	// Content area
	contentStyle := lipgloss.NewStyle().
		Width(a.width - 4).
		Height(contentHeight).
		Padding(1, 2)

	contentArea := contentStyle.Render(content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		contentArea,
		footer,
	)
}

func (a *App) renderHeader() string {
	// Compact header: WABAGO | status | path
	logo := styles.LogoStyle.Render("WABAGO")

	// Status indicator (compact)
	var status string
	if a.hasChanges {
		status = styles.NotifyWarningStyle.Render("*")
	} else {
		status = styles.NotifySuccessStyle.Render("~")
	}

	// Path info (truncate if needed)
	configPath := a.configPath
	maxPathLen := a.width - 30
	if maxPathLen < 20 {
		maxPathLen = 20
	}
	if len(configPath) > maxPathLen {
		configPath = "..." + configPath[len(configPath)-maxPathLen+3:]
	}
	pathInfo := styles.TextMutedStyle.Render(configPath)

	headerLine := logo + " " + status + " " + pathInfo

	headerStyle := styles.HeaderStyle.Width(a.width - 4)
	return headerStyle.Render(headerLine)
}

func (a *App) renderFooter() string {
	// Notification
	var notifyBar string
	if a.notification != "" {
		switch a.notificationType {
		case "success":
			notifyBar = styles.NotifySuccessStyle.Width(a.width - 4).Render(a.notification)
		case "warning":
			notifyBar = styles.NotifyWarningStyle.Width(a.width - 4).Render(a.notification)
		case "error":
			notifyBar = styles.NotifyErrorStyle.Width(a.width - 4).Render(a.notification)
		default:
			notifyBar = styles.NotifyInfoStyle.Width(a.width - 4).Render(a.notification)
		}
	}

	// Help line
	helpKeys := []string{
		styles.HelpKeyStyle.Render("?") + styles.HelpDescStyle.Render(" help"),
		styles.HelpKeyStyle.Render("↑↓") + styles.HelpDescStyle.Render(" navigate"),
		styles.HelpKeyStyle.Render("enter") + styles.HelpDescStyle.Render(" select"),
		styles.HelpKeyStyle.Render("esc") + styles.HelpDescStyle.Render(" back"),
		styles.HelpKeyStyle.Render("ctrl+s") + styles.HelpDescStyle.Render(" save"),
		styles.HelpKeyStyle.Render("q") + styles.HelpDescStyle.Render(" quit"),
	}

	helpLine := styles.StatusBarStyle.Width(a.width - 4).Render(strings.Join(helpKeys, "  "))

	if notifyBar != "" {
		return lipgloss.JoinVertical(lipgloss.Left, notifyBar, helpLine)
	}
	return helpLine
}

func (a *App) renderMainView() string {
	menuItems := a.getMainMenuItems()

	var items []string
	for i, item := range menuItems {
		icon := styles.CategoryIconStyle.Render(item.Icon)
		title := item.Title
		desc := styles.DescriptionStyle.Render(item.Description)

		if i == a.menuIndex {
			title = styles.MenuSelectedStyle.Render(title)
			icon = styles.ActiveItemStyle.Render(styles.IconSelected)
		} else {
			title = styles.MenuItemStyle.Render(title)
			icon = styles.ListItemStyle.Render(" ")
		}

		itemLine := lipgloss.JoinHorizontal(lipgloss.Center, icon, " ", title)
		items = append(items, lipgloss.JoinVertical(lipgloss.Left, itemLine, desc))
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	menuBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(menu)

	titleStyle := styles.TitleStyle.Copy().MarginBottom(1)
	title := titleStyle.Render("Main Menu")

	return lipgloss.JoinVertical(lipgloss.Left, title, menuBox)
}

func (a *App) renderBarSettingsView() string {
	title := styles.TitleStyle.Render("Bar Settings")

	fields := []string{
		"Position", "Layer", "Height", "Width",
		"Spacing", "Margin", "Mode", "Name",
	}

	var formItems []string
	for i, field := range fields {
		label := styles.LabelStyle.Render(field + ":")

		var input string
		if i < len(a.textInputs) {
			if i == a.fieldIndex {
				input = styles.FocusedInputStyle.Render(a.textInputs[i].View())
			} else {
				input = styles.InputStyle.Render(a.textInputs[i].View())
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Center, label, input)
		formItems = append(formItems, row)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)

	formBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(form)

	hint := styles.DescriptionStyle.Render("Use ↑↓/Tab to navigate fields, Esc to go back")

	return lipgloss.JoinVertical(lipgloss.Left, title, formBox, hint)
}

func (a *App) renderModulesListView() string {
	var titleText string
	switch a.view {
	case ViewModulesLeft:
		titleText = "Left Modules"
	case ViewModulesCenter:
		titleText = "Center Modules"
	case ViewModulesRight:
		titleText = "Right Modules"
	}

	title := styles.TitleStyle.Render(titleText)

	modules := a.getCurrentModulesList()

	if len(modules) == 0 {
		empty := styles.EmptyStateStyle.Render("No modules configured\nPress 'a' to add a module")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty)
	}

	var items []string
	for i, mod := range modules {
		moduleDef := config.GetModuleDefinition(mod)

		var desc string
		if moduleDef != nil {
			desc = moduleDef.Description
		}

		var item string
		if i == a.listIndex {
			icon := styles.IconSelected
			if a.moving {
				icon = styles.IconTriangle
			}
			item = styles.SelectedItemStyle.Render(fmt.Sprintf("%s %s", icon, mod))
		} else {
			item = styles.ListItemStyle.Render(fmt.Sprintf("  %s", mod))
		}

		if desc != "" {
			item = lipgloss.JoinVertical(lipgloss.Left, item, styles.DescriptionStyle.Render("  "+desc))
		}

		items = append(items, item)
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)

	listBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(list)

	var modeHint string
	if a.moving {
		modeHint = styles.TagActiveStyle.Render("MOVE MODE") + " Use ↑↓ to reorder, m to exit move mode"
	} else {
		modeHint = "a: add  d: delete  m: move  enter: edit"
	}
	hint := styles.DescriptionStyle.Render(modeHint)

	return lipgloss.JoinVertical(lipgloss.Left, title, listBox, hint)
}

func (a *App) renderModuleEditorView() string {
	title := styles.TitleStyle.Render("Edit Module: " + a.editingModule)

	moduleDef := config.GetModuleDefinition(a.editingModule)
	var moduleDesc string
	if moduleDef != nil {
		moduleDesc = styles.DescriptionStyle.Render(moduleDef.Description)
	}

	// Show scrollable list of properties
	startIdx := 0
	maxVisible := 10

	if a.fieldIndex >= maxVisible {
		startIdx = a.fieldIndex - maxVisible + 1
	}

	endIdx := startIdx + maxVisible
	if endIdx > len(a.moduleProperties) {
		endIdx = len(a.moduleProperties)
	}

	var formItems []string
	for i := startIdx; i < endIdx; i++ {
		propName := a.moduleProperties[i]
		label := styles.LabelStyle.Width(25).Render(propName + ":")

		var input string
		if i < len(a.textInputs) {
			if i == a.fieldIndex {
				input = styles.FocusedInputStyle.Render(a.textInputs[i].View())
			} else {
				input = styles.InputStyle.Render(a.textInputs[i].View())
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Center, label, input)
		formItems = append(formItems, row)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)

	formBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(form)

	scrollInfo := fmt.Sprintf("Showing %d-%d of %d properties", startIdx+1, endIdx, len(a.moduleProperties))
	hint := styles.DescriptionStyle.Render(scrollInfo + " | Use ↑↓/Tab to navigate, Esc to save and go back")

	return lipgloss.JoinVertical(lipgloss.Left, title, moduleDesc, formBox, hint)
}

func (a *App) renderModuleAddView() string {
	title := styles.TitleStyle.Render("Add Module")

	categories := config.GetModuleCategories()

	// Render category tabs
	var tabs []string
	for i, cat := range categories {
		if i == a.addModuleCategory {
			tabs = append(tabs, styles.ActiveTabStyle.Render(cat.Name))
		} else {
			tabs = append(tabs, styles.TabStyle.Render(cat.Name))
		}
	}
	tabBar := styles.TabBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))

	// Render modules in selected category
	var items []string
	if a.addModuleCategory < len(categories) {
		mods := categories[a.addModuleCategory].Modules
		for i, mod := range mods {
			var item string
			if i == a.addModuleIndex {
				item = styles.SelectedItemStyle.Render(fmt.Sprintf("%s %s", styles.IconSelected, mod.Name))
			} else {
				item = styles.ListItemStyle.Render(fmt.Sprintf("  %s", mod.Name))
			}
			desc := styles.DescriptionStyle.Render("  " + mod.Description)
			items = append(items, lipgloss.JoinVertical(lipgloss.Left, item, desc))
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)

	listBox := styles.BoxStyle.
		Width(a.width - 10).
		Height(styles.Min(len(items)*3+2, 15)).
		Render(list)

	hint := styles.DescriptionStyle.Render("Use ←→ to switch categories, ↑↓ to select module, Enter to add")

	return lipgloss.JoinVertical(lipgloss.Left, title, tabBar, listBox, hint)
}

func (a *App) renderStyleEditorView() string {
	title := styles.TitleStyle.Render("Style Editor (CSS)")

	// Show CSS content with line numbers
	lines := strings.Split(a.styleContent, "\n")
	maxLines := 20

	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var codeLines []string
	for i, line := range lines {
		lineNum := styles.LineNumberStyle.Render(fmt.Sprintf("%3d", i+1))
		code := styles.CodeStyle.Render(line)
		codeLines = append(codeLines, lipgloss.JoinHorizontal(lipgloss.Top, lineNum, code))
	}

	code := lipgloss.JoinVertical(lipgloss.Left, codeLines...)

	codeBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(code)

	if len(lines) < len(strings.Split(a.styleContent, "\n")) {
		truncated := styles.DescriptionStyle.Render(fmt.Sprintf("... and %d more lines", len(strings.Split(a.styleContent, "\n"))-maxLines))
		codeBox = lipgloss.JoinVertical(lipgloss.Left, codeBox, truncated)
	}

	hint := styles.DescriptionStyle.Render("Style file: " + a.stylePath + " | Press Esc to go back")

	return lipgloss.JoinVertical(lipgloss.Left, title, codeBox, hint)
}

func (a *App) renderBackupsView() string {
	title := styles.TitleStyle.Render("Configuration Backups")

	backups, _ := a.configManager.GetBackups()

	if len(backups) == 0 {
		empty := styles.EmptyStateStyle.Render("No backups available\nPress 'b' to create a backup")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty)
	}

	var items []string
	for i, backup := range backups {
		var item string
		if i == a.listIndex {
			item = styles.SelectedItemStyle.Render(fmt.Sprintf("%s %s", styles.IconSelected, backup))
		} else {
			item = styles.ListItemStyle.Render(fmt.Sprintf("  %s", backup))
		}
		items = append(items, item)
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)

	listBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(list)

	hint := styles.DescriptionStyle.Render("Press Enter to restore selected backup, b to create new backup")

	return lipgloss.JoinVertical(lipgloss.Left, title, listBox, hint)
}

func (a *App) renderHelpView() string {
	title := styles.TitleStyle.Render("Keyboard Shortcuts")

	sections := []struct {
		name string
		keys []struct{ key, desc string }
	}{
		{
			name: "Navigation",
			keys: []struct{ key, desc string }{
				{"↑ / k", "Move up"},
				{"↓ / j", "Move down"},
				{"← / h", "Move left / Previous category"},
				{"→ / l", "Move right / Next category"},
				{"Enter", "Select / Confirm"},
				{"Esc", "Go back / Cancel"},
				{"Tab", "Next field"},
			},
		},
		{
			name: "Actions",
			keys: []struct{ key, desc string }{
				{"a", "Add new module"},
				{"d / Delete", "Delete selected module"},
				{"m", "Toggle move mode (reorder modules)"},
				{"Ctrl+S", "Save configuration"},
				{"r", "Reload configuration from disk"},
				{"b", "Create backup"},
			},
		},
		{
			name: "General",
			keys: []struct{ key, desc string }{
				{"?", "Show/hide this help"},
				{"q / Ctrl+C", "Quit application"},
			},
		},
	}

	var sectionViews []string
	for _, section := range sections {
		sectionTitle := styles.CategoryStyle.Render(section.name)

		var keyRows []string
		for _, k := range section.keys {
			key := styles.HelpKeyStyle.Width(15).Render(k.key)
			desc := styles.HelpDescStyle.Render(k.desc)
			keyRows = append(keyRows, lipgloss.JoinHorizontal(lipgloss.Center, key, desc))
		}

		sectionContent := lipgloss.JoinVertical(lipgloss.Left, keyRows...)
		sectionViews = append(sectionViews, lipgloss.JoinVertical(lipgloss.Left, sectionTitle, sectionContent))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sectionViews...)

	helpBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(content)

	hint := styles.DescriptionStyle.Render("Press Esc or ? to close help")

	return lipgloss.JoinVertical(lipgloss.Left, title, helpBox, hint)
}
