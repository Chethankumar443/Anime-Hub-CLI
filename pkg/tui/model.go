package tui

import (
	"fmt"
	"os/exec"

	"animehub/pkg/provider"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var PlayerPath = "mpv"

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
		m.resultsList.SetSize(msg.Width, msg.Height-4)
		m.episodesList.SetSize(msg.Width, msg.Height-4)
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

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress esc to go back or ctrl+c to quit.", m.err)
	}

	if m.loading {
		return "Loading..."
	}

	switch m.state {
	case SearchState:
		return fmt.Sprintf(
			"Welcome to AnimeHub\n\n%s\n\n(Press Enter to search, ctrl+c to quit)",
			m.searchInput.View(),
		)
	case ResultsState:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.resultsList.View())
	case EpisodeSelectState:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.episodesList.View())
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
