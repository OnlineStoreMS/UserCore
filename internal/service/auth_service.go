package service

import (
	"errors"
	"strings"

	"usercore/internal/config"
	"usercore/internal/dto"
	jwtmgr "usercore/internal/pkg/jwt"
	"usercore/internal/model"
	"usercore/internal/pkg/password"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("邮箱或密码错误")
	ErrUserDisabled       = errors.New("账号已禁用")
	ErrTenantForbidden    = errors.New("无权访问该租户")
	ErrTenantRequired     = errors.New("请选择租户")
)

type AuthService struct {
	repos  *repo.Repos
	jwt    *jwtmgr.Manager
	appCfg *config.AppsConfig
}

func NewAuthService(repos *repo.Repos, jwt *jwtmgr.Manager, appCfg *config.AppsConfig) *AuthService {
	return &AuthService{repos: repos, jwt: jwt, appCfg: appCfg}
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.repos.User.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status != 1 || !password.Verify(user.Password, req.Password) {
		return nil, ErrInvalidCredentials
	}

	tenants, err := s.repos.User.ListTenantsForUser(user.ID)
	if err != nil {
		return nil, err
	}
	if user.IsPlatform == 0 && len(tenants) == 0 {
		return nil, ErrTenantForbidden
	}

	resp := &dto.LoginResponse{
		User:    toUserProfile(user),
		Tenants: toTenantBriefs(tenants),
	}

	tenantID := req.TenantID
	if tenantID == 0 {
		if user.IsPlatform == 1 && len(tenants) > 0 {
			tenantID = tenants[0].ID
		} else if len(tenants) == 1 {
			tenantID = tenants[0].ID
		} else {
			return resp, nil
		}
	}

	tenant, perms, err := s.issueForTenant(user, tenantID)
	if err != nil {
		return nil, err
	}
	token, exp, err := s.jwt.IssueAccess(jwtmgr.Claims{
		UserID:      user.ID,
		CompanyID:   tenant.CompanyID,
		TenantID:    tenant.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Permissions: perms,
		IsPlatform:  user.IsPlatform == 1,
	})
	if err != nil {
		return nil, err
	}
	resp.AccessToken = token
	resp.ExpiresAt = exp.Unix()
	resp.Tenant = *tenant
	resp.Permissions = perms
	return resp, nil
}

func (s *AuthService) SwitchTenant(userID uint64, isPlatform bool, tenantID uint64) (*dto.LoginResponse, error) {
	user, err := s.repos.User.GetByID(userID)
	if err != nil {
		return nil, err
	}
	tenant, perms, err := s.issueForTenant(user, tenantID)
	if err != nil {
		return nil, err
	}
	token, exp, err := s.jwt.IssueAccess(jwtmgr.Claims{
		UserID:      user.ID,
		CompanyID:   tenant.CompanyID,
		TenantID:    tenant.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Permissions: perms,
		IsPlatform:  isPlatform,
	})
	if err != nil {
		return nil, err
	}
	tenants, _ := s.repos.User.ListTenantsForUser(user.ID)
	return &dto.LoginResponse{
		AccessToken: token,
		ExpiresAt:   exp.Unix(),
		User:        toUserProfile(user),
		Tenant:      *tenant,
		Permissions: perms,
		Tenants:     toTenantBriefs(tenants),
	}, nil
}

func (s *AuthService) Me(claims *jwtmgr.Claims) (*dto.MeResponse, error) {
	user, err := s.repos.User.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	tenant, err := s.repos.Tenant.GetByID(claims.TenantID)
	if err != nil {
		return nil, err
	}
	tenants, _ := s.repos.User.ListTenantsForUser(user.ID)
	return &dto.MeResponse{
		User:        toUserProfile(user),
		Tenant:      toTenantBrief(tenant),
		Permissions: claims.Permissions,
		Tenants:     toTenantBriefs(tenants),
	}, nil
}

func (s *AuthService) ListApps(claims *jwtmgr.Claims) ([]dto.AppDTO, error) {
	apps, err := s.repos.App.ListForTenant(claims.TenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AppDTO, 0, len(apps))
	for _, app := range apps {
		if app.RequiredPerm != "" && !claims.IsPlatform && !hasPerm(claims.Permissions, app.RequiredPerm) {
			continue
		}
		url := app.URL
		if app.Code == "productcore" && s.appCfg.ProductCoreURL != "" {
			url = s.appCfg.ProductCoreURL
		}
		if app.Code == "supplycore" && s.appCfg.SupplyCoreURL != "" {
			url = s.appCfg.SupplyCoreURL
		}
		if app.Code == "aftersalescore" && s.appCfg.AfterSalesCoreURL != "" {
			url = s.appCfg.AfterSalesCoreURL
		}
		if app.Code == "storecore" && s.appCfg.StoreCoreURL != "" {
			url = s.appCfg.StoreCoreURL
		}
		out = append(out, dto.AppDTO{
			ID:          app.ID,
			Code:        app.Code,
			Name:        app.Name,
			Description: app.Description,
			Icon:        app.Icon,
			URL:         url,
			Sort:        app.Sort,
		})
	}
	return out, nil
}

func (s *AuthService) issueForTenant(user *model.User, tenantID uint64) (*dto.TenantBriefDTO, []string, error) {
	if tenantID == 0 {
		return nil, nil, ErrTenantRequired
	}
	tenant, err := s.repos.Tenant.GetByID(tenantID)
	if err != nil {
		return nil, nil, ErrTenantForbidden
	}
	if user.IsPlatform == 0 {
		ok, err := s.repos.User.IsMember(user.ID, tenantID)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, ErrTenantForbidden
		}
	}
	perms, err := s.repos.User.PermissionsForUser(tenantID, user.ID, user.IsPlatform == 1)
	if err != nil {
		return nil, nil, err
	}
	brief := toTenantBrief(tenant)
	return &brief, perms, nil
}

func toUserProfile(u *model.User) dto.UserProfileDTO {
	return dto.UserProfileDTO{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsPlatform:  u.IsPlatform == 1,
	}
}

func toTenantBrief(t *model.Tenant) dto.TenantBriefDTO {
	return dto.TenantBriefDTO{ID: t.ID, CompanyID: t.CompanyID, Name: t.Name, Code: t.Code}
}

func toTenantBriefs(list []model.Tenant) []dto.TenantBriefDTO {
	out := make([]dto.TenantBriefDTO, 0, len(list))
	for i := range list {
		out = append(out, toTenantBrief(&list[i]))
	}
	return out
}

func hasPerm(perms []string, code string) bool {
	for _, p := range perms {
		if p == code || p == "*" {
			return true
		}
	}
	return false
}

func HasPerm(perms []string, code string) bool { return hasPerm(perms, code) }
