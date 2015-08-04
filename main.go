package main

import (
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

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

	// init the user data worker
	u.UserInit()

	// init the analytics data worker
	m.AnalyticsInit()

	// channel for shutdown
	c := make(chan os.Signal, 10)

	// watch for shutdown signals to shutdown cleanly
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-c
		Shutdown()
	}()

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
	public.Use(m.Auth(m.SetAuthLevel().All()))
	public.Use(m.Analytics())
	public.Use(m.Cache())

	public.GET("/index/:ib/:page", c.IndexController)
	public.GET("/thread/:ib/:thread/:page", c.ThreadController)
	public.GET("/tag/:ib/:tag/:page", c.TagController)
	public.GET("/image/:ib/:id", c.ImageController)
	public.GET("/post/:ib/:thread/:id", c.PostController)
	public.GET("/tags/:ib", c.TagsController)
	public.GET("/directory/:ib", c.DirectoryController)
	public.GET("/taginfo/:id", c.TagInfoController)
	public.GET("/tagtypes", c.TagTypesController)
	public.GET("/pram", c.PramController)

	// user pages
	users := r.Group("/user")
	users.Use(m.Auth(m.SetAuthLevel().Registered()))

	users.GET("/whoami", c.UserController)
	users.GET("/favorite/:id", c.FavoriteController)
	users.GET("/favorites/:ib/:page", c.FavoritesController)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Settings.General.Address, config.Settings.General.Port),
		Handler: r,
	}

	gracehttp.Serve(s)

}

// called on sigterm or interrupt
func Shutdown() {

	fmt.Println("Shutting down...")

	// close the database connection
	fmt.Println("Closing database connection")
	err := u.CloseDb()
	if err != nil {
		fmt.Println(err)
	}

}
