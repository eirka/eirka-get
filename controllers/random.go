package controllers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-get/models"
)

// RandomController handles random images
func RandomController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// Initialize model struct
	m := &models.RandomModel{
		Ib: params[0],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("RandomController.Get")
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("RandomController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("RandomController.Marshal")
		return
	}

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Data(200, "application/json", output)

}
