package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"animehub/pkg/provider"
	"animehub/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	version = "1.0.0"
	commit  = "none"
	date    = "unknown"
)

var lookPath = exec.LookPath
var resolvedMPVPath = "mpv"

var fileExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkDependencies() error {
	// First check if 'mpv' is already in PATH
	path, err := lookPath("mpv")
	if err == nil {
		resolvedMPVPath = path
		return nil
	}

	// On Windows, check common default installation directories
	if runtime.GOOS == "windows" {
		var commonPaths []string
		if pf := os.Getenv("ProgramFiles"); pf != "" {
			commonPaths = append(commonPaths, filepath.Join(pf, `MPV Player\mpv.exe`))
		}
		if pfx86 := os.Getenv("ProgramFiles(x86)"); pfx86 != "" {
			commonPaths = append(commonPaths, filepath.Join(pfx86, `MPV Player\mpv.exe`))
		}
		if la := os.Getenv("LOCALAPPDATA"); la != "" {
			commonPaths = append(commonPaths, filepath.Join(la, `Microsoft\WinGet\Links\mpv.exe`))
		}
		// Hardcoded fallbacks just in case env vars are missing
		commonPaths = append(commonPaths,
			`C:\Program Files\MPV Player\mpv.exe`,
			`C:\Program Files (x86)\MPV Player\mpv.exe`,
		)

		for _, p := range commonPaths {
			if fileExists(p) {
				resolvedMPVPath = p
				return nil
			}
		}
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1)

	errorMessage := "Error: Dependency 'mpv' is missing.\n\n" +
		"Please install 'mpv' to stream video:\n" +
		"  - macOS:   brew install mpv\n" +
		"  - Linux:   sudo apt install mpv\n" +
		"  - Windows: winget install mpv"
	
	return errors.New(style.Render(errorMessage))
}

func main() {
	versionFlag := flag.Bool("version", false, "Print application version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("animehub version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Check mpv dependency first
	if err := checkDependencies(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
	actualPort := consumetManager.Port()
	consumetProv := provider.NewConsumetProvider(actualPort)
	fallbackManager := provider.NewFallbackManager(consumetProv)

	// 3. Initialize Bubbletea TUI Model
	tui.PlayerPath = resolvedMPVPath
	mainModel := tui.NewMainModel(fallbackManager)

	// 4. Start AltScreen Program
	p := tea.NewProgram(mainModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI loop: %v\n", err)
		os.Exit(1)
	}
}
