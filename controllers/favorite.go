package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-get/errors"
	"github.com/techjanitor/pram-get/models"
	u "github.com/techjanitor/pram-get/utils"
)

// FavoriteController handles favorited images
func FavoriteController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(u.User)

	// Initialize model struct
	m := &models.FavoriteModel{
		User: userdata.Id,
		Id:   params[0],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err != nil {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
