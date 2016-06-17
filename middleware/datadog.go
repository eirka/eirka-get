package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/datadog"
	"github.com/eirka/eirka-libs/user"
)

// DataDog sends various statistics to statsd
func DataDog() gin.HandlerFunc {
	return func(c *gin.Context) {

		// get userdata from session middleware
		userdata := c.MustGet("userdata").(user.User)

		// Start timer
		start := time.Now()

		c.Next()

		// Stop timer
		end := time.Now()
		// get request latency
		latency := end.Sub(start)

		// count a hit
		datadog.Client.Count("page.hits", 1, nil, 1)
		// count unique ips
		datadog.Client.Set("visitors.unique", c.ClientIP(), nil, 1)
		// count unique users
		datadog.Client.Set("users.unique", strconv.Itoa(int(userdata.ID)), nil, 1)
		// count request duration
		datadog.Client.Histogram("request.latency", latency.Seconds(), nil, 1)

		return

	}
}
