### Phase 1: Barebones Skeleton & API Foundations (Days 1–3)

* Initialize modules (`go mod init animehub`). Load basic runtime UI frameworks (`bubbletea`, `lipgloss`).
* Establish data models matching the AniList GraphQL layout schema.
* Assemble standard HTTP GET routing profiles pointing to the provider engine framework.

### Phase 2: Interactive TUI Construction (Days 4–7)

* Define core state representations: `SearchState`, `ResultsState`, and `EpisodeState`.
* Connect keybindings to handle lists, text input boxes, and options selectors.

### Phase 3: Playback Integration & Core Loop (Days 8–10)

* Implement execution calls passing stream URLs to the native playback application.
* Inject state protection boundaries using the framework's runtime execution methods to isolate the parent terminal's graphical layout buffers.

```go
package main

import (
	"os/exec"
	"github.com/charmbracelet/bubbletea"
)

type SessionState int
const (
	SearchState SessionState = iota
	ResultsState
	EpisodeSelectState
)

type Model struct {
	state           SessionState
	provider        AnimeProvider
	selectedEpisode Episode
	selectedLang    string
}

type playbackFinishedMsg struct{ err error }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" && m.state == EpisodeSelectState {
			// Resolve URL immediately prior to invocation to avoid early token death
			url, err := m.provider.GetStreamURL(m.selectedEpisode.ID, m.selectedLang)
			if err != nil {
				return m, func() tea.Msg { return playbackFinishedMsg{err: err} }
			}

			// --no-terminal isolates standard output descriptors from polluting terminal UI grids
			cmd := exec.Command("mpv", "--no-terminal", url)

			// Safely pause the TUI loop, pass control to mpv, and restore screen buffers on close
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return playbackFinishedMsg{err: err}
			})
		}
	}
	return m, nil
}

```

### Phase 4: Error Handling & Fallback Routine Hardening (Days 11–14)

* Integrate error-catching mechanics inside the playback response channel (`playbackFinishedMsg`) to handle sudden execution errors.
* Build automated lookup verification loops (`exec.LookPath`) to confirm that native platform player binaries are present before allowing execution streams.
