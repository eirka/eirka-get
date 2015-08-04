package middleware

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"time"

	//u "github.com/techjanitor/pram-get/utils"
)

var testout = bufio.NewWriter(os.Stdout)

// requesttype holds the data we want to capture
type RequestType struct {
	Ib        string
	Ip        string
	Path      string
	Status    int
	Latency   time.Duration
	Useragent string
	Referer   string
}

// Analytics will log requests in the database
func Analytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		// Start timer
		start := time.Now()
		// get request path
		path := req.URL.Path

		fmt.Fprintln(testout, "derp")

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
		}

		// print headers
		fmt.Fprintf(testout, "%s\n%s\n", request, req.Header)

	}
}
