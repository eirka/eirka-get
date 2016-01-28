package middleware

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"
	"strings"

	"github.com/eirka/eirka-libs/redis"
)

var (
	RedisKeyIndex = make(map[string]RedisKey)
	RedisKeys     = []RedisKey{
		{base: "index", fieldcount: 1, hash: true, expire: false},
		{base: "thread", fieldcount: 2, hash: true, expire: false},
		{base: "tag", fieldcount: 2, hash: true, expire: true},
		{base: "image", fieldcount: 1, hash: true, expire: false},
		{base: "post", fieldcount: 2, hash: true, expire: false},
		{base: "tags", fieldcount: 1, hash: true, expire: false},
		{base: "directory", fieldcount: 1, hash: true, expire: false},
		{base: "new", fieldcount: 1, hash: false, expire: true},
		{base: "popular", fieldcount: 1, hash: false, expire: true},
		{base: "favorited", fieldcount: 1, hash: false, expire: true},
		{base: "tagtypes", fieldcount: 0, hash: false, expire: false},
		{base: "imageboards", fieldcount: 0, hash: false, expire: true},
	}
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

		// bool for analytics middleware
		c.Set("cached", false)

		// break cache if there is a query
		if c.Request.URL.RawQuery != "" {
			c.Next()
			return
		}

		// redis circuitbreaker
		hystrix.Do("cache", func() (err error) {

			// Trim leading / from path and split
			params := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")

			// get the keyname
			key, ok := RedisKeyIndex[params[0]]
			if !ok {
				c.Next()
				return
			}

			// set the key minus the base
			key.SetKey(params[1:]...)

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

			return

		}, func(err error) error {
			c.Next()
			return nil
		})

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
	hashid     string
}

func (r *RedisKey) SetKey(ids ...string) {

	if r.fieldcount == 0 {
		r.key = r.base
		return
	}

	// create our key
	r.key = strings.Join([]string{r.base, strings.Join(ids[:r.fieldcount], ":")}, ":")

	// get our hash id
	if r.hash {
		r.hashid = strings.Join(ids[r.fieldcount:], "")
	}

	return
}

func (r *RedisKey) Get() (result []byte, err error) {

	if r.hash {
		return redis.RedisCache.HGet(r.key, r.hashid)
	} else {
		return redis.RedisCache.Get(r.key)
	}

	return
}

func (r *RedisKey) Set(data []byte) (err error) {

	if r.hash {
		err = redis.RedisCache.HMSet(r.key, r.hashid, data)
	} else {
		err = redis.RedisCache.Set(r.key, data)
	}

	if r.expire {
		return redis.RedisCache.Expire(r.key, 600)
	}

	return
}
