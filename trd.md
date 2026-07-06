### Language & Runtime Environment

* **Core:** Go 1.21+ (leveraging structured concurrency primitives and optimized standard library HTTP implementations).

### Core Dependencies

* `github.com/charmbracelet/bubbletea`: Declarative TUI framework based on the Elm Architecture.
* `github.com/charmbracelet/lipgloss`: Layout engine and terminal styling primitives.
* `github.com/go-resty/resty/v2`: Resilient HTTP client featuring built-in retries.
* `github.com/adrg/xdg`: Cross-platform compliance for standard user configurations.

### External APIs & Data Infrastructure

* **Metadata Layer:** AniList GraphQL API (`https://graphql.anilist.co`). Selected over Jikan for increased rate limits, speed, and standard structural mutations without mandatory API keys.
* **Source Aggregation Layer:** Consumet API instance. Handles complex JavaScript-based scraping, decryption routines, and source extraction into unified JSON formats.
* **Playback Core:** `mpv` Media Player. Chosen for out-of-the-box adaptive HLS/DASH streaming capabilities. Must be installed on the host platform and exposed via the system `$PATH`.

### Critical Performance Targets

* **Cold Boot Execution:** < 200ms to interactive TUI view state.
* **Data Aggregation Latency:** < 2 seconds for a standard indexed text query.
* **Runtime Memory Allocation:** < 50MB during idle state-machine execution.

### Hard Engineering Traps & Constraints

| Trap | Engineering Risk | Technical Mitigation Rule |
| --- | --- | --- |
| **Stream Token Expiration** | Video links expire within 15–60 minutes. Pre-fetching causes playback to fail mid-session. | **Never pre-fetch.** Resolve the final stream address *milliseconds* before spawning the player process. |
| **Upstream Rate Limiting** | AniList enforces a 90 requests/min ceiling. Jikan restricts to 3 requests/sec. | Implement internal request caching and exponential backoff wrappers inside the HTTP transport client. |
| **Terminal State Corruption** | Subprocesses hijacking `stdout` can break the terminal's alternative buffer screen state. | Suspend the parent Go TUI loop using explicit window-switching process boundaries during player handoff. |
