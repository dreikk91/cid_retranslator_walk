package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter for logging
type RateLimiter struct {
	rate       float64    // tokens per second
	burst      int        // maximum burst size
	tokens     float64    // current tokens
	lastUpdate time.Time  // last token update time
	mu         sync.Mutex // protects all fields
	suppressed int        // count of suppressed messages
}

// NewRateLimiter creates a new rate limiter
// rate: tokens per second (e.g., 1.0 = 1 log per second, 0.5 = 1 log per 2 seconds)
// burst: maximum number of tokens that can accumulate
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
		suppressed: 0,
	}
}

// Allow checks if a log message should be allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.lastUpdate = now

	// Add tokens based on elapsed time
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	// Check if we have a token available
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// RecordSuppressed increments the suppressed message counter
func (rl *RateLimiter) RecordSuppressed() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.suppressed++
}

// GetAndResetSuppressed returns the number of suppressed messages and resets the counter
func (rl *RateLimiter) GetAndResetSuppressed() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	count := rl.suppressed
	rl.suppressed = 0
	return count
}

// GetSuppressed returns the current count of suppressed messages without resetting
func (rl *RateLimiter) GetSuppressed() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.suppressed
}
