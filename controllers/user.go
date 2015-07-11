package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// UserType is the top level of the JSON response
type UserType struct {
	User u.User `json:"user"`
}

// UserController gets account info
func UserController(c *gin.Context) {

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(u.User)

	// Initialize response header
	response := UserType{}

	// seet userdata from auth middleware
	response.User = userdata

	// Marshal the structs into JSON
	output, err := json.Marshal(response)
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
