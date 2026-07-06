package config

import (
	"os"
	"testing"
	"time"

	"animehub/pkg/provider"
)

func TestSQLiteHistory(t *testing.T) {
	// Create isolated test environment directory
	tmpDir, err := os.MkdirTemp("", "animehub-history-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save existing environment variables to restore afterwards
	oldAppData := os.Getenv("APPDATA")
	oldHome := os.Getenv("HOME")

	os.Setenv("APPDATA", tmpDir)
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("APPDATA", oldAppData)
		os.Setenv("HOME", oldHome)
		ResetDB() // Reset global connection to prevent pollution
	}()

	ResetDB()

	// 1. Verify LoadHistory initially returns empty History
	hist, err := LoadHistory()
	if err != nil {
		t.Fatalf("failed to load initial history: %v", err)
	}
	if len(hist.Watchlist) != 0 || len(hist.Progress) != 0 {
		t.Errorf("expected empty watchlist and progress, got: %d watchitems, %d progressitems", len(hist.Watchlist), len(hist.Progress))
	}

	// 2. Add to Watchlist
	anime1 := provider.Anime{ID: "naruto-shippuden", Title: "Naruto Shippuden", Image: "http://example.com/naruto.png"}
	err = AddToWatchlist(anime1)
	if err != nil {
		t.Fatalf("failed to add to watchlist: %v", err)
	}

	// Attempt duplicate add (should be ignored)
	err = AddToWatchlist(anime1)
	if err != nil {
		t.Fatalf("failed during duplicate watchlist add: %v", err)
	}

	// Load and verify
	hist, err = LoadHistory()
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	if len(hist.Watchlist) != 1 || hist.Watchlist[0].ID != anime1.ID || hist.Watchlist[0].Image != anime1.Image {
		t.Errorf("expected watchlist size 1 with ID '%s' and Image '%s', got: %v", anime1.ID, anime1.Image, hist.Watchlist)
	}

	// 3. Update Progress
	prog1 := provider.PlaybackProgress{
		AnimeID:     "naruto-shippuden",
		EpisodeID:   "episode-1",
		EpisodeNum:  1,
		ElapsedSec:  100,
		DurationSec: 1200,
		Completed:   false,
	}
	err = UpdateProgress(prog1)
	if err != nil {
		t.Fatalf("failed to update progress: %v", err)
	}

	// Load and verify progress is stored
	hist, err = LoadHistory()
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	if len(hist.Progress) != 1 || hist.Progress[0].EpisodeID != prog1.EpisodeID {
		t.Errorf("expected progress list size 1 with EpisodeID '%s', got: %v", prog1.EpisodeID, hist.Progress)
	}
	if hist.Progress[0].ElapsedSec != 100 {
		t.Errorf("expected elapsed seconds 100, got: %d", hist.Progress[0].ElapsedSec)
	}

	// Update existing progress (same composite key: AnimeID + EpisodeID)
	prog1.ElapsedSec = 500
	err = UpdateProgress(prog1)
	if err != nil {
		t.Fatalf("failed to update existing progress: %v", err)
	}

	// Load and verify it updated instead of duplicating
	hist, err = LoadHistory()
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	if len(hist.Progress) != 1 {
		t.Errorf("expected progress list size 1 (upsert), got: %d", len(hist.Progress))
	}
	if hist.Progress[0].ElapsedSec != 500 {
		t.Errorf("expected elapsed seconds updated to 500, got: %d", hist.Progress[0].ElapsedSec)
	}

	// 4. SaveHistory (Bulk overwrite)
	anime2 := provider.Anime{ID: "one-piece", Title: "One Piece"}
	prog2 := provider.PlaybackProgress{
		AnimeID:     "one-piece",
		EpisodeID:   "episode-100",
		EpisodeNum:  100,
		ElapsedSec:  800,
		DurationSec: 1400,
		LastUpdated: time.Now(),
		Completed:   true,
	}

	newHist := &History{
		Watchlist: []provider.Anime{anime2},
		Progress:  []provider.PlaybackProgress{prog2},
	}
	err = SaveHistory(newHist)
	if err != nil {
		t.Fatalf("failed to save history bulk overwrite: %v", err)
	}

	// Load and verify overwrite
	hist, err = LoadHistory()
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	if len(hist.Watchlist) != 1 || hist.Watchlist[0].ID != "one-piece" {
		t.Errorf("expected overwritten watchlist with one-piece, got: %v", hist.Watchlist)
	}
	if len(hist.Progress) != 1 || hist.Progress[0].EpisodeID != "episode-100" {
		t.Errorf("expected overwritten progress list with episode-100, got: %v", hist.Progress)
	}

	// 5. Remove from Watchlist
	err = RemoveFromWatchlist("one-piece")
	if err != nil {
		t.Fatalf("failed to remove from watchlist: %v", err)
	}

	// Verify removal
	hist, err = LoadHistory()
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	if len(hist.Watchlist) != 0 {
		t.Errorf("expected watchlist to be empty after removal, got size: %d", len(hist.Watchlist))
	}
}
