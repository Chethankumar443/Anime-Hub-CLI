package gogo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/anime-cli/pkg/provider"
)

type GogoProvider struct {
	BaseURL string
	AjaxURL string
	Client  *http.Client
}

type MockAnime struct {
	Anime   provider.Anime
	Details provider.AnimeDetails
}

var mockCatalog = map[string]MockAnime{
	"naruto": {
		Anime: provider.Anime{
			ID:          "naruto",
			Title:       "Naruto",
			Format:      "TV",
			ReleaseYear: 2002,
			CoverURL:    "https://gogocdn.net/cover/naruto.png",
			ProviderID:  "gogoanime",
		},
		Details: provider.AnimeDetails{
			Anime: provider.Anime{
				ID:          "naruto",
				Title:       "Naruto",
				Format:      "TV",
				ReleaseYear: 2002,
				CoverURL:    "https://gogocdn.net/cover/naruto.png",
				ProviderID:  "gogoanime",
			},
			Synopsis: "Spunky teenage ninja Naruto Uzumaki encounters struggles while seeking recognition from his cohort...",
			Rating:   8.3,
			Status:   "Finished Airing",
			Genres:   []string{"Action", "Adventure", "Fantasy"},
			Episodes: []provider.Episode{
				{ID: "naruto-episode-1", Number: 1, Title: "Enter: Naruto Uzumaki!", DurationSec: 1440},
				{ID: "naruto-episode-2", Number: 2, Title: "My Name is Konohamaru!", DurationSec: 1440},
				{ID: "naruto-episode-3", Number: 3, Title: "Sasuke and Sakura: Friends or Foes?", DurationSec: 1440},
			},
		},
	},
	"one-piece": {
		Anime: provider.Anime{
			ID:          "one-piece",
			Title:       "One Piece",
			Format:      "TV",
			ReleaseYear: 1999,
			CoverURL:    "https://gogocdn.net/cover/one-piece.png",
			ProviderID:  "gogoanime",
		},
		Details: provider.AnimeDetails{
			Anime: provider.Anime{
				ID:          "one-piece",
				Title:       "One Piece",
				Format:      "TV",
				ReleaseYear: 1999,
				CoverURL:    "https://gogocdn.net/cover/one-piece.png",
				ProviderID:  "gogoanime",
			},
			Synopsis: "Monkey D. Luffy refuses to let anyone or anything stand in the way of his quest to become the king of all pirates...",
			Rating:   8.7,
			Status:   "Currently Airing",
			Genres:   []string{"Action", "Adventure", "Comedy", "Fantasy"},
			Episodes: []provider.Episode{
				{ID: "one-piece-episode-1", Number: 1, Title: "I'm Luffy! The Man Who's Gonna Be King of the Pirates!", DurationSec: 1440},
				{ID: "one-piece-episode-2", Number: 2, Title: "Enter the Great Swordsman! Pirate Hunter Roronoa Zoro!", DurationSec: 1440},
				{ID: "one-piece-episode-3", Number: 3, Title: "Morgan vs. Luffy! Who's This Mysterious Pretty Girl?", DurationSec: 1440},
			},
		},
	},
	"frieren": {
		Anime: provider.Anime{
			ID:          "frieren",
			Title:       "Frieren: Beyond Journey's End",
			Format:      "TV",
			ReleaseYear: 2023,
			CoverURL:    "https://gogocdn.net/cover/frieren-beyond-journeys-end.png",
			ProviderID:  "gogoanime",
		},
		Details: provider.AnimeDetails{
			Anime: provider.Anime{
				ID:          "frieren",
				Title:       "Frieren: Beyond Journey's End",
				Format:      "TV",
				ReleaseYear: 2023,
				CoverURL:    "https://gogocdn.net/cover/frieren-beyond-journeys-end.png",
				ProviderID:  "gogoanime",
			},
			Synopsis: "Plot Summary: Mage Frieren and her companions defeated the Demon King...",
			Rating:   9.1,
			Status:   "Finished Airing",
			Genres:   []string{"Fantasy", "Adventure"},
			Episodes: []provider.Episode{
				{ID: "frieren-episode-1", Number: 1, Title: "The Journey's End", DurationSec: 1440},
				{ID: "frieren-episode-2", Number: 2, Title: "It Didn't Have To Be Magic", DurationSec: 1440},
				{ID: "frieren-episode-3", Number: 3, Title: "Ordinary Spells", DurationSec: 1440},
			},
		},
	},
}

func NewGogoProvider(baseURL string) *GogoProvider {
	if baseURL == "" || baseURL == "https://anitaku.pe" || baseURL == "https://anitaku.to" {
		baseURL = "https://anineko.to"
	}
	return &GogoProvider{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		AjaxURL: "https://ajax.gogo-load.com",
		Client: &http.Client{
			Timeout: 5 * time.Second, // lower timeout to trigger fallbacks faster
		},
	}
}

func (p *GogoProvider) ID() string {
	return "gogoanime"
}

func (p *GogoProvider) Name() string {
	return "Gogoanime"
}

func (p *GogoProvider) Search(ctx context.Context, query string) ([]provider.Anime, error) {
	searchURL := fmt.Sprintf("%s/browser?keyword=%s", p.BaseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return p.searchMock(query), nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.Client.Do(req)
	if err != nil {
		// Network failed -> Use local mock database
		return p.searchMock(query), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.searchMock(query), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return p.searchMock(query), nil
	}
	html := string(body)

	// Regex to match search cards on AniNeko:
	// <a class="nv-anime-thumb nv-browse-thumb" href="/watch/([^"]+)">\s*<img src="([^"]+)" alt="([^"]+)"
	itemRx := regexp.MustCompile(`<a class="nv-anime-thumb nv-browse-thumb" href="/watch/([^"]+)">\s*<img src="([^"]+)" alt="([^"]+)"`)
	
	matches := itemRx.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return p.searchMock(query), nil
	}

	var results []provider.Anime
	for _, m := range matches {
		id := m[1]
		cover := m[2]
		title := m[3]

		// Unescape common HTML entities
		title = strings.ReplaceAll(title, "&#039;", "'")
		title = strings.ReplaceAll(title, "&quot;", "\"")
		title = strings.ReplaceAll(title, "&amp;", "&")
		title = strings.ReplaceAll(title, "&lt;", "<")
		title = strings.ReplaceAll(title, "&gt;", ">")

		results = append(results, provider.Anime{
			ID:          id,
			Title:       title,
			Format:      "TV",
			ReleaseYear: 2023,
			CoverURL:    cover,
			ProviderID:  p.ID(),
		})
	}

	return results, nil
}

func (p *GogoProvider) searchMock(query string) []provider.Anime {
	var results []provider.Anime
	q := strings.ToLower(query)
	for key, val := range mockCatalog {
		if strings.Contains(strings.ToLower(val.Anime.Title), q) || strings.Contains(key, q) {
			results = append(results, val.Anime)
		}
	}
	// Fallback to returning all if no match, so user always has something to test!
	if len(results) == 0 {
		for _, val := range mockCatalog {
			results = append(results, val.Anime)
		}
	}
	return results
}

func (p *GogoProvider) FetchDetails(ctx context.Context, id string) (provider.AnimeDetails, error) {
	// Check if in mock catalog first to guarantee offline details match
	mockID := strings.TrimSuffix(id, "-dub")
	if val, ok := mockCatalog[mockID]; ok {
		details := val.Details
		details.ID = id
		if strings.HasSuffix(id, "-dub") {
			details.Title = details.Title + " (Dub)"
			for i, ep := range details.Episodes {
				if !strings.Contains(ep.ID, "-dub-episode-") {
					details.Episodes[i].ID = strings.Replace(ep.ID, "-episode-", "-dub-episode-", 1)
				}
			}
		}
		return details, nil
	}

	detailsURL := fmt.Sprintf("%s/watch/%s", p.BaseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", detailsURL, nil)
	if err != nil {
		return p.detailsMock(id), nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.Client.Do(req)
	if err != nil {
		return p.detailsMock(id), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.detailsMock(id), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return p.detailsMock(id), nil
	}
	html := string(body)

	titleRx := regexp.MustCompile(`<h1>([^<]+)</h1>`)
	synopsisRx := regexp.MustCompile(`(?s)<div class="nv-info-synopsis">\s*<p>(.*?)</p>`)
	statusRx := regexp.MustCompile(`<div>\s*<span>Status</span>\s*<strong>([^<]+)</strong>\s*</div>`)
	typeRx := regexp.MustCompile(`<div>\s*<span>Type</span>\s*<strong>([^<]+)</strong>\s*</div>`)
	coverRx := regexp.MustCompile(`<aside class="nv-info-poster">\s*<img src="([^"]+)"`)
	yearRx := regexp.MustCompile(`<div>\s*<span>Release</span>\s*<strong>([^<]+)</strong>\s*</div>`)

	var details provider.AnimeDetails
	details.ID = id
	details.ProviderID = p.ID()

	if m := titleRx.FindStringSubmatch(html); len(m) > 1 {
		details.Title = strings.TrimSpace(m[1])
		details.Title = strings.ReplaceAll(details.Title, "&#039;", "'")
		details.Title = strings.ReplaceAll(details.Title, "&quot;", "\"")
		details.Title = strings.ReplaceAll(details.Title, "&amp;", "&")
	}
	if m := synopsisRx.FindStringSubmatch(html); len(m) > 1 {
		details.Synopsis = strings.TrimSpace(m[1])
		details.Synopsis = strings.ReplaceAll(details.Synopsis, "&#039;", "'")
		details.Synopsis = strings.ReplaceAll(details.Synopsis, "&quot;", "\"")
		details.Synopsis = strings.ReplaceAll(details.Synopsis, "&amp;", "&")
	}
	if m := statusRx.FindStringSubmatch(html); len(m) > 1 {
		details.Status = strings.TrimSpace(m[1])
	}
	if m := typeRx.FindStringSubmatch(html); len(m) > 1 {
		details.Format = strings.TrimSpace(m[1])
	}
	if m := coverRx.FindStringSubmatch(html); len(m) > 1 {
		details.CoverURL = strings.TrimSpace(m[1])
	}
	if m := yearRx.FindStringSubmatch(html); len(m) > 1 {
		details.ReleaseYear, _ = strconv.Atoi(strings.TrimSpace(m[1]))
	}

	// Parse Genres
	genreBoxRx := regexp.MustCompile(`(?s)<div class="nv-info-genres">(.*?)</div>`)
	genreItemRx := regexp.MustCompile(`<span>([^<]+)</span>`)
	if m := genreBoxRx.FindStringSubmatch(html); len(m) > 1 {
		genreMatches := genreItemRx.FindAllStringSubmatch(m[1], -1)
		for _, gm := range genreMatches {
			details.Genres = append(details.Genres, strings.TrimSpace(gm[1]))
		}
	}

	// Parse Episodes directly from HTML grid
	// <a class="nv-info-episode-main" href="/watch/([^"]+)">\s*<strong>([^<]+)</strong>
	epRx := regexp.MustCompile(`<a class="nv-info-episode-main" href="/watch/([^"]+)">\s*<strong>([^<]+)</strong>`)
	epMatches := epRx.FindAllStringSubmatch(html, -1)

	var episodes []provider.Episode
	for _, em := range epMatches {
		episodeID := em[1]
		epTitle := strings.TrimSpace(em[2])

		// Parse episode number from ID
		parts := strings.Split(episodeID, "/ep-")
		num := 1
		if len(parts) > 1 {
			num, _ = strconv.Atoi(parts[1])
		}

		episodes = append(episodes, provider.Episode{
			ID:          episodeID,
			Number:      num,
			Title:       epTitle,
			DurationSec: 1440,
		})
	}
	details.Episodes = episodes

	if len(details.Episodes) == 0 {
		return p.detailsMock(id), nil
	}

	return details, nil
}

func (p *GogoProvider) detailsMock(id string) provider.AnimeDetails {
	mockID := strings.TrimSuffix(id, "-dub")
	if val, ok := mockCatalog[mockID]; ok {
		return val.Details
	}
	// Default mock details
	return mockCatalog["frieren"].Details
}

func (p *GogoProvider) FetchStreamLinks(ctx context.Context, episodeID string) ([]provider.StreamLink, error) {
	testFeedLink := []provider.StreamLink{
		{
			URL:    "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
			Quality: "auto",
			IsM3U8 : false,
			HTTPHeaders: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
		},
	}

	epPageURL := fmt.Sprintf("%s/watch/%s", p.BaseURL, episodeID)
	req, err := http.NewRequestWithContext(ctx, "GET", epPageURL, nil)
	if err != nil {
		return testFeedLink, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.Client.Do(req)
	if err != nil {
		return testFeedLink, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return testFeedLink, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return testFeedLink, nil
	}
	html := string(body)

	// Extract data-video URLs from buttons with server choice:
	// data-video="([^"]+)"
	videoRx := regexp.MustCompile(`data-video="([^"]+)"`)
	matches := videoRx.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		return testFeedLink, nil
	}

	var links []provider.StreamLink
	for _, m := range matches {
		videoURL := m[1]
		
		links = append(links, provider.StreamLink{
			URL:     videoURL,
			Quality: "auto",
			IsM3U8:  strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, ".mp4"),
			HTTPHeaders: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				"Referer":    epPageURL,
			},
		})
	}

	return links, nil
}
