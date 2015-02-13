package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"

	"pram-get/config"
	e "pram-get/errors"
)

// ValidateParams will loop through the route parameters to make sure theyre uint
func ValidateParams() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Params != nil {

			var params []uint

			for _, param := range c.Params {

				pid, err := strconv.ParseUint(param.Value, 10, 0)
				if err != nil {
					c.JSON(e.ErrorMessage(e.ErrInvalidParam))
					c.Error(e.ErrInvalidParam, "Operation aborted")
					c.Abort()
					return
				} else if uint(pid) > config.Settings.Limits.ParamMaxSize {
					c.JSON(e.ErrorMessage(e.ErrInvalidParam))
					c.Error(e.ErrInvalidParam, "Operation aborted")
					c.Abort()
					return
				}

				params = append(params, uint(pid))

			}

			c.Set("params", params)

		}

		c.Next()

	}
}
