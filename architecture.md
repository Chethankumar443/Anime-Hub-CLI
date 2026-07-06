**The Embedded Node Binary Manager (Replacing Docker):**
Do not use Docker. Compile the Consumet Node.js app into a standalone binary using `pkg` or `nexe`, and manage it as a background process.

```go
type ConsumetManager struct {
    cmd  *exec.Cmd
    port string
}

func (cm *ConsumetManager) Start() error {
    if isPortInUse(cm.port) { return nil } // Already running
    
    cm.cmd = exec.Command(pathToConsumetBinary)
    cm.cmd.Env = append(os.Environ(), "PORT="+cm.port)
    
    // Hide console window on Windows
    if runtime.GOOS == "windows" {
        cm.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    }
    if err := cm.cmd.Start(); err != nil { return err }
    return cm.waitForReady()
}

func (cm *ConsumetManager) waitForReady() error {
    url := fmt.Sprintf("http://localhost:%s/health", cm.port)
    for i := 0; i < 30; i++ { // Poll for 15 seconds
        resp, err := http.Get(url)
        if err == nil && resp.StatusCode == 200 {
            resp.Body.Close()
            return nil
        }
        time.Sleep(500 * time.Millisecond)
    }
    return errors.New("Consumet failed to start")
}
```

**The Provider & Fallback Logic:**
```go
type FallbackManager struct { providers []AnimeProvider }

func (fm *FallbackManager) GetStreamURL(episodeID, lang string) (string, error) {
    for _, p := range fm.providers {
        url, err := p.GetStreamURL(episodeID, lang)
        if err == nil { return url, nil }
        log.Printf("Provider failed: %v, trying next...", err)
    }
    return "", errors.New("all providers failed")
}
```

**The TUI State Machine & MPV Handoff:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if msg, ok := msg.(tea.KeyMsg); ok {
        if msg.String() == "enter" && m.state == EpisodeSelectState {
            url, _ := m.provider.GetStreamURL(m.selectedEpisode.ID, m.selectedLang)
            cmd := exec.Command("mpv", "--no-terminal", url)
            // Safely pause TUI, run mpv, and resume
            return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
                return playbackFinishedMsg{err: err}
            })
        }
    }
    return m, nil
}
```
