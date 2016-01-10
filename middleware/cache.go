package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"

	"github.com/eirka/eirka-libs/redis"
)

// REDIS KEY STRUCTURE
//
// HASH:
// "thread:1:45"
// "post:1:45"
// "tag:1:4"
// "index:1"
// "image:1"
//
// REGULAR:
// "directory:1"
// "tags:1"
// "tagtypes"
// "pram"

const (
	// key expire seconds
	ExpireSeconds = 600
)

// list of keys to expire
var cacheKeyList = map[string]bool{
	"imageboards": true,
	"popular":     true,
	"new":         true,
	"favorited":   true,
	"tag":         true,
}

// Cache will check for the key in Redis and serve it. If not found, it will
// take the marshalled JSON from the controller and set it in Redis
func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {
		var result []byte
		var err error

		// bool for analytics middleware
		c.Set("cached", false)

		// break cache if there is a query
		if c.Request.URL.RawQuery != "" {
			c.Next()
			return
		}

		// Get request path
		path := c.Request.URL.Path

		// Trim leading / from path and split
		params := strings.Split(strings.Trim(path, "/"), "/")

		// Make key from path
		key := redisKey{}
		key.expireKey(params[0])
		key.generateKey(params...)

		// Initialize cache handle
		cache := redis.RedisCache

		if key.Hash {
			// Check to see if there is already a key we can serve
			result, err = cache.HGet(key.Key, key.Field)
			if err == redis.ErrCacheMiss {

				c.Next()

				// Check if there was an error from the controller
				_, controllerError := c.Get("controllerError")
				if controllerError {
					c.Abort()
					return
				}

				// Get data from controller
				data := c.MustGet("data").([]byte)

				// Set output to cache
				err = cache.HMSet(key.Key, key.Field, data)
				if err != nil {
					c.Error(err)
					c.Abort()
					return
				}

				return

			} else if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

		}

		if !key.Hash {
			// Check to see if there is already a key we can serve
			result, err = cache.Get(key.Key)
			if err == redis.ErrCacheMiss {

				c.Next()

				// Check if there was an error from the controller
				_, controllerError := c.Get("controllerError")
				if controllerError {
					c.Abort()
					return
				}

				// Get data from controller
				data := c.MustGet("data").([]byte)

				if key.Expire {

					// Set output to cache
					err = cache.SetEx(key.Key, ExpireSeconds, data)
					if err != nil {
						c.Error(err)
						c.Abort()
						return
					}

				} else {

					// Set output to cache
					err = cache.Set(key.Key, data)
					if err != nil {
						c.Error(err)
						c.Abort()
						return
					}

				}

				return

			} else if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

		}

		// if we made it this far then the page was cached
		c.Set("cached", true)

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(result)
		c.Abort()

	}

}

type redisKey struct {
	Key    string
	Field  string
	Hash   bool
	Expire bool
}

// Will take the params from the request and turn them into a key
func (r *redisKey) generateKey(params ...string) {

	var keys []string

	switch {
	// keys like pram, directory, and tags
	case len(params) <= 2:
		r.Key = strings.Join(params, ":")
	// index and image
	case len(params) == 3:
		keys = append(keys, params[0], params[1])
		r.Field = params[2]
		r.Hash = true
		r.Key = strings.Join(keys, ":")
	// thread, post, and tag
	case len(params) == 4:
		keys = append(keys, params[0], params[1], params[2])
		r.Field = params[3]
		r.Hash = true
		r.Key = strings.Join(keys, ":")
	}

	return

}

// Check if key should be expired
func (r *redisKey) expireKey(key string) {

	if cacheKeyList[strings.ToLower(key)] {
		r.Expire = true
	}

	return

}
