package middleware

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

type Permission uint

const (
	All Permission = 1 << iota
	Registered
	Moderators
	Admins
)

// checks for session cookie and handles permissions
func Auth(perm Permission) gin.HandlerFunc {
	return func(c *gin.Context) {

		// set default anonymous user
		user := u.User{
			Id:              1,
			Group:           1,
			IsAuthenticated: false,
		}

		// parse jwt token if its there
		token, err := jwt.ParseFromRequest(c.Request, func(token *jwt.Token) (interface{}, error) {

			// check alg to make sure its hmac
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// compare with secret from settings
			return []byte(config.Settings.Session.Secret), nil
		})
		if err != nil && err != jwt.ErrNoTokenInRequest {
			// if theres some jwt error then return unauth
			c.JSON(e.ErrorMessage(e.ErrUnauthorized))
			c.Error(err)
			c.Abort()
			return
		}

		// process token
		if token != nil {
			// if the token is valid set the data
			if err == nil && token.Valid {

				// get uid from jwt, cast to float
				uid, ok := token.Claims["user_id"].(float64)
				if !ok {
					c.JSON(e.ErrorMessage(e.ErrInternalError))
					c.Error(err)
					c.Abort()
					return
				}

				// set user id in user struct
				user.Id = uint(uid)

				// get the rest of the user info
				err = user.Info()
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error_message": err.Error()})
					c.Error(err)
					c.Abort()
					return
				}

			} else {
				c.JSON(e.ErrorMessage(e.ErrInternalError))
				c.Error(err)
				c.Abort()
				return
			}

		}

		// check if user meets set permissions
		if user.Group < perm {
			c.JSON(e.ErrorMessage(e.ErrUnauthorized))
			c.Error(e.ErrUnauthorized)
			c.Abort()
			return
		}

		// set user data
		c.Set("userdata", user)

		c.Next()

	}

}
