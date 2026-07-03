package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type GHRelease struct {
	TagName string `json:"tag_name"`
}

func CheckForUpdates(currentVersion string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/cheth/anime-cli/releases/latest")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking update server:", err)
		return
	}
	defer resp.Body.Close()

	var release GHRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding release schema:", err)
		return
	}

	if release.TagName != "v"+currentVersion && release.TagName != currentVersion {
		fmt.Printf("A new version is available: %s (Current: %s)\n", release.TagName, currentVersion)
		fmt.Printf("Run the install script or update via your system package manager.\n")
	} else {
		fmt.Println("anime-cli is already up to date.")
	}
}
