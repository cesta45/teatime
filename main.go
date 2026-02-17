package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabrielfornes/teatime/internal/storage"
	"github.com/gabrielfornes/teatime/internal/tui"
)

func main() {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing teatime: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(store)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running teatime: %v\n", err)
		os.Exit(1)
	}
}
