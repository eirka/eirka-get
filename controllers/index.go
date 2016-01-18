package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/eirka/eirka-libs/config"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/redis"
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

	// validate query parameter
	ut, err := validate.ValidateParam(threads)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err)
		return
	}

	// max for query params
	if ut > 20 || ut < 5 {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(e.ErrInvalidParam)
		return
	}

	// validate query parameter
	up, err := validate.ValidateParam(posts)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err)
		return
	}

	// max for query params
	if up > 20 {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(e.ErrInvalidParam)
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
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err)
		return
	} else if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.Set("controllerError", true)
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	key := redis.RedisKeyIndex["index"]

	key.SetKey(m.Ib).SetHashId(m.Page)

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
