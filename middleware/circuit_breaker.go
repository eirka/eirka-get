package middleware

import (
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState uint32

const (
	// StateClosed means the circuit is closed and requests flow normally
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit is open and requests bypass cache
	StateOpen
	// StateHalfOpen means we're testing if the circuit can be closed again
	StateHalfOpen
	// StateTest is a test state used for test coverage
	StateTest CircuitBreakerState = 99
)

// CircuitBreakerConfig holds the configuration for the circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures required to open the circuit
	FailureThreshold uint32
	// ResetTimeout is the duration the circuit stays open before moving to half-open
	ResetTimeout time.Duration
	// HalfOpenMaxRequests is the number of requests allowed through when half-open
	HalfOpenMaxRequests uint32
}

// DefaultCircuitBreakerConfig provides sensible defaults for the circuit breaker
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	FailureThreshold:    5,
	ResetTimeout:        10 * time.Second,
	HalfOpenMaxRequests: 3,
}

// CacheCircuitBreaker implements a simple circuit breaker pattern for Redis cache
type CacheCircuitBreaker struct {
	state           uint32
	mutex           sync.RWMutex
	config          CircuitBreakerConfig
	failures        uint32
	lastStateChange time.Time
	halfOpenCount   uint32
}

// NewCircuitBreaker creates a new circuit breaker with default configuration
func NewCircuitBreaker() *CacheCircuitBreaker {
	return NewCircuitBreakerWithConfig(DefaultCircuitBreakerConfig)
}

// NewCircuitBreakerWithConfig creates a new circuit breaker with the given configuration
func NewCircuitBreakerWithConfig(config CircuitBreakerConfig) *CacheCircuitBreaker {
	return &CacheCircuitBreaker{
		state:           uint32(StateClosed),
		config:          config,
		failures:        0,
		lastStateChange: time.Now(),
	}
}

// State returns the current state of the circuit breaker
func (cb *CacheCircuitBreaker) State() CircuitBreakerState {
	return CircuitBreakerState(atomic.LoadUint32(&cb.state))
}

// RecordSuccess records a successful Redis operation
func (cb *CacheCircuitBreaker) RecordSuccess() {
	state := cb.State()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Reset failure counter on success
	cb.failures = 0

	// If we're half-open and get a success, close the circuit
	if state == StateHalfOpen {
		cb.changeState(StateClosed)
	}
}

// RecordFailure records a failed Redis operation
func (cb *CacheCircuitBreaker) RecordFailure() {
	state := cb.State()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Increment failure counter
	newFailures := cb.failures + 1
	cb.failures = newFailures

	// Open circuit if we hit the threshold and circuit is currently closed
	if state == StateClosed && newFailures >= cb.config.FailureThreshold {
		cb.changeState(StateOpen)
	}
}

// AllowRequest checks if a request should use Redis cache or bypass it
func (cb *CacheCircuitBreaker) AllowRequest() bool {
	state := cb.State()

	switch state {
	case StateClosed:
		// When closed, all requests go through the normal caching flow
		return true

	case StateOpen:
		// When open, check if it's time to try half-open state
		cb.mutex.RLock()
		elapsed := time.Since(cb.lastStateChange)
		cb.mutex.RUnlock()

		if elapsed >= cb.config.ResetTimeout {
			// Try moving to half-open
			cb.mutex.Lock()
			if CircuitBreakerState(cb.state) == StateOpen {
				cb.changeState(StateHalfOpen)
			}
			cb.mutex.Unlock()

			// Allow the first few requests in half-open state
			count := atomic.AddUint32(&cb.halfOpenCount, 1)
			return count <= cb.config.HalfOpenMaxRequests
		}

		// In open state and not ready to test, bypass cache
		return false

	case StateHalfOpen:
		// Only allow a limited number of requests in half-open state
		count := atomic.AddUint32(&cb.halfOpenCount, 1)
		return count <= cb.config.HalfOpenMaxRequests

	default:
		// Unknown state, default to allowing the request
		return true
	}
}

// changeState changes the state of the circuit breaker
func (cb *CacheCircuitBreaker) changeState(newState CircuitBreakerState) {
	atomic.StoreUint32(&cb.state, uint32(newState))
	cb.lastStateChange = time.Now()

	// Reset half-open counter when changing state
	if newState == StateHalfOpen {
		atomic.StoreUint32(&cb.halfOpenCount, 0)
	}
}
