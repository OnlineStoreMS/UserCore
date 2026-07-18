package service

import (
	"errors"
	"sort"
	"strings"

	"usercore/internal/config"
	"usercore/internal/dto"
	"usercore/internal/model"
	jwtmgr "usercore/internal/pkg/jwt"
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
		out = append(out, dto.AppDTO{
			ID:          app.ID,
			Code:        app.Code,
			Name:        app.Name,
			Description: app.Description,
			Icon:        app.Icon,
			URL:         s.resolveAppURL(app),
			Sort:        app.Sort,
		})
	}
	s.applyUserAppOrder(claims.UserID, out)
	return out, nil
}

func (s *AuthService) SaveAppOrder(claims *jwtmgr.Claims, appIDs []uint64) error {
	if claims == nil || claims.UserID == 0 {
		return ErrInvalidCredentials
	}
	allowed, err := s.ListApps(claims)
	if err != nil {
		return err
	}
	allowedSet := make(map[uint64]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a.ID] = struct{}{}
	}
	filtered := make([]uint64, 0, len(appIDs))
	seen := make(map[uint64]struct{}, len(appIDs))
	for _, id := range appIDs {
		if id == 0 {
			continue
		}
		if _, ok := allowedSet[id]; !ok {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		filtered = append(filtered, id)
	}
	// append any allowed apps missing from payload (keep them after)
	for _, a := range allowed {
		if _, ok := seen[a.ID]; !ok {
			filtered = append(filtered, a.ID)
		}
	}
	return s.repos.App.ReplaceUserAppOrders(claims.UserID, filtered)
}

func (s *AuthService) applyUserAppOrder(userID uint64, apps []dto.AppDTO) {
	if userID == 0 || len(apps) == 0 {
		return
	}
	orders, err := s.repos.App.ListUserAppOrders(userID)
	if err != nil || len(orders) == 0 {
		return
	}
	orderMap := make(map[uint64]int, len(orders))
	for _, o := range orders {
		orderMap[o.AppID] = o.Sort
	}
	sort.SliceStable(apps, func(i, j int) bool {
		oi, oki := orderMap[apps[i].ID]
		oj, okj := orderMap[apps[j].ID]
		if oki && okj {
			if oi != oj {
				return oi > oj
			}
			return apps[i].ID < apps[j].ID
		}
		if oki != okj {
			return oki
		}
		if apps[i].Sort != apps[j].Sort {
			return apps[i].Sort > apps[j].Sort
		}
		return apps[i].ID < apps[j].ID
	})
	for i := range apps {
		if s, ok := orderMap[apps[i].ID]; ok {
			apps[i].Sort = s
		}
	}
}

func (s *AuthService) resolveAppURL(app model.Application) string {
	url := app.URL
	if s.appCfg == nil {
		return url
	}
	switch app.Code {
	case "productcore":
		if s.appCfg.ProductCoreURL != "" {
			return s.appCfg.ProductCoreURL
		}
	case "supplycore":
		if s.appCfg.SupplyCoreURL != "" {
			return s.appCfg.SupplyCoreURL
		}
	case "aftersalescore":
		if s.appCfg.AfterSalesCoreURL != "" {
			return s.appCfg.AfterSalesCoreURL
		}
	case "storecore":
		if s.appCfg.StoreCoreURL != "" {
			return s.appCfg.StoreCoreURL
		}
	case "warehousecore":
		if s.appCfg.WarehouseCoreURL != "" {
			return s.appCfg.WarehouseCoreURL
		}
	case "ordercore":
		if s.appCfg.OrderCoreURL != "" {
			return s.appCfg.OrderCoreURL
		}
	case "customercore":
		if s.appCfg.CustomerCoreURL != "" {
			return s.appCfg.CustomerCoreURL
		}
	case "mallcore":
		if s.appCfg.MallCoreURL != "" {
			return s.appCfg.MallCoreURL
		}
	case "shippingcore":
		if s.appCfg.ShippingCoreURL != "" {
			return s.appCfg.ShippingCoreURL
		}
	case "storesyncagent":
		if s.appCfg.StoreSyncAgentURL != "" {
			return s.appCfg.StoreSyncAgentURL
		}
	}
	return url
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
