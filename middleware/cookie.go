package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/techjanitor/pram-get/config"
)

// Generates the antispam cookie for pram-post
func AntiSpamCookie() gin.HandlerFunc {
	return func(c *gin.Context) {

		var cookie = &http.Cookie{
			Name:    config.Settings.Antispam.CookieName,
			Value:   config.Settings.Antispam.CookieValue,
			Expires: time.Now().Add(356 * 24 * time.Hour),
			Path:    "/",
		}

		http.SetCookie(c.Writer, cookie)

		c.Next()

	}
}
