package admin

import (
	"errors"
	"fmt"
	"net/http"

	"usercore/admin/middleware"
	"usercore/internal/dto"
	"usercore/internal/pkg/response"
	"usercore/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	auth      *service.AuthService
	users     *service.UserService
	roles     *service.RoleService
	tenants   *service.TenantService
	companies *service.CompanyService
}

func NewHandler(auth *service.AuthService, users *service.UserService, roles *service.RoleService, tenants *service.TenantService, companies *service.CompanyService) *Handler {
	return &Handler{auth: auth, users: users, roles: roles, tenants: tenants, companies: companies}
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.auth.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrTenantForbidden) {
			response.Fail(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if resp.AccessToken == "" {
		response.OK(c, resp)
		return
	}
	response.OK(c, resp)
}

func (h *Handler) SwitchTenant(c *gin.Context) {
	var req dto.SwitchTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	resp, err := h.auth.SwitchTenant(claims.UserID, claims.IsPlatform, req.TenantID)
	if err != nil {
		response.Fail(c, http.StatusForbidden, err.Error())
		return
	}
	response.OK(c, resp)
}

func (h *Handler) Me(c *gin.Context) {
	resp, err := h.auth.Me(middleware.Claims(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, resp)
}

func (h *Handler) ListApps(c *gin.Context) {
	list, err := h.auth.ListApps(middleware.Claims(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *Handler) SaveAppOrder(c *gin.Context) {
	var req dto.SaveAppOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.SaveAppOrder(middleware.Claims(c), req.AppIDs); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"saved": true})
}

func (h *Handler) ListUsers(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	claims := middleware.Claims(c)
	list, total, err := h.users.List(claims.TenantID, q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.Page(list, total, q.Page, q.PageSize))
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	u, err := h.users.Create(claims.TenantID, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, u)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	u, err := h.users.Update(claims.TenantID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			response.Fail(c, http.StatusForbidden, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, u)
}

func (h *Handler) RemoveUser(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	if err := h.users.RemoveFromTenant(claims.TenantID, id, claims.UserID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			response.Fail(c, http.StatusForbidden, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *Handler) ListRoles(c *gin.Context) {
	claims := middleware.Claims(c)
	list, err := h.roles.List(claims.TenantID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	r, err := h.roles.Create(claims.TenantID, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, r)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	r, err := h.roles.Update(claims.TenantID, id, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, r)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	if err := h.roles.Delete(claims.TenantID, id); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *Handler) ListPermissions(c *gin.Context) {
	list, err := h.roles.ListPermissions()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *Handler) ListTenants(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	list, total, err := h.tenants.List(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.Page(list, total, q.Page, q.PageSize))
}

func (h *Handler) CreateTenant(c *gin.Context) {
	var req dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := middleware.Claims(c)
	companyID := claims.CompanyID
	if req.CompanyID > 0 {
		companyID = req.CompanyID
	}
	if companyID == 0 {
		companyID = 1
	}
	t, err := h.tenants.Create(companyID, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, t)
}

func (h *Handler) UpdateTenant(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	var req dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.tenants.Update(id, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, t)
}

func (h *Handler) ListCompanies(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	list, total, err := h.companies.List(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.Page(list, total, q.Page, q.PageSize))
}

func (h *Handler) CreateCompany(c *gin.Context) {
	var req dto.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.companies.Create(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *Handler) UpdateCompany(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.companies.Update(id, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, item)
}

func parseID(c *gin.Context, param string) (uint64, error) {
	var id uint64
	if _, err := fmt.Sscan(c.Param(param), &id); err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
