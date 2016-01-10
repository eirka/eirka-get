package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var StartTime = time.Now()

// Uptime controllers shows proc uptime
func UptimeController(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"uptime": fmt.Sprintf("%.0fm", time.Since(StartTime).Minutes()),
	})

	return

}
