package controllers

import (
	"github.com/gin-gonic/gin"

	u "pram-get/utils"
)

// Uptime controllers shows proc uptime
func UptimeController(c *gin.Context) {

	c.JSON(200, gin.H{"uptime": u.GetTime()})

	return

}
