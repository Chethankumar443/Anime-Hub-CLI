package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/yourorg/anime-cli/pkg/config"
	"github.com/yourorg/anime-cli/pkg/player"
	"github.com/yourorg/anime-cli/pkg/provider"
	"github.com/yourorg/anime-cli/pkg/syncplay"
)

type ViewState int

const (
	SearchView ViewState = iota
	SelectAnimeView
	SelectProviderView
	SelectDubSubView
	SelectActionView
	SelectEpisodeView
	SelectPlayerView
	SelectQualityView
	PlaybackActiveView
	DownloadingView
)

type AnimeItem struct {
	Anime provider.Anime
}

func (i AnimeItem) Title() string       { return i.Anime.Title }
func (i AnimeItem) Description() string { return fmt.Sprintf("Year: %d | Format: %s", i.Anime.ReleaseYear, i.Anime.Format) }
func (i AnimeItem) FilterValue() string { return i.Anime.Title }

type EpisodeItem struct {
	Episode provider.Episode
}

func (i EpisodeItem) Title() string       { return i.Episode.Title }
func (i EpisodeItem) Description() string { return fmt.Sprintf("Duration: %dm", i.Episode.DurationSec/60) }
func (i EpisodeItem) FilterValue() string { return i.Episode.Title }

// Generic Selector Choice helper
type SelectionModel struct {
	Title    string
	Choices  []string
	Selected int
}

type MainModel struct {
	State        ViewState
	Width        int
	Height       int

	// Wizard choices state
	SearchQuery      string
	SelectedAnime    provider.Anime
	SelectedProvider string
	SelectedDubSub   string // Sub or Dub
	SelectedAction   string // Stream or Download
	SelectedEpisode  provider.Episode
	SelectedPlayer   string // MPV or VLC
	SelectedQuality  string // 1080p, 720p, etc.

	// In-memory choice selectors
	ProviderSelector SelectionModel
	DubSubSelector   SelectionModel
	ActionSelector   SelectionModel
	PlayerSelector   SelectionModel
	QualitySelector  SelectionModel

	// Persistence & Engines
	AppConfig    *config.Config
	AppHistory   *config.History
	Prov         provider.Provider
	Player       *player.PlayerManager
	IPC          *player.IPCClient
	Syncplay     *syncplay.SyncplayClient

	// UI Components
	SearchInput  textinput.Model
	SearchList   list.Model
	EpisodeList  list.Model

	// Active State Data
	AnimeDetails     provider.AnimeDetails
	PlaybackTick     player.PlaybackUpdate
	CoverImagePath   string // Cached cover path if rendering
	SidebarCoverPath string // Cover path for previewing sidebar items

	// Download progress
	DownloadPercent  float64
	DownloadBytes    int64
	DownloadTotal    int64
	DownloadSpeed    string

	// Control states
	Loading      bool
	ErrorMsg     string
	Styles       *Styles
	LastSaveTime time.Time
}

func NewMainModel(cfg *config.Config, hist *config.History, prov provider.Provider) MainModel {
	si := textinput.New()
	si.Placeholder = "Search for anime (e.g. Naruto, One Piece)..."
	si.Focus()
	si.CharLimit = 150
	si.Width = 50

	// Custom item delegates
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(DefaultStyles().AccentColor).BorderForeground(DefaultStyles().AccentColor)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(DefaultStyles().AccentColor).BorderForeground(DefaultStyles().AccentColor)

	searchList := list.New([]list.Item{}, d, 0, 0)
	searchList.Title = "Select Anime"
	searchList.SetShowHelp(false)

	epList := list.New([]list.Item{}, d, 0, 0)
	epList.Title = "Select Episode"
	epList.SetShowHelp(false)

	// Setup wizards options
	var provChoices []string
	for _, p := range provider.GetRegistry().List() {
		provChoices = append(provChoices, p.Name())
	}
	if len(provChoices) == 0 {
		provChoices = []string{"Gogoanime", "AllAnime"}
	}

	provSelector := SelectionModel{
		Title:   "SELECT STREAMING SOURCE / PROVIDER",
		Choices: provChoices,
	}

	dubSubSelector := SelectionModel{
		Title:   "SELECT AUDIO & TRANSLATION TYPE",
		Choices: []string{"Subtitled (Sub) - Original Japanese Audio", "English Dubbed (Dub) - English Voiceover"},
	}

	actionSelector := SelectionModel{
		Title:   "SELECT PREFERRED ACTION",
		Choices: []string{"Stream Video (External Player)", "Download Episode to Local Storage"},
	}

	playerSelector := SelectionModel{
		Title:   "SELECT MEDIA PLAYER CONTROLLER",
		Choices: []string{"MPV Media Player (Recommended, Sockets Control)", "VLC Player (HTTP Sync Engine)"},
	}

	qualitySelector := SelectionModel{
		Title:   "SELECT TARGET VIDEO RESOLUTION",
		Choices: []string{"1080p (Full High Definition)", "720p (High Definition)", "480p (Standard Definition)", "Auto (Source Default)"},
	}

	m := MainModel{
		State:            SearchView,
		AppConfig:        cfg,
		AppHistory:       hist,
		Prov:             prov,
		Player:           player.NewPlayerManager(cfg.DefaultPlayer),
		SearchInput:      si,
		SearchList:       searchList,
		EpisodeList:      epList,
		ProviderSelector: provSelector,
		DubSubSelector:   dubSubSelector,
		ActionSelector:   actionSelector,
		PlayerSelector:   playerSelector,
		QualitySelector:  qualitySelector,
		Styles:           DefaultStyles(),
	}

	return m
}
