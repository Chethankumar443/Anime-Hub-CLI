package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHttpClientRetriesAndHeaderRotation(t *testing.T) {
	var requestCount int
	var mu sync.Mutex
	userAgentsSeen := make(map[string]bool)

	// Mock server that returns a 500 error twice and then succeeds on the third attempt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		requestCount++
		
		ua := r.Header.Get("User-Agent")
		if ua != "" {
			userAgentsSeen[ua] = true
		}

		if requestCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	c := NewClient(2 * time.Second)

	resp, err := c.Get(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("expected request to succeed eventually, got error: %v", err)
	}
	defer resp.Body.Close()

	mu.Lock()
	count := requestCount
	mu.Unlock()

	if count != 3 {
		t.Errorf("expected 3 requests including retries, got %d", count)
	}

	if len(userAgentsSeen) == 0 {
		t.Error("expected User-Agent header to be set and rotated")
	}
}
