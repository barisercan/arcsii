package main

import (
	"fmt"
	"os"

	"github.com/barisercan/arcsii/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Get the target directory (current dir or specified)
	targetDir := "."
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	}

	p := tea.NewProgram(
		ui.NewModel(targetDir),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running arcsii: %v\n", err)
		os.Exit(1)
	}
}
