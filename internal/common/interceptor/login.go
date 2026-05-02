package interceptor

import (
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/gin-gonic/gin"
)

const loginUserKey = "loginUser"

// LoginInterceptor is a Gin middleware that validates JWT tokens.
// Compatible with Java's LoginInterceptor (ThreadLocal replacement via gin.Context).
func LoginInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}
		token := c.GetHeader("token")
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
			c.Abort()
			return
		}
		loginUser, err := util.ParseToken(token)
		if err != nil {
			response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
			c.Abort()
			return
		}
		c.Set(loginUserKey, loginUser)
		c.Next()
	}
}

// GetLoginUser retrieves the LoginUser from gin.Context.
func GetLoginUser(c *gin.Context) *model.LoginUser {
	val, exists := c.Get(loginUserKey)
	if !exists {
		return nil
	}
	return val.(*model.LoginUser)
}
