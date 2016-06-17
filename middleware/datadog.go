package middleware

import (
	"github.com/eirka/eirka-libs/datadog"
	"github.com/gin-gonic/gin"
)

// DataDog sends various statistics to statsd
func DataDog() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		// count a hit
		datadog.Client.Count("hit", 1, nil, 1)

		return

	}
}
