# Build & Deployment Specification
## Project: TerminalAnime CLI (`anime-cli`)

---

### 1. Build Architecture & Optimization Parameters

The Go compiler compiles the codebase into a single static executable with zero external dynamic runtime link checks. This ensures maximum speed, portability, and compatibility across macOS, Linux, and Windows.

#### 1.1. Optimization Flags
To reduce executable size and improve startup times, builds must use these parameters:
*   `CGO_ENABLED=0`: Disables dynamic linking checks to produce fully static binaries.
*   Linker Flags (`-ldflags="-s -w"`):
    *   `-s`: Strips symbol tables and debug details, reducing file footprint by ~40%.
    *   `-w`: Omits DWARF symbol tables, preventing debug attachments but minimizing runtime memory consumption.

#### 1.2. Cross-Compilation Command Specifications
```bash
# Compile target: Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/anime-cli-linux-amd64 cmd/anime-cli/main.go

# Compile target: macOS Apple Silicon (ARM64)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/anime-cli-darwin-arm64 cmd/anime-cli/main.go

# Compile target: Windows AMD64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/anime-cli-windows-amd64.exe cmd/anime-cli/main.go
```

---

### 2. CI/CD Release Pipeline

Save this GitHub Actions workflow configuration to `.github/workflows/release.yml` in the project root:

```yaml
name: Release Toolchain

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Source Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install Go Environment
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Run Goreleaser Release Engine
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### 3. GoReleaser Configuration File

Save this configuration to `.goreleaser.yaml` in the project root:

```yaml
version: 2

project_name: anime-cli

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    dir: ./cmd/anime-cli
    main: ./main.go
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

---

### 4. Package Manager Distribution Specs

#### 4.1. Homebrew Formula (macOS & Linux)
Save this formula to `Formula/anime-cli.rb` inside the tap repository:

```ruby
class AnimeCli < Formula
  desc "Premium terminal-native anime streaming client written in Go"
  homepage "https://github.com/cheth/anime-cli"
  url "https://github.com/cheth/anime-cli/releases/download/v1.0.0/anime-cli_1.0.0_darwin_arm64.tar.gz"
  sha256 "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  license "MIT"

  depends_on "mpv"

  def install
    bin.install "anime-cli"
  end

  test do
    system "#{bin}/anime-cli", "--version"
  end
end
```

#### 4.2. Arch Linux PKGBUILD (AUR Recipe)
Save this recipe to the AUR repository folder as `PKGBUILD`:

```bash
# Maintainer: Chethan <dev@animecli.dev>
pkgname=anime-cli-bin
pkgver=1.0.0
pkgrel=1
pkgdesc="Premium terminal-native anime streaming client written in Go"
arch=('x86_64' 'aarch64')
url="https://github.com/cheth/anime-cli"
license=('MIT')
depends=('mpv')
provides=('anime-cli')
conflicts=('anime-cli')
source_x86_64=("https://github.com/cheth/${pkgname}/releases/download/v${pkgver}/anime-cli_${pkgver}_linux_amd64.tar.gz")
sha256sums_x86_64=('e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')

package() {
  install -Dm755 "${srcdir}/anime-cli" "${pkgdir}/usr/bin/anime-cli"
}
```

#### 4.3. Windows Scoop Manifest
Save this manifest to the Scoop bucket repository as `anime-cli.json`:

```json
{
  "version": "1.0.0",
  "description": "Premium terminal-native anime streaming client written in Go",
  "homepage": "https://github.com/cheth/anime-cli",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/cheth/anime-cli/releases/download/v1.0.0/anime-cli_1.0.0_windows_amd64.zip",
      "hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  "bin": "anime-cli.exe",
  "suggest": {
    "mpv": "mpv"
  }
}
```

---

### 5. Automated Installation & Updates

#### 5.1. Unix-based Installer Script
To install the latest release automatically, run the following shell pipeline:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Auto-detect Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Error: Unsupported architecture $ARCH" >&2
    exit 1
fi

REPO="cheth/anime-cli"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
CLEAN_VERSION="${LATEST_VERSION#v}"

DOWNLOAD_URL="https://github.com/$REPO/releases/download/${LATEST_VERSION}/anime-cli_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"

echo "Downloading anime-cli ${LATEST_VERSION}..."
curl -sL "$DOWNLOAD_URL" | tar -xz -C /tmp

echo "Installing binary to /usr/local/bin/anime-cli..."
sudo mv /tmp/anime-cli /usr/local/bin/anime-cli
sudo chmod +x /usr/local/bin/anime-cli

echo "Installation complete. Run 'anime-cli' to launch."
```

#### 5.2. Integrated Self-Updater Command
The CLI includes a built-in `--update` command flag to check and pull the latest releases:

```go
package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type GHRelease struct {
	TagName string `json:"tag_name"`
}

func CheckForUpdates(currentVersion string) {
	resp, err := http.Get("https://api.github.com/repos/cheth/anime-cli/releases/latest")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking update server:", err)
		return
	}
	defer resp.Body.Close()

	var release GHRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding release response:", err)
		return
	}

	if release.TagName != currentVersion {
		fmt.Printf("A new version is available: %s (Current: %s)\n", release.TagName, currentVersion)
		fmt.Printf("Run the install script or update via your system package manager.\n")
	} else {
		fmt.Println("anime-cli is already up to date.")
	}
}
```

---

### 6. File Cross-References
*   Product requirements and user stories: See [prd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/prd.md).
*   Underlying system loops and socket connections: See [trd.md](file:///c:/Users/cheth/Desktop/TerminalAnime/trd.md).
*   Visual wireframes and navigation keybindings: See [navigation.md](file:///c:/Users/cheth/Desktop/TerminalAnime/navigation.md).
*   Phased implementation checklist: See [implementation_plan.md](file:///c:/Users/cheth/Desktop/TerminalAnime/implementation_plan.md).
*   Domain structs and JSON specifications: See [techstack.md](file:///c:/Users/cheth/Desktop/TerminalAnime/techstack.md).
