package lb

import (
	"sync"
	"time"
)

type RateLimiter struct {
	tokens         int
	maxTokens      int
	refillRate     int
	refillInterval time.Duration
	mu             sync.Mutex
}

func NewRateLimiter(maxTokens, refillRate int, refillInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		refillInterval: refillInterval,
	}
	go rl.refillTokens()
	return rl
}

func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(rl.refillInterval)
	for range ticker.C {
		rl.mu.Lock()
		rl.tokens += rl.refillRate
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}
