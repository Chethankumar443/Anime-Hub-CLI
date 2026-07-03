package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	// Palette Colors
	BgColor       lipgloss.TerminalColor
	FgColor       lipgloss.TerminalColor
	AccentColor   lipgloss.TerminalColor
	MutedColor    lipgloss.TerminalColor
	WarningColor  lipgloss.TerminalColor
	BorderColor   lipgloss.TerminalColor

	// Layout Styles
	DocStyle      lipgloss.Style
	HeaderStyle   lipgloss.Style
	FooterStyle   lipgloss.Style
	
	// Panel Grid
	ActivePanel   lipgloss.Style
	InactivePanel lipgloss.Style
	SidebarPanel  lipgloss.Style

	// Tabs Layout
	ActiveTab     lipgloss.Style
	InactiveTab   lipgloss.Style
	TabsRow       lipgloss.Style

	// List & Content Styles
	TitleStyle    lipgloss.Style
	CursorStyle   lipgloss.Style
	SelectedText  lipgloss.Style
	NormalText    lipgloss.Style
	ProgressFull  lipgloss.Style
	ProgressEmpty lipgloss.Style
	MetadataLabel lipgloss.Style
	MetadataValue lipgloss.Style
}

func DefaultStyles() *Styles {
	s := &Styles{
		BgColor:      lipgloss.Color("#09090B"), // Zinc 950
		FgColor:      lipgloss.Color("#F4F4F5"), // Zinc 100
		AccentColor:  lipgloss.Color("#10B981"), // Emerald 500
		MutedColor:   lipgloss.Color("#71717A"), // Zinc 500
		WarningColor: lipgloss.Color("#F59E0B"), // Amber 500
		BorderColor:  lipgloss.Color("#27272A"), // Zinc 800
	}

	s.DocStyle = lipgloss.NewStyle().
		Background(s.BgColor).
		Foreground(s.FgColor)

	s.HeaderStyle = lipgloss.NewStyle().
		Foreground(s.AccentColor).
		Bold(true).
		Padding(1, 2)

	s.FooterStyle = lipgloss.NewStyle().
		Foreground(s.MutedColor).
		Padding(0, 2).
		Height(1)

	s.ActivePanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.AccentColor).
		Padding(1, 2)

	s.InactivePanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.BorderColor).
		Padding(1, 2)

	s.SidebarPanel = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(s.BorderColor).
		Padding(0, 2)

	s.ActiveTab = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(s.AccentColor).
		Bold(true).
		Padding(0, 2).
		MarginRight(2)

	s.InactiveTab = lipgloss.NewStyle().
		Foreground(s.MutedColor).
		Background(lipgloss.Color("#18181B")). // Zinc 900
		Padding(0, 2).
		MarginRight(2)

	s.TabsRow = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(s.BorderColor).
		Padding(0, 2).
		MarginBottom(1)

	s.TitleStyle = lipgloss.NewStyle().
		Foreground(s.AccentColor).
		Bold(true)

	s.CursorStyle = lipgloss.NewStyle().
		Foreground(s.WarningColor).
		Bold(true)

	s.SelectedText = lipgloss.NewStyle().
		Foreground(s.FgColor).
		Bold(true)

	s.NormalText = lipgloss.NewStyle().
		Foreground(s.MutedColor)

	s.ProgressFull = lipgloss.NewStyle().
		Foreground(s.AccentColor)

	s.ProgressEmpty = lipgloss.NewStyle().
		Foreground(s.MutedColor)

	s.MetadataLabel = lipgloss.NewStyle().
		Foreground(s.MutedColor).
		Bold(true).
		Width(12)

	s.MetadataValue = lipgloss.NewStyle().
		Foreground(s.FgColor)

	return s
}
