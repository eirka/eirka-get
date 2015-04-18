package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"

	u "github.com/techjanitor/pram-get/utils"
)

// Uptime controllers shows proc uptime
func UptimeController(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"uptime": u.GetTime()})

	return

}
