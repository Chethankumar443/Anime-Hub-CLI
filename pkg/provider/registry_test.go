package provider

import (
	"context"
	"testing"
)

type dummyProvider struct{}

func (d *dummyProvider) ID() string {
	return "dummy"
}

func (d *dummyProvider) Name() string {
	return "Dummy"
}

func (d *dummyProvider) Search(ctx context.Context, query string) ([]Anime, error) {
	return nil, nil
}

func (d *dummyProvider) FetchDetails(ctx context.Context, id string) (AnimeDetails, error) {
	return AnimeDetails{}, nil
}

func (d *dummyProvider) FetchStreamLinks(ctx context.Context, episodeID string) ([]StreamLink, error) {
	return nil, nil
}

func TestRegistry(t *testing.T) {
	reg := GetRegistry()
	p := &dummyProvider{}
	reg.Register(p)

	got, err := reg.Get("dummy")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID() != "dummy" {
		t.Errorf("expected dummy ID, got %s", got.ID())
	}

	list := reg.List()
	found := false
	for _, lp := range list {
		if lp.ID() == "dummy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dummy to be in registry list")
	}
}
