# Product Requirements Document (PRD)
## Project: TerminalAnime CLI (`anime-cli`)

---

### 1. Introduction & Vision

`anime-cli` is a high-fidelity, terminal-native anime streaming client written in Go. The application is designed to bypass the resource-heavy overhead of modern web browsers, strip out advertisements, and provide a keyboard-driven viewing experience directly within a developer's terminal. 

The core philosophy of `anime-cli` is **absolute visual craftsmanship and zero layout jitter**. Instead of using generic, unstyled terminal library controls, the interface is treated as a premium canvas where layouts are computed dynamically down to individual terminal character cells.

#### 1.1. Core Objectives
*   **Zero-Jitter Interface:** The interface must scale smoothly when terminal windows are resized. Standard UI library components often break layout constraints or wrap text unpredictably. `anime-cli` uses explicit character-cell boundary mathematics to truncate, wrap, or collapse panels gracefully.
*   **"Ambient Horizon" Experience:** A context-aware state engine that adapts UI coloring and spatial focus depending on what the user is inspecting. The TUI changes colors depending on the genre of the active anime (e.g., Crimson for Action, Electric Cyan for Sci-Fi, Rose for Romance) and shifts panels dynamically to spotlight active work zones.
*   **Integrated Process Pipeline:** A robust media socket control layer that launches external media engines (`mpv` or `vlc`) as child processes, monitors playback progress via local IPC sockets, and updates history caches atomically without blocking the main event thread.
*   **Ultra-lightweight Portability:** Compiled into a single static Go binary with no database engine (SQLite, Redis) dependencies. State management and caching use native Go `encoding/json` serialization.
*   **High-Fidelity Graphic Previews:** A background graphics manager capable of checking terminal protocols (Sixel, Kitty Graphic Protocol, iTerm2 Inline Images, or ANSI Half-blocks) and drawing crisp, cached poster art directly inside the TUI panels.

---

### 2. Target Audience & Personas

*   **The Terminal Purist:** A developer or systems administrator who operates entirely within a terminal environment (using tools like `tmux`, `neovim`, and tiling window managers like `i3` or `sway`). They require lightning-fast startup times (<150ms), a minimal memory footprint (<30MB idle), and keyboard-only, vim-compatible controls.
*   **The Desktop Anime Fan:** A viewer looking to stream anime episodes without launching a web browser that consumes gigabytes of memory. They require high-quality video playback, automatic episode transition tracking, sub/dub selection, and a clean "Resume Watching" history loop.
*   **The Co-Viewer (Syncplay):** Groups of terminal users who want to watch anime episodes concurrently. They require low-latency synchronization of playback states (play, pause, seek) with peers.

---

### 3. Core Feature Matrix

| Feature ID | Feature Group | Description | Priority |
|---|---|---|---|
| **FEAT-001** | **UI/UX Foundation** | Charm Bubble Tea MVU framework styled exclusively with explicit Lip Gloss layouts, utilizing custom micro-borders and absolute layouts. | Must Have |
| **FEAT-002** | **Ambient Horizon** | State engine mapping media genres to terminal ANSI color schemes and shifting panel sizes dynamically according to focus states. | Must Have |
| **FEAT-003** | **Text Wrapping & Math** | A mathematical truncation and layout-clipping algorithm that measures cells in real-time, preventing layout breakage on resize. | Must Have |
| **FEAT-004** | **Process Sync Engine** | JSON-IPC connection manager for spawning `mpv`/`vlc` and syncing playback parameters over Unix domain sockets or Windows named pipes. | Must Have |
| **FEAT-005** | **Zero-DB Persistence** | Struct persistence written directly to disk via Go's native JSON engine for configuration, history tracking, and watchlist status. | Must Have |
| **FEAT-006** | **High-Fi Cover Art** | Rendering pipeline checking for Kitty, iTerm2, Sixel, or ANSI Half-block capabilities, outputting cached files locally. | Must Have |
| **FEAT-007** | **Syncplay Client** | Real-time playback synchronization with remote rooms over persistent WebSocket connections. | Should Have |

---

### 4. Detailed Functional Specifications

#### 4.1. The UI/UX Foundation & Border Mechanics
The interface is divided into a three-panel grid system that eliminates generic shell component styling.
*   **Rounded Micro-Borders:** Lip Gloss styles must use custom rounded border runes:
    *   Top-Left: `╭`
    *   Top-Right: `╮`
    *   Bottom-Right: `╯`
    *   Bottom-Left: `╰`
    *   Horizontal: `─`
    *   Vertical: `│`
*   **Explicit Sizing Calculations:** The application intercepts the window size message (`tea.WindowSizeMsg`) and calculates the exact integer width and height for every border container. Under no circumstances should size calculations result in decimal values, which causes double-line rendering bugs and container overflow.

#### 4.2. The "Ambient Horizon" Focus Engine
The interface changes color schemes depending on the genre of the selected anime. When a show details panel is focused:
*   **Action / Shonen:** Dynamic shift to Crimson Red (`#EF4444`) and Charcoal Grey (`#1F2937`).
*   **Sci-Fi / Mecha:** Dynamic shift to Electric Cyan (`#06B6D4`) and Deep Blue-Grey (`#0F172A`).
*   **Romance / Drama:** Dynamic shift to Soft Rose (`#EC4899`) and Cocoa Brown (`#2D1F21`).
*   **Fantasy / Isekai:** Dynamic shift to Forest Emerald (`#10B981`) and Deep Olive (`#064E3B`).
*   **Slice of Life / Comedy:** Dynamic shift to Warm Amber (`#F59E0B`) and Charcoal (`#1E293B`).
*   **Default State:** Zinc Grey (`#71717A`) and Matte Black (`#18181B`).

#### 4.3. Text-Wrapping and Truncation Math
To prevent visual line wrapping from pushing panels off-screen:
*   Every string displayed in a list item or detail panel must be processed through an explicit string width measurement function (handling double-width Unicode characters for East Asian scripts).
*   If a string width exceeds `AvailableWidth - BorderPadding`, it is truncated at `AvailableWidth - BorderPadding - 3` and suffixed with `...`.
*   Synopsis text blocks are wrapped using a word-wrap algorithm that splits text on whitespace boundaries and fits text lines into the computed height of the details container. If text lines exceed container height, the view adds an indicator scrollbar (`█` and `░`) and ignores overflow lines.

#### 4.4. Media Control Process Pipeline
The system spawns a background thread immediately upon starting playback.
*   **Launch Parameters:** Launches `mpv` with specialized parameters:
    `mpv --input-ipc-server=<SocketPath> --no-terminal --ontop --title="anime-cli: <Episode Title>"`
*   **JSON-IPC Daemon:** A background goroutine reads and writes to the socket stream. It queries `time-pos` (current playback second) and `duration` (total video duration) every 1 second.
*   **Progress Updating:** When the playback progress reaches the **90% threshold**, the episode is automatically marked as `Completed` in the history state. The TUI will show the progress bar as fully completed.
*   **Resume State:** If the player is closed before the 90% threshold is reached, the exact exit timestamp is stored. On subsequent TUI loads, the episode listing shows a visual indicator `[Resume at MM:SS]`.

#### 4.5. High-Fidelity Asset Pipeline
Images must be processed off the main thread to avoid introduction of user input lag.
*   **Graphics Capability Check:** On startup, the cache manager executes probe checks:
    1. Sends Kitty query escape sequence. If a response is received via stdin within 15ms, sets renderer mode to `Kitty`.
    2. Checks `TERM_PROGRAM` for `iTerm.app`. If positive, sets mode to `iTerm2`.
    3. Queries Sixel device attributes. If terminal replies with Sixel code support, sets mode to `Sixel`.
    4. Fallback: Sets mode to `HalfBlock` (ANSI Truecolor block characters).
*   **LRU Disk Cache:** Images downloaded from API providers are saved to `~/.cache/anime-cli/images/`.
*   **Cache Eviction Rule:** When the image folder exceeds 250 Megabytes, the background engine sorts all images in the folder by their access timestamp (read from a localized history file `~/.cache/anime-cli/access.json`) and deletes the oldest entries until the folder size drops below 180 Megabytes.

---

### 5. Non-Functional Requirements (NFR)

*   **Startup Speed:** Cold startup (invoking binary to initial TUI paint) must be completed in under 150 milliseconds on standard test machines (x87_64 or Apple Silicon, SSD storage).
*   **Resource Utilization:**
    *   Idle Memory (RSS): Under 30 Megabytes.
    *   Active Playback Monitoring: Under 50 Megabytes.
    *   CPU Idle Usage: Less than 1% utilization.
*   **Cross-Platform Portability:**
    *   Single static binary compilation target.
    *   Windows builds must utilize Named Pipes (`\\.\pipe\anime-cli-playback-{PID}`) rather than Unix sockets, and support Windows Command Prompt and PowerShell colors.
    *   Linux and macOS builds must support Unix sockets under `/tmp/` or `/var/run/`.
*   **Robust State Recovery:** The application must write files atomically (writing first to `.tmp` files and executing rename calls). If the configuration file is corrupted (e.g. invalid JSON), the application must automatically write the default configuration, preserve the backup, and log the event.

---

### 6. File Cross-References
*   Detailed technical architecture, data flows, and state models: See [trd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/trd.md).
*   TUI navigation layouts, panels, and keyboard shortcuts: See [navigation.md](file:///c:/Users/cheth/Desktop/TerminalAnime/navigation.md).
*   Step-by-step development phases and milestones: See [implementation_plan.md](file:///c:/Users/cheth/Desktop/TerminalAnime/implementation_plan.md).
*   API structures, JSON files schemas, and directory layout: See [techstack.md](file:///c:/Users/cheth/Desktop/TerminalAnime/techstack.md).
*   CI/CD pipelines, package manager specs, and release scripts: See [deployment_plan.md](file:///c:/Users/cheth/Desktop/TerminalAnime/deployment_plan.md).
