package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	"github.com/techjanitor/pram-libs/auth"
	e "github.com/techjanitor/pram-libs/errors"

	"github.com/techjanitor/pram-get/models"
)

// FavoritesController handles favorites pages
func FavoritesController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(auth.User)

	// Initialize model struct
	m := &models.FavoritesModel{
		User: userdata.Id,
		Ib:   params[0],
		Page: params[1],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err)
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
