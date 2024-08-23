package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-get/models"
)

// TagSearchController handles search requests for tags
func TagSearchController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get search query if its there
	search := c.Query("search")

	// there needs to be a search term obviously
	if search == "" {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(e.ErrInvalidParam).SetMeta("TagSearchController.SearchTermInvalid")
		return
	}

	// Initialize model struct
	m := &models.TagSearchModel{
		Ib:   params[0],
		Term: search,
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("TagSearchController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("TagSearchController.Marshal")
		return
	}

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Data(http.StatusOK, "application/json", output)

}
