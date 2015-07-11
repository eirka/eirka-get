package main

import (
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"net/http"
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

	// Get limits and stuff from database
	u.GetDatabaseSettings()

	// Print out config
	config.Print()

}

func main() {
	r := gin.Default()

	// Adds CORS headers
	r.Use(m.CORS())

	// get current user info
	r.GET("/whoami", m.Auth(m.SetAuthLevel().All()), c.UserController)

	r.GET("/uptime", c.UptimeController)
	r.NoRoute(c.ErrorController)

	controllers := r.Group("/")
	// Adds antispam cookie to requests
	controllers.Use(m.AntiSpamCookie())
	// Makes sure params are uint
	controllers.Use(m.ValidateParams())
	// Caches requests in Redis
	controllers.Use(m.Cache())

	controllers.GET("/index/:ib/:page", c.IndexController)
	controllers.GET("/thread/:ib/:thread/:page", c.ThreadController)
	controllers.GET("/tag/:ib/:tag/:page", c.TagController)
	controllers.GET("/taginfo/:id", c.TagInfoController)
	controllers.GET("/directory/:ib", c.DirectoryController)
	controllers.GET("/image/:ib/:id", c.ImageController)
	controllers.GET("/post/:ib/:thread/:id", c.PostController)
	controllers.GET("/tags/:ib", c.TagsController)
	controllers.GET("/tagtypes", c.TagTypesController)
	controllers.GET("/pram", c.PramController)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Settings.General.Address, config.Settings.General.Port),
		Handler: r,
	}

	gracehttp.Serve(s)

}
