package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"animehub/pkg/provider"
	"animehub/pkg/tui"
)

var (
	version = "1.0.0"
	commit  = "none"
	date    = "unknown"
)

func main() {
	versionFlag := flag.Bool("version", false, "Print application version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("animehub version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// 1. Initialize Consumet Manager (Embedded Node Binary)
	// We use port 3000 as default
	consumetManager := provider.NewConsumetManager("3000")
	
	// Try starting the manager, fallback gracefully if already running
	if err := consumetManager.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Consumet: %v\n", err)
		os.Exit(1)
	}
	defer consumetManager.Stop()

	// 2. Initialize the Provider ecosystem
	consumetProv := provider.NewConsumetProvider("3000")
	fallbackManager := provider.NewFallbackManager(consumetProv)

	// 3. Initialize Bubbletea TUI Model
	mainModel := tui.NewMainModel(fallbackManager)

	// 4. Start AltScreen Program
	p := tea.NewProgram(mainModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI loop: %v\n", err)
		os.Exit(1)
	}
}
