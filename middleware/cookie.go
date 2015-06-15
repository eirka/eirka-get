package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/techjanitor/pram-get/config"
)

// Generates the antispam cookie for pram-post
// This needs to be able to do multiple domains
func AntiSpamCookie() gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie := &http.Cookie{
			Name:     config.Settings.Antispam.CookieName,
			Value:    config.Settings.Antispam.CookieValue,
			Expires:  time.Now().Add(356 * 24 * time.Hour),
			Path:     "/",
			HttpOnly: true,
		}

		http.SetCookie(c.Writer, cookie)

		c.Next()

	}
}
