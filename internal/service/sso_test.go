package service_test

import (
	"testing"
	"time"

	"usercore/internal/config"
	"usercore/internal/dto"
	"usercore/internal/model"
	jwtmgr "usercore/internal/pkg/jwt"
	"usercore/internal/pkg/password"
	"usercore/internal/repo"
	"usercore/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSSOTest(t *testing.T) (*service.AuthService, *jwtmgr.Claims, string) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Company{},
		&model.Tenant{},
		&model.User{},
		&model.TenantMember{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.Application{},
		&model.UserAppOrder{},
		&model.AppTenantGrant{},
		&model.SSOAuthCode{},
		&model.RefreshSession{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	company := &model.Company{Name: "Demo Co", Code: "demo", Status: 1}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	tenant := &model.Tenant{CompanyID: company.ID, Name: "Demo Tenant", Code: "demo", Status: 1}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("tenant: %v", err)
	}
	hash, err := password.Hash("demo123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	user := &model.User{
		Email: "admin@demo.com", Password: hash, DisplayName: "Admin", Status: 1, IsPlatform: 1,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	if err := db.Create(&model.TenantMember{TenantID: tenant.ID, UserID: user.ID, Status: 1}).Error; err != nil {
		t.Fatalf("member: %v", err)
	}
	app := &model.Application{
		Code: "productcore", Name: "商品", URL: "http://localhost:5173",
		Sort: 1, Enabled: 1, RequiredPerm: "product:read",
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("app: %v", err)
	}
	perm := &model.Permission{Code: "product:read", Name: "查看商品", AppCode: "productcore"}
	if err := db.Create(perm).Error; err != nil {
		t.Fatalf("perm: %v", err)
	}

	repos := repo.New(db)
	jwt := jwtmgr.NewManager("test-secret", 120, 168)
	appsCfg := &config.AppsConfig{ProductCoreURL: "http://localhost:5173"}
	svc := service.NewAuthService(repos, jwt, appsCfg, 60)

	claims := &jwtmgr.Claims{
		UserID:      user.ID,
		CompanyID:   company.ID,
		TenantID:    tenant.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Permissions: []string{"product:read"},
		IsPlatform:  true,
	}
	return svc, claims, "http://localhost:5173/auth/callback"
}

func TestAuthorizeAndExchangeSSO(t *testing.T) {
	svc, claims, callback := setupSSOTest(t)

	authz, err := svc.AuthorizeSSO(claims, dto.SSOAuthorizeRequest{AppCode: "productcore"})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if authz.Code == "" || authz.RedirectURI != callback || authz.ExpiresIn != 60 {
		t.Fatalf("unexpected authorize resp: %+v", authz)
	}

	tok, err := svc.ExchangeSSOCode(dto.SSOTokenRequest{Code: authz.Code, RedirectURI: callback})
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if tok.AccessToken == "" || tok.RefreshToken == "" {
		t.Fatalf("missing tokens: %+v", tok)
	}
}

func TestSSOCodeSingleUse(t *testing.T) {
	svc, claims, callback := setupSSOTest(t)
	authz, err := svc.AuthorizeSSO(claims, dto.SSOAuthorizeRequest{AppCode: "productcore"})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if _, err := svc.ExchangeSSOCode(dto.SSOTokenRequest{Code: authz.Code, RedirectURI: callback}); err != nil {
		t.Fatalf("first exchange: %v", err)
	}
	if _, err := svc.ExchangeSSOCode(dto.SSOTokenRequest{Code: authz.Code, RedirectURI: callback}); err == nil {
		t.Fatal("expected second exchange to fail")
	} else if err != service.ErrInvalidSSOCode {
		t.Fatalf("want ErrInvalidSSOCode, got %v", err)
	}
}

func TestSSORedirectMismatch(t *testing.T) {
	svc, claims, _ := setupSSOTest(t)
	authz, err := svc.AuthorizeSSO(claims, dto.SSOAuthorizeRequest{AppCode: "productcore"})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if _, err := svc.ExchangeSSOCode(dto.SSOTokenRequest{
		Code: authz.Code, RedirectURI: "http://evil.example/auth/callback",
	}); err == nil {
		t.Fatal("expected redirect mismatch to fail")
	} else if err != service.ErrInvalidSSOCode {
		t.Fatalf("want ErrInvalidSSOCode, got %v", err)
	}
}

func TestSSOAuthorizeByRedirectURI(t *testing.T) {
	svc, claims, callback := setupSSOTest(t)
	authz, err := svc.AuthorizeSSO(claims, dto.SSOAuthorizeRequest{RedirectURI: callback})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if authz.RedirectURI != callback {
		t.Fatalf("redirect: %s", authz.RedirectURI)
	}
}

func TestSSOAuthorizeForbiddenApp(t *testing.T) {
	svc, claims, _ := setupSSOTest(t)
	if _, err := svc.AuthorizeSSO(claims, dto.SSOAuthorizeRequest{AppCode: "nosuch"}); err == nil {
		t.Fatal("expected forbidden")
	} else if err != service.ErrSSOAppForbidden {
		t.Fatalf("want ErrSSOAppForbidden, got %v", err)
	}
}

func TestSSOAuthorizeOpenRedirectBlocked(t *testing.T) {
	svc, claims, _ := setupSSOTest(t)
	if _, err := svc.AuthorizeSSO(claims, dto.SSOAuthorizeRequest{
		RedirectURI: "http://evil.example/auth/callback",
	}); err == nil {
		t.Fatal("expected forbidden open redirect")
	} else if err != service.ErrSSOAppForbidden {
		t.Fatalf("want ErrSSOAppForbidden, got %v", err)
	}
}

func TestSSOCodeExpired(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = db.AutoMigrate(
		&model.Company{}, &model.Tenant{}, &model.User{}, &model.TenantMember{},
		&model.Role{}, &model.Permission{}, &model.RolePermission{}, &model.UserRole{},
		&model.Application{}, &model.UserAppOrder{}, &model.AppTenantGrant{}, &model.SSOAuthCode{}, &model.RefreshSession{},
	)
	company := &model.Company{Name: "C", Code: "c", Status: 1}
	_ = db.Create(company)
	tenant := &model.Tenant{CompanyID: company.ID, Name: "T", Code: "t", Status: 1}
	_ = db.Create(tenant)
	hash, _ := password.Hash("demo123")
	user := &model.User{Email: "u@demo.com", Password: hash, DisplayName: "U", Status: 1, IsPlatform: 1}
	_ = db.Create(user)
	_ = db.Create(&model.TenantMember{TenantID: tenant.ID, UserID: user.ID, Status: 1})
	_ = db.Create(&model.Application{
		Code: "productcore", Name: "P", URL: "http://localhost:5173", Enabled: 1, RequiredPerm: "product:read",
	})
	_ = db.Create(&model.Permission{Code: "product:read", Name: "r", AppCode: "productcore"})

	callback := "http://localhost:5173/auth/callback"
	short := service.NewAuthService(repo.New(db), jwtmgr.NewManager("s", 120, 168), &config.AppsConfig{
		ProductCoreURL: "http://localhost:5173",
	}, 1)
	cl := &jwtmgr.Claims{
		UserID: user.ID, CompanyID: company.ID, TenantID: tenant.ID,
		Email: user.Email, Permissions: []string{"product:read"}, IsPlatform: true,
	}
	authz, err := short.AuthorizeSSO(cl, dto.SSOAuthorizeRequest{AppCode: "productcore"})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := short.ExchangeSSOCode(dto.SSOTokenRequest{Code: authz.Code, RedirectURI: callback}); err == nil {
		t.Fatal("expected expired code to fail")
	} else if err != service.ErrInvalidSSOCode {
		t.Fatalf("want ErrInvalidSSOCode, got %v", err)
	}
}
