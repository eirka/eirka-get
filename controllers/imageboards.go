package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-get/models"
)

// ImageboardsController handles eirka index pages
func ImageboardsController(c *gin.Context) {

	// Initialize model struct
	m := &models.ImageboardsModel{}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("ImageboardsController.Get")
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ImageboardsController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.MarshalIndent(m.Result, "", "  ")
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ImageboardsController.Marshal")
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
