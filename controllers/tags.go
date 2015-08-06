package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-get/errors"
	"github.com/techjanitor/pram-get/models"
)

// TagsController handles tags pages
func TagsController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get search query if its there
	search := c.Query("search")

	// pick what model we want
	if search == "" {
		// Initialize model struct
		m := &models.TagsModel{
			Ib: params[0],
		}
	} else {
		// Initialize model struct
		m := &models.TagSearchModel{
			Ib:   params[0],
			Term: search,
		}
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
