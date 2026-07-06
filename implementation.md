**V1 Implementation Steps:**
1.  `go mod init` and install dependencies.
2.  Write the AniList GraphQL client.
3.  Build the 3-state Bubbletea TUI (Search -> Results -> Episodes).
4.  Implement the `mpv` handoff using `tea.ExecProcess`.

**V2 Implementation Steps:**
1.  Abstract the Consumet client into the `Provider` interface.
2.  Build the `FallbackManager`.
3.  Integrate SQLite for history.
4.  Add terminal image rendering logic.

**V3 Implementation Steps:**
1.  Implement OAuth2 device flow for AniList.
2.  Build the auto-updater using GitHub Releases API.

**Professional CI/CD Pipeline (GitHub Actions):**
Create `.github/workflows/ci.yml`:
```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true
      - name: Run Linter
        uses: golangci/golangci-lint-action@v3
      - name: Run Tests
        run: go test -v -race ./...
```
Create `.github/workflows/release.yml` using **GoReleaser** to automatically build and attach binaries to GitHub tags.
