package restapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/service"
	"net/http"
)

const KeyCurrentUser = "__current_user"

func CheckToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if len(token) == 0 {
			var err error
			token, err = c.Cookie("accessToken")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, newMessageVo(errors.New("token not found")))
				return
			}
		}

		user, err := service.CheckToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, newMessageVo(err))
			return
		}

		if user == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, newMessageVo(errors.New("bad token")))
			return
		}

		c.Set(KeyCurrentUser, user)
		c.Next()
	}
}

func GetCurrentUser(ctx *gin.Context) *entities.User {
	val, found := ctx.Get(KeyCurrentUser)
	if !found {
		return nil
	}
	return val.(*entities.User)
}
