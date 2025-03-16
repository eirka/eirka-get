package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-get/models"
)

// DirectoryController handles directory pages
func DirectoryController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// Initialize model struct
	m := &models.DirectoryModel{
		Ib:   params[0],
		Page: params[1],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("DirectoryController.Get")
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DirectoryController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DirectoryController.Marshal")
		return
	}

	// Check if this is a cacheMiss (request from cache middleware)
	if _, ok := c.Get("cacheMiss"); ok {
		// Get the data callback function and use it to send the data
		if callback, ok := c.Get("setDataCallback"); ok {
			callback.(func([]byte, error))(output, nil)
		}
	}

	// Always write the response back to the client
	c.Data(http.StatusOK, "application/json", output)
}
