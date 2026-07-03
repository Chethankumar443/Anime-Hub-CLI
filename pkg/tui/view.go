package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourorg/anime-cli/pkg/cache"
	"github.com/mattn/go-runewidth"
)

const asciiLogo = `   
   ▄████████ ███▄▄▄▄    ▄█    ▄▄▄▄███▄▄▄▄      ▄████████         ▄█    █▄    ███    █▄  ▀█████████▄       
  ███    ███ ███▀▀▀██▄ ███  ▄██▀▀▀███▀▀▀██▄   ███    ███        ███    ███   ███    ███   ███    ███      
  ███    ███ ███   ███ ███▌ ███   ███   ███   ███    █▀         ███    ███   ███    ███   ███    ███      
  ███    ███ ███   ███ ███▌ ███   ███   ███  ▄███▄▄▄           ▄███▄▄▄▄███▄▄ ███    ███  ▄███▄▄▄██▀       
▀███████████ ███   ███ ███▌ ███   ███   ███ ▀▀███▀▀▀          ▀▀███▀▀▀▀███▀  ███    ███ ▀▀███▀▀▀██▄       
  ███    ███ ███   ███ ███  ███   ███   ███   ███    █▄         ███    ███   ███    ███   ███    ██▄      
  ███    ███ ███   ███ ███  ███   ███   ███   ███    ███        ███    ███   ███    ███   ███    ███      
  ███    █▀   ▀█   █▀  █▀    ▀█   ███   █▀    ██████████        ███    █▀    ████████▀  ▄█████████▀       `

// Centered and themed ASCII logo renderer
func (m MainModel) renderASCIILogo() string {
	lines := strings.Split(asciiLogo, "\n")
	var styledLines []string
	for _, line := range lines {
		padding := (m.Width - runewidth.StringWidth(line)) / 2
		if padding < 0 {
			padding = 0
		}
		styledLines = append(styledLines, strings.Repeat(" ", padding)+m.Styles.TitleStyle.Render(line))
	}
	return strings.Join(styledLines, "\n") + "\n"
}

// Layout Dimensions calculations ensuring zero layout jitter
func (m MainModel) getLayoutDimensions() (int, int) {
	showLogo := m.Width >= 115 && m.Height >= 26
	var logoHeight int
	if showLogo {
		logoHeight = 9 // 8 lines logo + 1 blank line
	}

	panelWidth := m.Width - 8
	panelHeight := m.Height - 11 - logoHeight

	if panelWidth < 10 {
		panelWidth = 80
	}
	if panelHeight < 5 {
		panelHeight = 20
	}
	return panelWidth, panelHeight
}

func (m MainModel) View() string {
	var body string

	if m.Loading {
		return m.Styles.DocStyle.Render("\n\n   ▲ Connecting to video CDN server... Please wait.\n\n")
	}

	if m.ErrorMsg != "" {
		body = m.Styles.DocStyle.Render(fmt.Sprintf("\n\n   Error encountered: %s\n   Press Esc to step back, or Ctrl+C to quit.\n\n", m.ErrorMsg))
		return m.Styles.DocStyle.Render(lipgloss.JoinVertical(lipgloss.Top, m.headerView(), body, m.footerView()))
	}

	panelWidth, panelHeight := m.getLayoutDimensions()

	switch m.State {
	case SearchView:
		body = m.searchView()
	case SelectAnimeView:
		body = m.selectAnimeView()
	case SelectProviderView:
		body = renderSelectionWithSidebar(m, m.ProviderSelector, panelWidth, panelHeight)
	case SelectDubSubView:
		body = renderSelectionWithSidebar(m, m.DubSubSelector, panelWidth, panelHeight)
	case SelectActionView:
		body = renderSelectionWithSidebar(m, m.ActionSelector, panelWidth, panelHeight)
	case SelectEpisodeView:
		body = m.selectEpisodeView()
	case SelectPlayerView:
		body = renderSelectionWithSidebar(m, m.PlayerSelector, panelWidth, panelHeight)
	case SelectQualityView:
		body = renderSelectionWithSidebar(m, m.QualitySelector, panelWidth, panelHeight)
	case PlaybackActiveView:
		body = m.playbackView()
	case DownloadingView:
		body = m.downloadingView()
	default:
		body = "\n\n   Unknown State\n\n"
	}

	return m.Styles.DocStyle.Render(lipgloss.JoinVertical(lipgloss.Left, m.headerView(), body, m.footerView()))
}

func (m MainModel) headerView() string {
	var subtext string
	switch m.State {
	case SearchView:
		subtext = "1. SEARCH ENGINE"
	case SelectAnimeView:
		subtext = "2. CHOOSE SERIES"
	case SelectProviderView:
		subtext = "3. CHOOSE PROVIDER"
	case SelectDubSubView:
		subtext = "4. CHOOSE SUB/DUB"
	case SelectActionView:
		subtext = "5. CHOOSE ACTION"
	case SelectEpisodeView:
		subtext = "6. CHOOSE EPISODE"
	case SelectPlayerView:
		subtext = "7. CHOOSE MEDIA PLAYER"
	case SelectQualityView:
		subtext = "7. CHOOSE QUALITY"
	case PlaybackActiveView:
		subtext = "8. PLAYBACK ENGAGED"
	case DownloadingView:
		subtext = "8. FILE DOWNLOAD LIVE"
	}

	var titleText string
	showLogo := m.Width >= 115 && m.Height >= 26
	if showLogo {
		titleText = m.renderASCIILogo()
	} else {
		titleText = m.Styles.HeaderStyle.Render("▲ ANIME DOWNLOAD & STREAM ENGINE")
	}

	progressText := m.Styles.InactiveTab.Render(subtext)
	
	return lipgloss.JoinVertical(lipgloss.Left, titleText, m.Styles.TabsRow.Width(m.Width-4).Render("  "+progressText))
}

func (m MainModel) footerView() string {
	var help string
	switch m.State {
	case SearchView:
		help = "enter Search Query  |  ctrl+c Quit"
	case PlaybackActiveView:
		help = "Stream active  |  Close player window to return to episodes"
	case DownloadingView:
		help = "Downloading file  |  esc Cancel download and return"
	default:
		help = "j/k Scroll  |  enter Select Option  |  esc Back to previous step"
	}
	return m.Styles.FooterStyle.Width(m.Width - 4).Render(help)
}

func (m MainModel) searchView() string {
	panelWidth, panelHeight := m.getLayoutDimensions()

	searchBox := m.Styles.ActivePanel.Width(panelWidth).Height(3).Render(
		"Enter Anime Search Query: " + m.SearchInput.View(),
	)

	return lipgloss.JoinVertical(lipgloss.Left, "\n", searchBox, "\n", m.Styles.InactivePanel.Width(panelWidth).Height(panelHeight-6).Render("\n\n   Use the search bar above to fetch indices.\n   Supports multi-source scrapers and subtitles filters."))
}

func (m MainModel) selectAnimeView() string {
	panelWidth, panelHeight := m.getLayoutDimensions()
	halfWidth := (panelWidth / 2) - 1

	leftContent := m.Styles.ActivePanel.Width(halfWidth).Height(panelHeight).Render(m.SearchList.View())
	
	var rightContent string
	if m.SearchList.SelectedItem() != nil {
		anime := m.SearchList.SelectedItem().(AnimeItem).Anime
		
		var imgPreview string
		if m.CoverImagePath != "" {
			imgPreview = cache.RenderImage(m.CoverImagePath, halfWidth-4, panelHeight-10)
		} else {
			imgPreview = "\n\n        [ Image Loading... ]"
		}

		rightContent = m.Styles.InactivePanel.Width(halfWidth).Height(panelHeight).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				m.Styles.TitleStyle.Render(anime.Title),
				"\n",
				fmt.Sprintf("%s %s", m.Styles.MetadataLabel.Render("Year:"), m.Styles.MetadataValue.Render(fmt.Sprintf("%d", anime.ReleaseYear))),
				fmt.Sprintf("%s %s", m.Styles.MetadataLabel.Render("Format:"), m.Styles.MetadataValue.Render(anime.Format)),
				"\n",
				imgPreview,
			),
		)
	} else {
		rightContent = m.Styles.InactivePanel.Width(halfWidth).Height(panelHeight).Render("\n\n   No anime selected.")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
}

func (m MainModel) selectEpisodeView() string {
	panelWidth, panelHeight := m.getLayoutDimensions()
	halfWidth := (panelWidth / 2) - 1

	var leftContent string
	if m.CoverImagePath != "" {
		leftContent = m.Styles.InactivePanel.Width(halfWidth).Height(panelHeight).Render(
			cache.RenderImage(m.CoverImagePath, halfWidth-4, panelHeight-4),
		)
	} else {
		leftContent = m.Styles.InactivePanel.Width(halfWidth).Height(panelHeight).Render("\n\n\n         [ Artwork Cover ]")
	}

	m.EpisodeList.SetSize(halfWidth-4, panelHeight-14)
	
	detailsHeader := lipgloss.JoinVertical(lipgloss.Left,
		m.Styles.TitleStyle.Render(m.SelectedAnime.Title),
		fmt.Sprintf("Provider: %s", m.SelectedProvider),
		fmt.Sprintf("Audio:    %s", m.SelectedDubSub),
		fmt.Sprintf("Action:   %s", m.SelectedAction),
		"\n------------------------------------------------------\n",
	)

	rightContent := m.Styles.ActivePanel.Width(halfWidth).Height(panelHeight).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			detailsHeader,
			m.EpisodeList.View(),
		),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
}

func (m MainModel) playbackView() string {
	panelWidth, panelHeight := m.getLayoutDimensions()

	progressBar := "..."
	percent := 0.0
	if m.PlaybackTick.DurationSec > 0 {
		percent = float64(m.PlaybackTick.ElapsedSec) / float64(m.PlaybackTick.DurationSec)
		barWidth := panelWidth - 40
		if barWidth < 10 {
			barWidth = 10
		}
		
		filledLen := int(percent * float64(barWidth))
		emptyLen := barWidth - filledLen

		filledBar := m.Styles.ProgressFull.Render(strings.Repeat("=", filledLen))
		emptyBar := m.Styles.ProgressEmpty.Render(strings.Repeat("-", emptyLen))
		progressBar = fmt.Sprintf("[%s>%s] %.0f%% (%d/%d sec)", filledBar, emptyBar, percent*100, m.PlaybackTick.ElapsedSec, m.PlaybackTick.DurationSec)
	}

	var statusText string
	if m.PlaybackTick.Paused {
		statusText = m.Styles.CursorStyle.Render("PAUSED")
	} else {
		statusText = m.Styles.TitleStyle.Render("PLAYING")
	}

	playbackDetail := fmt.Sprintf(
		"\n\n\n\n               Active Stream Integration Live\n\n"+
			"               Show:    %s\n"+
			"               Episode: %s\n"+
			"               Player:  %s\n"+
			"               Status:  %s\n\n\n"+
			"               %s\n\n\n\n",
		m.SelectedAnime.Title, m.SelectedEpisode.Title, m.SelectedPlayer, statusText, progressBar,
	)

	return m.Styles.ActivePanel.Width(panelWidth).Height(panelHeight).Render(playbackDetail)
}

func (m MainModel) downloadingView() string {
	panelWidth, panelHeight := m.getLayoutDimensions()

	barWidth := panelWidth - 40
	if barWidth < 10 {
		barWidth = 10
	}
	
	filledLen := int(m.DownloadPercent * float64(barWidth))
	emptyLen := barWidth - filledLen

	filledBar := m.Styles.ProgressFull.Render(strings.Repeat("=", filledLen))
	emptyBar := m.Styles.ProgressEmpty.Render(strings.Repeat("-", emptyLen))
	
	progressBar := fmt.Sprintf("[%s>%s] %.0f%%", filledBar, emptyBar, m.DownloadPercent*100)
	
	sizeProgress := "Calculating..."
	if m.DownloadTotal > 0 {
		sizeProgress = fmt.Sprintf("%.2f MB / %.2f MB", float64(m.DownloadBytes)/(1024*1024), float64(m.DownloadTotal)/(1024*1024))
	} else if m.DownloadBytes > 0 {
		sizeProgress = fmt.Sprintf("%.2f MB downloaded", float64(m.DownloadBytes)/(1024*1024))
	}

	downloadDetail := fmt.Sprintf(
		"\n\n\n\n               File Download Progress Live\n\n"+
			"               Show:    %s\n"+
			"               Episode: %s\n"+
			"               Quality: %s\n"+
			"               Speed:   %s\n"+
			"               Status:  %s\n\n\n"+
			"               %s\n\n\n\n",
		m.SelectedAnime.Title, m.SelectedEpisode.Title, m.SelectedQuality, m.Styles.CursorStyle.Render(m.DownloadSpeed), m.Styles.TitleStyle.Render(sizeProgress), progressBar,
	)

	return m.Styles.ActivePanel.Width(panelWidth).Height(panelHeight).Render(downloadDetail)
}

func renderSelectionWithSidebar(m MainModel, sel SelectionModel, width, height int) string {
	halfWidth := (width / 2) - 1

	// Left: selection items
	var sb strings.Builder
	sb.WriteString("\n  ")
	sb.WriteString(m.Styles.TitleStyle.Render(sel.Title))
	sb.WriteString("\n\n")

	for i, choice := range sel.Choices {
		if i == sel.Selected {
			sb.WriteString("  ")
			sb.WriteString(m.Styles.CursorStyle.Render("▶ "))
			sb.WriteString(m.Styles.SelectedText.Render("[x] " + choice))
			sb.WriteString("\n")
		} else {
			sb.WriteString("    ")
			sb.WriteString(m.Styles.NormalText.Render("[ ] " + choice))
			sb.WriteString("\n")
		}
	}
	leftContent := m.Styles.ActivePanel.Width(halfWidth).Height(height).Render(sb.String())

	// Right: Anime details & artwork preview
	var imgPreview string
	if m.CoverImagePath != "" {
		imgPreview = cache.RenderImage(m.CoverImagePath, halfWidth-4, height-10)
	} else {
		imgPreview = "\n\n        [ Image Loading... ]"
	}

	rightContent := m.Styles.InactivePanel.Width(halfWidth).Height(height).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.Styles.TitleStyle.Render(m.SelectedAnime.Title),
			"\n",
			fmt.Sprintf("%s %s", m.Styles.MetadataLabel.Render("Year:"), m.Styles.MetadataValue.Render(fmt.Sprintf("%d", m.SelectedAnime.ReleaseYear))),
			fmt.Sprintf("%s %s", m.Styles.MetadataLabel.Render("Format:"), m.Styles.MetadataValue.Render(m.SelectedAnime.Format)),
			"\n",
			imgPreview,
		),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftContent, " ", rightContent)
}

