package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "pram-get/errors"
	"pram-get/models"
)

// TagsController handles tags pages
func TagsController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// Initialize model struct
	m := &models.TagsModel{
		Ib: params[0],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err, "Operation aborted")
		return
	}
	if err != nil {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err, "Operation aborted")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err, "Operation aborted")
		return
	}

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
