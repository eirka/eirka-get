package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/datadog"
	"github.com/eirka/eirka-libs/user"
)

// DataDog sends various statistics to statsd
func DataDog() gin.HandlerFunc {
	return func(c *gin.Context) {

		// get userdata from session middleware
		userdata := c.MustGet("userdata").(user.User)

		c.Next()

		// count a hit
		datadog.Client.Count("page.hits", 1, nil, 1)
		// count unique ips
		datadog.Client.Set("visitors.unique", c.ClientIP(), nil, 1)
		// count unique users
		datadog.Client.Set("users.unique", strconv.Itoa(int(userdata.ID)), nil, 1)

		return

	}
}
