package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"runtime"

	"github.com/techjanitor/pram-get/config"
	c "github.com/techjanitor/pram-get/controllers"
	m "github.com/techjanitor/pram-get/middleware"
	u "github.com/techjanitor/pram-get/utils"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Get start time
	u.InitTime()

	// Set up DB connection
	u.NewDb()

	// Set up Redis connection
	u.NewRedisCache()

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

	r.Run(fmt.Sprintf("%s:%d", config.Settings.General.Address, config.Settings.General.Port))

}
