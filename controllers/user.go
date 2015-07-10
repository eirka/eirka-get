package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-get/errors"
	"github.com/techjanitor/pram-get/models"
	u "github.com/techjanitor/pram-get/utils"
)

// UserController gets account info
func UserController(c *gin.Context) {

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(u.User)

	// Initialize model struct
	m := &models.UserModel{
		Id: userdata.Id,
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err)
		return
	}
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

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
