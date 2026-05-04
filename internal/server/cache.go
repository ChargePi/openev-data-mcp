package server

import (
	"context"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type cachedEntry struct {
	contents  []mcp.ResourceContents
	expiresAt time.Time
}

// ResourceCache is a TTL-based resource handler middleware.
// Each unique request URI gets its own cache entry.
type ResourceCache struct {
	mu      sync.RWMutex
	entries map[string]cachedEntry
	ttl     time.Duration
}

func newResourceCache(ttl time.Duration) *ResourceCache {
	return &ResourceCache{
		entries: make(map[string]cachedEntry),
		ttl:     ttl,
	}
}

// Middleware implements mcpserver.ResourceHandlerMiddleware: it serves from
// the cache when a valid entry exists and delegates to next otherwise.
func (c *ResourceCache) Middleware(next mcpserver.ResourceHandlerFunc) mcpserver.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		c.mu.RLock()
		if entry, ok := c.entries[req.Params.URI]; ok && time.Now().Before(entry.expiresAt) {
			c.mu.RUnlock()
			return entry.contents, nil
		}
		c.mu.RUnlock()

		result, err := next(ctx, req)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.entries[req.Params.URI] = cachedEntry{
			contents:  result,
			expiresAt: time.Now().Add(c.ttl),
		}
		c.mu.Unlock()

		return result, nil
	}
}