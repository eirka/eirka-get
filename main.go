package main

import (
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"

	"github.com/techjanitor/pram-libs/auth"
	"github.com/techjanitor/pram-libs/config"
	"github.com/techjanitor/pram-libs/cors"
	"github.com/techjanitor/pram-libs/db"
	"github.com/techjanitor/pram-libs/redis"
	"github.com/techjanitor/pram-libs/validate"

	local "github.com/techjanitor/pram-get/config"
	c "github.com/techjanitor/pram-get/controllers"
	m "github.com/techjanitor/pram-get/middleware"
)

var (
	version = "1.0.5"
)

func init() {
	// Database connection settings
	dbase := db.Database{

		User:           local.Settings.Database.User,
		Password:       local.Settings.Database.Password,
		Proto:          local.Settings.Database.Proto,
		Host:           local.Settings.Database.Host,
		Database:       local.Settings.Database.Database,
		MaxIdle:        local.Settings.Database.MaxIdle,
		MaxConnections: local.Settings.Database.MaxConnections,
	}

	// Set up DB connection
	dbase.NewDb()

	// Get limits and stuff from database
	config.GetDatabaseSettings()

	// redis settings
	r := redis.Redis{
		// Redis address and max pool connections
		Protocol:       local.Settings.Redis.Protocol,
		Address:        local.Settings.Redis.Address,
		MaxIdle:        local.Settings.Redis.MaxIdle,
		MaxConnections: local.Settings.Redis.MaxConnections,
	}

	// Set up Redis connection
	r.NewRedisCache()

	// set auth middleware secret
	auth.Secret = local.Settings.Session.Secret

	// print the starting info
	StartInfo()

	// Print out config
	config.Print()

	// Print out config
	local.Print()

	// set cors domains
	cors.SetDomains(local.Settings.CORS.Sites, strings.Split("GET", ","))

}

func main() {
	r := gin.Default()

	// add CORS headers
	r.Use(cors.CORS())
	// validate all route parameters
	r.Use(validate.ValidateParams())

	r.GET("/uptime", c.UptimeController)
	r.NoRoute(c.ErrorController)

	// public cached pages
	public := r.Group("/")
	public.Use(m.AntiSpamCookie())
	public.Use(auth.Auth(auth.All))
	public.Use(m.Analytics())
	public.Use(m.Cache())

	public.GET("/:ib/index/:page", c.IndexController)
	public.GET("/:ib/thread/:thread/:page", c.ThreadController)
	public.GET("/:ib/tag/:tag/:page", c.TagController)
	public.GET("/:ib/image/:id", c.ImageController)
	public.GET("/:ib/post/:thread/:id", c.PostController)
	public.GET("/:ib/tags/:page", c.TagsController)
	public.GET("/:ib/tagsearch", c.TagSearchController)
	public.GET("/:ib/directory", c.DirectoryController)
	public.GET("/:ib/popular", c.PopularController)
	public.GET("/:ib/new", c.NewController)
	public.GET("/:ib/favorited", c.FavoritedController)
	public.GET("/:ib/taginfo/:id", c.TagInfoController)
	public.GET("/:ib/tagtypes", c.TagTypesController)
	public.GET("/:ib/imageboards", c.ImageboardsController)

	// user pages
	users := r.Group("/user")
	users.Use(auth.Auth(auth.Registered))

	users.GET("/:ib/whoami", c.UserController)
	users.GET("/:ib/favorite/:id", c.FavoriteController)
	users.GET("/:ib/favorites/:page", c.FavoritesController)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", local.Settings.Get.Address, local.Settings.Get.Port),
		Handler: r,
	}

	gracehttp.Serve(s)

}

func StartInfo() {

	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n\n", "PRAM-GET")
	fmt.Printf("%-20v%40v\n", "Version", version)
	fmt.Println(strings.Repeat("*", 60))

}
