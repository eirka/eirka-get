package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/user"
)

// list of keys record
var analyticsKey = map[string]bool{
	"index":     true,
	"thread":    true,
	"tag":       true,
	"image":     true,
	"tags":      true,
	"directory": true,
	"popular":   true,
}

type requestType struct {
	Ib        string
	IP        string
	User      uint
	Path      string
	ItemKey   string
	ItemValue string
	Status    int
	Latency   time.Duration
	Cached    bool
}

// Analytics will log requests in the database
func Analytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		// get userdata from session middleware
		userdata := c.MustGet("userdata").(user.User)

		// Make key from path
		key := generateKey(req.URL.Path)

		// skip if we're not recording this key
		if !analyticsKey[strings.ToLower(key.Key)] {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		// get request latency
		latency := end.Sub(start)

		// Check if there was an error from the controller
		_, controllerError := c.Get("controllerError")
		if controllerError {
			c.Abort()
			return
		}

		// set our data
		request := requestType{
			Ib:        c.Param("ib"),
			IP:        c.ClientIP(),
			User:      userdata.ID,
			Path:      req.URL.Path,
			Status:    c.Writer.Status(),
			Latency:   latency,
			ItemKey:   key.Key,
			ItemValue: key.Value,
			Cached:    c.MustGet("cached").(bool),
		}

		// fire and forget
		go insertRecord(request)

	}
}

func insertRecord(request requestType) (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// input data
	_, err = dbase.Exec(`INSERT INTO analytics (ib_id, user_id, request_ip, request_path, request_status, request_latency, request_itemkey, request_itemvalue, request_cached, request_time) VALUES (?,?,?,?,?,?,?,?,?,NOW())`,
		request.Ib, request.User, request.IP, request.Path, request.Status, request.Latency, request.ItemKey, request.ItemValue, request.Cached)
	if err != nil {
		return
	}

	return

}

type itemKey struct {
	Key   string
	Value string
}

// Will take the params from the request and turn them into a key
func generateKey(path string) itemKey {

	// Trim leading / from path and split
	params := strings.Split(strings.Trim(path, "/"), "/")
	// new item key
	r := itemKey{}

	if params != nil {
		switch {
		case len(params) <= 2:
			r.Key = params[0]
			r.Value = "1"
		case len(params) >= 3:
			r.Key = params[0]
			r.Value = params[2]
		}
	}

	return r

}
