package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// UptimeController shows proc uptime
func UptimeController(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"uptime": fmt.Sprintf("%.0fm", time.Since(startTime).Minutes()),
	})

	return

}
