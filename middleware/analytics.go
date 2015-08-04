package middleware

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"

	u "github.com/techjanitor/pram-get/utils"
)

// holds our prepared statement
var analyticsStmt *sql.Stmt

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

func init() {
	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		panic(err)
	}

	analyticsStmt, err := db.Prepare("INSERT INTO analytics (ib_id, request_time, request_ip, request_path, request_status) VALUES (?,NOW(),?,?,?)")
	if err != nil {
		panic(err)
	}

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

		// input data
		_, err = analyticsStmt.Exec(request.Ib, request.Ip, request.Path, request.Status)
		if err != nil {
			return
		}

	}
}
