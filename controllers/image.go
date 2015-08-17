package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-get/errors"
	"github.com/techjanitor/pram-get/models"
	u "github.com/techjanitor/pram-get/utils"
)

// ImageController handles image pages
func ImageController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(u.User)

	// Initialize model struct
	m := &models.ImageModel{
		User: userdata.Id,
		Ib:   params[0],
		Id:   params[1],
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

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
