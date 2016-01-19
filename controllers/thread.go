package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"

	"github.com/eirka/eirka-get/models"
)

// ThreadController handles thread pages
func ThreadController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// how many posts per page
	posts := c.DefaultQuery("posts", strconv.Itoa(int(config.Settings.Limits.PostsPerPage)))

	up, err := validate.ValidateParam(posts)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("IndexController.ValidateQueryParams")
		return
	}

	// Initialize model struct
	m := &models.ThreadModel{
		Ib:     params[0],
		Thread: params[1],
		Page:   params[2],
		Posts:  validate.Clamp(up, 100, 20),
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("ThreadController.Get")
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ThreadController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ThreadController.Marshal")
		return
	}

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
