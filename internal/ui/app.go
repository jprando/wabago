// ui implements the Bubble Tea Model interface methods (Init, Update, View) and rendering logic for the application.
package ui

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jprando/wabago/internal/config"
	"github.com/jprando/wabago/internal/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// modifiedItemDelegate is a custom list delegate that highlights modified items
type modifiedItemDelegate struct {
	list.DefaultDelegate
}

func newModifiedItemDelegate() modifiedItemDelegate {
	d := modifiedItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	d.ShowDescription = true
	d.SetSpacing(0)
	return d
}

func (d modifiedItemDelegate) Render(w io.Writer, model list.Model, index int, item list.Item) {
	// Check if this is a PropertyItem and if it's modified
	if propItem, ok := item.(PropertyItem); ok && propItem.isModified {
		// Render with modified style
		title := d.Styles.NormalTitle.
			Foreground(lipgloss.Color("226")).  // Yellow/orange color for modified items
			Bold(true).
			Render(propItem.Title())

		desc := d.Styles.NormalDesc.
			Foreground(lipgloss.Color("243")).
			Render(propItem.Description())

		// Check if this item is selected
		if index == model.Index() {
			title = d.Styles.SelectedTitle.
				Foreground(lipgloss.Color("226")).
				Bold(true).
				Render("» " + propItem.Title())
			desc = d.Styles.SelectedDesc.
				Foreground(lipgloss.Color("243")).
				Render(propItem.Description())
		}

		fmt.Fprint(w, lipgloss.JoinVertical(lipgloss.Left, title, desc))
		return
	}

	// Default rendering for non-modified items
	d.DefaultDelegate.Render(w, model, index, item)
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
		a.notification = fmt.Sprintf("Load error: %v", err)
		a.notificationType = "error"
	} else {
		a.config = cfg
		a.notification = ""
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
	currentView := a.view
	model, cmd := a.handleUpdate(msg)

	// Force full redraw if view changed
	if app, ok := model.(*App); ok && app.view != currentView {
		return model, tea.Batch(cmd, tea.ClearScreen)
	}
	return model, cmd
}

// handleUpdate handles messages (internal)
func (a *App) handleUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Save original message for views that need it (like List filtering)
	originalMsg := msg

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

		// Module Editor needs original tea.Msg for list filtering
		// If filtering is active, skip global keys and pass directly to view
		if a.view == ViewModuleEditor {
			if a.propertyList.FilterState() == list.Filtering {
				// Skip global keys while filtering - pass message directly
				return a.updateModuleEditorView(originalMsg)
			}
		}

		// Bar Settings needs original tea.Msg for list filtering
		// If filtering is active, skip global keys and pass directly to view
		if a.view == ViewBarSettings {
			if a.barSettingsList.FilterState() == list.Filtering {
				// Skip global keys while filtering - pass message directly
				return a.updateBarSettingsView(originalMsg)
			}
		}

		// Main Menu needs original tea.Msg for list filtering
		// If filtering is active, skip global keys and pass directly to view
		if a.view == ViewMain {
			if a.mainMenuList.FilterState() == list.Filtering {
				// Skip global keys while filtering - pass message directly
				return a.updateMainView(originalMsg)
			}
		}

		// Modules List needs original tea.Msg for list filtering
		// If filtering is active, skip global keys and pass directly to view
		if a.view == ViewModulesLeft || a.view == ViewModulesCenter || a.view == ViewModulesRight {
			if a.modulesList.FilterState() == list.Filtering {
				// Skip global keys while filtering - pass message directly
				return a.updateModulesListView(originalMsg)
			}
		}

		// Global keys (processed only when NOT filtering)
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

		// Module Editor needs original tea.Msg for list filtering
		// After global keys are handled, pass the original message
		if a.view == ViewModuleEditor {
			return a.updateModuleEditorView(originalMsg)
		}

		// Bar Settings needs original tea.Msg for list filtering
		// After global keys are handled, pass the original message
		if a.view == ViewBarSettings {
			return a.updateBarSettingsView(originalMsg)
		}

		// Main Menu needs original tea.Msg for list filtering
		// After global keys are handled, pass the original message
		if a.view == ViewMain {
			return a.updateMainView(originalMsg)
		}

		// Modules List needs original tea.Msg for list filtering
		// After global keys are handled, pass the original message
		if a.view == ViewModulesLeft || a.view == ViewModulesCenter || a.view == ViewModulesRight {
			return a.updateModulesListView(originalMsg)
		}

		// View-specific handling
		switch a.view {
		case ViewModuleAdd:
			return a.updateModuleAddView(msg)
		case ViewStyleEditor:
			return a.updateStyleEditorView(msg)
		case ViewBackups:
			return a.updateBackupsView(msg)
		case ViewModuleCatalog:
			return a.updateModuleCatalogView(msg)
		case ViewDiff:
			return a.updateDiffView(msg)
		case ViewRestoreConfirm:
			return a.updateRestoreConfirmView(msg)
		case ViewModal:
			return a.updateModalView(msg)
		case ViewHelp:
			return a.updateHelpView(msg)
		}
	
	case restoreConfirmMsg:
		a.view = ViewRestoreConfirm
		return a, nil
	}

	// Update text inputs if any (only for modal input now)
	var cmd tea.Cmd
	if a.modal.isActive && a.modal.mode == ModalTypeInput {
		a.modal.input, cmd = a.modal.input.Update(msg)
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

func (a *App) reloadWaybar() tea.Cmd {
	return func() tea.Msg {
		// 1. Save current config first
		if err := a.configManager.SaveConfig(a.config); err != nil {
			a.notification = fmt.Sprintf("Failed to save before reload: %v", err)
			a.notificationType = "error"
			return nil
		}
		if err := a.configManager.SaveStyle(a.styleContent); err != nil {
			a.notification = fmt.Sprintf("Failed to save style before reload: %v", err)
			a.notificationType = "error"
			return nil
		}

		// 2. Send signal to Waybar
		// Using pkill -SIGUSR2 waybar
		cmd := exec.Command("pkill", "-SIGUSR2", "waybar")
		if err := cmd.Run(); err != nil {
			// If pkill fails, maybe waybar isn't running?
			// Let's check
			if err := exec.Command("pgrep", "waybar").Run(); err != nil {
				a.notification = "Waybar is not running. Reload failed."
				a.notificationType = "warning"
				return nil
			}
			a.notification = fmt.Sprintf("Failed to signal waybar: %v", err)
			a.notificationType = "error"
			return nil
		}

		// 3. Wait and Verify
		time.Sleep(1 * time.Second)

		// Check if waybar is still running
		if err := exec.Command("pgrep", "waybar").Run(); err != nil {
			// Waybar died!
			// Return a message that triggers the Restore Confirm View
			return restoreConfirmMsg{}
		}

		a.notification = "Waybar reloaded successfully"
		a.notificationType = "success"
		return nil
	}
}



func (a *App) openSelectModal(title string, options []string, targetIndex int, currentVal string) {
	// Add "Clear Value" option at the end
	extendedOptions := make([]string, len(options)+1)
	copy(extendedOptions, options)
	extendedOptions[len(options)] = "─── Clear Value ───"

	a.modal = ModalState{
		isActive:      true,
		mode:          ModalTypeSelect,
		title:         title,
		options:       extendedOptions,
		index:         0,
		targetIndex:   targetIndex,
		originalValue: currentVal,
	}
	// Try to find current value in options to select it
	for i, opt := range options {
		if opt == currentVal {
			a.modal.index = i
			break
		}
	}
	a.view = ViewModal
}

func (a *App) openInputModal(title string, targetIndex int, currentVal string) {
	ti := textinput.New()
	ti.SetValue(currentVal)
	ti.Focus()
	ti.Width = 40

	a.modal = ModalState{
		isActive:      true,
		mode:          ModalTypeInput,
		title:         title,
		input:         ti,
		originalValue: currentVal,
		targetIndex:   targetIndex,
	}
	a.view = ViewModal
}

func (a *App) updateModalView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.modal.mode == ModalTypeSelect {
		// Select Mode: Both Esc and Backspace close the modal
		switch {
		case key.Matches(msg, a.keyMap.Back):
			// Cancel
			a.view = a.previousView
			a.modal.isActive = false
			return a, nil
		case key.Matches(msg, a.keyMap.Up):
			if a.modal.index > 0 {
				a.modal.index--
			}
		case key.Matches(msg, a.keyMap.Down):
			if a.modal.index < len(a.modal.options)-1 {
				a.modal.index++
			}
		case key.Matches(msg, a.keyMap.Enter):
			// Select
			if a.modal.index < len(a.modal.options) {
				a.applyModalValue(a.modal.options[a.modal.index])
			}
			a.view = a.previousView
			a.modal.isActive = false
			return a, nil
		}
	} else {
		// Input Mode: Only Esc closes the modal, backspace edits text
		switch {
		case msg.String() == "esc":
			// Cancel - only Esc closes, not backspace
			a.view = a.previousView
			a.modal.isActive = false
			return a, nil
		case msg.String() == "ctrl+d":
			// Clear value - Ctrl+D
			a.applyModalValue("")
			a.view = a.previousView
			a.modal.isActive = false
			return a, nil
		case key.Matches(msg, a.keyMap.Enter):
			// Confirm
			a.applyModalValue(a.modal.input.Value())
			a.view = a.previousView
			a.modal.isActive = false
			return a, nil
		default:
			// Pass all other keys (including backspace) to the input
			var cmd tea.Cmd
			a.modal.input, cmd = a.modal.input.Update(msg)
			return a, cmd
		}
	}
	return a, nil
}

func (a *App) applyModalValue(value string) {
	a.hasChanges = true
	
	switch a.previousView {
	case ViewBarSettings:
		a.updateBarSetting(a.modal.targetIndex, value)
	case ViewModuleEditor:
		a.updateModuleProperty(a.modal.targetIndex, value)
	case ViewModuleCatalog:
		if a.modal.targetIndex == -1 {
			// Module Actions
			switch value {
			case "Configure":
				a.view = ViewModuleEditor
				a.initPropertyList()
			case "Remove":
				// Remove from all lists
				var newLeft, newCenter, newRight []string
				for _, m := range a.config.ModulesLeft { if m != a.editingModule { newLeft = append(newLeft, m) } }
				for _, m := range a.config.ModulesCenter { if m != a.editingModule { newCenter = append(newCenter, m) } }
				for _, m := range a.config.ModulesRight { if m != a.editingModule { newRight = append(newRight, m) } }
				a.config.ModulesLeft = newLeft
				a.config.ModulesCenter = newCenter
				a.config.ModulesRight = newRight
				a.hasChanges = true
			}
		} else if a.modal.targetIndex == -2 {
			// Add Module To
			var position string
			switch value {
			case "Left":
				a.config.ModulesLeft = append(a.config.ModulesLeft, a.editingModule)
				a.hasChanges = true
				position = "Left"
			case "Center":
				a.config.ModulesCenter = append(a.config.ModulesCenter, a.editingModule)
				a.hasChanges = true
				position = "Center"
			case "Right":
				a.config.ModulesRight = append(a.config.ModulesRight, a.editingModule)
				a.hasChanges = true
				position = "Right"
			}

			// Initialize module config if it doesn't exist
			// This ensures the module will be saved with its configuration
			if value != "Cancel" && value != "" {
				if _, exists := a.config.Modules[a.editingModule]; !exists {
					a.config.Modules[a.editingModule] = make(map[string]interface{})
				}

				// Show notification with reminder to save
				a.notification = fmt.Sprintf("Added '%s' to Modules %s - Press Ctrl+S to save!", a.editingModule, position)
				a.notificationType = "success"
			}
		}
	}
}

func (a *App) renderModalView() string {
	// We want to render the background view slightly dimmed or just overlay the modal
	// Since lipgloss doesn't support layering fully in a single pass without complex layout,
	// we will render the modal centered on the screen.
	
	title := styles.ModalTitleStyle.Render(a.modal.title)
	
	var content string
	
	if a.modal.mode == ModalTypeSelect {
		var options []string
		for i, opt := range a.modal.options {
			if i == a.modal.index {
				options = append(options, styles.MenuSelectedStyle.Render(opt))
			} else {
				options = append(options, styles.MenuItemStyle.Render(opt))
			}
		}
		content = lipgloss.JoinVertical(lipgloss.Left, options...)
	} else {
		// Input Mode
		original := styles.DescriptionStyle.Render(fmt.Sprintf("Current: %s", a.modal.originalValue))
		input := styles.FocusedInputStyle.Render(a.modal.input.View())

		confirm := styles.ButtonStyle.Render("[Enter] Confirm")
		clear := styles.ButtonStyle.Render("[Ctrl+D] Clear")
		cancel := styles.ButtonDangerStyle.Render("[Esc] Cancel")
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, confirm, "  ", clear, "  ", cancel)

		content = lipgloss.JoinVertical(lipgloss.Center, original, "", input, "", buttons)
	}
	
	// Calculate height to center
	fullContent := lipgloss.JoinVertical(lipgloss.Center, title, content)
	modalBox := styles.ModalStyle.Render(fullContent)
	
	// Center vertically and horizontally
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, modalBox)
}

func (a *App) updateBarSetting(index int, value string) {
	// Handle clear value
	if value == "─── Clear Value ───" || value == "" {
		switch index {
		case 0:
			a.config.Position = ""
		case 1:
			a.config.Layer = ""
		case 2:
			a.config.Height = 0
		case 3:
			a.config.Width = 0
		case 4:
			a.config.Spacing = 0
		case 5:
			a.config.Margin = ""
		case 6:
			a.config.Mode = ""
		case 7:
			a.config.Name = ""
		}
		a.hasChanges = true
		// Refresh the list to show cleared value
		a.initBarSettingsList()
		return
	}

	// Set value
	switch index {
	case 0:
		a.config.Position = value
	case 1:
		a.config.Layer = value
	case 2:
		fmt.Sscanf(value, "%d", &a.config.Height)
	case 3:
		fmt.Sscanf(value, "%d", &a.config.Width)
	case 4:
		fmt.Sscanf(value, "%d", &a.config.Spacing)
	case 5:
		a.config.Margin = value
	case 6:
		a.config.Mode = value
	case 7:
		a.config.Name = value
	}

	a.hasChanges = true
	// Refresh the list to show updated value
	a.initBarSettingsList()
}

func (a *App) updateModuleProperty(index int, value string) {
	if index >= len(a.moduleProperties) {
		return
	}

	prop := a.moduleProperties[index]
	moduleConfig := a.config.GetModuleConfig(a.editingModule)

	// Check if user wants to clear the value
	if value == "─── Clear Value ───" || value == "" {
		// Remove the property from config
		delete(moduleConfig, prop.Name)
		a.config.SetModuleConfig(a.editingModule, moduleConfig)
		a.hasChanges = true
		// Refresh the list to show cleared value
		a.initPropertyList()
		return
	}

	// Parse value based on type
	switch prop.Type {
	case "integer":
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			moduleConfig[prop.Name] = intVal
		} else {
			// If it fails to parse, keep as string
			moduleConfig[prop.Name] = value
		}
	case "number":
		var floatVal float64
		if _, err := fmt.Sscanf(value, "%f", &floatVal); err == nil {
			moduleConfig[prop.Name] = floatVal
		} else {
			moduleConfig[prop.Name] = value
		}
	case "boolean":
		var boolVal bool
		if _, err := fmt.Sscanf(value, "%t", &boolVal); err == nil {
			moduleConfig[prop.Name] = boolVal
		} else {
			moduleConfig[prop.Name] = value
		}
	default:
		moduleConfig[prop.Name] = value
	}

	a.config.SetModuleConfig(a.editingModule, moduleConfig)
	a.hasChanges = true
	// Refresh the list to show updated value
	a.initPropertyList()
}

func (a *App) updateMainView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle Enter to select menu item (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Enter) && a.mainMenuList.FilterState() != list.Filtering {
			selectedItem := a.mainMenuList.SelectedItem()
			if menuItem, ok := selectedItem.(MenuItem); ok {
				// Special handling for Reload Waybar
				if menuItem.TitleText == "Reload Waybar" {
					return a, a.reloadWaybar()
				}

				a.view = menuItem.View
				a.listIndex = 0
				a.fieldIndex = 0

				// Initialize viewport if entering style editor
				if menuItem.View == ViewStyleEditor {
					a.initStyleViewport()
				}
				// Initialize diff if entering diff view
				if menuItem.View == ViewDiff {
					a.initDiffView()
					a.initChangesTable()
				}
				// Initialize bar settings list if entering bar settings
				if menuItem.View == ViewBarSettings {
					a.initBarSettingsList()
				}
				// Initialize modules list if entering modules views
				if menuItem.View == ViewModulesLeft || menuItem.View == ViewModulesCenter || menuItem.View == ViewModulesRight {
					a.initModulesList()
				}
			}
			return a, nil
		}
	}

	// Pass message to list for navigation and filtering
	a.mainMenuList, cmd = a.mainMenuList.Update(msg)
	return a, cmd
}

func (a *App) updateBarSettingsView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle back key separately (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Back) && a.barSettingsList.FilterState() != list.Filtering {
			a.view = ViewMain
			return a, nil
		}

		// Handle Delete to clear property value (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Delete) && a.barSettingsList.FilterState() != list.Filtering {
			selectedItem := a.barSettingsList.SelectedItem()
			if propItem, ok := selectedItem.(PropertyItem); ok {
				prop := propItem.property

				// Clear the bar setting value
				switch prop.Name {
				case "position":
					a.config.Position = ""
				case "layer":
					a.config.Layer = ""
				case "height":
					a.config.Height = 0
				case "width":
					a.config.Width = 0
				case "spacing":
					a.config.Spacing = 0
				case "margin":
					a.config.Margin = ""
				case "mode":
					a.config.Mode = ""
				case "name":
					a.config.Name = ""
				}

				a.hasChanges = true
				a.initBarSettingsList()

				a.notification = fmt.Sprintf("Cleared value for '%s'", prop.Name)
				a.notificationType = "success"
			}
			return a, nil
		}

		// Handle Enter to edit property (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Enter) && a.barSettingsList.FilterState() != list.Filtering {
			selectedItem := a.barSettingsList.SelectedItem()
			if propItem, ok := selectedItem.(PropertyItem); ok {
				prop := propItem.property

				// Get current value
				var currentVal string
				switch prop.Name {
				case "position":
					currentVal = a.config.Position
				case "layer":
					currentVal = a.config.Layer
				case "height":
					currentVal = fmt.Sprintf("%d", a.config.Height)
				case "width":
					currentVal = fmt.Sprintf("%d", a.config.Width)
				case "spacing":
					currentVal = fmt.Sprintf("%d", a.config.Spacing)
				case "margin":
					currentVal = a.config.Margin
				case "mode":
					currentVal = a.config.Mode
				case "name":
					currentVal = a.config.Name
				}

				// Store the index for applying the value
				a.fieldIndex = a.barSettingsList.Index()
				a.previousView = ViewBarSettings

				if len(prop.Options) > 0 {
					a.openSelectModal("Select "+prop.Name, prop.Options, a.fieldIndex, currentVal)
				} else {
					a.openInputModal("Edit "+prop.Name, a.fieldIndex, currentVal)
				}
			}
			return a, nil
		}
	}

	// Pass message to list for navigation and filtering
	a.barSettingsList, cmd = a.barSettingsList.Update(msg)
	return a, cmd
}

func (a *App) updateModulesListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Don't process special keys while filtering
		if a.modulesList.FilterState() != list.Filtering {
			switch {
			case key.Matches(keyMsg, a.keyMap.Back):
				a.moving = false
				a.view = ViewMain
				return a, nil

			case key.Matches(keyMsg, a.keyMap.Move):
				modules := a.getCurrentModulesList()
				if len(modules) > 0 {
					a.moving = !a.moving
					a.notification = "Move mode: " + map[bool]string{true: "ON", false: "OFF"}[a.moving]
					a.notificationType = "info"
				}
				return a, nil

			case key.Matches(keyMsg, a.keyMap.Up):
				if a.moving {
					idx := a.modulesList.Index()
					if idx > 0 {
						a.swapModules(idx, idx-1)
						a.hasChanges = true
						// Refresh list and move selection up
						a.initModulesList()
						a.modulesList.Select(idx - 1)
					}
					return a, nil
				}

			case key.Matches(keyMsg, a.keyMap.Down):
				if a.moving {
					modules := a.getCurrentModulesList()
					idx := a.modulesList.Index()
					if idx < len(modules)-1 {
						a.swapModules(idx, idx+1)
						a.hasChanges = true
						// Refresh list and move selection down
						a.initModulesList()
						a.modulesList.Select(idx + 1)
					}
					return a, nil
				}

			case key.Matches(keyMsg, a.keyMap.Enter):
				selectedItem := a.modulesList.SelectedItem()
				if moduleItem, ok := selectedItem.(ModuleListItem); ok {
					a.editingModule = moduleItem.moduleName
					a.view = ViewModuleEditor
					a.initPropertyList()
				}
				return a, nil

			case key.Matches(keyMsg, a.keyMap.Add):
				a.view = ViewModuleAdd
				a.addModuleCategory = 0
				a.addModuleIndex = 0
				a.previousView = a.view
				return a, nil

			case key.Matches(keyMsg, a.keyMap.Delete):
				idx := a.modulesList.Index()
				modules := a.getCurrentModulesList()
				if len(modules) > 0 && idx < len(modules) {
					moduleName := modules[idx]
					a.deleteModule(idx)
					a.hasChanges = true
					// Refresh list
					a.initModulesList()
					// Adjust selection if needed
					newModules := a.getCurrentModulesList()
					if idx >= len(newModules) && idx > 0 {
						a.modulesList.Select(idx - 1)
					}
					a.notification = fmt.Sprintf("Removed module '%s'", moduleName)
					a.notificationType = "success"
				}
				return a, nil
			}
		}
	}

	// Pass message to list for navigation and filtering
	a.modulesList, cmd = a.modulesList.Update(msg)
	return a, cmd
}

func (a *App) updateModuleEditorView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle back key separately (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Back) && a.propertyList.FilterState() != list.Filtering {
			// No apply needed, changes applied immediately via modal
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
			return a, nil
		}

		// Handle Delete to clear property value (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Delete) && a.propertyList.FilterState() != list.Filtering {
			selectedItem := a.propertyList.SelectedItem()
			if propItem, ok := selectedItem.(PropertyItem); ok {
				prop := propItem.property
				moduleConfig := a.config.GetModuleConfig(a.editingModule)

				// Remove the property from config
				delete(moduleConfig, prop.Name)
				a.config.SetModuleConfig(a.editingModule, moduleConfig)
				a.hasChanges = true

				// Refresh the list to show cleared value
				a.initPropertyList()

				// Show notification
				a.notification = fmt.Sprintf("Cleared value for '%s'", prop.Name)
				a.notificationType = "success"
			}
			return a, nil
		}

		// Handle Enter to edit property (but not while filtering)
		if key.Matches(keyMsg, a.keyMap.Enter) && a.propertyList.FilterState() != list.Filtering {
			selectedItem := a.propertyList.SelectedItem()
			if propItem, ok := selectedItem.(PropertyItem); ok {
				prop := propItem.property
				moduleConfig := a.config.GetModuleConfig(a.editingModule)

				// Get current value
				var currentVal string
				if val, ok := moduleConfig[prop.Name]; ok {
					currentVal = fmt.Sprintf("%v", val)
				}

				// Store the index for applying the value
				a.fieldIndex = a.propertyList.Index()
				a.previousView = ViewModuleEditor

				if len(prop.Options) > 0 {
					a.openSelectModal("Select "+prop.Name, prop.Options, a.fieldIndex, currentVal)
				} else {
					a.openInputModal("Edit "+prop.Name, a.fieldIndex, currentVal)
				}
			}
			return a, nil
		}
	}

	// Pass message to list for navigation and filtering
	a.propertyList, cmd = a.propertyList.Update(msg)
	return a, cmd
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

func (a *App) updateModuleCatalogView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	categories := config.GetModuleCategories()

	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = ViewMain
	case key.Matches(msg, a.keyMap.Left):
		if a.catalogCategory > 0 {
			a.catalogCategory--
			a.catalogIndex = 0
		}
	case key.Matches(msg, a.keyMap.Right):
		if a.catalogCategory < len(categories)-1 {
			a.catalogCategory++
			a.catalogIndex = 0
		}
	case key.Matches(msg, a.keyMap.Up):
		if a.catalogIndex > 0 {
			a.catalogIndex--
		}
	case key.Matches(msg, a.keyMap.Down):
		if a.catalogCategory < len(categories) {
			mods := categories[a.catalogCategory].Modules
			if a.catalogIndex < len(mods)-1 {
				a.catalogIndex++
			}
		}
	case key.Matches(msg, a.keyMap.Enter):
		if a.catalogCategory < len(categories) {
			mods := categories[a.catalogCategory].Modules
			if a.catalogIndex < len(mods) {
				moduleName := mods[a.catalogIndex].Name
				// Handle custom/ and group/ modules
				if strings.HasSuffix(moduleName, "/") {
					moduleName = moduleName + "name"
				}
				
				// Check if active
				isActive := false
				for _, m := range a.config.ModulesLeft { if m == moduleName { isActive = true; break } }
				if !isActive { for _, m := range a.config.ModulesCenter { if m == moduleName { isActive = true; break } } }
				if !isActive { for _, m := range a.config.ModulesRight { if m == moduleName { isActive = true; break } } }

				a.editingModule = moduleName // Set target module for actions
				
				if isActive {
					a.openSelectModal("Module Actions", []string{"Configure", "Remove"}, -1, "")
				} else {
					a.openSelectModal("Add Module To", []string{"Left", "Center", "Right", "Cancel"}, -2, "")
				}
			}
		}
	}
	return a, nil
}

func (a *App) renderModuleCatalogView() string {
	title := styles.TitleStyle.Render("Module Catalog")

	categories := config.GetModuleCategories()

	// Render category tabs
	var tabs []string
	for i, cat := range categories {
		if i == a.catalogCategory {
			tabs = append(tabs, styles.ActiveTabStyle.Render(cat.Name))
		} else {
			tabs = append(tabs, styles.TabStyle.Render(cat.Name))
		}
	}
	tabBar := styles.TabBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))

	// Calculate available height
	listHeight := a.contentHeight - lipgloss.Height(title) - lipgloss.Height(tabBar) - 2
	if listHeight < 5 { listHeight = 5 }
	
	itemsPerView := (listHeight - 4) / 2
	if itemsPerView < 1 { itemsPerView = 1 }

	// Windowing logic for items
	startIdx := 0
	if a.catalogIndex >= itemsPerView {
		startIdx = a.catalogIndex - itemsPerView + 1
	}
	
	var items []string
	if a.catalogCategory < len(categories) {
		mods := categories[a.catalogCategory].Modules
		
		endIdx := startIdx + itemsPerView
		if endIdx > len(mods) {
			endIdx = len(mods)
			if endIdx - startIdx < itemsPerView {
				startIdx = endIdx - itemsPerView
				if startIdx < 0 { startIdx = 0 }
			}
		}

		for i := startIdx; i < endIdx; i++ {
			mod := mods[i]
			moduleName := mod.Name
			if strings.HasSuffix(moduleName, "/") { moduleName = moduleName + "name" }

			// Check status
			status := "[ ]"
			activeColor := styles.TextMuted
			for _, m := range a.config.ModulesLeft { if m == moduleName { status = "[L]"; activeColor = styles.Primary; break } }
			if status == "[ ]" { for _, m := range a.config.ModulesCenter { if m == moduleName { status = "[C]"; activeColor = styles.Primary; break } } }
			if status == "[ ]" { for _, m := range a.config.ModulesRight { if m == moduleName { status = "[R]"; activeColor = styles.Primary; break } } }

			statusTag := lipgloss.NewStyle().Foreground(activeColor).Render(status)
			
			var item string
			if i == a.catalogIndex {
				item = styles.SelectedItemStyle.Render(fmt.Sprintf("%s %s %s", styles.IconSelected, statusTag, mod.Name))
			} else {
				item = styles.ListItemStyle.Render(fmt.Sprintf("  %s %s", statusTag, mod.Name))
			}
			desc := styles.DescriptionStyle.Render("  " + mod.Description)
			items = append(items, lipgloss.JoinVertical(lipgloss.Left, item, desc))
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)

	listBox := styles.BoxStyle.
		Width(a.width - 10).
		Height(listHeight).
		Render(list)

	hint := styles.DescriptionStyle.Render("Use ←→ categories, ↑↓ select. Enter: Toggle/Configure")

	return lipgloss.JoinVertical(lipgloss.Left, title, tabBar, listBox, hint)
}

func (a *App) initStyleViewport() {
	// Calculate available height for viewport
	headerHeight := 3 // title + padding
	footerHeight := 2 // hint
	viewportHeight := a.contentHeight - headerHeight - footerHeight
	if viewportHeight < 4 {
		viewportHeight = 4
	}

	// Create viewport with content
	vp := viewport.New(a.width-10, viewportHeight)

	// Format content with line numbers
	lines := strings.Split(a.styleContent, "\n")
	var formattedLines []string
	for i, line := range lines {
		lineNum := fmt.Sprintf("%3d │ ", i+1)
		formattedLines = append(formattedLines, lineNum+line)
	}

	vp.SetContent(strings.Join(formattedLines, "\n"))
	vp.YPosition = 0

	a.styleViewport = vp
}

func (a *App) initDiffView() {
	// Calculate diff between current state and saved files
	diff, err := a.configManager.GetDiff(a.config, a.styleContent)
	if err != nil {
		a.diffContent = fmt.Sprintf("Error calculating diff: %v", err)
		a.notification = "Failed to calculate diff"
		a.notificationType = "error"
	} else {
		a.diffContent = diff
	}
	a.listIndex = 0 // Reset scroll position
}

func (a *App) initChangesTable() {
	// Load saved configuration for comparison
	savedConfig, err := a.configManager.LoadConfig()
	if err != nil {
		// If can't load saved config, show error
		columns := []table.Column{
			{Title: "Erro", Width: 60},
		}
		rows := []table.Row{
			{fmt.Sprintf("Não foi possível carregar configuração salva: %v", err)},
		}
		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
		)
		a.changesTable = t
		return
	}

	// Define columns
	columns := []table.Column{
		{Title: "Contexto", Width: 25},
		{Title: "Atributo", Width: 20},
		{Title: "Valor Novo", Width: 25},
		{Title: "Valor Antigo", Width: 25},
	}

	// Collect changes
	var rows []table.Row

	// Compare bar settings
	rows = append(rows, a.compareBarSettings(savedConfig)...)

	// Compare module lists
	rows = append(rows, a.compareModuleLists(savedConfig)...)

	// Compare module configurations
	rows = append(rows, a.compareModuleConfigs(savedConfig)...)

	// If no changes, show a message
	if len(rows) == 0 {
		rows = append(rows, table.Row{"", "Nenhuma mudança detectada", "", ""})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(a.contentHeight - 8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	a.changesTable = t
}

// compareBarSettings compares bar-level settings between saved and current config
func (a *App) compareBarSettings(saved *config.WaybarConfig) []table.Row {
	var rows []table.Row

	// Compare position
	if a.config.Position != saved.Position {
		rows = append(rows, table.Row{
			"Bar Settings",
			"position",
			a.config.Position,
			saved.Position,
		})
	}

	// Compare layer
	if a.config.Layer != saved.Layer {
		rows = append(rows, table.Row{
			"Bar Settings",
			"layer",
			a.config.Layer,
			saved.Layer,
		})
	}

	// Compare height
	if a.config.Height != saved.Height {
		rows = append(rows, table.Row{
			"Bar Settings",
			"height",
			fmt.Sprintf("%d", a.config.Height),
			fmt.Sprintf("%d", saved.Height),
		})
	}

	// Compare width
	if a.config.Width != saved.Width {
		rows = append(rows, table.Row{
			"Bar Settings",
			"width",
			fmt.Sprintf("%d", a.config.Width),
			fmt.Sprintf("%d", saved.Width),
		})
	}

	// Compare spacing
	if a.config.Spacing != saved.Spacing {
		rows = append(rows, table.Row{
			"Bar Settings",
			"spacing",
			fmt.Sprintf("%d", a.config.Spacing),
			fmt.Sprintf("%d", saved.Spacing),
		})
	}

	// Compare margin
	if a.config.Margin != saved.Margin {
		rows = append(rows, table.Row{
			"Bar Settings",
			"margin",
			a.config.Margin,
			saved.Margin,
		})
	}

	// Compare mode
	if a.config.Mode != saved.Mode {
		rows = append(rows, table.Row{
			"Bar Settings",
			"mode",
			a.config.Mode,
			saved.Mode,
		})
	}

	// Compare name
	if a.config.Name != saved.Name {
		rows = append(rows, table.Row{
			"Bar Settings",
			"name",
			a.config.Name,
			saved.Name,
		})
	}

	return rows
}

// compareModuleLists compares module lists between saved and current config
func (a *App) compareModuleLists(saved *config.WaybarConfig) []table.Row {
	var rows []table.Row

	// Compare modules-left
	if !slicesEqual(a.config.ModulesLeft, saved.ModulesLeft) {
		rows = append(rows, table.Row{
			"Modules",
			"modules-left",
			fmt.Sprintf("%v", a.config.ModulesLeft),
			fmt.Sprintf("%v", saved.ModulesLeft),
		})
	}

	// Compare modules-center
	if !slicesEqual(a.config.ModulesCenter, saved.ModulesCenter) {
		rows = append(rows, table.Row{
			"Modules",
			"modules-center",
			fmt.Sprintf("%v", a.config.ModulesCenter),
			fmt.Sprintf("%v", saved.ModulesCenter),
		})
	}

	// Compare modules-right
	if !slicesEqual(a.config.ModulesRight, saved.ModulesRight) {
		rows = append(rows, table.Row{
			"Modules",
			"modules-right",
			fmt.Sprintf("%v", a.config.ModulesRight),
			fmt.Sprintf("%v", saved.ModulesRight),
		})
	}

	return rows
}

// compareModuleConfigs compares individual module configurations
func (a *App) compareModuleConfigs(saved *config.WaybarConfig) []table.Row {
	var rows []table.Row

	// Get all modules from both configs
	allModules := make(map[string]bool)
	for moduleName := range a.config.Modules {
		allModules[moduleName] = true
	}
	for moduleName := range saved.Modules {
		allModules[moduleName] = true
	}

	// Compare each module
	for moduleName := range allModules {
		currentConfig := a.config.GetModuleConfig(moduleName)
		savedConfig := saved.GetModuleConfig(moduleName)

		// Get all properties from both configs
		allProps := make(map[string]bool)
		for prop := range currentConfig {
			allProps[prop] = true
		}
		for prop := range savedConfig {
			allProps[prop] = true
		}

		// Compare each property
		for prop := range allProps {
			currentVal := currentConfig[prop]
			savedVal := savedConfig[prop]

			// Compare values
			if !valuesEqual(currentVal, savedVal) {
				rows = append(rows, table.Row{
					moduleName,
					prop,
					formatValue(currentVal),
					formatValue(savedVal),
				})
			}
		}
	}

	return rows
}

// slicesEqual compares two string slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// valuesEqual compares two interface{} values
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// formatValue formats an interface{} value for display
func formatValue(v interface{}) string {
	if v == nil {
		return "(empty)"
	}
	return fmt.Sprintf("%v", v)
}

func (a *App) updateStyleEditorView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = ViewMain
		return a, nil
	case key.Matches(msg, a.keyMap.Up):
		a.styleViewport, cmd = a.styleViewport.Update(msg)
		return a, cmd
	case key.Matches(msg, a.keyMap.Down):
		a.styleViewport, cmd = a.styleViewport.Update(msg)
		return a, cmd
	}

	// Pass other keys to viewport for navigation
	a.styleViewport, cmd = a.styleViewport.Update(msg)
	return a, cmd
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

func (a *App) updateDiffView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = ViewMain
		return a, nil
	}

	// Pass other keys to table for navigation
	a.changesTable, cmd = a.changesTable.Update(msg)
	return a, cmd
}

func (a *App) updateRestoreConfirmView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keyMap.Enter):
		// Restore
		return a, a.restoreStartupBackup()
	case key.Matches(msg, a.keyMap.Back):
		// Cancel
		a.view = ViewMain
		a.notification = "Restore cancelled. Waybar might be in a broken state."
		a.notificationType = "warning"
	}
	// Also accept 'y' and 'n' specifically
	switch msg.String() {
	case "y", "Y":
		return a, a.restoreStartupBackup()
	case "n", "N":
		a.view = ViewMain
		a.notification = "Restore cancelled. Waybar might be in a broken state."
		a.notificationType = "warning"
	}
	return a, nil
}

func (a *App) restoreStartupBackup() tea.Cmd {
	return func() tea.Msg {
		if a.startupConfigBackup == "" || a.startupStyleBackup == "" {
			a.notification = "No startup backup found!"
			a.notificationType = "error"
			a.view = ViewMain
			return nil
		}

		// Use RestoreBackup from configManager? 
		// But RestoreBackup expects a filename in BackupDir.
		// CreateNamedBackup returns full paths.
		// However, ConfigManager logic is filename relative to BackupDir.
		
		configName := filepath.Base(a.startupConfigBackup)
		styleName := filepath.Base(a.startupStyleBackup)
		
		// Restore config
		if err := a.configManager.RestoreBackup(configName); err != nil {
			a.notification = fmt.Sprintf("Failed to restore config: %v", err)
			a.notificationType = "error"
			a.view = ViewMain
			return nil
		}
		
		// Restore style
		if err := a.configManager.RestoreBackup(styleName); err != nil {
			a.notification = fmt.Sprintf("Failed to restore style: %v", err)
			a.notificationType = "error"
			a.view = ViewMain
			return nil
		}
		
		// Reload internal state
		a.loadConfig()
		
		// Try to reload Waybar again
		exec.Command("pkill", "-SIGUSR2", "waybar").Run()
		
		a.notification = "Configuration restored to startup state and Waybar reloaded."
		a.notificationType = "success"
		a.view = ViewMain
		return nil
	}
}

func (a *App) renderRestoreConfirmView() string {
	title := styles.TitleStyle.Render("Waybar Error Detected")
	
	msg := styles.NotifyErrorStyle.Render("Waybar failed to reload or crashed!")
	desc := styles.DescriptionStyle.Render("Do you want to restore the configuration to how it was when you started the application?")
	
	confirm := styles.ButtonStyle.Render("Press [Enter] or [y] to Restore")
	cancel := styles.ButtonDangerStyle.Render("Press [Esc] or [n] to Cancel")
	
	content := lipgloss.JoinVertical(lipgloss.Center, msg, "", desc, "", lipgloss.JoinHorizontal(lipgloss.Center, confirm, cancel))
	
	box := styles.BoxStyle.
		Width(a.width - 10).
		Align(lipgloss.Center).
		Padding(2).
		Render(content)
		
	return lipgloss.JoinVertical(lipgloss.Left, title, box)
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
		{TitleText: "Bar Settings", DescriptionText: "Configure bar position, size, and behavior", Icon: styles.IconGear, View: ViewBarSettings},
		{TitleText: "Module Catalog", DescriptionText: "Browse, enable and configure native modules", Icon: styles.IconPlugin, View: ViewModuleCatalog},
		{TitleText: "Modules Left", DescriptionText: "Manage left-aligned modules", Icon: styles.IconArrowLeft, View: ViewModulesLeft},
		{TitleText: "Modules Center", DescriptionText: "Manage center-aligned modules", Icon: styles.IconDot, View: ViewModulesCenter},
		{TitleText: "Modules Right", DescriptionText: "Manage right-aligned modules", Icon: styles.IconArrowRight, View: ViewModulesRight},
		{TitleText: "Style Editor", DescriptionText: "Edit CSS styling", Icon: styles.IconPalette, View: ViewStyleEditor},
		{TitleText: "Review Changes", DescriptionText: "See pending changes vs saved config", Icon: styles.IconInfo, View: ViewDiff},
		{TitleText: "Backups", DescriptionText: "Manage configuration backups", Icon: styles.IconFolder, View: ViewBackups},
		{TitleText: "Reload Waybar", DescriptionText: "Force Waybar to reload config (verify & restore if failed)", Icon: styles.IconRefresh, View: ViewMain}, // ViewMain as placeholder, handled specially
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

	// Initialize module config if it doesn't exist
	// This ensures the module will be saved with its configuration
	if _, exists := a.config.Modules[name]; !exists {
		a.config.Modules[name] = make(map[string]interface{})
	}
}

func (a *App) initPropertyList() {
	moduleDef := config.GetModuleDefinition(a.editingModule)
	moduleConfig := a.config.GetModuleConfig(a.editingModule)

	// Load saved config for comparison
	savedConfig, err := a.configManager.LoadConfig()
	var savedModuleConfig map[string]interface{}
	if err == nil {
		savedModuleConfig = savedConfig.GetModuleConfig(a.editingModule)
	}

	// Clear notification
	a.notification = ""

	// Combine common and module-specific properties with deduplication
	propMap := make(map[string]config.PropertyDefinition)

	// Add module-specific properties first (they take priority)
	if moduleDef != nil {
		for _, prop := range moduleDef.Properties {
			propMap[prop.Name] = prop
		}
	}

	// Add common properties (skip if already exists)
	for _, prop := range config.CommonProperties() {
		if _, exists := propMap[prop.Name]; !exists {
			propMap[prop.Name] = prop
		}
	}

	// Convert map to slice
	var allProps []config.PropertyDefinition
	for _, prop := range propMap {
		allProps = append(allProps, prop)
	}

	// Sort properties: prioritized by having a value, then alphabetical
	sort.Slice(allProps, func(i, j int) bool {
		hasValI := false
		if _, ok := moduleConfig[allProps[i].Name]; ok {
			hasValI = true
		}
		hasValJ := false
		if _, ok := moduleConfig[allProps[j].Name]; ok {
			hasValJ = true
		}

		if hasValI != hasValJ {
			return hasValI
		}
		return allProps[i].Name < allProps[j].Name
	})

	a.moduleProperties = make([]config.PropertyDefinition, len(allProps))
	copy(a.moduleProperties, allProps)
	a.fieldIndex = 0

	// Create list items
	items := make([]list.Item, len(allProps))
	for i, prop := range allProps {
		var value, savedValue string
		if val, ok := moduleConfig[prop.Name]; ok {
			value = fmt.Sprintf("%v", val)
		}
		if err == nil {
			if val, ok := savedModuleConfig[prop.Name]; ok {
				savedValue = fmt.Sprintf("%v", val)
			}
		}

		items[i] = PropertyItem{
			property:   prop,
			value:      value,
			isModified: err == nil && value != savedValue,
		}
	}

	// Initialize list with custom delegate for highlighting modified items
	delegate := newModifiedItemDelegate()

	l := list.New(items, delegate, 0, 0)
	l.Title = "Edit Module: " + a.editingModule
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We'll use custom help
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	a.propertyList = l
}

func (a *App) initBarSettingsList() {
	// Define bar settings properties
	barSettings := []config.PropertyDefinition{
		{Name: "position", Description: "Bar position on screen", Options: []string{"top", "bottom", "left", "right"}},
		{Name: "layer", Description: "Window layer", Options: []string{"top", "bottom", "overlay", "background"}},
		{Name: "height", Description: "Bar height in pixels"},
		{Name: "width", Description: "Bar width in pixels"},
		{Name: "spacing", Description: "Spacing between modules"},
		{Name: "margin", Description: "Margin around bar"},
		{Name: "mode", Description: "Display mode", Options: []string{"dock", "hide", "invisible", "overlay"}},
		{Name: "name", Description: "Bar name/identifier"},
	}

	// Load saved config for comparison
	savedConfig, err := a.configManager.LoadConfig()

	// Create list items with current values
	items := make([]list.Item, len(barSettings))
	for i, prop := range barSettings {
		var value, savedValue string
		switch prop.Name {
		case "position":
			value = a.config.Position
			if err == nil {
				savedValue = savedConfig.Position
			}
		case "layer":
			value = a.config.Layer
			if err == nil {
				savedValue = savedConfig.Layer
			}
		case "height":
			value = fmt.Sprintf("%d", a.config.Height)
			if err == nil {
				savedValue = fmt.Sprintf("%d", savedConfig.Height)
			}
		case "width":
			value = fmt.Sprintf("%d", a.config.Width)
			if err == nil {
				savedValue = fmt.Sprintf("%d", savedConfig.Width)
			}
		case "spacing":
			value = fmt.Sprintf("%d", a.config.Spacing)
			if err == nil {
				savedValue = fmt.Sprintf("%d", savedConfig.Spacing)
			}
		case "margin":
			value = a.config.Margin
			if err == nil {
				savedValue = savedConfig.Margin
			}
		case "mode":
			value = a.config.Mode
			if err == nil {
				savedValue = savedConfig.Mode
			}
		case "name":
			value = a.config.Name
			if err == nil {
				savedValue = savedConfig.Name
			}
		}

		items[i] = PropertyItem{
			property:   prop,
			value:      value,
			isModified: err == nil && value != savedValue,
		}
	}

	// Initialize list with custom delegate for highlighting modified items
	delegate := newModifiedItemDelegate()

	l := list.New(items, delegate, 0, 0)
	l.Title = "Bar Settings"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	a.barSettingsList = l
}

func (a *App) initMainMenuList() {
	menuItems := a.getMainMenuItems()

	// Convert to list items
	items := make([]list.Item, len(menuItems))
	for i, item := range menuItems {
		items[i] = item
	}

	// Initialize list with filtering enabled
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(0)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Waybar Configuration Editor"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	a.mainMenuList = l
}

func (a *App) initModulesList() {
	modules := a.getCurrentModulesList()

	// Convert to list items
	items := make([]list.Item, len(modules))
	for i, moduleName := range modules {
		items[i] = ModuleListItem{moduleName: moduleName}
	}

	// Initialize list with filtering enabled
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(0)

	l := list.New(items, delegate, 0, 0)

	// Set title based on current view
	var title string
	switch a.view {
	case ViewModulesLeft:
		title = "Modules Left"
	case ViewModulesCenter:
		title = "Modules Center"
	case ViewModulesRight:
		title = "Modules Right"
	}

	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	a.modulesList = l
}

// applyBarSettings and applyModuleSettings are removed as updates happen via modals

// View renders the application

// View renders the application
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	// Calculate available content height
	headerHeight := lipgloss.Height(a.renderHeader())
	footerHeight := lipgloss.Height(a.renderFooter())
	// 2 is for top/bottom padding of contentArea defined in renderLayout
	a.contentHeight = a.height - headerHeight - footerHeight - 2
	if a.contentHeight < 0 {
		a.contentHeight = 0
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
	case ViewDiff:
		content = a.renderDiffView()
	case ViewModuleCatalog:
		content = a.renderModuleCatalogView()
	case ViewRestoreConfirm:
			content = a.renderRestoreConfirmView()
		case ViewModal:
			content = a.renderModalView()
		case ViewHelp:
			content = a.renderHelpView()
		}
		
		// If modal, we don't use standard layout because we want full screen centering
		// But standard layout provides header/footer. 
		// renderModalView uses Place which fills screen.
		if a.view == ViewModal {
			return content
		}
	
		return a.renderLayout(content)}

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
	// Status indicator - simple colored character
	var statusChar string
	if a.hasChanges {
		statusChar = lipgloss.NewStyle().Foreground(styles.Warning).Render("*")
	} else {
		statusChar = lipgloss.NewStyle().Foreground(styles.Success).Render("~")
	}

	// Path info (truncate if needed)
	configPath := a.configPath
	maxPathLen := a.width - 20
	if maxPathLen < 15 {
		maxPathLen = 15
	}
	if len(configPath) > maxPathLen {
		configPath = "..." + configPath[len(configPath)-maxPathLen+3:]
	}

	// Build header as simple string
	header := fmt.Sprintf("WABAGO %s %s", statusChar, configPath)

	return styles.HeaderStyle.Render(header)
}

// getStatusSummary returns context-specific status for the current view
func (a *App) getStatusSummary() string {
	var info string

	switch a.view {
	case ViewMain:
		// Main menu: show general config info
		info = fmt.Sprintf("%d modules loaded | config: %s", len(a.config.Modules), filepath.Base(a.configPath))

	case ViewBarSettings:
		// Bar settings: show current bar configuration
		pos := a.config.Position
		if pos == "" {
			pos = "top"
		}
		info = fmt.Sprintf("position: %s | height: %d | spacing: %d", pos, a.config.Height, a.config.Spacing)

	case ViewModulesLeft:
		// Left modules: show count and selected
		count := len(a.config.ModulesLeft)
		if count > 0 && a.listIndex < count {
			info = fmt.Sprintf("%d modules | selected: %s", count, a.config.ModulesLeft[a.listIndex])
		} else {
			info = fmt.Sprintf("%d modules", count)
		}

	case ViewModulesCenter:
		// Center modules: show count and selected
		count := len(a.config.ModulesCenter)
		if count > 0 && a.listIndex < count {
			info = fmt.Sprintf("%d modules | selected: %s", count, a.config.ModulesCenter[a.listIndex])
		} else {
			info = fmt.Sprintf("%d modules", count)
		}

	case ViewModulesRight:
		// Right modules: show count and selected
		count := len(a.config.ModulesRight)
		if count > 0 && a.listIndex < count {
			info = fmt.Sprintf("%d modules | selected: %s", count, a.config.ModulesRight[a.listIndex])
		} else {
			info = fmt.Sprintf("%d modules", count)
		}

	case ViewModuleEditor:
		// Module editor: show module name and configured properties count
		moduleConfig := a.config.GetModuleConfig(a.editingModule)
		info = fmt.Sprintf("editing: %s | %d properties configured", a.editingModule, len(moduleConfig))

	case ViewModuleAdd:
		// Add module: show selected category
		categories := config.GetModuleCategories()
		if a.addModuleCategory < len(categories) {
			info = fmt.Sprintf("category: %s | %d modules available", categories[a.addModuleCategory].Name, len(categories[a.addModuleCategory].Modules))
		}

	case ViewStyleEditor:
		// Style editor: show file info
		lines := len(strings.Split(a.styleContent, "\n"))
		info = fmt.Sprintf("style.css | %d lines", lines)

	case ViewBackups:
		// Backups: show count
		backups, _ := a.configManager.GetBackups()
		info = fmt.Sprintf("%d backups available", len(backups))

	case ViewHelp:
		info = "keyboard shortcuts reference"

	default:
		info = fmt.Sprintf("%d modules | L:%d C:%d R:%d", len(a.config.Modules), len(a.config.ModulesLeft), len(a.config.ModulesCenter), len(a.config.ModulesRight))
	}

	// Add unsaved changes indicator if there are changes
	if a.hasChanges {
		info += " " + styles.NotifyWarningStyle.Render("● Unsaved")
	}

	return styles.StatusBarStyle.Render(info)
}

func (a *App) renderFooter() string {
	// Status line - always shows summary info
	var statusLine string
	if a.notification != "" {
		// Show notification if present
		switch a.notificationType {
		case "success":
			statusLine = styles.NotifySuccessStyle.Render(a.notification)
		case "warning":
			statusLine = styles.NotifyWarningStyle.Render(a.notification)
		case "error":
			statusLine = styles.NotifyErrorStyle.Render(a.notification)
		default:
			statusLine = styles.NotifyInfoStyle.Render(a.notification)
		}
	} else {
		// Show default status summary
		statusLine = a.getStatusSummary()
	}

	// Help line - keyboard shortcuts
	helpKeys := []string{
		styles.HelpKeyStyle.Render("?") + styles.HelpDescStyle.Render(" help"),
		styles.HelpKeyStyle.Render("↑↓") + styles.HelpDescStyle.Render(" navigate"),
		styles.HelpKeyStyle.Render("enter") + styles.HelpDescStyle.Render(" select"),
		styles.HelpKeyStyle.Render("esc") + styles.HelpDescStyle.Render(" back"),
		styles.HelpKeyStyle.Render("ctrl+s") + styles.HelpDescStyle.Render(" save"),
		styles.HelpKeyStyle.Render("r") + styles.HelpDescStyle.Render(" reload"),
		styles.HelpKeyStyle.Render("q") + styles.HelpDescStyle.Render(" quit"),
	}
	helpLine := styles.StatusBarStyle.Render(strings.Join(helpKeys, "  "))

	return lipgloss.JoinVertical(lipgloss.Left, statusLine, helpLine)
}

func (a *App) renderMainView() string {
	// Set list dimensions
	listHeight := a.contentHeight - 4
	if listHeight < 10 {
		listHeight = 10
	}
	a.mainMenuList.SetSize(a.width-10, listHeight)

	// Render list
	listView := a.mainMenuList.View()

	hint := styles.DescriptionStyle.Render("↑↓ navigate | Enter select | / filter | q quit")

	return lipgloss.JoinVertical(lipgloss.Left, listView, "", hint)
}

func (a *App) renderBarSettingsView() string {
	// Set list dimensions
	listHeight := a.contentHeight - 4
	if listHeight < 10 {
		listHeight = 10
	}
	a.barSettingsList.SetSize(a.width-10, listHeight)

	// Render list
	listView := a.barSettingsList.View()

	hint := styles.DescriptionStyle.Render("↑↓ navigate | Enter edit | / filter | Delete clear | Esc back")

	return lipgloss.JoinVertical(lipgloss.Left, listView, "", hint)
}

func (a *App) renderModulesListView() string {
	// Set list dimensions
	listHeight := a.contentHeight - 4
	if listHeight < 10 {
		listHeight = 10
	}
	a.modulesList.SetSize(a.width-10, listHeight)

	// Render list
	listView := a.modulesList.View()

	// Add hint with move mode status
	var modeHint string
	if a.moving {
		modeHint = styles.TagActiveStyle.Render("MOVE MODE") + " ↑↓ reorder | m exit move"
	} else {
		modeHint = "↑↓ navigate | Enter edit | / filter | a add | d delete | m move | Esc back"
	}
	hint := styles.DescriptionStyle.Render(modeHint)

	return lipgloss.JoinVertical(lipgloss.Left, listView, "", hint)
}

func (a *App) renderModuleEditorView() string {
	// Set list dimensions
	listHeight := a.contentHeight - 4
	if listHeight < 10 {
		listHeight = 10
	}
	a.propertyList.SetSize(a.width-4, listHeight)

	// Render the list
	listView := a.propertyList.View()

	// Add hint
	hint := styles.DescriptionStyle.Render("↑↓ navigate | Enter edit | / filter | Delete clear | Esc back")

	return lipgloss.JoinVertical(lipgloss.Left, listView, "", hint)
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

	// Calculate available height
	// Title: 2 lines
	// Tabs: 2 lines (approx)
	// Hint: 1 line
	// Box overhead: 4 lines
	// Total overhead: ~9 lines
	listHeight := a.contentHeight - lipgloss.Height(title) - lipgloss.Height(tabBar) - 2
	if listHeight < 5 {
		listHeight = 5
	}
	
	itemsPerView := (listHeight - 4) / 2
	if itemsPerView < 1 {
		itemsPerView = 1
	}

	// Windowing logic for items
	startIdx := 0
	if a.addModuleIndex >= itemsPerView {
		startIdx = a.addModuleIndex - itemsPerView + 1
	}
	
	var items []string
	if a.addModuleCategory < len(categories) {
		mods := categories[a.addModuleCategory].Modules
		
		endIdx := startIdx + itemsPerView
		if endIdx > len(mods) {
			endIdx = len(mods)
			if endIdx - startIdx < itemsPerView {
				startIdx = endIdx - itemsPerView
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		for i := startIdx; i < endIdx; i++ {
			mod := mods[i]
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
		Height(listHeight).
		Render(list)

	hint := styles.DescriptionStyle.Render("Use ←→ to switch categories, ↑↓ to select module, Enter to add")

	return lipgloss.JoinVertical(lipgloss.Left, title, tabBar, listBox, hint)
}

func (a *App) renderStyleEditorView() string {
	title := styles.TitleStyle.Render("Style Editor (CSS)")

	// Render viewport
	viewportContent := a.styleViewport.View()

	// Info line showing position
	totalLines := strings.Count(a.styleContent, "\n") + 1
	scrollInfo := fmt.Sprintf("Lines: %d | Scroll: %d%%", totalLines, int(a.styleViewport.ScrollPercent()*100))
	hint := styles.DescriptionStyle.Render("Style: " + a.stylePath + " | " + scrollInfo + " | ↑↓/PgUp/PgDn: scroll | Esc: back")

	return lipgloss.JoinVertical(lipgloss.Left, title, viewportContent, hint)
}

func (a *App) renderBackupsView() string {
	title := styles.TitleStyle.Render("Configuration Backups")
	
	// Calculate available height
	listHeight := a.contentHeight - lipgloss.Height(title) - 2
	if listHeight < 4 {
		listHeight = 4
	}
	
	itemsPerView := listHeight - 4
	if itemsPerView < 1 {
		itemsPerView = 1
	}

	backups, _ := a.configManager.GetBackups()

	if len(backups) == 0 {
		empty := styles.EmptyStateStyle.Render("No backups available\nPress 'b' to create a backup")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty)
	}
	
	// Windowing
	startIdx := 0
	if a.listIndex >= itemsPerView {
		startIdx = a.listIndex - itemsPerView + 1
	}
	endIdx := startIdx + itemsPerView
	if endIdx > len(backups) {
		endIdx = len(backups)
		if endIdx-startIdx < itemsPerView {
			startIdx = endIdx - itemsPerView
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	var items []string
	for i := startIdx; i < endIdx; i++ {
		backup := backups[i]
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
		Height(listHeight).
		Render(list)

	hint := styles.DescriptionStyle.Render("Press Enter to restore selected backup, b to create new backup")

	return lipgloss.JoinVertical(lipgloss.Left, title, listBox, hint)
}

func (a *App) renderDiffView() string {
	title := styles.TitleStyle.Render("Configuration Changes")

	// Render table
	tableView := a.changesTable.View()

	hint := styles.DescriptionStyle.Render("Press Esc to go back | ↑↓ to navigate | / to filter by Atributo")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", tableView, "", hint)
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
