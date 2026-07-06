**V1 Distribution:**
*   Use GoReleaser to compile stripped binaries (`ldflags: -s -w`) for Linux, macOS, and Windows.
*   First-run check: `exec.LookPath("mpv")`. If missing, print a styled error with OS-specific install commands and exit.

**V2 Distribution:**
*   Publish to Homebrew (via a custom tap) and Scoop (via a manifest).
*   Add shell completions (Bash, Zsh, Fish) generated via Cobra or a custom script.

**V3 Distribution:**
*   Implement the `anime-hub update` command. It downloads the new binary, verifies the checksum, replaces the current executable, and re-execs.

**The 6 Mandatory Security Patches (Robustness Requirements):**
A professional does not just write code; they secure the execution environment. You must implement these 6 patches:

1.  **Supply Chain Security (Dependency Scanning):** Integrate `Dependabot` or `Trivy` in your GitHub Actions. *Why:* Go modules can have vulnerabilities. You need automated PRs to update vulnerable dependencies.
2.  **Execution Safety (Safe `mpv` Spawning):** Never pass user input directly into `exec.Command` without sanitization. *Why:* If a pirate site returns a malicious URL with shell injection characters (e.g., `; rm -rf /`), and you use `sh -c`, you get pwned. Always pass arguments as an array to `exec.Command`, never as a single string to a shell.
3.  **Data at Rest (Config Permissions):** When creating the local SQLite DB or JSON config in `~/.config/animehub`, set file permissions to `0600` (read/write for owner only). *Why:* Prevents other users on a shared machine from reading watch history or OAuth tokens.
4.  **Network Security (TLS Enforcement):** Configure your `resty` HTTP client to enforce TLS 1.2+ and verify certificates. *Why:* Prevents Man-in-the-Middle (MitM) attacks on public Wi-Fi where a malicious actor could intercept API requests and return poisoned video URLs.
5.  **Update Integrity (Checksum Verification):** In your V3 auto-updater, never execute a downloaded binary without first downloading the `checksums.txt` file from the GitHub release, calculating the SHA256 of the downloaded binary, and comparing them. *Why:* Prevents a compromised GitHub account or CDN from pushing a malicious binary to your users.
6.  **Input Validation (GraphQL/API):** Validate all search queries before sending them to the API. Strip control characters and limit query length to 100 characters. *Why:* Prevents accidental Denial of Service (DoS) on the free AniList/Consumet APIs and prevents malformed requests from crashing your TUI parser.
