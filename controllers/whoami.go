package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/user"

	"github.com/eirka/eirka-get/models"
)

// UserController gets account info
func UserController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(user.User)

	// Initialize model struct
	m := &models.UserModel{
		User: userdata.Id,
		Ib:   params[0],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err != nil {
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
