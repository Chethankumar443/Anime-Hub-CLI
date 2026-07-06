**V1: The Functional Prototype (Weeks 1-2)**
*   **Goal:** Prove the terminal state machine and `mpv` handoff work.
*   **Scope:** Text-only UI. Single metadata API (AniList). Single video provider (Consumet). Basic Sub/Dub toggle. Local JSON watch history.
*   **Success Metric:** A user can search "One Piece", select episode 1, and watch it in `mpv` without the terminal breaking.

**V2: The Resilient Product (Weeks 3-4)**
*   **Goal:** Eliminate single points of failure.
*   **Scope:** Multi-provider fallback routing. Stream URL auto-refresh. Terminal image support (Kitty/Sixel) for cover art. 
*   **Success Metric:** If the primary video provider goes down, the user experiences zero downtime; the tool silently switches to the backup provider.

**V3: The Ecosystem Tool (Weeks 5-6)**
*   **Goal:** Deep integration with the user's anime tracking.
*   **Scope:** AniList OAuth2 integration to auto-mark episodes as watched. Auto-updater binary replacement. 
*   **Success Metric:** The tool acts as a complete frontend for the user's AniList account.
