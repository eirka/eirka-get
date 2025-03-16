package middleware

import (
	"errors"
	"testing"
	"time"

	"github.com/eirka/eirka-libs/redis"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Use performRequest from analytics_test.go

func TestCache(t *testing.T) {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(Cache())

	router.GET("/index/:ib/:page", func(c *gin.Context) {
		// Properly use the callback mechanism that the cache middleware expects
		if _, ok := c.Get("cacheMiss"); ok {
			// Get the data callback function and use it to send the data
			if callback, ok := c.Get("setDataCallback"); ok {
				callback.(func([]byte, error))([]byte("cache data"), nil)
			}
		}

		c.String(200, "not cached")
	})

	router.GET("/thread/:ib/:thread/:page", func(c *gin.Context) {
		c.Set("controllerError", true)
		c.String(500, "BAD!!")
	})

	router.GET("/nocache/:id", func(c *gin.Context) {
		c.String(200, "OK")
	})

	redis.NewRedisMock()

	// Reset the circuit breaker before tests
	CircuitBreaker = NewCircuitBreaker()

	// break cache with a query string
	query := performRequest(router, "GET", "/index/1/2?what=2")

	assert.Equal(t, query.Code, 200, "HTTP request code should match")

	// a path with no matching cache string
	nocache := performRequest(router, "GET", "/nocache/1")

	assert.Equal(t, nocache.Code, 200, "HTTP request code should match")

	// a bad key
	bad := performRequest(router, "GET", "/thread/1/1")

	assert.Equal(t, bad.Code, 400, "HTTP request code should match")

	// a controller that returns a controllerError
	redis.Cache.Mock.Command("HGET", "thread:1:1", "1").Expect(nil)

	controllererror := performRequest(router, "GET", "/thread/1/1/1")

	assert.Equal(t, controllererror.Code, 500, "HTTP request code should match")
	assert.Equal(t, controllererror.Body.String(), "BAD!!", "Body should match")

	// get a cached query
	redis.Cache.Mock.Command("HGET", "index:1", "2").Expect("cached")

	cached := performRequest(router, "GET", "/index/1/2")

	assert.Equal(t, cached.Body.String(), "cached", "Body should match")
	assert.Equal(t, cached.Code, 200, "HTTP request code should match")

	// get a cache miss and set data
	redis.Cache.Mock.Command("HGET", "index:1", "3")
	redis.Cache.Mock.Command("HMSET", "index:1", "3", []byte("cache data"))

	getdata := performRequest(router, "GET", "/index/1/3")

	assert.Equal(t, getdata.Body.String(), "not cached", "Body should match")
	assert.Equal(t, getdata.Code, 200, "HTTP request code should match")

	// a redis get error
	redis.Cache.Mock.Command("HGET", "index:1", "4").ExpectError(errors.New("get error"))

	badget := performRequest(router, "GET", "/index/1/4")

	assert.Equal(t, badget.Body.String(), "not cached", "Body should match")
	assert.Equal(t, badget.Code, 200, "HTTP request code should match")

	// a redis set error
	redis.Cache.Mock.Command("HGET", "index:1", "4").Expect(nil)
	redis.Cache.Mock.Command("HMSET", "index:1", "4", []byte("cache data")).ExpectError(errors.New("set error"))

	badset := performRequest(router, "GET", "/index/1/4")

	assert.Equal(t, badset.Body.String(), "not cached", "Body should match")
	assert.Equal(t, badset.Code, 200, "HTTP request code should match")
}

func TestCacheCircuitBreaker(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// Create a test circuit breaker with a lower threshold for testing
	testConfig := CircuitBreakerConfig{
		FailureThreshold:    2,
		ResetTimeout:        100 * time.Millisecond,
		HalfOpenMaxRequests: 1,
	}

	// Use a local circuit breaker for testing to avoid affecting other tests
	cb := NewCircuitBreakerWithConfig(testConfig)

	// Test initial state
	assert.Equal(t, StateClosed, cb.State(), "Initial state should be Closed")

	// Test failure count increment
	cb.RecordFailure()
	assert.Equal(t, StateClosed, cb.State(), "State should still be Closed after 1 failure")

	// Test open circuit after threshold
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State(), "State should be Open after 2 failures")

	// Test circuit bypassing
	assert.False(t, cb.AllowRequest(), "Request should be denied when circuit is Open")

	// Test half-open after timeout
	time.Sleep(150 * time.Millisecond)
	assert.True(t, cb.AllowRequest(), "First request should be allowed in Half-Open state")
	assert.Equal(t, StateHalfOpen, cb.State(), "State should be Half-Open after timeout")

	// Test request limiting in half-open state
	assert.False(t, cb.AllowRequest(), "Second request should be denied in Half-Open state")

	// Test circuit closing after success
	cb.RecordSuccess()
	assert.Equal(t, StateClosed, cb.State(), "State should be Closed after success in Half-Open state")

	// Test failure count reset on success
	cb.RecordFailure()
	assert.Equal(t, StateClosed, cb.State(), "State should still be Closed after 1 failure")
	cb.RecordSuccess()
	cb.RecordFailure()
	assert.Equal(t, StateClosed, cb.State(), "State should still be Closed after success resets failure count")
}

// TestHalfOpenRequests tests the behavior of HalfOpenMaxRequests
func TestHalfOpenRequests(t *testing.T) {
	// Configure a circuit breaker with low reset timeout
	config := CircuitBreakerConfig{
		FailureThreshold:    1,
		ResetTimeout:        10 * time.Millisecond,
		HalfOpenMaxRequests: 2, // Allow exactly 2 requests in half-open state
	}
	cb := NewCircuitBreakerWithConfig(config)

	// Force it into open state
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State(), "Circuit should be open after failure")

	// Wait for timeout to move to half-open
	time.Sleep(15 * time.Millisecond)

	// First request in half-open should be allowed
	assert.True(t, cb.AllowRequest(), "First request in half-open should be allowed")

	// Second request in half-open should still be allowed
	assert.True(t, cb.AllowRequest(), "Second request in half-open should be allowed")

	// Third request should be denied
	assert.False(t, cb.AllowRequest(), "Third request in half-open should be denied")
}

// TestZeroMaxRequests tests behavior with zero HalfOpenMaxRequests
func TestZeroMaxRequests(t *testing.T) {
	// Configure with zero max requests in half-open
	config := CircuitBreakerConfig{
		FailureThreshold:    1,
		ResetTimeout:        10 * time.Millisecond,
		HalfOpenMaxRequests: 0, // Allow NO requests in half-open state
	}
	cb := NewCircuitBreakerWithConfig(config)

	// Force it into open state
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State(), "Circuit should be open after failure")

	// Wait for timeout to move to half-open
	time.Sleep(15 * time.Millisecond)

	// No requests should be allowed when max is zero
	assert.False(t, cb.AllowRequest(), "No requests should be allowed with zero max requests")
}

// TestDefaultState tests the default state case in AllowRequest
func TestDefaultState(t *testing.T) {
	// Create a circuit breaker
	cb := NewCircuitBreaker()

	// Force its state to be the test state
	cb.changeState(StateTest)

	// Default behavior for unknown state is to allow requests
	assert.True(t, cb.AllowRequest(), "Unknown state should allow requests by default")
}

// TestCacheEmptyPath tests that requests with empty paths bypass caching
func TestCacheEmptyPath(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(Cache())

	// Track if handler was called
	var handlerCalled bool

	router.GET("/", func(c *gin.Context) {
		handlerCalled = true
		c.String(200, "empty")
	})

	redis.NewRedisMock()

	// Make request to empty path
	resp := performRequest(router, "GET", "/")

	// Check handler was called (cache should be bypassed)
	assert.True(t, handlerCalled, "Handler should be called for empty path")
	assert.Equal(t, 200, resp.Code, "Should return 200 status")
	assert.Equal(t, "empty", resp.Body.String(), "Should return correct body")
}

// TestCircuitBreakerRecovery tests the complete recovery cycle of the circuit breaker
// This ensures the circuit breaker properly transitions from open to closed state
// when Redis becomes available again
func TestCircuitBreakerRecovery(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// Use a test configuration with short timeouts for faster testing
	testConfig := CircuitBreakerConfig{
		FailureThreshold:    2,                     // Open after 2 failures
		ResetTimeout:        10 * time.Millisecond, // Try half-open quickly
		HalfOpenMaxRequests: 1,                     // Allow one test request
	}

	// Create a new circuit breaker with test config
	CircuitBreaker = NewCircuitBreakerWithConfig(testConfig)

	router := gin.New()
	router.Use(Cache())

	// Create a test handler that uses the cache middleware correctly
	router.GET("/index/:ib/:page", func(c *gin.Context) {
		// Set a shorter timeout for tests
		c.Set("testTimeout", 50*time.Millisecond)

		// For a cache miss, use the callback to return data
		if _, ok := c.Get("cacheMiss"); ok {
			if callback, ok := c.Get("setDataCallback"); ok {
				callback.(func([]byte, error))([]byte("data from db"), nil)
			}
		}

		// Return non-cached response
		c.String(200, "not cached")
	})

	// Set up Redis mock
	redis.NewRedisMock()

	// 1. First, simulate Redis working properly
	redis.Cache.Mock.Command("HGET", "index:1", "1").Expect("cached data")
	resp := performRequest(router, "GET", "/index/1/1")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "cached data", resp.Body.String())
	assert.Equal(t, StateClosed, CircuitBreaker.State())

	// 2. Now simulate Redis failures to trigger circuit open
	redis.Cache.Mock.Command("HGET", "index:1", "2").ExpectError(errors.New("Redis connection error"))
	resp = performRequest(router, "GET", "/index/1/2")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "not cached", resp.Body.String())

	redis.Cache.Mock.Command("HGET", "index:1", "2").ExpectError(errors.New("Redis connection error"))
	resp = performRequest(router, "GET", "/index/1/2")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "not cached", resp.Body.String())

	// Circuit should now be open
	assert.Equal(t, StateOpen, CircuitBreaker.State())

	// 3. Make more requests - they should bypass Redis completely when circuit is open
	resp = performRequest(router, "GET", "/index/1/3")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "not cached", resp.Body.String())

	// 4. Wait for reset timeout to allow half-open state
	time.Sleep(20 * time.Millisecond)

	// 5. Test half-open failure and back to open state
	// Simulate Redis still being down when the circuit first tries half-open
	redis.Cache.Mock.Command("HGET", "index:1", "4").ExpectError(errors.New("Redis still down"))

	resp = performRequest(router, "GET", "/index/1/4")
	assert.Equal(t, 200, resp.Code)

	// Circuit should go back to open after failure in half-open state
	assert.Equal(t, StateOpen, CircuitBreaker.State())

	// Wait for another reset timeout period
	time.Sleep(20 * time.Millisecond)

	// 6. Now simulate Redis working again - this should allow recovery
	redis.Cache.Mock.Command("HGET", "index:1", "5").Expect(nil) // Cache miss
	redis.Cache.Mock.Command("HMSET", "index:1", "5", []byte("data from db")).Expect("OK")

	resp = performRequest(router, "GET", "/index/1/5")
	assert.Equal(t, 200, resp.Code)

	// First successful operation keeps us in half-open state
	assert.Equal(t, StateHalfOpen, CircuitBreaker.State())

	// Directly record a success to ensure circuit closes
	CircuitBreaker.RecordSuccess()

	// Circuit should now be closed again after the successful Redis operations
	assert.Equal(t, StateClosed, CircuitBreaker.State())

	// 7. Final verification - Redis should be used normally again
	redis.Cache.Mock.Command("HGET", "index:1", "6").Expect("cached again")
	resp = performRequest(router, "GET", "/index/1/6")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "cached again", resp.Body.String())
}
