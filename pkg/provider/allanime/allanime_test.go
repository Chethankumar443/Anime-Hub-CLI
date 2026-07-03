package allanime

import (
	"context"
	"testing"
)

func TestAllAnimeProvider(t *testing.T) {
	p := NewAllAnimeProvider("")

	if p.ID() != "allanime" {
		t.Errorf("expected ID allanime, got %s", p.ID())
	}

	if p.Name() != "AllAnime" {
		t.Errorf("expected Name AllAnime, got %s", p.Name())
	}

	ctx := context.Background()

	animes, err := p.Search(ctx, "frieren")
	if err != nil {
		t.Fatalf("expected no error searching, got %v", err)
	}
	if len(animes) == 0 {
		t.Fatal("expected at least one anime returned in mock search")
	}

	details, err := p.FetchDetails(ctx, "frieren")
	if err != nil {
		t.Fatalf("expected no error fetching details, got %v", err)
	}
	if len(details.Episodes) == 0 {
		t.Fatal("expected episodes in details")
	}

	links, err := p.FetchStreamLinks(ctx, "frieren-episode-1")
	if err != nil {
		t.Fatalf("expected no error fetching stream links, got %v", err)
	}
	if len(links) == 0 {
		t.Fatal("expected at least one stream link")
	}
}
