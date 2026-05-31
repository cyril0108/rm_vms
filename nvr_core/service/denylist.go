package service

import (
	"sync"
	"time"
)

// InMemoryDenylist provides a thread-safe cache for revoked JWTs
type InMemoryDenylist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

// NewInMemoryDenylist initializes the cache and starts the background cleanup task
func NewInMemoryDenylist(cleanupInterval time.Duration) *InMemoryDenylist {
	d := &InMemoryDenylist{
		tokens: make(map[string]time.Time),
	}
	
	// Start a background goroutine to sweep expired tokens
	go d.startCleanupLoop(cleanupInterval)
	
	return d
}

// Revoke adds a token's JTI to the denylist until its natural expiration time
func (d *InMemoryDenylist) Revoke(jti string, exp time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tokens[jti] = exp
}

// IsRevoked safely checks if a JTI exists in the denylist and is still active
func (d *InMemoryDenylist) IsRevoked(jti string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	exp, exists := d.tokens[jti]
	if !exists {
		return false
	}
	
	// If it's in the list but the time has passed, it's naturally expired anyway
	return time.Now().Before(exp)
}

// startCleanupLoop periodically purges tokens that have passed their expiration time
func (d *InMemoryDenylist) startCleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		d.mu.Lock()
		now := time.Now()
		for jti, exp := range d.tokens {
			if now.After(exp) {
				delete(d.tokens, jti)
			}
		}
		d.mu.Unlock()
	}
}