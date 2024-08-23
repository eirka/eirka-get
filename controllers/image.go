package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-get/models"
)

// ImageController handles image pages
func ImageController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// Initialize model struct
	m := &models.ImageModel{
		Ib: params[0],
		ID: params[1],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("ImageController.Get")
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ImageController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ImageController.Marshal")
		return
	}

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Data(http.StatusOK, "application/json", output)

}
