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
	defaultAllowHeaders = []string{"Origin", "Accept", "Content-Type"}
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
		res := c.Writer
		origin := req.Header.Get("Origin")

		// Set origin header from sites config
		if isAllowedSite(origin) {
			res.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			res.Header().Set("Access-Control-Allow-Origin", "")
		}

		if req.Method == "OPTIONS" {

			// Add allowed method header
			res.Header().Set("Access-Control-Allow-Methods", "GET")

			// Add allowed headers header
			res.Header().Set("Access-Control-Allow-Headers", strings.Join(defaultAllowHeaders, ","))

			c.AbortWithStatus(http.StatusOK)

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
