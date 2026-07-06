package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"animehub/pkg/provider"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type History struct {
	Watchlist []provider.Anime            `json:"watchlist"`
	Progress  []provider.PlaybackProgress `json:"progress"`
}

type WatchlistItem struct {
	ID        string `gorm:"primaryKey"`
	Title     string
	Image     string
	CreatedAt time.Time
}

type PlaybackProgressModel struct {
	AnimeID     string    `gorm:"primaryKey;index"`
	EpisodeID   string    `gorm:"primaryKey"`
	EpisodeNum  int
	ElapsedSec  int
	DurationSec int
	LastUpdated time.Time
	Completed   bool
}

var (
	db     *gorm.DB
	dbLock sync.Mutex
)

// GetDatabasePath returns the path to history.db
func GetDatabasePath() string {
	return filepath.Join(GetAppDir(), "history.db")
}

// ResetDB closes and resets the global database connection (useful for unit tests)
func ResetDB() {
	dbLock.Lock()
	defer dbLock.Unlock()
	db = nil
}

func initDB() (*gorm.DB, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if db != nil {
		return db, nil
	}

	path := GetDatabasePath()
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	var err error
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate schemas
	err = db.AutoMigrate(&WatchlistItem{}, &PlaybackProgressModel{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// LoadHistory reads watch history from the SQLite database.
func LoadHistory() (*History, error) {
	database, err := initDB()
	if err != nil {
		return nil, err
	}

	var watchitems []WatchlistItem
	if err := database.Find(&watchitems).Error; err != nil {
		return nil, err
	}

	var progressitems []PlaybackProgressModel
	if err := database.Find(&progressitems).Error; err != nil {
		return nil, err
	}

	watchlist := make([]provider.Anime, len(watchitems))
	for i, item := range watchitems {
		watchlist[i] = provider.Anime{
			ID:    item.ID,
			Title: item.Title,
			Image: item.Image,
		}
	}

	progress := make([]provider.PlaybackProgress, len(progressitems))
	for i, item := range progressitems {
		progress[i] = provider.PlaybackProgress{
			AnimeID:     item.AnimeID,
			EpisodeID:   item.EpisodeID,
			EpisodeNum:  item.EpisodeNum,
			ElapsedSec:  item.ElapsedSec,
			DurationSec: item.DurationSec,
			LastUpdated: item.LastUpdated,
			Completed:   item.Completed,
		}
	}

	return &History{
		Watchlist: watchlist,
		Progress:  progress,
	}, nil
}

// SaveHistory writes the history data back to the database.
func SaveHistory(hist *History) error {
	database, err := initDB()
	if err != nil {
		return err
	}

	return database.Transaction(func(tx *gorm.DB) error {
		// Delete all existing items first to overwrite completely
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&WatchlistItem{}).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&PlaybackProgressModel{}).Error; err != nil {
			return err
		}

		// Insert new items
		for _, item := range hist.Watchlist {
			watchitem := WatchlistItem{
				ID:        item.ID,
				Title:     item.Title,
				Image:     item.Image,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(&watchitem).Error; err != nil {
				return err
			}
		}

		for _, item := range hist.Progress {
			progressitem := PlaybackProgressModel{
				AnimeID:     item.AnimeID,
				EpisodeID:   item.EpisodeID,
				EpisodeNum:  item.EpisodeNum,
				ElapsedSec:  item.ElapsedSec,
				DurationSec: item.DurationSec,
				LastUpdated: item.LastUpdated,
				Completed:   item.Completed,
			}
			if err := tx.Create(&progressitem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateProgress updates the play progress for a specific episode.
func UpdateProgress(progress provider.PlaybackProgress) error {
	database, err := initDB()
	if err != nil {
		return err
	}

	progressitem := PlaybackProgressModel{
		AnimeID:     progress.AnimeID,
		EpisodeID:   progress.EpisodeID,
		EpisodeNum:  progress.EpisodeNum,
		ElapsedSec:  progress.ElapsedSec,
		DurationSec: progress.DurationSec,
		LastUpdated: time.Now(),
		Completed:   progress.Completed,
	}

	return database.Save(&progressitem).Error
}

// AddToWatchlist appends an anime entry to the watchlist.
func AddToWatchlist(anime provider.Anime) error {
	database, err := initDB()
	if err != nil {
		return err
	}

	var existing WatchlistItem
	err = database.Where("id = ?", anime.ID).First(&existing).Error
	if err == nil {
		return nil // Already in watchlist
	}

	watchitem := WatchlistItem{
		ID:        anime.ID,
		Title:     anime.Title,
		Image:     anime.Image,
		CreatedAt: time.Now(),
	}

	return database.Create(&watchitem).Error
}

// RemoveFromWatchlist removes an anime entry from the watchlist.
func RemoveFromWatchlist(animeID string) error {
	database, err := initDB()
	if err != nil {
		return err
	}

	return database.Where("id = ?", animeID).Delete(&WatchlistItem{}).Error
}
