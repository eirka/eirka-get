package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strings"

	"github.com/eirka/eirka-libs/redis"
)

var (
	expire        = 600
	RedisKeyIndex = make(map[string]RedisKey)
	RedisKeys     = []RedisKey{
		{base: "index", fieldcount: 1, hash: true, expire: false},
		{base: "image", fieldcount: 1, hash: true, expire: false},
		{base: "tags", fieldcount: 1, hash: true, expire: false},
		{base: "tag", fieldcount: 2, hash: true, expire: true},
		{base: "thread", fieldcount: 2, hash: true, expire: false},
		{base: "post", fieldcount: 2, hash: true, expire: false},
		{base: "directory", fieldcount: 1, hash: false, expire: false},
		{base: "favorited", fieldcount: 1, hash: false, expire: true},
		{base: "new", fieldcount: 1, hash: false, expire: true},
		{base: "popular", fieldcount: 1, hash: false, expire: true},
		{base: "imageboards", fieldcount: 1, hash: false, expire: true},
	}
	cache = redis.RedisCache
)

func init() {
	// key index map
	for _, key := range RedisKeys {
		RedisKeyIndex[key.base] = key
	}
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

		// Trim leading / from path and split
		params := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")

		// get the keyname
		key := RedisKeyIndex[params[0]]
		if key == nil {
			c.Next()
			return
		}

		// set the key minus the base
		key.SetKey(params[1:])

		result, err = key.Get()
		if err == redis.ErrCacheMiss {
			// go to the controller
			c.Next()

			// Check if there was an error from the controller
			_, controllerError := c.Get("controllerError")
			if controllerError {
				c.Abort()
				return
			}

			err = key.Set(c.MustGet("data").([]byte))
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

		} else if err != nil {
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

type RedisKey struct {
	base       string
	fieldcount int
	hash       bool
	expire     bool
	key        string
	hashid     uint
}

func (r *RedisKey) SetKey(ids ...string) {

	// create our key
	r.key = strings.Join([]string{r.base, strings.Join(ids[:r.fieldcount], ":")}, ":")

	// get our hash id
	if r.hash {
		r.hashid = ids[r.fieldcount:]
	}

	return
}

func (r *RedisKey) Get() (result []byte, err error) {

	if r.hash {
		return cache.HGet(r.key, r.hashid)
	} else {
		return cache.Get(r.key)
	}

	return
}

func (r *RedisKey) Set(data []byte) (err error) {

	if r.hash {
		return cache.HMSet(r.key, r.hashid, data)
	} else {
		if r.expire {
			return cache.SetEx(r.key, expire, data)
		} else {
			return cache.Set(r.key, data)
		}
	}

	return
}
