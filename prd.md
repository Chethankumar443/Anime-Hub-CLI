**Product Vision:**
A blazing-fast, terminal-based anime hub built in Go that allows users to search, select, and stream anime directly from their command line, bypassing the need for a web browser.

**Version Roadmap (Practical Improvements):**
*   **Version 1 (The Working Prototype):** 
    *   *Goal:* It plays a video. 
    *   *Scope:* Text-only TUI. Single provider (Consumet via Embedded Node Binary). Basic Sub/Dub toggle. `mpv` handoff. Local JSON history.
*   **Version 2 (The Resilient Tool):**
    *   *Goal:* It doesn't break when things fail.
    *   *Scope:* Multi-provider fallback routing. Stream URL expiration handling (auto-refresh). Terminal image support (Kitty/Sixel) for cover art. SQLite for history.
*   **Version 3 (The Power User Tool):**
    *   *Goal:* It replaces the web browser entirely.
    *   *Scope:* AniList OAuth2 integration for tracking watch status. Auto-updater binary replacement. Custom keybindings.

**Out of Scope:**
*   Rendering video frames directly inside the terminal (too slow, poor UX).
*   Downloading episodes for offline viewing.
*   User account creation (local-only configuration).
