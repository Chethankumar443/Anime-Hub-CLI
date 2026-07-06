**Phase 1: V1 Implementation (Days 1-10)**
1.  `go mod init animehub`. Install Bubbletea, Lipgloss, Resty.
2.  Compile Consumet to standalone binaries using `pkg`.
3.  Implement `ConsumetManager` to download (if missing), start, and poll the embedded binary.
4.  Build the 3-state Bubbletea TUI (Search -> Results -> Episodes).
5.  Implement `tea.ExecProcess` for `mpv` handoff.

**Phase 2: V2 Implementation (Days 11-20)**
1.  Abstract providers into the `FallbackManager` interface.
2.  Migrate local JSON history to SQLite (`glebarez/sqlite`).
3.  Implement terminal image rendering (Kitty/Sixel) for cover art.

**Phase 3: V3 Implementation (Days 21+)**
1.  Implement AniList OAuth2 device flow.
2.  Build the `anime-hub update` command with SHA256 verification.

**Professional CI/CD Pipeline (GitHub Actions):**
```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.21', cache: true }
      - uses: golangci/golangci-lint-action@v3
      - run: go test -v -race ./...
```

**AI IDE Workflow (Cursor/Windsurf):**
Create a `.cursorrules` file in the root:
> "You are an expert Go developer. Always use `golangci-lint` standards. Prefer composition over inheritance. Do not write raw HTML scrapers; use the defined Provider interfaces. Keep functions under 40 lines. Never use `panic()` in application logic."
