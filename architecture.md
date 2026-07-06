**V1 Architecture: The Monolith**
*   **Pattern:** MVC (Model-View-Controller) via Bubbletea.
*   **Data Flow:** `User Input` -> `Bubbletea Update` -> `Resty HTTP Client` -> `AniList/Consumet API` -> `Parse JSON` -> `Update Model` -> `Bubbletea View`.
*   **Playback Logic:** `Model` holds the selected episode ID. On `Enter`, `Update` calls `GetStreamURL`, creates `exec.Command("mpv", url)`, and returns `tea.ExecProcess`.

**V2 Architecture: The Plugin System**
*   **Pattern:** Strategy Pattern for Providers.
*   **Logic:** 
    ```go
    type ProviderRegistry struct {
        providers []VideoProvider
    }
    // Iterates through providers. If Provider A returns 404 or 500, 
    // it catches the error and tries Provider B.
    ```
*   **State Management:** Move from in-memory JSON to SQLite. Create a `HistoryService` interface that the Bubbletea model calls via channels to prevent blocking the UI thread during DB writes.

**V3 Architecture: Event-Driven Sync**
*   **Pattern:** Observer Pattern.
*   **Logic:** When `mpv` exits, the `playbackFinishedMsg` is sent. The `Update` function not only resumes the TUI but also fires an asynchronous event to the `AniListSyncService` to update the user's watch progress via GraphQL mutation.
