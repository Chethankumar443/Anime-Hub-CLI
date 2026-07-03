package tui

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/yourorg/anime-cli/pkg/cache"
	"github.com/yourorg/anime-cli/pkg/config"
	"github.com/yourorg/anime-cli/pkg/player"
	"github.com/yourorg/anime-cli/pkg/provider"
	"github.com/yourorg/anime-cli/pkg/syncplay"
)

type searchCompletedMsg struct {
	results []provider.Anime
	err     error
}

type animeDetailsLoadedMsg struct {
	details provider.AnimeDetails
	err     error
}

type imageDownloadedMsg struct {
	path string
	err  error
}

type downloadProgressMsg struct {
	Progress cache.DownloadProgress
	Chan     <-chan cache.DownloadProgress
}

type playerClosedMsg struct{}

type playerTickMsg struct {
	tick player.PlaybackUpdate
}

type syncplayEventMsg syncplay.SyncEvent

type playerStartedMsg struct {
	IPC      *player.IPCClient
	Syncplay *syncplay.SyncplayClient
}

func watchSyncplayEvents(client *syncplay.SyncplayClient) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-client.Events()
		if !ok {
			return nil
		}
		return syncplayEventMsg(ev)
	}
}

func triggerSearch(prov provider.Provider, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := prov.Search(context.Background(), query)
		return searchCompletedMsg{results: results, err: err}
	}
}

func triggerDetails(prov provider.Provider, animeID string) tea.Cmd {
	return func() tea.Msg {
		details, err := prov.FetchDetails(context.Background(), animeID)
		return animeDetailsLoadedMsg{details: details, err: err}
	}
}

func triggerImageDownload(url string) tea.Cmd {
	return func() tea.Msg {
		path, err := cache.DownloadImage(context.Background(), url)
		return imageDownloadedMsg{path: path, err: err}
	}
}

func watchPlayerUpdates(ipc *player.IPCClient) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-ipc.Updates()
		if !ok || update.Closed {
			return playerClosedMsg{}
		}
		return playerTickMsg{tick: update}
	}
}

func watchDownloadProgress(ch <-chan cache.DownloadProgress) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-ch
		if !ok {
			return downloadProgressMsg{Progress: cache.DownloadProgress{Done: true}}
		}
		return downloadProgressMsg{Progress: progress, Chan: ch}
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) getHydratedAnimeID() string {
	id := m.SelectedAnime.ID
	isDubSelected := m.DubSubSelector.Selected == 1 // English Dubbed (Dub)

	if m.SelectedProvider == "Gogoanime (Decrypted CDN Mirror)" || m.Prov.ID() == "gogoanime" {
		if isDubSelected {
			if !strings.HasSuffix(id, "-dub") {
				id = id + "-dub"
			}
		} else {
			id = strings.TrimSuffix(id, "-dub")
		}
	}
	return id
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ErrorMsg != "" {
			if msg.String() == "esc" {
				m.ErrorMsg = ""
				m.Loading = false
				if m.State == SelectPlayerView || m.State == PlaybackActiveView || m.State == SelectQualityView || m.State == DownloadingView {
					m.State = SelectEpisodeView
				}
				return m, nil
			}
			if msg.String() == "ctrl+c" {
				if m.IPC != nil {
					m.IPC.Close()
				}
				return m, tea.Quit
			}
			return m, nil // Ignore other keys on error screen
		}

		switch msg.String() {
		case "ctrl+c":
			if m.IPC != nil {
				m.IPC.Close()
			}
			return m, tea.Quit
		}

		// Handle key presses based on active screen wizard state
		switch m.State {
		case SearchView:
			switch msg.String() {
			case "enter":
				query := m.SearchInput.Value()
				if query != "" {
					m.Loading = true
					m.ErrorMsg = ""
					m.SearchQuery = query
					return m, triggerSearch(m.Prov, query)
				}
			case "esc":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.SearchInput, cmd = m.SearchInput.Update(msg)
			cmds = append(cmds, cmd)

		case SelectAnimeView:
			oldSel := m.SearchList.SelectedItem()
			switch msg.String() {
			case "esc":
				m.State = SearchView
				m.SearchInput.Focus()
				m.CoverImagePath = ""
				return m, nil
			case "enter":
				if sel := m.SearchList.SelectedItem(); sel != nil {
					m.SelectedAnime = sel.(AnimeItem).Anime
					m.State = SelectProviderView
					m.ProviderSelector.Selected = 0
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.SearchList, cmd = m.SearchList.Update(msg)
			cmds = append(cmds, cmd)

			newSel := m.SearchList.SelectedItem()
			if newSel != nil && (oldSel == nil || oldSel.(AnimeItem).Anime.ID != newSel.(AnimeItem).Anime.ID) {
				anime := newSel.(AnimeItem).Anime
				if anime.CoverURL != "" {
					cmds = append(cmds, triggerImageDownload(anime.CoverURL))
				} else {
					m.CoverImagePath = ""
				}
			}

		case SelectProviderView:
			switch msg.String() {
			case "esc":
				m.State = SelectAnimeView
				return m, nil
			case "up", "k":
				if m.ProviderSelector.Selected > 0 {
					m.ProviderSelector.Selected--
				}
			case "down", "j":
				if m.ProviderSelector.Selected < len(m.ProviderSelector.Choices)-1 {
					m.ProviderSelector.Selected++
				}
			case "enter":
				m.SelectedProvider = m.ProviderSelector.Choices[m.ProviderSelector.Selected]
				
				// Lookup and switch active provider dynamically
				for _, p := range provider.GetRegistry().List() {
					if p.Name() == m.SelectedProvider {
						m.Prov = p
						break
					}
				}

				m.State = SelectDubSubView
				m.DubSubSelector.Selected = 0
				return m, nil
			}

		case SelectDubSubView:
			switch msg.String() {
			case "esc":
				m.State = SelectProviderView
				return m, nil
			case "up", "k":
				if m.DubSubSelector.Selected > 0 {
					m.DubSubSelector.Selected--
				}
			case "down", "j":
				if m.DubSubSelector.Selected < len(m.DubSubSelector.Choices)-1 {
					m.DubSubSelector.Selected++
				}
			case "enter":
				m.SelectedDubSub = m.DubSubSelector.Choices[m.DubSubSelector.Selected]
				m.State = SelectActionView
				m.ActionSelector.Selected = 0
				return m, nil
			}

		case SelectActionView:
			switch msg.String() {
			case "esc":
				m.State = SelectDubSubView
				return m, nil
			case "up", "k":
				if m.ActionSelector.Selected > 0 {
					m.ActionSelector.Selected--
				}
			case "down", "j":
				if m.ActionSelector.Selected < len(m.ActionSelector.Choices)-1 {
					m.ActionSelector.Selected++
				}
			case "enter":
				m.SelectedAction = m.ActionSelector.Choices[m.ActionSelector.Selected]
				m.Loading = true
				return m, triggerDetails(m.Prov, m.getHydratedAnimeID())
			}

		case SelectEpisodeView:
			switch msg.String() {
			case "esc":
				m.State = SelectActionView
				m.CoverImagePath = ""
				return m, nil
			case "enter":
				if sel := m.EpisodeList.SelectedItem(); sel != nil {
					m.SelectedEpisode = sel.(EpisodeItem).Episode
					if m.SelectedAction == m.ActionSelector.Choices[0] { // Stream
						m.State = SelectPlayerView
						m.PlayerSelector.Selected = 0
					} else { // Download
						m.State = SelectQualityView
						m.QualitySelector.Selected = 0
					}
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.EpisodeList, cmd = m.EpisodeList.Update(msg)
			cmds = append(cmds, cmd)

		case SelectPlayerView:
			switch msg.String() {
			case "esc":
				m.State = SelectEpisodeView
				return m, nil
			case "up", "k":
				if m.PlayerSelector.Selected > 0 {
					m.PlayerSelector.Selected--
				}
			case "down", "j":
				if m.PlayerSelector.Selected < len(m.PlayerSelector.Choices)-1 {
					m.PlayerSelector.Selected++
				}
			case "enter":
				m.SelectedPlayer = m.PlayerSelector.Choices[m.PlayerSelector.Selected]
				m.Loading = true
				
				// Launch Stream playback
				return m, func() tea.Msg {
					links, err := m.Prov.FetchStreamLinks(context.Background(), m.SelectedEpisode.ID)
					if err != nil {
						return animeDetailsLoadedMsg{err: err}
					}
					if len(links) == 0 {
						return animeDetailsLoadedMsg{err: fmt.Errorf("no streams found")}
					}

					var ipcPath string
					if runtime.GOOS == "windows" {
						rPort := rand.Intn(16383) + 49152
						ipcPath = fmt.Sprintf("127.0.0.1:%d", rPort)
					} else {
						ipcPath = filepath.Join(os.TempDir(), fmt.Sprintf("anime-cli-ipc-%d.sock", os.Getpid()))
						_ = os.Remove(ipcPath)
					}

					cmd, err := m.Player.StartPlayer(links[0].URL, ipcPath, links[0].HTTPHeaders)
					if err != nil {
						return animeDetailsLoadedMsg{err: err}
					}

					ipc := player.NewIPCClient(ipcPath)
					if strings.Contains(strings.ToLower(m.SelectedPlayer), "vlc") {
						ipc.PlayerName = "vlc"
					} else {
						ipc.PlayerName = "mpv"
					}
					if err := ipc.Connect(context.Background()); err != nil {
						_ = cmd.Process.Kill()
						return animeDetailsLoadedMsg{err: err}
					}

					ipc.StartMonitoring(context.Background())

					var syncClient *syncplay.SyncplayClient
					if m.AppConfig.SyncplayServerURL != "" {
						syncClient = syncplay.NewSyncplayClient(
							m.AppConfig.SyncplayServerURL,
							"general",
							"user",
						)
						_ = syncClient.Connect(context.Background())
						syncClient.Start(context.Background())
					}

					return playerStartedMsg{
						IPC:      ipc,
						Syncplay: syncClient,
					}
				}
			}

		case SelectQualityView:
			switch msg.String() {
			case "esc":
				m.State = SelectEpisodeView
				return m, nil
			case "up", "k":
				if m.QualitySelector.Selected > 0 {
					m.QualitySelector.Selected--
				}
			case "down", "j":
				if m.QualitySelector.Selected < len(m.QualitySelector.Choices)-1 {
					m.QualitySelector.Selected++
				}
			case "enter":
				m.SelectedQuality = m.QualitySelector.Choices[m.QualitySelector.Selected]
				m.State = DownloadingView
				m.DownloadPercent = 0
				m.DownloadSpeed = "0 B/s"

				// Trigger background download goroutine
				progressChan := make(chan cache.DownloadProgress, 10)
				
				// Launch download fetch
				cmds = append(cmds, func() tea.Msg {
					links, err := m.Prov.FetchStreamLinks(context.Background(), m.SelectedEpisode.ID)
					if err != nil {
						return downloadProgressMsg{Progress: cache.DownloadProgress{Err: err}}
					}
					if len(links) == 0 {
						return downloadProgressMsg{Progress: cache.DownloadProgress{Err: fmt.Errorf("stream links not found")}}
					}

					// Resolve local path target (~/Downloads/AnimeName/Episode.mp4)
					home, _ := os.UserHomeDir()
					destFolder := filepath.Join(home, "Downloads", "anime-cli", m.SelectedAnime.Title)
					destFile := filepath.Join(destFolder, fmt.Sprintf("Episode_%d.mp4", m.SelectedEpisode.Number))

					go cache.DownloadVideoFile(context.Background(), links[0].URL, destFile, progressChan)

					return downloadProgressMsg{Chan: progressChan}
				})

				return m, tea.Batch(append(cmds, watchDownloadProgress(progressChan))...)
			}

		case PlaybackActiveView:
			// View goes idle during playback, waiting for player closed signal
			
		case DownloadingView:
			if msg.String() == "esc" {
				m.State = SelectEpisodeView
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		
		listWidth := (msg.Width / 2) - 4
		listHeight := msg.Height - 12
		m.SearchList.SetSize(listWidth, listHeight)
		m.EpisodeList.SetSize(listWidth, listHeight)

	case searchCompletedMsg:
		m.Loading = false
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		
		var items []list.Item
		for _, a := range msg.results {
			items = append(items, AnimeItem{Anime: a})
		}
		m.SearchList.SetItems(items)
		m.State = SelectAnimeView
		m.CoverImagePath = ""

		if len(items) > 0 {
			firstAnime := items[0].(AnimeItem).Anime
			if firstAnime.CoverURL != "" {
				return m, triggerImageDownload(firstAnime.CoverURL)
			}
		}

	case animeDetailsLoadedMsg:
		m.Loading = false
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.AnimeDetails = msg.details
		
		var items []list.Item
		for _, ep := range msg.details.Episodes {
			items = append(items, EpisodeItem{Episode: ep})
		}
		m.EpisodeList.SetItems(items)
		m.State = SelectEpisodeView

		if msg.details.CoverURL != "" {
			return m, triggerImageDownload(msg.details.CoverURL)
		}

	case imageDownloadedMsg:
		if msg.err == nil {
			m.CoverImagePath = msg.path
		}

	case downloadProgressMsg:
		if msg.Progress.Err != nil {
			m.ErrorMsg = msg.Progress.Err.Error()
			return m, nil
		}

		m.DownloadPercent = msg.Progress.Percent
		m.DownloadBytes = msg.Progress.Bytes
		m.DownloadTotal = msg.Progress.Total
		m.DownloadSpeed = msg.Progress.Speed

		if msg.Progress.Done {
			// Finished downloading, direct return to episodes
			m.State = SelectEpisodeView
			return m, nil
		}

		if msg.Chan != nil {
			return m, watchDownloadProgress(msg.Chan)
		}

	case playerStartedMsg:
		m.Loading = false
		m.IPC = msg.IPC
		m.Syncplay = msg.Syncplay
		m.State = PlaybackActiveView

		var batch []tea.Cmd
		if msg.IPC != nil {
			batch = append(batch, watchPlayerUpdates(msg.IPC))
		}
		if msg.Syncplay != nil {
			batch = append(batch, watchSyncplayEvents(msg.Syncplay))
		}
		return m, tea.Batch(batch...)

	case syncplayEventMsg:
		if m.IPC != nil {
			ctx := context.Background()
			switch msg.Event {
			case "pause":
				_ = m.IPC.SetPause(ctx, true)
			case "play":
				_ = m.IPC.SetPause(ctx, false)
			case "seek":
				_ = m.IPC.Seek(ctx, msg.ElapsedSec)
			}
		}
		if m.Syncplay != nil {
			return m, watchSyncplayEvents(m.Syncplay)
		}

	case playerTickMsg:
		m.Loading = false
		m.State = PlaybackActiveView
		m.PlaybackTick = msg.tick

		if msg.tick.ElapsedSec > 0 && m.Syncplay != nil {
			_ = m.Syncplay.SendEvent(syncplay.SyncEvent{
				Event:      "status",
				Username:   m.Syncplay.Username,
				Room:       m.Syncplay.Room,
				ElapsedSec: msg.tick.ElapsedSec,
			})
		}

		// Throttled save every 30 seconds to prevent massive I/O overhead & GC pauses
		if msg.tick.ElapsedSec > 0 && msg.tick.DurationSec > 0 {
			if m.LastSaveTime.IsZero() {
				m.LastSaveTime = time.Now()
			} else if time.Since(m.LastSaveTime) >= 30*time.Second {
				m.LastSaveTime = time.Now()
				progress := provider.PlaybackProgress{
					AnimeID:     m.SelectedAnime.ID,
					EpisodeID:   m.SelectedEpisode.ID,
					ElapsedSec:  msg.tick.ElapsedSec,
					DurationSec: msg.tick.DurationSec,
					Completed:   float64(msg.tick.ElapsedSec) >= float64(msg.tick.DurationSec)*0.9,
				}
				cmds = append(cmds, func() tea.Msg {
					_ = config.UpdateProgress(progress)
					return nil
				})
			}
		}

		if m.IPC != nil {
			return m, tea.Batch(append(cmds, watchPlayerUpdates(m.IPC))...)
		}

	case playerClosedMsg:
		// Save final progress on close
		if m.PlaybackTick.ElapsedSec > 0 && m.PlaybackTick.DurationSec > 0 {
			progress := provider.PlaybackProgress{
				AnimeID:     m.SelectedAnime.ID,
				EpisodeID:   m.SelectedEpisode.ID,
				ElapsedSec:  m.PlaybackTick.ElapsedSec,
				DurationSec: m.PlaybackTick.DurationSec,
				Completed:   float64(m.PlaybackTick.ElapsedSec) >= float64(m.PlaybackTick.DurationSec)*0.9,
			}
			_ = config.UpdateProgress(progress)
		}
		m.State = SelectEpisodeView
		m.PlaybackTick = player.PlaybackUpdate{}
		if m.IPC != nil {
			m.IPC.Close()
			m.IPC = nil
		}
		if m.Syncplay != nil {
			m.Syncplay.Close()
			m.Syncplay = nil
		}
	}

	return m, tea.Batch(cmds...)
}
