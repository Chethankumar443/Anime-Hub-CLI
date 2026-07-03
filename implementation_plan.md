# Implementation Plan & Milestones
## Project: TerminalAnime CLI (`anime-cli`)

---

### Phase 1: Project Foundations & Structure
Set up the initial Go application structures, file system persistence drivers, and verification frameworks.

#### Phase 1 Tasks
1.  **Initialize Project Modules:** Run `go mod init github.com/cheth/anime-cli` and download standard dependencies:
    ```bash
    go get github.com/charmbracelet/bubbletea@v0.25.0
    go get github.com/charmbracelet/lipgloss@v0.9.0
    go get github.com/charmbracelet/bubbles@v0.17.0
    go get github.com/mattn/go-runewidth@v0.0.15
    ```
2.  **Define Directory Structure:** Create the following directory layout:
    ```bash
    mkdir -p cmd/anime-cli pkg/config pkg/player pkg/provider pkg/cache pkg/tui
    ```
3.  **Implement JSON Persistence Engine:**
    *   Create [config.go](file:///c:/Users/cheth/Desktop/TerminalAnime/pkg/config/config.go) to read and write configs. Include OS-specific path parsing (using `os.UserConfigDir` and `os.UserHomeDir`).
    *   Create [history.go](file:///c:/Users/cheth/Desktop/TerminalAnime/pkg/config/history.go) to manage watchlist entries and atomic file writes.

#### Phase 1 Verification & Testing
*   **Unit Tests:** Run tests checking file locking and configuration recovery on invalid JSON formats:
    ```bash
    go test -v ./pkg/config/...
    ```
*   **Expected Output:** Tests must verify that a corrupt `config.json` is backed up to `config.json.bak`, and a fresh, correct default configuration is written.

#### Phase 1 Deliverables
*   `go.mod` and `go.sum` containing exact dependency versions.
*   `pkg/config/config.go` containing type declarations and file management logic.
*   `pkg/config/history.go` containing watchlist structs.

---

### Phase 2: Modular Provider Interface & Network Ingestion
Construct the scraper interface and develop the decryption engine to unpack protected video streams.

#### Phase 2 Tasks
1.  **Define Provider Interface Spec:** Create [provider.go](file:///c:/Users/cheth/Desktop/TerminalAnime/pkg/provider/provider.go) defining search, details query, and stream extraction hooks.
2.  **HTTP Client Wrapper:** Write an HTTP helper that sets random User-Agents, sets appropriate request headers, and handles HTTP connection retries.
3.  **Implement AES Decryption Engine:** Create `pkg/provider/gogo/decrypt.go` to decrypt stream URLs using AES-128-CBC.

#### Phase 2 Verification & Testing
*   **Integration Tests:** Verify scraper parsers compile and extract correct streaming links from mock pages:
    ```bash
    go test -v ./pkg/provider/...
    ```
*   **Expected Output:** Extracting a video link must return valid, non-expired M3U8 links with correct stream header tokens.

#### Phase 2 Deliverables
*   `pkg/provider/provider.go` defining scraper interfaces.
*   `pkg/provider/gogo/parser.go` parsing Gogoanime elements.
*   `pkg/provider/gogo/decrypt.go` handling AES decryption.

---

### Phase 3: Player Lifecycle & IPC Control Loop
Establish external process managers to handle child processes and socket threads.

#### Phase 3 Tasks
1.  **Socket Handshake Setup:** Implement connection loops in `pkg/player/ipc.go` supporting Unix sockets and Windows named pipes.
2.  **Process Manager Logic:** Implement process creation routines in `pkg/player/manager.go`. Run command parameter wrappers to launch `mpv` or `vlc` with IPC flags enabled.
3.  **Goroutine Monitor:** Create background listener threads that poll elapsed playback times every second and push progress structs to the TUI channel.

#### Phase 3 Verification & Testing
*   **IPC Loop Tests:** Run a test script to spawn MPV, connect to socket, send a query command, and verify return structure:
    ```bash
    go test -v ./pkg/player/...
    ```
*   **Expected Output:** Handshake dials must connect successfully in under 250ms, and the read thread must capture frame ticks cleanly.

#### Phase 3 Deliverables
*   `pkg/player/ipc.go` managing socket handshakes and reads.
*   `pkg/player/manager.go` handling external child process execution.

---

### Phase 4: Bubbletea TUI Engine & Styling
Integrate views, key bindings, and layout elements.

#### Phase 4 Tasks
1.  **Structure Main Model:** Define states, focuses, and child components in `pkg/tui/model.go`.
2.  **Develop View Layouts:** Write border containers and custom progress indicators in `pkg/tui/view.go`.
3.  **Implement Event Routers:** Handle keyboard inputs, list updates, and IPC channel messages in `pkg/tui/update.go`.
4.  **Genre Themes:** Define Lipgloss color palettes in `pkg/tui/styles.go`.

#### Phase 4 Verification & Testing
*   **TUI Render Audit:** Run interactive developer builds to test scaling and panel wrapping:
    ```bash
    go run ./cmd/anime-cli/main.go
    ```
*   **Expected Output:** Resizing the window must not wrap headers or break panel borders. Columns must adjust width proportionally.

#### Phase 4 Deliverables
*   `pkg/tui/model.go` containing state models.
*   `pkg/tui/update.go` coordinating input loops.
*   `pkg/tui/view.go` formatting terminal views.
*   `pkg/tui/styles.go` declaring genre-specific colors.

---

### Phase 5: Image Caching & Graphic Previews
Implement capability checks and rasterizers for graphic-supported terminals.

#### Phase 5 Tasks
1.  **Download Engine:** Create async image downloading helpers in `pkg/cache/download.go`.
2.  **Eviction Engine:** Code size-aware LRU cache cleanup routines in `pkg/cache/eviction.go`.
3.  **Capability Check & Renderer:** Implement Kitty, Sixel, and iTerm2 adapters, and write ANSI Half-block fallbacks in `pkg/cache/render.go`.

#### Phase 5 Verification & Testing
*   **Cache Limits Tests:** Verify eviction functions delete oldest items when folder exceeds 250MB:
    ```bash
    go test -v ./pkg/cache/...
    ```
*   **Expected Output:** Cleanup scripts must clear cache folder down to 180MB without leaving corrupted images.

#### Phase 5 Deliverables
*   `pkg/cache/download.go` handling file caching.
*   `pkg/cache/eviction.go` limiting folder size.
*   `pkg/cache/render.go` managing graphics conversion protocols.

---

### Phase 6: Syncplay & Release Optimization
Sync playback coordinates between users and compile release builds.

#### Phase 6 Tasks
1.  **WebSocket Client:** Write synchronization logic in `pkg/syncplay/client.go` to exchange coordinates with a Syncplay server.
2.  **Resource Audit:** Profile memory and CPU utilization to verify constraints are met.
3.  **Release Build configurations:** Set up compilation matrices using GoReleaser.

#### Phase 6 Verification & Testing
*   **Syncplay Tests:** Verify WebSocket client connects to mock servers and formats message properties:
    ```bash
    go test -v ./pkg/syncplay/...
    ```
*   **Expected Output:** Command frames (Play, Pause, Seek) must sync between clients with less than 50ms latency.

#### Phase 6 Deliverables
*   `pkg/syncplay/client.go` managing WebSocket traffic.
*   `.goreleaser.yaml` containing compilation matrices.

---

### File Cross-References
*   Product requirements and user stories: See [prd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/prd.md).
*   Underlying system loops and socket connections: See [trd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/trd.md).
*   Visual wireframes and navigation keybindings: See [navigation.md](file:///c:/Users/cheth/Desktop/TerminalAnime/navigation.md).
*   Domain structs and JSON specifications: See [techstack.md](file:///c:/Users/cheth/Desktop/TerminalAnime/techstack.md).
*   CI/CD pipelines and deployment scripts: See [deployment_plan.md](file:///c:/Users/cheth/Desktop/TerminalAnime/deployment_plan.md).
