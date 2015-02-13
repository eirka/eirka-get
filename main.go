package main

import (
	"github.com/gin-gonic/gin"
	"runtime"

	"pram-get/config"
	c "pram-get/controllers"
	m "pram-get/middleware"
	u "pram-get/utils"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Get start time
	u.InitTime()

	conf := config.Settings

	db := conf.Database

	// Set up DB connection
	u.NewDb(db.DbUser, db.DbPassword, db.DbProto, db.DbHost, db.DbDatabase, db.DbMaxIdle, db.DbMaxConnections)

	redis := conf.Redis

	// Set up Redis connection
	u.NewRedisCache(redis.RedisAddress, redis.RedisProtocol, redis.RedisMaxIdle, redis.RedisMaxConnections)

}

func main() {
	r := gin.Default()

	r.Use(gin.ForwardedFor("127.0.0.1/32"))

	controllers := r.Group("/")
	// Adds antispam cookie to requests
	controllers.Use(m.AntiSpamCookie())
	// Makes sure params are uint
	controllers.Use(m.ValidateParams())
	// Caches requests in Redis
	controllers.Use(m.Cache())

	controllers.GET("/index/:ib/:page", c.IndexController)
	controllers.GET("/thread/:thread/:page", c.ThreadController)
	controllers.GET("/tag/:tag/:page", c.TagController)
	controllers.GET("/directory/:ib", c.DirectoryController)
	controllers.GET("/image/:id", c.ImageController)
	controllers.GET("/post/:thread/:id", c.PostController)
	controllers.GET("/tags/:ib", c.TagsController)
	controllers.GET("/tagtypes", c.TagTypesController)

	r.GET("/uptime", c.UptimeController)
	r.NoRoute(c.ErrorController)

	r.Run("127.0.0.1:5010")

}
