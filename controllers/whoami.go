package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	"github.com/techjanitor/pram-libs/auth"
	e "github.com/techjanitor/pram-libs/errors"
)

// UserType is the top level of the JSON response
type UserType struct {
	User auth.User `json:"user"`
}

// UserController gets account info
func UserController(c *gin.Context) {

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(auth.User)

	// Initialize response header
	response := UserType{}

	// seet userdata from auth middleware
	response.User = userdata

	// Marshal the structs into JSON
	output, err := json.Marshal(response)
	if err != nil {
		c.Set("controllerError", true)
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
