package router

import (
	"usercore/admin"
	"usercore/admin/middleware"
	"usercore/internal/config"
	jwtmgr "usercore/internal/pkg/jwt"
	"usercore/internal/repo"
	"usercore/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), cors(cfg))

	repos := repo.New(db)
	jwt := jwtmgr.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTLMinutes, cfg.JWT.RefreshTTLHours)
	authSvc := service.NewAuthService(repos, jwt, &cfg.Apps)
	userSvc := service.NewUserService(repos)
	roleSvc := service.NewRoleService(repos)
	tenantSvc := service.NewTenantService(repos)
	companySvc := service.NewCompanyService(repos)
	h := admin.NewHandler(authSvc, userSvc, roleSvc, tenantSvc, companySvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "usercore"})
	})

	v1 := r.Group("/api/v1")
	jwtAuth := middleware.JWTAuth(jwt)
	admin.RegisterRoutes(v1, h, jwtAuth)

	return r
}

func cors(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.CORS.AllowOrigins
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin == ""
		for _, o := range origins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}
		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
