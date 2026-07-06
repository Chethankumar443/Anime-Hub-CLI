**V1 Tech Stack:**
*   **Runtime:** Go 1.21.
*   **TUI:** `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles`.
*   **HTTP:** `go-resty/resty/v2`.
*   **APIs:** AniList GraphQL (Metadata), Consumet REST (Video).

**V2 Tech Stack Additions:**
*   **Terminal Graphics:** `eduardoejp/termimg` (for Kitty/Sixel protocol detection and rendering).
*   **Local DB:** `glebarez/sqlite` (Replace JSON history with a real database for concurrent read/writes).

**V3 Tech Stack Additions:**
*   **Auth:** `go-oauth2/oauth2` (For AniList token management).
*   **Cryptography:** `crypto/sha256` (For verifying auto-updated binaries).

**Hard Constraints (Apply to all versions):**
1.  **No Raw Scraping:** Never parse HTML in Go. If a provider requires it, use the Consumet Node.js API.
2.  **Stream Freshness:** Video URLs expire. Fetch the URL in the `Update` function immediately before calling `tea.ExecProcess`. Never cache the video URL.
