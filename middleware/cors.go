package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"

	"github.com/techjanitor/pram-get/config"
)

var validSites = map[string]bool{}

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
		origin := req.RemoteAddr

		// Set origin header from sites config
		//if isAllowedSite(origin) {
		//	res.Header().Set("Access-Control-Allow-Origin", origin)
		//} else {
		//	res.Header().Set("Access-Control-Allow-Origin", "")
		//}

		res.Header().Set("Access-Control-Allow-Origin", origin)

		// Add allowed method header
		res.Header().Set("Access-Control-Allow-Methods", "GET")

		c.Next()

	}
}

// Check if origin is allowed
func isAllowedSite(host string) bool {

	if validSites[strings.ToLower(host)] {
		return true
	}

	return false

}
