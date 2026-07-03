package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourorg/anime-cli/internal/config"
	"github.com/yourorg/anime-cli/internal/persistence"
	pkgconfig "github.com/yourorg/anime-cli/pkg/config"
	"github.com/yourorg/anime-cli/pkg/provider"
	"github.com/yourorg/anime-cli/pkg/provider/allanime"
	"github.com/yourorg/anime-cli/pkg/provider/gogo"
	"github.com/yourorg/anime-cli/pkg/tui"
)

var (
	version = "1.0.0"
	commit  = "none"
	date    = "unknown"
)

func main() {
	updateFlag := flag.Bool("update", false, "Check for available updates")
	versionFlag := flag.Bool("version", false, "Print application version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("anime-cli version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	if *updateFlag {
		tui.CheckForUpdates(version)
		os.Exit(0)
	}

	// 1. Load Configurations Persistent settings
	cfg, err := pkgconfig.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configurations: %v\n", err)
		os.Exit(1)
	}

	// 2. Load Watch History Data
	hist, err := pkgconfig.LoadHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading watch history database: %v\n", err)
		os.Exit(1)
	}

	// 3. Initialize internal/persistence and config stores
	home, _ := os.UserHomeDir()
	internalDir := filepath.Join(home, ".config", "anime-cli-internal")
	_ = os.MkdirAll(internalDir, 0755)

	_, err = persistence.NewStore(filepath.Join(internalDir, "state.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize app state store: %v\n", err)
	}

	_, err = config.NewConfigManager(filepath.Join(internalDir, "prefs.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize user preferences: %v\n", err)
	}

	// 4. Initialize Provider Registry
	reg := provider.GetRegistry()
	
	gogoSettings := cfg.ProvidersConfig["gogoanime"]
	gogoProvider := gogo.NewGogoProvider(gogoSettings.BaseURL)
	reg.Register(gogoProvider)

	var allBaseURL string
	if allSettings, ok := cfg.ProvidersConfig["allanime"]; ok {
		allBaseURL = allSettings.BaseURL
	}
	allProvider := allanime.NewAllAnimeProvider(allBaseURL)
	reg.Register(allProvider)

	// Get default provider to boot model
	defaultProv, err := reg.Get("gogoanime")
	if err != nil {
		defaultProv = gogoProvider
	}

	// 5. Initialize Bubbletea TUI Model
	mainModel := tui.NewMainModel(cfg, hist, defaultProv)

	// 6. Start AltScreen Program
	p := tea.NewProgram(mainModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI loop: %v\n", err)
		os.Exit(1)
	}
}
