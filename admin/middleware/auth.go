package middleware

import (
	"net/http"
	"strings"

	jwtmgr "usercore/internal/pkg/jwt"
	"usercore/internal/pkg/response"
	"usercore/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ContextClaims = "auth_claims"
)

func JWTAuth(jwt *jwtmgr.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "请先登录")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwt.ParseAccess(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set(ContextClaims, claims)
		c.Next()
	}
}

func RequirePerm(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := Claims(c)
		if claims == nil {
			response.Fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		if claims.IsPlatform || service.HasPerm(claims.Permissions, code) {
			c.Next()
			return
		}
		response.Fail(c, http.StatusForbidden, "权限不足")
		c.Abort()
	}
}

func Claims(c *gin.Context) *jwtmgr.Claims {
	v, ok := c.Get(ContextClaims)
	if !ok {
		return nil
	}
	claims, ok := v.(*jwtmgr.Claims)
	if !ok {
		return nil
	}
	return claims
}
