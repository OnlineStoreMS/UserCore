package admin

import (
	"usercore/admin/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(g *gin.RouterGroup, h *Handler, jwtAuth gin.HandlerFunc) {
	g.POST("/auth/login", h.Login)

	auth := g.Group("")
	auth.Use(jwtAuth)
	auth.GET("/auth/me", h.Me)
	auth.POST("/auth/switch-tenant", h.SwitchTenant)
	auth.GET("/apps", h.ListApps)
	auth.GET("/permissions", h.ListPermissions)

	tenantAdmin := auth.Group("")
	tenantAdmin.Use(middleware.RequirePerm("tenant:admin"))
	tenantAdmin.GET("/users", h.ListUsers)
	tenantAdmin.POST("/users", h.CreateUser)
	tenantAdmin.PUT("/users/:id", h.UpdateUser)
	tenantAdmin.DELETE("/users/:id", h.RemoveUser)
	tenantAdmin.GET("/roles", h.ListRoles)
	tenantAdmin.POST("/roles", h.CreateRole)
	tenantAdmin.PUT("/roles/:id", h.UpdateRole)
	tenantAdmin.DELETE("/roles/:id", h.DeleteRole)

	platform := auth.Group("")
	platform.Use(middleware.RequirePlatform())
	platform.GET("/companies", h.ListCompanies)
	platform.POST("/companies", h.CreateCompany)
	platform.PUT("/companies/:id", h.UpdateCompany)
	platform.GET("/tenants", h.ListTenants)
	platform.POST("/tenants", h.CreateTenant)
	platform.PUT("/tenants/:id", h.UpdateTenant)
}
