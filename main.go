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
	users.Use(auth.Auth(auth.Registered))

	users.GET("/whoami", c.UserController)
	users.GET("/favorite/:id", c.FavoriteController)
	users.GET("/favorites/:ib/:page", c.FavoritesController)

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
