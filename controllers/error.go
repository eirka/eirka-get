package controllers

import (
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-libs/errors"
)

// Handles error messages for wrong routes
func ErrorController(c *gin.Context) {

	c.JSON(e.ErrorMessage(e.ErrNotFound))

	return

}
