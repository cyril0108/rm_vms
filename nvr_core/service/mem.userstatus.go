package service

import (
	"sync"
	"time"
)

// UserStatusMap provides a thread-safe in-memory cache to track user sessions
// and enforce instant permission revocation.
type UserStatusMap struct {
	mu       sync.RWMutex
	statuses map[int64]int64 // Key: UserID, Value: Unix Timestamp of last update
}

// NewUserStatusMap initializes the thread-safe map
func NewUserStatusMap() *UserStatusMap {
	return &UserStatusMap{
		statuses: make(map[int64]int64),
	}
}

// Login registers a user in the allowlist with the current timestamp.
// Call this when a user logs in via credentials OR successfully refreshes their token.
func (m *UserStatusMap) Login(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[userID] = time.Now().Unix()
}

// UpdatePermissions instantly invalidates all currently active JWTs for a user
// by updating their timestamp. Call this when an admin changes their role.
func (m *UserStatusMap) UpdatePermissions(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update the timestamp to now. Any JWT issued before this exact second will be rejected.
	m.statuses[userID] = time.Now().Unix()
}

// Logout completely removes a user from the active allowlist.
// Call this on explicit logout or account deactivation.
func (m *UserStatusMap) LogoutAll(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.statuses, userID)
}

// IsValid checks if a user is on the allowlist and if their JWT is still fresh.
func (m *UserStatusMap) IsValid(userID int64, tokenIssuedAt int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lastUpdate, exists := m.statuses[userID]
	if !exists {
		// User is not on the allowlist (logged out or server rebooted)
		return false
	}

	// If the token was issued BEFORE the last update, it is stale and invalid.
	// We use >= to ensure tokens issued at the exact same second as the update remain valid.
	return tokenIssuedAt >= lastUpdate
}