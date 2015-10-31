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

	// add CORS headers
	r.Use(m.CORS())
	// validate all route parameters
	r.Use(m.ValidateParams())

	r.GET("/uptime", c.UptimeController)
	r.NoRoute(c.ErrorController)

	// public cached pages
	public := r.Group("/")
	public.Use(m.AntiSpamCookie())
	public.Use(m.Auth(m.All))
	public.Use(m.Analytics())
	public.Use(m.Cache())

	public.GET("/index/:ib/:page", c.IndexController)
	public.GET("/thread/:ib/:thread/:page", c.ThreadController)
	public.GET("/tag/:ib/:tag/:page", c.TagController)
	public.GET("/image/:ib/:id", c.ImageController)
	public.GET("/post/:ib/:thread/:id", c.PostController)
	public.GET("/tags/:ib/:page", c.TagsController)
	public.GET("/tagsearch/:ib", c.TagSearchController)
	public.GET("/directory/:ib", c.DirectoryController)
	public.GET("/popular/:ib", c.PopularController)
	public.GET("/new/:ib", c.NewController)
	public.GET("/favorited/:ib", c.FavoritedController)
	public.GET("/taginfo/:id", c.TagInfoController)
	public.GET("/tagtypes", c.TagTypesController)
	public.GET("/pram", c.PramController)

	// user pages
	users := r.Group("/user")
	users.Use(m.Auth(m.Registered))

	users.GET("/whoami", c.UserController)
	users.GET("/favorite/:id", c.FavoriteController)
	users.GET("/favorites/:ib/:page", c.FavoritesController)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Settings.Get.Address, config.Settings.Get.Port),
		Handler: r,
	}

	gracehttp.Serve(s)

}
