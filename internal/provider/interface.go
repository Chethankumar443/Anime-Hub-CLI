package provider

import (
	"context"
	"time"
)

type Anime struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Format      string   `json:"format"`
	ReleaseYear int      `json:"release_year"`
	CoverURL    string   `json:"cover_url"`
	ProviderID  string   `json:"provider_id"`
}

type AnimeDetails struct {
	Anime
	Synopsis string    `json:"synopsis"`
	Rating   float64   `json:"rating"`
	Status   string    `json:"status"`
	Genres   []string  `json:"genres"`
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	DurationSec int    `json:"duration_sec"`
}

type StreamLink struct {
	URL         string            `json:"url"`
	Quality     string            `json:"quality"`
	IsM3U8      bool              `json:"is_m3u8"`
	HTTPHeaders map[string]string `json:"http_headers"`
}

type PlaybackProgress struct {
	AnimeID     string    `json:"anime_id"`
	EpisodeID   string    `json:"episode_id"`
	EpisodeNum  int       `json:"episode_num"`
	ElapsedSec  int       `json:"elapsed_sec"`
	DurationSec int       `json:"duration_sec"`
	LastUpdated time.Time `json:"last_updated"`
	Completed   bool      `json:"completed"`
}

type Provider interface {
	ID() string
	Name() string
	Search(ctx context.Context, query string) ([]Anime, error)
	FetchDetails(ctx context.Context, id string) (AnimeDetails, error)
	FetchStreamLinks(ctx context.Context, episodeID string) ([]StreamLink, error)
}
