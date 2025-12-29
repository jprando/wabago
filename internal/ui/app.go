// ui implements the Bubble Tea Model interface methods (Init, Update, View) and rendering logic for the application.
package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jprando/wabago/internal/config"
	"github.com/jprando/wabago/internal/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
	a.modal = ModalState{
		isActive:    true,
		mode:        ModalTypeSelect,
		title:       title,
		options:     options,
		index:       0,
		targetIndex: targetIndex,
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
	switch {
	case key.Matches(msg, a.keyMap.Back):
		// Cancel
		a.view = a.previousView
		a.modal.isActive = false
		return a, nil
	}

	if a.modal.mode == ModalTypeSelect {
		switch {
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
		// Input Mode
		switch {
		case key.Matches(msg, a.keyMap.Enter):
			// Confirm
			a.applyModalValue(a.modal.input.Value())
			a.view = a.previousView
			a.modal.isActive = false
			return a, nil
		default:
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
				a.initModuleEditorInputs()
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
			switch value {
			case "Left":
				a.config.ModulesLeft = append(a.config.ModulesLeft, a.editingModule)
				a.hasChanges = true
			case "Center":
				a.config.ModulesCenter = append(a.config.ModulesCenter, a.editingModule)
				a.hasChanges = true
			case "Right":
				a.config.ModulesRight = append(a.config.ModulesRight, a.editingModule)
				a.hasChanges = true
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
		cancel := styles.ButtonDangerStyle.Render("[Esc] Cancel")
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, confirm, cancel)
		
		content = lipgloss.JoinVertical(lipgloss.Center, original, "", input, "", buttons)
	}
	
	// Calculate height to center
	fullContent := lipgloss.JoinVertical(lipgloss.Center, title, content)
	modalBox := styles.ModalStyle.Render(fullContent)
	
	// Center vertically and horizontally
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, modalBox)
}

func (a *App) updateBarSetting(index int, value string) {
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
}

func (a *App) updateModuleProperty(index int, value string) {
	if index >= len(a.moduleProperties) {
		return
	}
	
	prop := a.moduleProperties[index]
	moduleConfig := a.config.GetModuleConfig(a.editingModule)
	
	// Parse value based on type
	switch prop.Type {
	case "integer":
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			moduleConfig[prop.Name] = intVal
		} else {
			// Try to handle empty or invalid as removal or string?
			// For now, if it fails to parse but is not empty, keep as string or ignore
			if value != "" {
				moduleConfig[prop.Name] = value 
			}
		}
	case "number":
		var floatVal float64
		if _, err := fmt.Sscanf(value, "%f", &floatVal); err == nil {
			moduleConfig[prop.Name] = floatVal
		} else {
			if value != "" {
				moduleConfig[prop.Name] = value
			}
		}
	case "boolean":
		var boolVal bool
		if _, err := fmt.Sscanf(value, "%t", &boolVal); err == nil {
			moduleConfig[prop.Name] = boolVal
		} else {
			if value != "" {
				moduleConfig[prop.Name] = value
			}
		}
	default:
		moduleConfig[prop.Name] = value
	}
	
	a.config.SetModuleConfig(a.editingModule, moduleConfig)
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
			item := menuItems[a.menuIndex]
			if item.Title == "Reload Waybar" {
				return a, a.reloadWaybar()
			}
			a.view = item.View
			a.listIndex = 0
			a.fieldIndex = 0
			// Inputs are no longer initialized here for bar settings
		}
	}
	return a, nil
}

func (a *App) updateBarSettingsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keyMap.Back):
		a.view = ViewMain
	case key.Matches(msg, a.keyMap.Up):
		if a.fieldIndex > 0 {
			a.fieldIndex--
		}
	case key.Matches(msg, a.keyMap.Down), key.Matches(msg, a.keyMap.Tab):
		if a.fieldIndex < 7 { // 8 fields, 0-7
			a.fieldIndex++
		}
	case key.Matches(msg, a.keyMap.Enter):
		a.previousView = ViewBarSettings
		
		// Get current value
		var currentVal string
		switch a.fieldIndex {
		case 0: currentVal = a.config.Position
		case 1: currentVal = a.config.Layer
		case 2: currentVal = fmt.Sprintf("%d", a.config.Height)
		case 3: currentVal = fmt.Sprintf("%d", a.config.Width)
		case 4: currentVal = fmt.Sprintf("%d", a.config.Spacing)
		case 5: currentVal = a.config.Margin
		case 6: currentVal = a.config.Mode
		case 7: currentVal = a.config.Name
		}
		
		switch a.fieldIndex {
		case 0: // Position
			a.openSelectModal("Select Position", []string{"top", "bottom", "left", "right"}, 0, currentVal)
		case 1: // Layer
			a.openSelectModal("Select Layer", []string{"top", "bottom", "overlay", "background"}, 1, currentVal)
		case 6: // Mode
			a.openSelectModal("Select Mode", []string{"dock", "hide", "invisible", "overlay"}, 6, currentVal)
		default:
			// Input modal for others
			fieldNames := []string{"Position", "Layer", "Height", "Width", "Spacing", "Margin", "Mode", "Name"}
			a.openInputModal("Edit "+fieldNames[a.fieldIndex], a.fieldIndex, currentVal)
		}
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
	case key.Matches(msg, a.keyMap.Up):
		if a.fieldIndex > 0 {
			a.fieldIndex--
		}
	case key.Matches(msg, a.keyMap.Down), key.Matches(msg, a.keyMap.Tab):
		if a.fieldIndex < len(a.moduleProperties)-1 {
			a.fieldIndex++
		}
	case key.Matches(msg, a.keyMap.Enter):
		if a.fieldIndex < len(a.moduleProperties) {
			prop := a.moduleProperties[a.fieldIndex]
			moduleConfig := a.config.GetModuleConfig(a.editingModule)
			
			// Get current value
			var currentVal string
			if val, ok := moduleConfig[prop.Name]; ok {
				currentVal = fmt.Sprintf("%v", val)
				// Clean up quotes if needed? fmt.Sprintf("%v") usually does fine for simple types.
			}
			
			a.previousView = ViewModuleEditor
			if len(prop.Options) > 0 {
				a.openSelectModal("Select "+prop.Name, prop.Options, a.fieldIndex, currentVal)
			} else {
				a.openInputModal("Edit "+prop.Name, a.fieldIndex, currentVal)
			}
		}
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
		{Title: "Bar Settings", Description: "Configure bar position, size, and behavior", Icon: styles.IconGear, View: ViewBarSettings},
		{Title: "Module Catalog", Description: "Browse, enable and configure native modules", Icon: styles.IconPlugin, View: ViewModuleCatalog},
		{Title: "Modules Left", Description: "Manage left-aligned modules", Icon: styles.IconArrowLeft, View: ViewModulesLeft},
		{Title: "Modules Center", Description: "Manage center-aligned modules", Icon: styles.IconDot, View: ViewModulesCenter},
		{Title: "Modules Right", Description: "Manage right-aligned modules", Icon: styles.IconArrowRight, View: ViewModulesRight},
		{Title: "Style Editor", Description: "Edit CSS styling", Icon: styles.IconPalette, View: ViewStyleEditor},
		{Title: "Review Changes", Description: "See pending changes vs saved config", Icon: styles.IconInfo, View: ViewDiff},
		{Title: "Backups", Description: "Manage configuration backups", Icon: styles.IconFolder, View: ViewBackups},
		{Title: "Reload Waybar", Description: "Force Waybar to reload config (verify & restore if failed)", Icon: styles.IconRefresh, View: ViewMain}, // ViewMain as placeholder, handled specially
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

func (a *App) initModuleEditorInputs() {
	moduleDef := config.GetModuleDefinition(a.editingModule)
	moduleConfig := a.config.GetModuleConfig(a.editingModule)

	// Clear notification - status bar will show module info
	a.notification = ""

	// Combine common and module-specific properties
	// Put module-specific properties first for better UX
	var allProps []config.PropertyDefinition
	if moduleDef != nil {
		allProps = append(allProps, moduleDef.Properties...)
	}
	allProps = append(allProps, config.CommonProperties()...)

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
			return hasValI // true comes first
		}
		return allProps[i].Name < allProps[j].Name
	})

	a.moduleProperties = make([]config.PropertyDefinition, len(allProps))
	copy(a.moduleProperties, allProps)
	a.fieldIndex = 0
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
		
		var value string
		switch i {
		case 0: value = a.config.Position
		case 1: value = a.config.Layer
		case 2: value = fmt.Sprintf("%d", a.config.Height)
		case 3: value = fmt.Sprintf("%d", a.config.Width)
		case 4: value = fmt.Sprintf("%d", a.config.Spacing)
		case 5: value = a.config.Margin
		case 6: value = a.config.Mode
		case 7: value = a.config.Name
		}
		
		var input string
		if i == a.fieldIndex {
			input = styles.FocusedInputStyle.Render(value)
			// Add cursor/edit hint?
			input = lipgloss.JoinHorizontal(lipgloss.Left, input, " ", styles.ActiveItemStyle.Render(styles.IconEdit))
		} else {
			input = styles.InputStyle.Render(value)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Center, label, input)
		formItems = append(formItems, row)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)

	formBox := styles.BoxStyle.
		Width(a.width - 10).
		Render(form)

	hint := styles.DescriptionStyle.Render("Use ↑↓ to navigate, Enter to edit, Esc to go back")

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
	
	// Calculate available height for the list
	// Title: ~2 lines
	// Hint: ~1 line
	// Box padding/border: ~2 lines (vertical padding) + 2 lines (border) = 4
	// Total overhead: ~7 lines
	listHeight := a.contentHeight - lipgloss.Height(title) - 2 // 1 for hint, 1 buffer
	if listHeight < 4 {
		listHeight = 4
	}

	modules := a.getCurrentModulesList()

	if len(modules) == 0 {
		empty := styles.EmptyStateStyle.Render("No modules configured\nPress 'a' to add a module")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty)
	}

	// Calculate max visible items
	// Assume each item is approx 2 lines (name + description)
	itemsPerView := (listHeight - 4) / 2
	if itemsPerView < 1 {
		itemsPerView = 1
	}

	// Windowing logic
	startIdx := 0
	if a.listIndex >= itemsPerView {
		startIdx = a.listIndex - itemsPerView + 1
	}
	endIdx := startIdx + itemsPerView
	if endIdx > len(modules) {
		endIdx = len(modules)
		// Adjust start if we have space
		if endIdx-startIdx < itemsPerView {
			startIdx = endIdx - itemsPerView
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	var items []string
	for i := startIdx; i < endIdx; i++ {
		mod := modules[i]
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
		Height(listHeight). // Set explicit height to fill space
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

	// Calculate available height
	// Title: ~2 lines
	// Desc: ~1-3 lines
	// Hint: ~2-3 lines
	// Box overhead: ~4 lines
	// Total overhead: ~10 lines
	
	formHeight := a.contentHeight - lipgloss.Height(title) - lipgloss.Height(moduleDesc) - 3 // 3 for hint
	if formHeight < 5 {
		formHeight = 5
	}
	
	maxVisible := (formHeight - 4) // Approx 1 line per field

	// Show scrollable list of properties
	startIdx := 0
	
	if a.fieldIndex >= maxVisible {
		startIdx = a.fieldIndex - maxVisible + 1
	}

	endIdx := startIdx + maxVisible
	if endIdx > len(a.moduleProperties) {
		endIdx = len(a.moduleProperties)
		// Adjust start if we have space
		if endIdx-startIdx < maxVisible {
			startIdx = endIdx - maxVisible
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}
	
	moduleConfig := a.config.GetModuleConfig(a.editingModule)

	var formItems []string
	for i := startIdx; i < endIdx; i++ {
		prop := a.moduleProperties[i]
		propName := prop.Name

		// Get value from config directly
		valueStr := ""
		if val, ok := moduleConfig[propName]; ok {
			valueStr = fmt.Sprintf("%v", val)
		}

		label := styles.LabelStyle.Width(25).Render(propName + ":")

		var input string
		displayVal := valueStr
		isPlaceholder := false

		if valueStr == "" {
			if len(prop.Options) > 0 {
				displayVal = strings.Join(prop.Options, ", ")
				isPlaceholder = true
			} else if prop.Type == "boolean" {
				displayVal = "true, false"
				isPlaceholder = true
			}
			
			// Truncate placeholder if too long
			if isPlaceholder && len(displayVal) > 40 {
				displayVal = displayVal[:37] + "..."
			}
		}

		if i == a.fieldIndex {
			style := styles.FocusedInputStyle
			if isPlaceholder {
				style = style.Copy().Foreground(styles.TextDim)
			}
			input = style.Render(displayVal)
			// Edit icon
			input = lipgloss.JoinHorizontal(lipgloss.Left, input, " ", styles.ActiveItemStyle.Render(styles.IconEdit))
		} else {
			if isPlaceholder {
				input = styles.PlaceholderStyle.Render(displayVal)
			} else {
				input = styles.InputStyle.Render(displayVal)
			}
		}
		
		// Note: We removed live validation visualization here because it's now handled in modal or on input.
		// If we want to show validation status of stored config, we'd need to re-validate here.
		// For now, simple display is consistent with "read-only view".

		row := lipgloss.JoinHorizontal(lipgloss.Center, label, input)
		formItems = append(formItems, row)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)

	formBox := styles.BoxStyle.
		Width(a.width - 10).
		Height(formHeight).
		Render(form)

	scrollInfo := fmt.Sprintf("Showing %d-%d of %d properties", startIdx+1, endIdx, len(a.moduleProperties))

	// Add options hint
	var extraHint string
	if a.fieldIndex < len(a.moduleProperties) {
		focusedProp := a.moduleProperties[a.fieldIndex]
		if len(focusedProp.Options) > 0 {
			extraHint = fmt.Sprintf("\nAllowed values: %s", strings.Join(focusedProp.Options, ", "))
		} else if focusedProp.Type != "" && focusedProp.Type != "string" {
			extraHint = fmt.Sprintf("\nType: %s", focusedProp.Type)
		}
	}

	hint := styles.DescriptionStyle.Render(scrollInfo + extraHint + " | Use ↑↓ to navigate, Enter to edit, Esc to go back")

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

	// Calculate available height
	listHeight := a.contentHeight - lipgloss.Height(title) - 2
	if listHeight < 4 {
		listHeight = 4
	}
	
	maxLines := listHeight - 4
	if maxLines < 1 {
		maxLines = 1
	}

	// Show CSS content with line numbers
	lines := strings.Split(a.styleContent, "\n")
	
	// Since we don't have scrolling logic in style editor yet (it's read only here as per comment),
	// we just show first N lines or truncated note.
	// Implementing basic scrolling would be good but for now we follow request to fill space.
	
	visibleLines := lines
	if len(lines) > maxLines {
		visibleLines = lines[:maxLines]
	}

	var codeLines []string
	for i, line := range visibleLines {
		lineNum := styles.LineNumberStyle.Render(fmt.Sprintf("%3d", i+1))
		code := styles.CodeStyle.Render(line)
		codeLines = append(codeLines, lipgloss.JoinHorizontal(lipgloss.Top, lineNum, code))
	}

	code := lipgloss.JoinVertical(lipgloss.Left, codeLines...)

	codeBox := styles.BoxStyle.
		Width(a.width - 10).
		Height(listHeight).
		Render(code)

	if len(lines) > maxLines {
		// If we truncated, maybe show that in hint or inside box?
		// But codeBox has fixed height now, so it will just be filled.
		// If we append "...", it might overflow.
	}

	hint := styles.DescriptionStyle.Render("Style file: " + a.stylePath + " | Press Esc to go back")

	return lipgloss.JoinVertical(lipgloss.Left, title, codeBox, hint)
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
