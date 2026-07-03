package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yourorg/anime-cli/pkg/provider"
)

type History struct {
	Watchlist []provider.Anime             `json:"watchlist"`
	Progress  []provider.PlaybackProgress `json:"progress"`
}

var (
	historyLock sync.RWMutex
	appHistory  *History
)

// GetHistoryPath returns the path to history.json
func GetHistoryPath() string {
	return filepath.Join(GetAppDir(), "history.json")
}

// LoadHistory reads watch history from disk.
func LoadHistory() (*History, error) {
	historyLock.Lock()
	defer historyLock.Unlock()

	if appHistory != nil {
		return appHistory, nil
	}

	path := GetHistoryPath()
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			appHistory = &History{
				Watchlist: []provider.Anime{},
				Progress:  []provider.PlaybackProgress{},
			}
			_ = saveHistoryLocked(appHistory)
			return appHistory, nil
		}
		return nil, err
	}
	defer file.Close()

	var hist History
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&hist); err != nil {
		if err == io.EOF {
			appHistory = &History{
				Watchlist: []provider.Anime{},
				Progress:  []provider.PlaybackProgress{},
			}
			return appHistory, nil
		}
		return nil, err
	}

	if hist.Watchlist == nil {
		hist.Watchlist = []provider.Anime{}
	}
	if hist.Progress == nil {
		hist.Progress = []provider.PlaybackProgress{}
	}

	appHistory = &hist
	return appHistory, nil
}

// SaveHistory writes the history data back to disk.
func SaveHistory(hist *History) error {
	historyLock.Lock()
	defer historyLock.Unlock()
	appHistory = hist
	return saveHistoryLocked(hist)
}

func saveHistoryLocked(hist *History) error {
	path := GetHistoryPath()
	tmpPath := path + ".tmp"

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(hist); err != nil {
		return err
	}

	file.Close() // Close before rename
	return os.Rename(tmpPath, path)
}

// UpdateProgress updates the play progress for a specific episode.
func UpdateProgress(progress provider.PlaybackProgress) error {
	hist, err := LoadHistory()
	if err != nil {
		return err
	}

	historyLock.Lock()
	defer historyLock.Unlock()

	// Update existing progress or insert new
	found := false
	for i, p := range hist.Progress {
		if p.AnimeID == progress.AnimeID && p.EpisodeID == progress.EpisodeID {
			hist.Progress[i].ElapsedSec = progress.ElapsedSec
			hist.Progress[i].DurationSec = progress.DurationSec
			hist.Progress[i].LastUpdated = time.Now()
			hist.Progress[i].Completed = progress.Completed
			found = true
			break
		}
	}

	if !found {
		progress.LastUpdated = time.Now()
		hist.Progress = append(hist.Progress, progress)
	}

	return saveHistoryLocked(hist)
}

// AddToWatchlist appends an anime entry to the watchlist.
func AddToWatchlist(anime provider.Anime) error {
	hist, err := LoadHistory()
	if err != nil {
		return err
	}

	historyLock.Lock()
	defer historyLock.Unlock()

	for _, item := range hist.Watchlist {
		if item.ID == anime.ID && item.ProviderID == anime.ProviderID {
			return nil // Already in watchlist
		}
	}

	hist.Watchlist = append(hist.Watchlist, anime)
	return saveHistoryLocked(hist)
}

// RemoveFromWatchlist removes an anime entry from the watchlist.
func RemoveFromWatchlist(animeID string) error {
	hist, err := LoadHistory()
	if err != nil {
		return err
	}

	historyLock.Lock()
	defer historyLock.Unlock()

	newWatchlist := make([]provider.Anime, 0, len(hist.Watchlist))
	for _, item := range hist.Watchlist {
		if item.ID != animeID {
			newWatchlist = append(newWatchlist, item)
		}
	}
	hist.Watchlist = newWatchlist

	return saveHistoryLocked(hist)
}
