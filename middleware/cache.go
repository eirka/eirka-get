package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"

	u "github.com/techjanitor/pram-get/utils"
)

// Cache will check for the key in Redis and serve it. If not found, it will
// take the marshalled JSON from the controller and set it in Redis
func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {
		var result []byte
		var err error

		// Get request path
		path := c.Request.URL.Path

		// Trim leading / from path and split
		params := strings.Split(strings.Trim(path, "/"), "/")

		// Make key from path
		key := redisKey{}
		key.expireKey(params[0])
		key.generateKey(params...)

		// Initialize cache handle
		cache := u.RedisCache

		if key.Hash {
			// Check to see if there is already a key we can serve
			result, err = cache.HGet(key.Key, key.Field)
			if err == u.ErrCacheMiss {
				c.Next()

				// Check if there was an error from the controller
				controllerError, _ := c.Get("controllerError")
				if controllerError != nil {
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

			}
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

		}

		if !key.Hash {
			// Check to see if there is already a key we can serve
			result, err = cache.Get(key.Key)
			if err == u.ErrCacheMiss {
				c.Next()

				// Check if there was an error from the controller
				controllerError, _ := c.Get("controllerError")
				if controllerError != nil {
					c.Abort()
					return
				}

				// Get data from controller
				data := c.MustGet("data").([]byte)

				if key.Expire {

					// Set output to cache
					err = cache.SetEx(key.Key, 60, data)
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

			}
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

		}

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
	case len(params) == 1 || len(params) == 2:
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

	keyList := map[string]bool{
		"image": true,
		"pram":  true,
		"tag":   true,
	}

	if keyList[strings.ToLower(key)] {
		r.Expire = true
	}

	return

}
