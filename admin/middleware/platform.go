package middleware

import (
	"net/http"

	"usercore/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func RequirePlatform() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := Claims(c)
		if claims == nil || !claims.IsPlatform {
			response.Fail(c, http.StatusForbidden, "需要平台管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
