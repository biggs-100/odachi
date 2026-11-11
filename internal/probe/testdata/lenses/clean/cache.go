package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// attempts is how many times a fetch is tried before it gives up.
const attempts = 3

// fetchTimeout bounds one remote call so a peer that never answers cannot hold
// the caller forever.
const fetchTimeout = 5 * time.Second

// maxResponseSize is the largest body a fetch accepts, in bytes.
const maxResponseSize = 1 << 20

// newCache returns a cache that fetches through a bounded HTTP client.
func newCache() *cache {
	return &cache{client: &http.Client{Timeout: fetchTimeout}}
}

type cache struct {
	mu     sync.Mutex
	m      map[string]string
	client *http.Client
}

func (c *cache) get(k string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[k]
	return v, ok
}

func (c *cache) set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = map[string]string{}
	}
	c.m[k] = v
}

func (c *cache) fetch(ctx context.Context, url string) (string, error) {
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		body, err := c.once(ctx, url)
		if err == nil {
			return body, nil
		}
		last = err
	}
	return "", fmt.Errorf("fetch after %d attempts: %w", attempts, last)
}

func (c *cache) once(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build the request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("remote answered %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return "", fmt.Errorf("read the response body: %w", err)
	}
	if len(body) > maxResponseSize {
		return "", fmt.Errorf("response body is larger than %d bytes", maxResponseSize)
	}
	return string(body), nil
}
