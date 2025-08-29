package bot

import (
	"sync"
	"time"
)

// RateLimiter implements per-user rate limiting
type RateLimiter struct {
	limits map[int64]*UserLimit
	mu     sync.RWMutex
	// Cleanup old entries periodically
	cleanupTicker *time.Ticker
	done          chan bool
}

// UserLimit tracks rate limiting for a specific user
type UserLimit struct {
	LastRequest time.Time
	Count       int
	Blocked     bool
	BlockedAt   time.Time
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:        make(map[int64]*UserLimit),
		cleanupTicker: time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
		done:          make(chan bool),
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Allow checks if a user is allowed to make a request
func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.limits[userID]

	// If user doesn't exist, create new limit
	if !exists {
		rl.limits[userID] = &UserLimit{
			LastRequest: now,
			Count:       1,
			Blocked:     false,
		}
		return true
	}

	// Check if user is blocked
	if limit.Blocked {
		// Unblock after 10 minutes
		if now.Sub(limit.BlockedAt) > 10*time.Minute {
			limit.Blocked = false
			limit.Count = 0
		} else {
			return false
		}
	}

	// Reset counter if more than 1 minute has passed
	if now.Sub(limit.LastRequest) > time.Minute {
		limit.Count = 0
		limit.LastRequest = now
	}

	// Increment counter
	limit.Count++
	limit.LastRequest = now

	// Block if too many requests
	if limit.Count > 20 { // Max 20 requests per minute
		limit.Blocked = true
		limit.BlockedAt = now
		return false
	}

	return true
}

// GetUserLimit returns the current limit for a user
func (rl *RateLimiter) GetUserLimit(userID int64) *UserLimit {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.limits[userID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	return &UserLimit{
		LastRequest: limit.LastRequest,
		Count:       limit.Count,
		Blocked:     limit.Blocked,
		BlockedAt:   limit.BlockedAt,
	}
}

// cleanupRoutine periodically removes old entries
func (rl *RateLimiter) cleanupRoutine() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.done:
			return
		}
	}
}

// cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for userID, limit := range rl.limits {
		// Remove entries older than 1 hour
		if now.Sub(limit.LastRequest) > time.Hour {
			delete(rl.limits, userID)
		}
	}
}

// Stop stops the rate limiter and cleans up resources
func (rl *RateLimiter) Stop() {
	rl.cleanupTicker.Stop()
	close(rl.done)
}
