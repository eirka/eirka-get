package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
	"time"

	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/user"
)

// list of keys to skip
var analyticsKeyList = map[string]bool{
	"taginfo":     true,
	"tags":        true,
	"post":        true,
	"tagtypes":    true,
	"imageboards": true,
	"new":         true,
	"favorited":   true,
}

// requesttype holds the data we want to capture
type RequestType struct {
	Ib        string
	Ip        string
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
		// get request path
		path := req.URL.Path
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

		// get userdata from session middleware
		userdata := c.MustGet("userdata").(user.User)

		// get cached state from cache middleware
		cached := c.MustGet("cached").(bool)

		// get the ib
		ib := c.Param("ib")

		// fire and forget
		go func() {

			// Trim leading / from path and split
			params := strings.Split(strings.Trim(path, "/"), "/")

			// Make key from path
			key := itemKey{}
			key.generateKey(params...)

			if !skipKey(params[0]) {

				// set our data
				request := RequestType{
					Ib:        ib,
					Ip:        c.ClientIP(),
					User:      userdata.Id,
					Path:      path,
					Status:    c.Writer.Status(),
					Latency:   latency,
					ItemKey:   key.Key,
					ItemValue: key.Value,
					Cached:    cached,
				}

				// Get Database handle
				dbase, err := db.GetDb()
				if err != nil {
					c.Error(err)
					c.Abort()
					return
				}

				// prepare query for analytics table
				ps1, err := dbase.Prepare("INSERT INTO analytics (ib_id, user_id, request_ip, request_path, request_status, request_latency, request_itemkey, request_itemvalue, request_cached, request_time) VALUES (?,?,?,?,?,?,?,?,?,NOW())")
				if err != nil {
					c.Error(err)
					c.Abort()
					return
				}
				defer ps1.Close()

				// input data
				_, err = ps1.Exec(request.Ib, request.User, request.Ip, request.Path, request.Status, request.Latency, request.ItemKey, request.ItemValue, request.Cached)
				if err != nil {
					c.Error(err)
					c.Abort()
					return
				}

			}

		}()

	}
}

type itemKey struct {
	Key   string
	Value string
}

// Will take the params from the request and turn them into a key
func (r *itemKey) generateKey(params ...string) {

	switch {
	case len(params) <= 2:
		r.Key = params[0]
		r.Value = "1"
	case len(params) >= 3:
		r.Key = params[0]
		r.Value = params[2]
	}

	return

}

// Check if key should be skipped
func skipKey(key string) bool {

	if analyticsKeyList[strings.ToLower(key)] {
		return true
	}

	return false

}
