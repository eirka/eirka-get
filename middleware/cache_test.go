package middleware

import (
	"errors"
	"testing"

	"github.com/eirka/eirka-libs/redis"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(Cache())

	router.GET("/index/:ib/:page", func(c *gin.Context) {
		// Hand off data to cache middleware
		c.Set("data", []byte("cache data"))

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
