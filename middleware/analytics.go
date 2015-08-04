package middleware

import (
	"github.com/gin-gonic/gin"
	"time"

	u "github.com/techjanitor/pram-get/utils"
)

// requesttype holds the data we want to capture
type RequestType struct {
	Ib        string
	Ip        string
	Path      string
	Status    int
	Latency   time.Duration
	Useragent string
	Referer   string
	Country   string
}

// Analytics will log requests in the database
func Analytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		// Start timer
		start := time.Now()
		// get request path
		path := req.URL.Path

		// Process request
		c.Next()

		// get the ib
		ib := c.Param("ib")

		// abort if theres no ib
		if ib == "" {
			c.Abort()
			return
		}

		// Stop timer
		end := time.Now()
		// get request latency
		latency := end.Sub(start)

		// set our data
		request := RequestType{
			Ib:        ib,
			Ip:        c.ClientIP(),
			Path:      path,
			Status:    c.Writer.Status(),
			Latency:   latency,
			Useragent: req.UserAgent(),
			Referer:   req.Referer(),
			Country:   req.Header.Get("CF-IPCountry"),
		}

		// Get Database handle
		db, err := u.GetDb()
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// prepare query for analytics table
		ps1, err := db.Prepare("INSERT INTO analytics (ib_id, request_time, request_ip, request_path, request_status) VALUES (?,NOW(),?,?,?)")
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// input data
		_, err = ps1.Exec(request.Ib, request.Ip, request.Path, request.Status)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

	}
}
