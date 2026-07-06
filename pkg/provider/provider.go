package provider

import (
	"errors"
	"log"
)

type Anime struct {
	ID    string
	Title string
}

type Episode struct {
	ID     string
	Number int
}

// AnimeProvider abstracts individual content scraper implementations.
type AnimeProvider interface {
	Search(query string) ([]Anime, error)
	GetEpisodes(animeID string) ([]Episode, error)
	GetStreamURL(episodeID string, lang string) (string, error)
}

// FallbackManager orchestrates sequential provider evaluation.
type FallbackManager struct {
	providers []AnimeProvider
}

// NewFallbackManager creates a new manager with the given providers
func NewFallbackManager(providers ...AnimeProvider) *FallbackManager {
	return &FallbackManager{
		providers: providers,
	}
}

func (fm *FallbackManager) Search(query string) ([]Anime, error) {
	for _, p := range fm.providers {
		results, err := p.Search(query)
		if err == nil {
			return results, nil
		}
		log.Printf("Primary provider Search failure: %v. Escalating to alternative route...", err)
	}
	return nil, errors.New("exhausted all available provider routes without resolving search")
}

func (fm *FallbackManager) GetEpisodes(animeID string) ([]Episode, error) {
	for _, p := range fm.providers {
		results, err := p.GetEpisodes(animeID)
		if err == nil {
			return results, nil
		}
		log.Printf("Primary provider GetEpisodes failure: %v. Escalating to alternative route...", err)
	}
	return nil, errors.New("exhausted all available provider routes without resolving episodes")
}

func (fm *FallbackManager) GetStreamURL(episodeID, lang string) (string, error) {
	for _, p := range fm.providers {
		url, err := p.GetStreamURL(episodeID, lang)
		if err == nil {
			return url, nil
		}
		log.Printf("Primary provider GetStreamURL failure: %v. Escalating to alternative route...", err)
	}
	return "", errors.New("exhausted all available provider routes without resolving stream")
}
