package middleware

import (
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/gin-gonic/gin"
)

// RpcTokenMiddleware validates the rpc-token header for inter-service calls.
func RpcTokenMiddleware(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("rpc-token")
		if token != expectedToken {
			c.AbortWithStatusJSON(200, response.BuildError("invalid rpc-token"))
			return
		}
		c.Next()
	}
}
