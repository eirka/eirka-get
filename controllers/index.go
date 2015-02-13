package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-get/errors"
	"github.com/techjanitor/pram-get/models"
)

// IndexController handles index pages
func IndexController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// Initialize model struct
	m := &models.IndexModel{
		Ib:   params[0],
		Page: params[1],
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
