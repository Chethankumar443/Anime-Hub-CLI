**Hard Constraints (What will break if you ignore them):**
1.  **Stream URL Expiration:** Free provider APIs generate video URLs that expire in 15 to 60 minutes. *Rule:* Never fetch the video URL when the app starts. Fetch it *milliseconds* before launching `mpv`.
2.  **API Rate Limiting:** AniList limits to 90 requests per minute. *Rule:* Implement exponential backoff and request caching in your Go HTTP client.
3.  **Terminal State Corruption:** When `mpv` runs, it takes over the terminal. *Rule:* You must use Bubbletea's `tea.ExecProcess`, not standard `os/exec`, to properly suspend and resume the TUI.

**The 6 Mandatory Security Patches:**
1.  **Supply Chain Security:** Integrate `Dependabot` in GitHub Actions to auto-update vulnerable Go modules.
2.  **Execution Safety:** Never pass user input directly into `exec.Command` via a shell. Always pass arguments as an array to prevent shell injection.
3.  **Data at Rest:** Set local config/SQLite file permissions to `0600` (owner read/write only) to protect OAuth tokens and history.
4.  **Network Security:** Enforce TLS 1.2+ and verify certificates in the `resty` HTTP client to prevent MitM attacks on public Wi-Fi.
5.  **Update Integrity:** In the V3 auto-updater, verify the SHA256 checksum of the downloaded binary against the GitHub release `checksums.txt` before executing it.
6.  **Input Validation:** Strip control characters and limit search queries to 100 characters to prevent DoS on free APIs.
