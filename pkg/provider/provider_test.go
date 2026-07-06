package provider

import (
	"errors"
	"testing"
)

type mockProvider struct {
	SearchFunc       func(query string) ([]Anime, error)
	GetEpisodesFunc  func(animeID string) ([]Episode, error)
	GetStreamURLFunc func(episodeID string, lang string) (string, error)
}

func (m *mockProvider) Search(query string) ([]Anime, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(query)
	}
	return nil, nil
}

func (m *mockProvider) GetEpisodes(animeID string) ([]Episode, error) {
	if m.GetEpisodesFunc != nil {
		return m.GetEpisodesFunc(animeID)
	}
	return nil, nil
}

func (m *mockProvider) GetStreamURL(episodeID string, lang string) (string, error) {
	if m.GetStreamURLFunc != nil {
		return m.GetStreamURLFunc(episodeID, lang)
	}
	return "", nil
}

func TestFallbackManager_GetStreamURL_ExpiredFallback(t *testing.T) {
	callCount := 0
	p1 := &mockProvider{
		GetStreamURLFunc: func(episodeID string, lang string) (string, error) {
			callCount++
			if callCount == 1 {
				return "https://valid-stream-url.com/play", nil
			}
			return "", errors.New("404 Not Found")
		},
	}

	p2 := &mockProvider{
		GetStreamURLFunc: func(episodeID string, lang string) (string, error) {
			return "https://fallback-stream-url.com/play", nil
		},
	}

	fm := NewFallbackManager(p1, p2)

	// First call should resolve successfully using Provider 1
	url1, err := fm.GetStreamURL("ep1", "sub")
	if err != nil {
		t.Fatalf("expected no error on first call, got: %v", err)
	}
	if url1 != "https://valid-stream-url.com/play" {
		t.Errorf("expected p1 url, got: %s", url1)
	}

	// Second call: Provider 1 fails (expired / 404), FallbackManager should try Provider 2 and succeed
	url2, err := fm.GetStreamURL("ep1", "sub")
	if err != nil {
		t.Fatalf("expected no error on second call (with fallback), got: %v", err)
	}
	if url2 != "https://fallback-stream-url.com/play" {
		t.Errorf("expected fallback to p2 url, got: %s", url2)
	}
}

func TestFallbackManager_Search_Fallback(t *testing.T) {
	p1 := &mockProvider{
		SearchFunc: func(query string) ([]Anime, error) {
			return nil, errors.New("search failed")
		},
	}
	p2 := &mockProvider{
		SearchFunc: func(query string) ([]Anime, error) {
			return []Anime{{ID: "anime-1", Title: "Test Anime"}}, nil
		},
	}

	fm := NewFallbackManager(p1, p2)
	results, err := fm.Search("test")
	if err != nil {
		t.Fatalf("expected search to succeed, got error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "anime-1" {
		t.Errorf("expected results from p2, got: %v", results)
	}
}

func TestFallbackManager_GetEpisodes_Fallback(t *testing.T) {
	p1 := &mockProvider{
		GetEpisodesFunc: func(animeID string) ([]Episode, error) {
			return nil, errors.New("episodes failed")
		},
	}
	p2 := &mockProvider{
		GetEpisodesFunc: func(animeID string) ([]Episode, error) {
			return []Episode{{ID: "ep-1", Number: 1}}, nil
		},
	}

	fm := NewFallbackManager(p1, p2)
	results, err := fm.GetEpisodes("anime-1")
	if err != nil {
		t.Fatalf("expected episodes to succeed, got error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "ep-1" {
		t.Errorf("expected results from p2, got: %v", results)
	}
}
