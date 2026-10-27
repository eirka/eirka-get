package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/pidfile"
	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/cors"
	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/redis"
	"github.com/eirka/eirka-libs/status"
	"github.com/eirka/eirka-libs/user"
	"github.com/eirka/eirka-libs/validate"

	local "github.com/eirka/eirka-get/config"
	c "github.com/eirka/eirka-get/controllers"
	m "github.com/eirka/eirka-get/middleware"
)

// configure performs production runtime setup: the pid file, database and redis
// connection pools, runtime settings loaded from the database, and CORS domains.
// It runs from main (not an init func) so that test binaries — which import this
// package — do not try to write a pid file or open database/redis sockets.
func configure() {

	var err error

	// create pid file
	pidfile.SetPidfilePath("/run/eirka/eirka-get.pid")

	err = pidfile.Write()
	if err != nil {
		panic("Could not write pid file")
	}

	if local.Settings != nil {
		// Database connection settings
		dbase := db.Database{

			User:           local.Settings.Database.User,
			Password:       local.Settings.Database.Password,
			Proto:          local.Settings.Database.Protocol,
			Host:           local.Settings.Database.Host,
			Database:       local.Settings.Database.Database,
			MaxIdle:        local.Settings.Get.DatabaseMaxIdle,
			MaxConnections: local.Settings.Get.DatabaseMaxConnections,
		}

		// Set up DB connection
		dbase.NewDb()

		// Get limits and stuff from database
		config.GetDatabaseSettings()

		// redis settings
		r := redis.Redis{
			// Redis address and max pool connections
			Protocol:       local.Settings.Redis.Protocol,
			Address:        local.Settings.Redis.Host,
			MaxIdle:        local.Settings.Get.RedisMaxIdle,
			MaxConnections: local.Settings.Get.RedisMaxConnections,
		}

		// Set up Redis connection
		r.NewRedisCache()

		// set cors domains
		cors.SetDomains(local.Settings.CORS.Sites, strings.Split("GET", ","))
	} else {
		panic("Could not initialize settings")
	}

}

// setupRouter builds the gin engine with every route and its middleware.
func setupRouter() *gin.Engine {
	r := gin.Default()

	// c.ClientIP() is recorded by the analytics middleware, so it must not be
	// client controlled. gin's default trusts every peer and returns the leftmost
	// X-Forwarded-For entry, which Cloudflare passes through from the client.
	// Instead, trust only nginx on loopback, and only its CF-Connecting-IP
	// header, which Cloudflare always overwrites. Any other peer resolves to its
	// own TCP address.
	r.RemoteIPHeaders = []string{"CF-Connecting-IP"}
	err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	if err != nil {
		panic("Could not set trusted proxies")
	}

	// add CORS headers
	r.Use(cors.CORS())
	// validate all route parameters
	r.Use(validate.ValidateParams())

	r.GET("/status", status.StatusController)
	r.NoRoute(c.ErrorController)

	// public cached pages
	public := r.Group("/")
	public.Use(user.Auth(false))
	public.Use(m.Analytics())
	public.Use(m.Cache())

	public.GET("/index/:ib/:page", c.IndexController)
	public.GET("/thread/:ib/:thread/:page", c.ThreadController)
	public.GET("/tag/:ib/:tag/:page", c.TagController)
	public.GET("/image/:ib/:id", c.ImageController)
	public.GET("/random/image/:ib", c.RandomController)
	public.GET("/post/:ib/:thread/:id", c.PostController)
	public.GET("/tags/:ib/:page", c.TagsController)
	public.GET("/tagsearch/:ib", c.TagSearchController)
	public.GET("/threadsearch/:ib", c.ThreadSearchController)
	public.GET("/directory/:ib/:page", c.DirectoryController)
	public.GET("/popular/:ib", c.PopularController)
	public.GET("/new/:ib", c.NewController)
	public.GET("/favorited/:ib", c.FavoritedController)
	public.GET("/tagtypes", c.TagTypesController)
	public.GET("/imageboards", c.ImageboardsController)
	public.GET("/whoami/:ib", c.WhoAmIController)

	// user pages
	users := r.Group("/user")
	users.Use(user.Auth(true))

	users.GET("/favorite/:id", c.FavoriteController)
	users.GET("/favorites/:ib/:page", c.FavoritesController)

	return r
}

func main() {
	configure()

	r := setupRouter()

	if local.Settings != nil {
		s := &http.Server{
			Addr:              fmt.Sprintf("%s:%d", local.Settings.Get.Host, local.Settings.Get.Port),
			ReadHeaderTimeout: 2 * time.Second,
			Handler:           r,
		}

		err := gracehttp.Serve(s)
		if err != nil {
			panic("Could not start server")
		}
	} else {
		panic("Could not initialize settings")
	}
}
