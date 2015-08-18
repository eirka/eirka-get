package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strings"

	"github.com/techjanitor/pram-get/config"
)

var (
	validSites          = map[string]bool{}
	defaultAllowHeaders = []string{"Origin", "Accept", "Content-Type", "Authorization"}
)

func init() {

	// add valid sites to map
	for _, site := range config.Settings.CORS.Sites {
		validSites[site] = true
	}

}

// CORS will set the headers for Cross-origin resource sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		req := c.Request
		origin := req.Header.Get("Origin")

		// Set origin header from sites config
		if isAllowedSite(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "")
		}

		c.Header("Access-Control-Allow-Credentials", "true")

		if req.Method == "OPTIONS" {

			// Add allowed method header
			c.Header("Access-Control-Allow-Methods", "GET")

			// Add allowed headers header
			c.Header("Access-Control-Allow-Headers", strings.Join(defaultAllowHeaders, ","))

			c.AbortWithStatus(http.StatusOK)

			return

		} else {

			c.Next()

		}

	}
}

// Check if origin is allowed
func isAllowedSite(host string) bool {

	// Get the host from the origin
	parsed, err := url.Parse(host)
	if err != nil {
		return false
	}

	if validSites[strings.ToLower(parsed.Host)] {
		return true
	}

	return false

}
