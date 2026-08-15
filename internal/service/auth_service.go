package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

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
	ErrInvalidRefresh     = errors.New("刷新凭证无效或已过期，请重新登录")
	ErrSSOAppForbidden    = errors.New("无权进入该应用")
	ErrInvalidSSOCode     = errors.New("授权码无效或已过期")
	ErrSSORequestInvalid  = errors.New("请指定 appCode 或 redirectUri")
)

type AuthService struct {
	repos      *repo.Repos
	jwt        *jwtmgr.Manager
	appCfg     *config.AppsConfig
	ssoCodeTTL time.Duration
}

func NewAuthService(repos *repo.Repos, jwt *jwtmgr.Manager, appCfg *config.AppsConfig, ssoCodeTTLSeconds int) *AuthService {
	ttl := time.Duration(ssoCodeTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return &AuthService{repos: repos, jwt: jwt, appCfg: appCfg, ssoCodeTTL: ttl}
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

	tenants, err := s.listTenantsForSession(user)
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
	access, refresh, exp, err := s.issueTokenPair(user, tenant, perms, user.IsPlatform == 1)
	if err != nil {
		return nil, err
	}
	resp.AccessToken = access
	resp.RefreshToken = refresh
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
	access, refresh, exp, err := s.issueTokenPair(user, tenant, perms, isPlatform)
	if err != nil {
		return nil, err
	}
	tenants, err := s.listTenantsForSession(user)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp.Unix(),
		User:         toUserProfile(user),
		Tenant:       *tenant,
		Permissions:  perms,
		Tenants:      toTenantBriefs(tenants),
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (*dto.RefreshResponse, error) {
	claims, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if claims.ID == "" {
		return nil, ErrInvalidRefresh
	}
	now := time.Now()
	sess, err := s.repos.App.GetRefreshSessionByJTIHash(hashSSOCode(claims.ID))
	if err != nil || sess.RevokedAt != nil || !sess.ExpiresAt.After(now) {
		// reuse detection: revoked+replaced → invalidate family by revoking... simplified: just reject
		return nil, ErrInvalidRefresh
	}
	user, err := s.repos.User.GetByID(claims.UserID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if user.Status != 1 {
		return nil, ErrUserDisabled
	}
	tenant, perms, err := s.issueForTenant(user, claims.TenantID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	access, refresh, exp, err := s.issueTokenPairRotating(user, tenant, perms, user.IsPlatform == 1, claims.ID)
	if err != nil {
		return nil, err
	}
	return &dto.RefreshResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp.Unix(),
	}, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	claims, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil || claims.ID == "" {
		return nil
	}
	_ = s.repos.App.RevokeRefreshSessionByJTIHash(hashSSOCode(claims.ID), time.Now())
	return nil
}

func (s *AuthService) issueTokenPair(
	user *model.User,
	tenant *dto.TenantBriefDTO,
	perms []string,
	isPlatform bool,
) (accessToken, refreshToken string, exp time.Time, err error) {
	return s.issueTokenPairRotating(user, tenant, perms, isPlatform, "")
}

func (s *AuthService) issueTokenPairRotating(
	user *model.User,
	tenant *dto.TenantBriefDTO,
	perms []string,
	isPlatform bool,
	oldRefreshJTI string,
) (accessToken, refreshToken string, exp time.Time, err error) {
	accessToken, exp, err = s.jwt.IssueAccess(jwtmgr.Claims{
		UserID:      user.ID,
		CompanyID:   tenant.CompanyID,
		TenantID:    tenant.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Permissions: perms,
		IsPlatform:  isPlatform,
	})
	if err != nil {
		return "", "", time.Time{}, err
	}
	var jti string
	refreshToken, jti, refreshExp, err := s.jwt.IssueRefresh(user.ID, tenant.ID)
	if err != nil {
		return "", "", time.Time{}, err
	}
	now := time.Now()
	next := &model.RefreshSession{
		JTIHash:   hashSSOCode(jti),
		UserID:    user.ID,
		TenantID:  tenant.ID,
		ExpiresAt: refreshExp,
		CreatedAt: now,
	}
	if oldRefreshJTI != "" {
		if err := s.repos.App.RotateRefreshSession(hashSSOCode(oldRefreshJTI), next, now); err != nil {
			return "", "", time.Time{}, ErrInvalidRefresh
		}
	} else {
		if err := s.repos.App.CreateRefreshSession(next); err != nil {
			return "", "", time.Time{}, err
		}
	}
	return accessToken, refreshToken, exp, nil
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
	tenants, err := s.listTenantsForSession(user)
	if err != nil {
		return nil, err
	}
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
	case "materialcore":
		if s.appCfg.MaterialCoreURL != "" {
			return s.appCfg.MaterialCoreURL
		}
	case "catalogcore":
		if s.appCfg.CatalogCoreURL != "" {
			return s.appCfg.CatalogCoreURL
		}
	case "quotecore":
		if s.appCfg.QuoteCoreURL != "" {
			return s.appCfg.QuoteCoreURL
		}
	case "todocenter":
		if s.appCfg.TodoCenterURL != "" {
			return s.appCfg.TodoCenterURL
		}
	case "selfcore":
		if s.appCfg.SelfCoreURL != "" {
			return s.appCfg.SelfCoreURL
		}
	}
	return url
}

func appBaseURL(appURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(appURL), "/")
	if strings.HasSuffix(trimmed, "/auth/callback") {
		return strings.TrimSuffix(trimmed, "/auth/callback")
	}
	return trimmed
}

func appCallbackURL(appURL string) string {
	return appBaseURL(appURL) + "/auth/callback"
}

func normalizeRedirectURI(uri string) string {
	return strings.TrimRight(strings.TrimSpace(uri), "/")
}

func hashSSOCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func generateSSOCode() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// AuthorizeSSO 为当前会话签发一次性授权码，供子应用 /auth/callback 换票。
func (s *AuthService) AuthorizeSSO(claims *jwtmgr.Claims, req dto.SSOAuthorizeRequest) (*dto.SSOAuthorizeResponse, error) {
	if claims == nil || claims.UserID == 0 || claims.TenantID == 0 {
		return nil, ErrInvalidCredentials
	}
	appCode := strings.TrimSpace(req.AppCode)
	redirectURI := normalizeRedirectURI(req.RedirectURI)
	if appCode == "" && redirectURI == "" {
		return nil, ErrSSORequestInvalid
	}

	allowed, err := s.ListApps(claims)
	if err != nil {
		return nil, err
	}
	var target *dto.AppDTO
	for i := range allowed {
		a := &allowed[i]
		callback := normalizeRedirectURI(appCallbackURL(a.URL))
		if appCode != "" && a.Code == appCode {
			target = a
			break
		}
		if redirectURI != "" && callback == redirectURI {
			target = a
			break
		}
	}
	if target == nil {
		return nil, ErrSSOAppForbidden
	}
	callback := normalizeRedirectURI(appCallbackURL(target.URL))

	plain, err := generateSSOCode()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := &model.SSOAuthCode{
		CodeHash:    hashSSOCode(plain),
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		AppCode:     target.Code,
		RedirectURI: callback,
		ExpiresAt:   now.Add(s.ssoCodeTTL),
		CreatedAt:   now,
	}
	if err := s.repos.App.CreateSSOAuthCode(row); err != nil {
		return nil, err
	}
	return &dto.SSOAuthorizeResponse{
		Code:        plain,
		ExpiresIn:   int(s.ssoCodeTTL.Seconds()),
		RedirectURI: callback,
	}, nil
}

// ExchangeSSOCode 核销一次性 code 并重新签发 access/refresh。
func (s *AuthService) ExchangeSSOCode(req dto.SSOTokenRequest) (*dto.RefreshResponse, error) {
	code := strings.TrimSpace(req.Code)
	redirectURI := normalizeRedirectURI(req.RedirectURI)
	if code == "" || redirectURI == "" {
		return nil, ErrInvalidSSOCode
	}
	row, err := s.repos.App.ConsumeSSOAuthCode(hashSSOCode(code), redirectURI, time.Now())
	if err != nil {
		return nil, ErrInvalidSSOCode
	}
	user, err := s.repos.User.GetByID(row.UserID)
	if err != nil {
		return nil, ErrInvalidSSOCode
	}
	if user.Status != 1 {
		return nil, ErrInvalidSSOCode
	}
	tenant, perms, err := s.issueForTenant(user, row.TenantID)
	if err != nil {
		return nil, ErrInvalidSSOCode
	}
	access, refresh, exp, err := s.issueTokenPair(user, tenant, perms, user.IsPlatform == 1)
	if err != nil {
		return nil, err
	}
	return &dto.RefreshResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp.Unix(),
	}, nil
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

// listTenantsForSession: 平台管理员返回全部启用租户，便于跨租户切换；普通用户仅返回成员租户。
func (s *AuthService) listTenantsForSession(user *model.User) ([]model.Tenant, error) {
	if user.IsPlatform == 1 {
		return s.repos.Tenant.ListActive()
	}
	return s.repos.User.ListTenantsForUser(user.ID)
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
