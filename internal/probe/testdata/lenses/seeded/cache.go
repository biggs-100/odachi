package probe

import (
	"context"
	"io"
	"net/http"
	"sync"
)

type cache struct {
	mu sync.Mutex
	m  map[string]string
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
	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			return string(b), nil
		}
	}
}
