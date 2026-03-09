package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"netbird-tui/internal/client"
	"netbird-tui/internal/ui"
)

func main() {
	socketPath := "unix:///var/run/netbird.sock"
	if len(os.Args) > 1 {
		socketPath = os.Args[1]
	}

	c, err := client.New(socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to NetBird daemon: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure NetBird is running and you have access to %s\n", socketPath)
		os.Exit(1)
	}
	defer c.Close()

	model := ui.New(c)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
