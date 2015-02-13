package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"pram-get/config"
)

var cookie = &http.Cookie{
	Name:    config.Settings.Antispam.CookieName,
	Value:   config.Settings.Antispam.CookieValue,
	Expires: time.Now().Add(356 * 24 * time.Hour),
	Path:    "/",
}

func AntiSpamCookie() gin.HandlerFunc {
	return func(c *gin.Context) {

		http.SetCookie(c.Writer, cookie)

		c.Next()

	}
}
