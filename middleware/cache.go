package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/redis"
)

// Group is the global singleflight group for cache requests
var Group = singleflight.Group{}

// CircuitBreaker is the cache circuit breaker
var CircuitBreaker = NewCircuitBreaker()

// Cache is a middleware that implements Redis caching with singleflight pattern and
// circuit breaker for resilience. It works as follows:
//
//  1. First checks if the circuit breaker allows using Redis (bypasses cache if Redis is failing)
//  2. When a request comes in, it checks if the response is in Redis cache
//  3. If found in cache, it serves the cached response immediately
//  4. If not in cache, it uses singleflight to ensure only ONE database query is made
//     regardless of how many concurrent requests are trying to access the same resource
//  5. All concurrent requests for the same resource wait for the first one to complete
//  6. Once the data is retrieved, it's cached in Redis and returned to all waiting clients
//  7. Redis failures are tracked by the circuit breaker which will temporarily bypass
//     the cache if Redis is experiencing problems
//
// This approach significantly reduces database load under high concurrency while
// maintaining responsiveness for clients even when Redis is experiencing issues.
func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set default cached status for analytics middleware
		c.Set("cached", false)
		// Set circuit breaker state for analytics/monitoring
		c.Set("circuitState", CircuitBreaker.State())

		// Skip caching for requests with query parameters
		// This ensures dynamic queries aren't incorrectly cached
		if c.Request.URL.RawQuery != "" {
			c.Next()
			return
		}

		// Parse the request path to generate the cache key
		// Example: "/index/1/2" becomes ["index", "1", "2"]
		request := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")

		// Skip caching for empty paths
		if len(request) == 0 {
			c.Next()
			return
		}

		// Get the base key name from the first path segment
		// This maps to Redis hash structures ("index", "thread", etc.)
		// These are the current controllers that should be cached:
		// controllers/directory.go
		// controllers/favorited.go
		// controllers/image.go
		// controllers/imageboards.go
		// controllers/index.go
		// controllers/new.go
		// controllers/popular.go
		// controllers/post.go
		// controllers/tag.go
		// controllers/tags.go
		// controllers/tagtypes.go
		// controllers/thread.go

		key := redis.NewKey(request[0])
		if key == nil {
			// If the key type isn't recognized, bypass caching
			c.Next()
			return
		}

		// Set the full key with all path parameters
		// Example: For "/index/1/2", key becomes "index" with field "1:2"
		key = key.SetKey(request[1:]...)

		// Generate a unique singleflight key for request deduplication
		// This identifies identical requests that should share the same database query
		sfKey := strings.Join(append([]string{request[0]}, request[1:]...), ":")

		// Get the current circuit state and whether the breaker allows this request
		circuitState := CircuitBreaker.State()
		allowRequest := CircuitBreaker.AllowRequest()

		// Set analytics flag about the circuit state
		c.Set("circuitBreakerState", circuitState)

		// If circuit breaker does not allow using Redis, bypass Redis completely
		// This properly respects both open state and the limited request count in half-open state
		if !allowRequest {
			c.Set("circuitBreakerActive", true)
			c.Next()
			return
		}

		// If we're in half-open state and allowed through, this is a test request
		// This allows monitoring if the test request succeeds or fails
		if circuitState == StateHalfOpen {
			c.Set("circuitTesting", true)
		}

		// -------------------------------------------------------------------------
		// STEP 1: Check if the response is already in Redis cache
		// -------------------------------------------------------------------------
		result, err := key.Get()

		// Handle case where the key couldn't be constructed properly
		if err == redis.ErrKeyNotSet {
			c.JSON(e.ErrorMessage(e.ErrInvalidParam))
			c.Error(err).SetMeta("Cache.KeyNotSet")
			c.Abort()
			return
		}

		// CACHE HIT: If the result is in cache, serve it and stop processing
		if err == nil {
			// Record success with circuit breaker
			CircuitBreaker.RecordSuccess()

			c.Set("cached", true)
			c.Data(http.StatusOK, "application/json", result)
			c.Abort()
			return
		}

		// Log any unexpected Redis errors and record failure with circuit breaker
		if err != redis.ErrCacheMiss {
			c.Error(err).SetMeta("Cache.Redis.Get")

			// Record Redis failure with circuit breaker
			CircuitBreaker.RecordFailure()
		}

		// -------------------------------------------------------------------------
		// STEP 2: Handle cache miss with singleflight pattern
		// -------------------------------------------------------------------------

		// Tell the controller this is a cache miss so it knows to use the callback
		c.Set("cacheMiss", true)

		// Set a timeout to avoid hanging indefinitely on failed requests
		// In production use 10 seconds, but for tests check for a test timeout
		requestTimeout := 10 * time.Second
		// Allow tests to override the timeout for faster test execution
		if testTimeout, exists := c.Get("testTimeout"); exists {
			requestTimeout = testTimeout.(time.Duration)
		}

		// Use singleflight to deduplicate concurrent requests for the same resource
		// This ensures only ONE database query is made regardless of concurrent request count
		data, err, shared := Group.Do(sfKey, func() (interface{}, error) {
			// Create channels for the controller to communicate its results back to us
			resultChan := make(chan []byte, 1)
			errorChan := make(chan error, 1)

			// Set a callback that the controller will use to pass data back to the middleware
			// The controller calls this with the JSON data after querying the database
			c.Set("setDataCallback", func(data []byte, controllerErr error) {
				if controllerErr != nil {
					errorChan <- controllerErr
					return
				}

				// Validate that the controller returned proper JSON before caching it
				if !json.Valid(data) {
					errorChan <- errors.New("invalid JSON from controller")
					return
				}

				// Only attempt to cache if circuit breaker still allows it
				// This ensures we respect the circuit breaker's decision on using Redis
				if CircuitBreaker.AllowRequest() {
					// Store the result in Redis cache for future requests
					if err := key.Set(data); err != nil {
						// If caching fails, we can still return the data to the client
						// but we log the error for monitoring and record the failure
						c.Error(err).SetMeta("Cache.Redis.Set")
						CircuitBreaker.RecordFailure()
					} else {
						// Record successful cache operation which will close the circuit
						// if we're in half-open state
						CircuitBreaker.RecordSuccess()
					}
				}

				// Send the data back to the singleflight function
				resultChan <- data
			})

			// Execute the next middleware/controller in the chain
			// This will trigger the database query in the controller
			c.Next()

			// If the controller explicitly set an error flag, report it
			if _, ok := c.Get("controllerError"); ok {
				return nil, fmt.Errorf("controller returned an error")
			}

			// Wait for the controller to send data through our callback
			// or timeout after the request timeout
			select {
			case data := <-resultChan:
				return data, nil
			case err := <-errorChan:
				return nil, err
			case <-time.After(requestTimeout):
				return nil, fmt.Errorf("controller timed out after %v", requestTimeout)
			}
		})

		// Log whether this request was deduplicated by singleflight
		if shared {
			c.Set("sharedRequest", true)
		}

		// Handle any errors from the singleflight execution
		if err != nil {
			c.Error(err).SetMeta("Cache.SingleFlight")
			// Avoid sending multiple responses if the controller already sent one
			if c.Writer.Written() {
				c.Abort()
				return
			}
			c.JSON(e.ErrorMessage(e.ErrInternalError))
			c.Abort()
			return
		}

		// Get the JSON data returned by the singleflight function
		jsonData, ok := data.([]byte)
		if !ok {
			c.Error(errors.New("invalid data type from singleflight")).SetMeta("Cache.InvalidData")
			c.JSON(e.ErrorMessage(e.ErrInternalError))
			c.Abort()
			return
		}

		// Only write the response if the controller hasn't already done so
		// This handles the case where the controller may have written directly to the client
		if !c.Writer.Written() {
			c.Data(http.StatusOK, "application/json", jsonData)
		}

		// Stop further middleware execution
		c.Abort()
	}
}
