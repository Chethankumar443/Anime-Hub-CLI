**The Embedded Node Binary Distribution (No Docker):**
1.  Compile `consumet-win.exe`, `consumet-mac`, and `consumet-linux` using `pkg`.
2.  Host these binaries in your GitHub Releases.
3.  In your Go CLI's `main.go`, implement a "First Run Downloader":
    ```go
    func ensureConsumetBinary() error {
        // Check ~/.config/animehub/consumet.exe
        // If missing, download from GitHub Releases based on runtime.GOOS
        // Set chmod +x on Linux/Mac
    }
    ```

**The `mpv` Distribution Reality:**
You cannot bundle `mpv`. On first run, check `exec.LookPath("mpv")`. If missing, print a Lipgloss-styled error with `brew install mpv`, `sudo apt install mpv`, or `winget install mpv` and exit gracefully.

**Auto-Update Logic (V3):**
1.  Query GitHub Releases API for the latest tag.
2.  Compare to local version (injected via `ldflags`).
3.  Download new binary to a temp file.
4.  Verify SHA256 against `checksums.txt`.
5.  Replace current executable and `exec` the new binary.

**Build Pipeline (GoReleaser):**
```yaml
builds:
  - ldflags:
      - -s -w
      - -X main.version={{.Version}}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
```
This ensures your binary is stripped of debug symbols (keeping it under 15MB) and has the version number baked in for the auto-updater.
