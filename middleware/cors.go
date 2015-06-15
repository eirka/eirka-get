package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
)

// CORS will set the headers for Cross-origin resource sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		res := c.Writer

		// Set origin header from sites config
		res.Header().Set("Access-Control-Allow-Origin", strings.Join(config.Settings.CORS.Sites, " "))

		// Add allowed method header
		res.Header().Set("Access-Control-Allow-Methods", "GET")

		c.Next()

	}
}
