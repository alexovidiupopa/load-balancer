package lb

import (
	"sync"
	"time"
)

type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	state        CircuitBreakerState
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
	mu           sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        Closed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Closed:
		return true
	case Open:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = HalfOpen
			return true
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = Closed
	cb.failures = 0
}

func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.maxFailures {
		cb.state = Open
	}
}
