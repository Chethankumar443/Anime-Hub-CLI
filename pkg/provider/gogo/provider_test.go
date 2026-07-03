package gogo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGogoProviderSearch(t *testing.T) {
	mockSearchHTML := `
	<article class="nv-anime-card nv-browse-card">
		<a class="nv-anime-thumb nv-browse-thumb" href="/watch/frieren-beyond-journeys-end">
			<img src="https://cdn.anizara.store/cover/frieren-beyond-journeys-end.webp" alt="Frieren: Beyond Journey's End" loading="lazy">
		</a>
	</article>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockSearchHTML))
	}))
	defer server.Close()

	prov := NewGogoProvider(server.URL)
	prov.BaseURL = server.URL // Override self-healing redirection for offline test server

	results, err := prov.Search(context.Background(), "frieren")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0]
	if res.ID != "frieren-beyond-journeys-end" {
		t.Errorf("expected ID frieren-beyond-journeys-end, got %s", res.ID)
	}
	if res.Title != "Frieren: Beyond Journey's End" {
		t.Errorf("expected Title Frieren: Beyond Journey's End, got %s", res.Title)
	}
	if res.CoverURL != "https://cdn.anizara.store/cover/frieren-beyond-journeys-end.webp" {
		t.Errorf("expected CoverURL, got %s", res.CoverURL)
	}
}

func TestGogoProviderFetchDetails(t *testing.T) {
	mockDetailsHTML := `
	<h1>Frieren: Beyond Journey's End</h1>
	<div><span>Status</span><strong>Completed</strong></div>
	<div><span>Type</span><strong>TV</strong></div>
	<div><span>Release</span><strong>2023</strong></div>
	<aside class="nv-info-poster">
		<img src="https://cdn.anizara.store/cover/frieren.jpg" alt="" />
	</aside>
	<div class="nv-info-genres">
		<span>Fantasy</span>
		<span>Adventure</span>
	</div>
	<div class="nv-info-synopsis">
		<p>Plot Summary: Mage Frieren and her companions defeated the Demon King...</p>
	</div>
	<a class="nv-info-episode-main" href="/watch/frieren-beyond-journeys-end/ep-1"><strong>Episode 1</strong></a>
	<a class="nv-info-episode-main" href="/watch/frieren-beyond-journeys-end/ep-2"><strong>Episode 2</strong></a>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		
		switch r.URL.Path {
		case "/watch/frieren-beyond-journeys-end":
			_, _ = w.Write([]byte(mockDetailsHTML))
		default:
			_, _ = w.Write([]byte(""))
		}
	}))
	defer server.Close()

	prov := NewGogoProvider(server.URL)
	prov.BaseURL = server.URL // Override self-healing redirect for test

	details, err := prov.FetchDetails(context.Background(), "frieren-beyond-journeys-end")
	if err != nil {
		t.Fatalf("fetch details failed: %v", err)
	}

	if details.Title != "Frieren: Beyond Journey's End" {
		t.Errorf("expected Title Frieren: Beyond Journey's End, got %s", details.Title)
	}
	if details.Format != "TV" {
		t.Errorf("expected Format TV, got %s", details.Format)
	}
	if details.Status != "Completed" {
		t.Errorf("expected Status Completed, got %s", details.Status)
	}
	if details.CoverURL != "https://cdn.anizara.store/cover/frieren.jpg" {
		t.Errorf("expected CoverURL, got %s", details.CoverURL)
	}
	if len(details.Genres) != 2 || details.Genres[0] != "Fantasy" {
		t.Errorf("expected Genres to contain Fantasy and Adventure")
	}

	if len(details.Episodes) != 2 {
		t.Fatalf("expected 2 episodes, got %d", len(details.Episodes))
	}
	if details.Episodes[0].ID != "frieren-beyond-journeys-end/ep-1" || details.Episodes[0].Number != 1 {
		t.Errorf("expected first episode to be number 1, got ID: %s", details.Episodes[0].ID)
	}
	if details.Episodes[1].ID != "frieren-beyond-journeys-end/ep-2" || details.Episodes[1].Number != 2 {
		t.Errorf("expected second episode to be number 2, got ID: %s", details.Episodes[1].ID)
	}
}
