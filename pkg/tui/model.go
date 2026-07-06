package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"animehub/pkg/provider"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var PlayerPath = "mpv"

const asciiLogo = `   █████████               ███                              █████   █████            █████       
  ███▒▒▒▒▒███             ▒▒▒                              ▒▒███   ▒▒███            ▒▒███        
 ▒███    ▒███  ████████   ████  █████████████    ██████     ▒███    ▒███  █████ ████ ▒███████    
 ▒███████████ ▒▒███▒▒███ ▒▒███ ▒▒███▒▒███▒▒███  ███▒▒███    ▒███████████ ▒▒███ ▒███  ▒███▒▒███   
 ▒███▒▒▒▒▒███  ▒███ ▒███  ▒███  ▒███ ▒███ ▒███ ▒███████     ▒███▒▒▒▒▒███  ▒███ ▒███  ▒███ ▒███   
 ▒███    ▒███  ▒███ ▒███  ▒███  ▒███ ▒███ ▒███ ▒███▒▒▒      ▒███    ▒███  ▒███ ▒███  ▒███ ▒███   
 █████   █████ ████ █████ █████ █████▒███ █████▒▒██████     █████   █████ ▒▒████████ ████████    
▒▒▒▒▒   ▒▒▒▒▒ ▒▒▒▒ ▒▒▒▒▒ ▒▒▒▒▒ ▒▒▒▒▒ ▒▒▒ ▒▒▒▒▒  ▒▒▒▒▒▒     ▒▒▒▒▒   ▒▒▒▒▒   ▒▒▒▒▒▒▒▒ ▒▒▒▒▒▒▒▒`

const smallLogo = `   __ _  _ _  _ _ _  ____   _  _  _ _  ___ 
  / _ \| | | || | | |/ ___) | || |/ | |/ _ \
 |  __/| | | || | | |  ___| | || |\_/| |  __/
  \___)|_|_|_||_|_|_|\____) |_||_||__|_|\___)
             A N I M E   H U B`

type SessionState int

const (
	SearchState SessionState = iota
	ResultsState
	EpisodeSelectState
)

type AnimeItem struct {
	Anime provider.Anime
}

func (i AnimeItem) Title() string       { return i.Anime.Title }
func (i AnimeItem) Description() string { return "ID: " + i.Anime.ID }
func (i AnimeItem) FilterValue() string { return i.Anime.Title }

type EpisodeItem struct {
	Episode provider.Episode
}

func (i EpisodeItem) Title() string       { return fmt.Sprintf("Episode %d", i.Episode.Number) }
func (i EpisodeItem) Description() string { return "ID: " + i.Episode.ID }
func (i EpisodeItem) FilterValue() string { return fmt.Sprintf("Episode %d", i.Episode.Number) }

type playbackFinishedMsg struct{ err error }

type searchResultMsg struct {
	results []provider.Anime
	err     error
}

type episodesResultMsg struct {
	episodes []provider.Episode
	err      error
}

type Model struct {
	state           SessionState
	provider        provider.AnimeProvider
	selectedAnime   provider.Anime
	selectedEpisode provider.Episode
	selectedLang    string

	searchInput  textinput.Model
	resultsList  list.Model
	episodesList list.Model

	loading bool
	err     error

	terminalWidth  int
	terminalHeight int
}

func NewMainModel(prov provider.AnimeProvider) Model {
	si := textinput.New()
	si.Placeholder = "Search for anime (e.g. Naruto)..."
	si.Focus()
	si.CharLimit = 150
	si.Width = 50

	rl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	rl.Title = "Search Results"
	rl.SetShowHelp(false)

	el := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	el.Title = "Select Episode"
	el.SetShowHelp(false)

	return Model{
		state:        SearchState,
		provider:     prov,
		selectedLang: "sub", // default
		searchInput:  si,
		resultsList:  rl,
		episodesList: el,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.resultsList.SetSize(msg.Width-4, msg.Height-12)
		m.episodesList.SetSize(msg.Width-4, msg.Height-12)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.loading {
			return m, nil // Block input while loading
		}

		switch m.state {
		case SearchState:
			if msg.String() == "enter" {
				query := m.searchInput.Value()
				if query != "" {
					m.loading = true
					m.err = nil
					cmds = append(cmds, m.searchAnime(query))
				}
			} else {
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case ResultsState:
			if msg.String() == "enter" {
				if item, ok := m.resultsList.SelectedItem().(AnimeItem); ok {
					m.selectedAnime = item.Anime
					m.loading = true
					m.err = nil
					cmds = append(cmds, m.getEpisodes(item.Anime.ID))
				}
			} else if msg.String() == "esc" {
				m.state = SearchState
			} else {
				m.resultsList, cmd = m.resultsList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case EpisodeSelectState:
			if msg.String() == "enter" {
				if item, ok := m.episodesList.SelectedItem().(EpisodeItem); ok {
					m.selectedEpisode = item.Episode

					url, err := m.provider.GetStreamURL(m.selectedEpisode.ID, m.selectedLang)
					if err != nil {
						m.err = err
						return m, nil
					}

					// --no-terminal isolates standard output descriptors from polluting terminal UI grids
					cmd := exec.Command(PlayerPath, "--no-terminal", url)

					// Safely pause the TUI loop, pass control to mpv, and restore screen buffers on close
					return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
						return playbackFinishedMsg{err: err}
					})
				}
			} else if msg.String() == "esc" {
				m.state = ResultsState
			} else {
				m.episodesList, cmd = m.episodesList.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			items := make([]list.Item, len(msg.results))
			for i, r := range msg.results {
				items[i] = AnimeItem{Anime: r}
			}
			m.resultsList.SetItems(items)
			m.state = ResultsState
		}

	case episodesResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			items := make([]list.Item, len(msg.episodes))
			for i, e := range msg.episodes {
				items[i] = EpisodeItem{Episode: e}
			}
			m.episodesList.SetItems(items)
			m.state = EpisodeSelectState
		}

	case playbackFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) getLogo() string {
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Cyan / blue accent
		Bold(true)

	if m.terminalWidth >= 100 {
		return logoStyle.Render(asciiLogo)
	} else if m.terminalWidth >= 48 {
		return logoStyle.Render(smallLogo)
	}
	return logoStyle.Render("=== ANIME HUB ===")
}

func (m Model) renderWithHeader(content string) string {
	logo := m.getLogo()

	if m.terminalWidth == 0 {
		return fmt.Sprintf("%s\n\n%s", logo, content)
	}

	// Center the logo lines
	var logoLines []string
	for _, line := range strings.Split(logo, "\n") {
		logoLines = append(logoLines, lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, line))
	}
	centeredLogo := strings.Join(logoLines, "\n")

	// Center the content lines
	var contentLines []string
	for _, line := range strings.Split(content, "\n") {
		contentLines = append(contentLines, lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, line))
	}
	centeredContent := strings.Join(contentLines, "\n")

	return fmt.Sprintf("%s\n\n%s", centeredLogo, centeredContent)
}

func (m Model) View() string {
	if m.err != nil {
		errorContent := fmt.Sprintf("Error: %v\n\nPress esc to go back or ctrl+c to quit.", m.err)
		return m.renderWithHeader(errorContent)
	}

	if m.loading {
		return m.renderWithHeader("Loading...")
	}

	switch m.state {
	case SearchState:
		searchContent := fmt.Sprintf(
			"Welcome to AnimeHub\n\n%s\n\n(Press Enter to search, ctrl+c to quit)",
			m.searchInput.View(),
		)
		return m.renderWithHeader(searchContent)
	case ResultsState:
		return m.renderWithHeader(m.resultsList.View())
	case EpisodeSelectState:
		return m.renderWithHeader(m.episodesList.View())
	}

	return ""
}

func (m Model) searchAnime(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.provider.Search(query)
		return searchResultMsg{results: results, err: err}
	}
}

func (m Model) getEpisodes(animeID string) tea.Cmd {
	return func() tea.Msg {
		episodes, err := m.provider.GetEpisodes(animeID)
		return episodesResultMsg{episodes: episodes, err: err}
	}
}
