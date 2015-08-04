package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"

	//u "github.com/techjanitor/pram-get/utils"
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

		// get the ib
		ib := c.Param("ib")

		// Process request
		c.Next()

		// abort if theres no ib
		if ib == "" {
			c.Error(fmt.Errorf("no ib???"))
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

		// print headers
		fmt.Fprintf(gin.DefaultWriter, "%s\n%s\n", request, req.Header)

	}
}
