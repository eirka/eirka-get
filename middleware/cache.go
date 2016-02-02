package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"

	"github.com/eirka/eirka-libs/redis"
)

// Cache will check for the key in Redis and serve it. If not found, it will
// take the marshalled JSON from the controller and set it in Redis
func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {

		// bool for analytics middleware
		c.Set("cached", false)

		// break cache if there is a query
		if c.Request.URL.RawQuery != "" {
			c.Next()
			return
		}

		// Trim leading / from path and split
		params := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")

		// get the keyname
		key, ok := redis.RedisKeyIndex[params[0]]
		if !ok {
			c.Next()
			return
		}

		// set the key minus the base
		key.SetKey(params[1:]...)

		// check the cache
		result, err := key.Get()
		if result == nil {
			// go to the controller if it wasnt found
			c.Next()

			// Check if there was an error from the controller
			_, controllerError := c.Get("controllerError")
			if controllerError {
				c.Abort()
				return
			}

			// set the data returned from the controller
			err = key.Set(c.MustGet("data").([]byte))
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

			// unlock shared mutex
			redis.RedisCache.Unlock()

			return

		}
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// if we made it this far then the page was cached
		c.Set("cached", true)

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(result)
		c.Abort()
		return
	}

}
