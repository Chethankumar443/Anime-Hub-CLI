package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

type Client struct {
	HTTPClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &Client{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

func (c *Client) DoWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	ua := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	backoff := 500 * time.Millisecond
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		resp, err = c.HTTPClient.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK || (resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests) {
				return resp, nil
			}
			resp.Body.Close()
			err = fmt.Errorf("temporary server error: status %d", resp.StatusCode)
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, err)
}

func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.DoWithRetry(req, 3)
}
