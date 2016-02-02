package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/eirka/eirka-libs/config"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"

	"github.com/eirka/eirka-get/models"
)

// IndexController handles index pages
func IndexController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// how many threads per index page
	threads := c.DefaultQuery("threads", strconv.Itoa(int(config.Settings.Limits.ThreadsPerPage)))
	// how many posts per thread
	posts := c.DefaultQuery("posts", strconv.Itoa(int(config.Settings.Limits.PostsPerThread)))

	// query param must be uint
	ut, err := validate.ValidateParam(threads)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("IndexController.ValidateQueryParams")
		return
	}

	up, err := validate.ValidateParam(posts)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("IndexController.ValidateQueryParams")
		return
	}

	// Initialize model struct
	m := &models.IndexModel{
		Ib:      params[0],
		Page:    params[1],
		Threads: validate.Clamp(ut, 20, 5),
		Posts:   validate.Clamp(up, 10, 0),
	}

	// Get the model which outputs JSON
	err = m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("IndexController.Get")
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("IndexController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("IndexController.Marshal")
		return
	}

	// Hand off data to cache middleware
	c.Set("data", output)

	c.Data(200, "application/json", output)

	return

}
