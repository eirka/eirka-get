package middleware

import (
	"github.com/gin-gonic/gin"
	"time"

	u "github.com/techjanitor/pram-get/utils"
)

// requesttype holds the data we want to capture
type RequestType struct {
	Ib      string
	Ip      string
	User    uint
	Path    string
	Status  int
	Latency time.Duration
}

// Analytics will log requests in the database
func Analytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		// get userdata from session middleware
		userdata := c.MustGet("userdata").(u.User)
		// Start timer
		start := time.Now()
		// get request path
		path := req.URL.Path

		// Process request
		c.Next()

		// get the ib
		ib := c.Param("ib")

		// Stop timer
		end := time.Now()
		// get request latency
		latency := end.Sub(start)

		// set our data
		request := RequestType{
			Ib:      ib,
			Ip:      c.ClientIP(),
			User:    userdata.Id,
			Path:    path,
			Status:  c.Writer.Status(),
			Latency: latency,
		}

		// Get Database handle
		db, err := u.GetDb()
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// prepare query for analytics table
		ps1, err := db.Prepare("INSERT INTO analytics (ib_id, user_id, request_ip, request_path, request_status, request_latency, request_time) VALUES (?,?,?,?,?,?,NOW())")
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// input data
		_, err = ps1.Exec(request.Ib, request.User, request.Ip, request.Path, request.Status, request.Latency)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

	}
}
