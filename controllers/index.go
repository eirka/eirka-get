package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	"github.com/techjanitor/pram-get/models"
	u "github.com/techjanitor/pram-get/utils"
)

// IndexController handles index pages
func IndexController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// how many threads per index page
	threads := c.DefaultQuery("threads", strconv.Itoa(int(config.Settings.Limits.ThreadsPerPage)))
	// how many posts per thread
	posts := c.DefaultQuery("posts", strconv.Itoa(int(config.Settings.Limits.PostsPerThread)))

	// validate query parameter
	ut, err := u.ValidateParam(threads)
	if err != nil {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err)
		return
	}

	// validate query parameter
	up, err := u.ValidateParam(posts)
	if err != nil {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err)
		return
	}

	// Initialize model struct
	m := &models.IndexModel{
		Ib:      params[0],
		Page:    params[1],
		Threads: ut,
		Posts:   up,
	}

	// Get the model which outputs JSON
	err = m.Get()
	if err == e.ErrNotFound {
		c.Set("controllerError", err)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err)
		return
	} else if err != nil {
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
