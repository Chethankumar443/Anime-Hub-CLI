# Technology Stack & Domain Models
## Project: TerminalAnime CLI (`anime-cli`)

---

### 1. Technology Selection

#### 1.1. Core Runtime & TUI Engine
*   **Language:** Go (Golang) v1.22+ (supporting structured logging via `slog` and native slice helpers).
*   **TUI MVU Framework:** `github.com/charmbracelet/bubbletea` v0.25.0 (managing event routing, input polling, and lifecycle rendering).
*   **Styling Engine:** `github.com/charmbracelet/lipgloss` v0.9.0 (providing border runic styles, text padding, and layout color attributes).
*   **UI Components:** `github.com/charmbracelet/bubbles` v0.17.0 (providing list widgets, textinputs, and loading spinners).

#### 1.2. Supplementary Libraries
*   **Text Width Calculations:** `github.com/mattn/go-runewidth` v0.0.15 (essential for calculating correct cell lengths of multi-byte CJK titles).
*   **Real-time Co-Viewing:** `github.com/gorilla/websocket` v1.5.0 (managing bi-directional socket connections with low-latency Syncplay servers).
*   **HTML Scraping Parsers:** `golang.org/x/net/html` v0.20.0 (parsing DOM structures returned from anime web streams).

---

### 2. Domain Models & Struct Schemas

The application defines these unified data models in `pkg/provider/provider.go` and `pkg/config/config.go` to share state across scraper interfaces, configuration files, and views.

```go
package models

import "time"

// Anime represents a summary listing of a show returned from search or featured listings.
type Anime struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Format      string `json:"format"`       // TV, Movie, OVA, Special
	ReleaseYear int    `json:"release_year"`
	CoverURL    string `json:"cover_url"`
	ProviderID  string `json:"provider_id"`
}

// AnimeDetails contains expanded metadata and lists of episodes.
type AnimeDetails struct {
	Anime
	Synopsis string    `json:"synopsis"`
	Rating   float64   `json:"rating"`
	Status   string    `json:"status"` // Ongoing, Completed
	Genres   []string  `json:"genres"`
	Episodes []Episode `json:"episodes"`
}

// Episode represents an individual video entry.
type Episode struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	DurationSec int    `json:"duration_sec"`
}

// StreamLink represents resolved CDN stream endpoints.
type StreamLink struct {
	URL         string            `json:"url"`
	Quality     string            `json:"quality"` // 1080p, 720p, 360p, auto
	IsM3U8      bool              `json:"is_m3u8"`
	HTTPHeaders map[string]string `json:"http_headers"` // Spoofed client headers
}

// PlaybackProgress tracks elapsed watch time.
type PlaybackProgress struct {
	AnimeID     string    `json:"anime_id"`
	EpisodeID   string    `json:"episode_id"`
	EpisodeNum  int       `json:"episode_num"`
	ElapsedSec  int       `json:"elapsed_sec"`
	DurationSec int       `json:"duration_sec"`
	LastUpdated time.Time `json:"last_updated"`
	Completed   bool      `json:"completed"`
}

// AppConfig represents global user configurations.
type AppConfig struct {
	Theme             string                    `json:"theme"`
	DefaultPlayer     string                    `json:"default_player"`      // mpv, vlc
	PreferredQuality  string                    `json:"preferred_quality"`    // 1080p, 720p, auto
	BufferSizeMB      int                       `json:"buffer_size_mb"`
	SyncplayServerURL string                    `json:"syncplay_server_url"`
	DefaultProvider   string                    `json:"default_provider"`    // gogoanime
	ProvidersConfig   map[string]ProviderParams `json:"providers_config"`
}

// ProviderParams stores endpoint configurations for scrapers.
type ProviderParams struct {
	BaseURL    string `json:"base_url"`
	TimeoutSec int    `json:"timeout_sec"`
}

// WatchHistory represents watch lists and progress history stored locally.
type WatchHistory struct {
	Watchlist []Anime            `json:"watchlist"`
	Progress  []PlaybackProgress `json:"progress"`
}
```

---

### 3. File System Persistence Format

Configurations and watch logs are saved on disk using Go's `encoding/json` standard encoder.

#### 3.1. Main Configuration File (`config.json`)
```json
{
  "theme": "dark-emerald",
  "default_player": "mpv",
  "preferred_quality": "1080p",
  "buffer_size_mb": 64,
  "syncplay_server_url": "wss://sync.animecli.dev",
  "default_provider": "gogoanime",
  "providers_config": {
    "gogoanime": {
      "base_url": "https://anitaku.pe",
      "timeout_sec": 10
    }
  }
}
```

#### 3.2. Watch History and Progress File (`history.json`)
```json
{
  "watchlist": [
    {
      "id": "frieren-beyond-journeys-end",
      "title": "Frieren: Beyond Journey's End",
      "format": "TV",
      "release_year": 2023,
      "cover_url": "https://images.animecli.dev/frieren.jpg",
      "provider_id": "gogoanime"
    }
  ],
  "progress": [
    {
      "anime_id": "frieren-beyond-journeys-end",
      "episode_id": "frieren-episode-1",
      "episode_num": 1,
      "elapsed_sec": 765,
      "duration_sec": 1440,
      "last_updated": "2026-07-03T09:50:00Z",
      "completed": false
    }
  ]
}
```

---

### 4. IPC Commands Payload Schemas

Communication with external players is performed via JSON-RPC. Every outgoing string payload must terminate with a newline character (`\n`).

#### 4.1. Retrieve Playback Coordinate (`time-pos`)
*   **Command Payload:**
    ```json
    {"command": ["get_property", "time-pos"], "request_id": 2001}
    ```
*   **Response Payload:**
    ```json
    {"data": 765.432, "error": "success", "request_id": 2001}
    ```

#### 4.2. Request Seek Execution
*   **Command Payload:**
    ```json
    {"command": ["seek", 300, "absolute"], "request_id": 2002}
    ```
*   **Response Payload:**
    ```json
    {"error": "success", "request_id": 2002}
    ```

#### 4.3. Trigger Video Pause State
*   **Command Payload:**
    ```json
    {"command": ["set_property", "pause", true], "request_id": 2003}
    ```
*   **Response Payload:**
    ```json
    {"error": "success", "request_id": 2003}
    ```

---

### 5. Standard Project Directory Structure

```
/anime-cli
├── cmd/
│   └── anime-cli/
│       └── main.go         # Entry point, initializes components and TUI loop
├── pkg/
│   ├── config/
│   │   ├── config.go       # LoadConfig, SaveConfig, default JSON configurations
│   │   └── history.go      # LoadHistory, SaveHistory, watchlist state management
│   ├── player/
│   │   ├── ipc.go          # Unix socket / Windows pipe dialers and monitoring thread
│   │   └── manager.go      # Spawns MPV child process with IPC parameters
│   ├── provider/
│   │   ├── provider.go     # Scraper interfaces and data structures
│   │   └── gogo/
│   │       ├── parser.go   # Scrapes search pages and episode indexes
│   │       └── decrypt.go  # Decrypts stream keys using AES-128-CBC
│   ├── cache/
│   │   ├── download.go     # Asynchronous file download engine
│   │   ├── render.go       # Translates images to Sixel, Kitty, or Half-block protocols
│   │   └── eviction.go     # Size-aware LRU cache deletion routine
│   ├── syncplay/
│   │   └── client.go       # WebSocket client for coordinating video events
│   └── tui/
│       ├── model.go        # Declares Charm models and sub-components
│       ├── update.go       # Vim keybinding event routers and state transitions
│       ├── view.go         # Formats UI panels and rounded borders
│       └── styles.go       # Lipgloss color definitions and theme styles
├── go.mod                  # Application modular dependencies configurations
└── go.sum                  # Package checksum hashes
```

---

### 6. File Cross-References
*   Product requirements and user stories: See [prd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/prd.md).
*   Underlying system loops and socket connections: See [trd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/trd.md).
*   Visual wireframes and navigation keybindings: See [navigation.md](file:///c:/Users/cheth/Desktop/TerminalAnime/navigation.md).
*   Phased implementation checklist: See [implementation_plan.md](file:///c:/Users/cheth/Desktop/TerminalAnime/implementation_plan.md).
*   CI/CD pipelines and deployment scripts: See [deployment_plan.md](file:///c:/Users/cheth/Desktop/TerminalAnime/deployment_plan.md).
