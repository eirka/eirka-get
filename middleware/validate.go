package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// ValidateParams will loop through the route parameters to make sure theyre uint
func ValidateParams() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Params != nil {

			var params []uint

			for _, param := range c.Params {

				pid, err := u.ValidateParam(param.Value)
				if err != nil {
					c.JSON(e.ErrorMessage(e.ErrInvalidParam))
					c.Error(e.ErrInvalidParam)
					c.Abort()
					return
				}

				params = append(params, pid)

			}

			c.Set("params", params)

		}

		c.Next()

	}
}
