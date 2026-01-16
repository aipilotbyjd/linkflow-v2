package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests")
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxFailures   int
	resetTimeout  time.Duration
	halfOpenMax   int

	mu            sync.Mutex
	state         State
	failures      int
	successes     int
	lastFailure   time.Time
	halfOpenCalls int
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures  int
	ResetTimeout time.Duration
	HalfOpenMax  int
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
		HalfOpenMax:  3,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  config.MaxFailures,
		resetTimeout: config.ResetTimeout,
		halfOpenMax:  config.HalfOpenMax,
		state:        StateClosed,
	}
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.allowRequest(); err != nil {
		return err
	}

	err := fn()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenCalls = 0
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		if cb.halfOpenCalls >= cb.halfOpenMax {
			return ErrTooManyRequests
		}
		cb.halfOpenCalls++
		return nil
	}

	return nil
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.successes = 0
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures && cb.state == StateClosed {
			cb.state = StateOpen
		} else if cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
	} else {
		cb.successes++

		if cb.state == StateHalfOpen && cb.successes >= cb.halfOpenMax {
			cb.state = StateClosed
			cb.failures = 0
		}
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return map[string]interface{}{
		"name":        cb.name,
		"state":       cb.state.String(),
		"failures":    cb.failures,
		"successes":   cb.successes,
		"last_failure": cb.lastFailure,
	}
}
