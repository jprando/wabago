package main

import (
	"fmt"
	"os"

	"github.com/jprando/wabago/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	app := ui.NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
