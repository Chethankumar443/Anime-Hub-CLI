package allanime

import (
	"context"
	"strings"
	"time"

	"github.com/yourorg/anime-cli/pkg/http"
	"github.com/yourorg/anime-cli/internal/provider"
)

type AllAnimeProvider struct {
	BaseURL string
	Client  *http.Client
}

func NewAllAnimeProvider(baseURL string) *AllAnimeProvider {
	if baseURL == "" {
		baseURL = "https://allanime.to"
	}
	return &AllAnimeProvider{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Client:  http.NewClient(5 * time.Second),
	}
}

func (p *AllAnimeProvider) ID() string {
	return "allanime"
}

func (p *AllAnimeProvider) Name() string {
	return "AllAnime"
}

func (p *AllAnimeProvider) Search(ctx context.Context, query string) ([]provider.Anime, error) {
	var results []provider.Anime
	q := strings.ToLower(query)
	
	catalog := map[string]provider.Anime{
		"frieren": {
			ID:          "frieren",
			Title:       "Frieren: Beyond Journey's End",
			Format:      "TV",
			ReleaseYear: 2023,
			CoverURL:    "https://gogocdn.net/cover/frieren-beyond-journeys-end.png",
			ProviderID:  p.ID(),
		},
		"one-piece": {
			ID:          "one-piece",
			Title:       "One Piece",
			Format:      "TV",
			ReleaseYear: 1999,
			CoverURL:    "https://gogocdn.net/cover/one-piece.png",
			ProviderID:  p.ID(),
		},
		"naruto": {
			ID:          "naruto",
			Title:       "Naruto",
			Format:      "TV",
			ReleaseYear: 2002,
			CoverURL:    "https://gogocdn.net/cover/naruto.png",
			ProviderID:  p.ID(),
		},
	}

	for key, val := range catalog {
		if strings.Contains(strings.ToLower(val.Title), q) || strings.Contains(key, q) {
			results = append(results, val)
		}
	}
	if len(results) == 0 {
		for _, val := range catalog {
			results = append(results, val)
		}
	}
	return results, nil
}

func (p *AllAnimeProvider) FetchDetails(ctx context.Context, id string) (provider.AnimeDetails, error) {
	episodes := []provider.Episode{
		{ID: id + "-episode-1", Number: 1, Title: "Episode 1", DurationSec: 1440},
		{ID: id + "-episode-2", Number: 2, Title: "Episode 2", DurationSec: 1440},
		{ID: id + "-episode-3", Number: 3, Title: "Episode 3", DurationSec: 1440},
	}

	title := "Frieren: Beyond Journey's End"
	if strings.Contains(id, "one-piece") {
		title = "One Piece"
	} else if strings.Contains(id, "naruto") {
		title = "Naruto"
	}

	if strings.HasSuffix(id, "-dub") {
		title = title + " (Dub)"
	}

	return provider.AnimeDetails{
		Anime: provider.Anime{
			ID:          id,
			Title:       title,
			Format:      "TV",
			ReleaseYear: 2023,
			CoverURL:    "https://gogocdn.net/cover/frieren-beyond-journeys-end.png",
			ProviderID:  p.ID(),
		},
		Synopsis: "Mock synopsis for Allanime provider details.",
		Rating:   8.9,
		Status:   "Finished Airing",
		Genres:   []string{"Action", "Fantasy"},
		Episodes: episodes,
	}, nil
}

func (p *AllAnimeProvider) FetchStreamLinks(ctx context.Context, episodeID string) ([]provider.StreamLink, error) {
	return []provider.StreamLink{
		{
			URL:     "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
			Quality: "auto",
			IsM3U8:  false,
			HTTPHeaders: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
		},
	}, nil
}
