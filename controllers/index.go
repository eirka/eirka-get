package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
	threads := c.DefaultQuery("threads", strconv.FormatUint(uint64(config.Settings.Limits.ThreadsPerPage), 10))
	// how many posts per thread
	posts := c.DefaultQuery("posts", strconv.FormatUint(uint64(config.Settings.Limits.PostsPerThread), 10))

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
