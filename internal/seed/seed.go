package seed

import (
	"log"

	"usercore/internal/model"
	"usercore/internal/pkg/password"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

func Run(db *gorm.DB, productCoreURL, supplyCoreURL, afterSalesCoreURL string) {
	repos := repo.New(db)
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		log.Printf("seed check failed: %v", err)
		return
	}
	if count > 0 {
		log.Println("seed skipped: users already exist")
		return
	}
	if err := seedAll(db, repos, productCoreURL, supplyCoreURL, afterSalesCoreURL); err != nil {
		log.Printf("seed failed: %v", err)
		return
	}
	log.Println("seed completed")
}

func seedAll(db *gorm.DB, repos *repo.Repos, productCoreURL, supplyCoreURL, afterSalesCoreURL string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		r := repo.New(tx)

		if err := seedPermissions(r); err != nil {
			return err
		}

		company := model.Company{Name: "演示集团", Code: "demo-corp", Status: 1}
		if err := r.Company.Create(&company); err != nil {
			return err
		}

		tenantFashion := model.Tenant{CompanyID: company.ID, Name: "服装商品库", Code: "fashion", Status: 1}
		tenantDigital := model.Tenant{CompanyID: company.ID, Name: "3C 商品库", Code: "digital", Status: 1}
		if err := r.Tenant.Create(&tenantFashion); err != nil {
			return err
		}
		if err := r.Tenant.Create(&tenantDigital); err != nil {
			return err
		}

		if err := seedBuiltinRoles(r, tenantFashion.ID); err != nil {
			return err
		}
		if err := seedBuiltinRoles(r, tenantDigital.ID); err != nil {
			return err
		}

		pw, err := password.Hash("demo123")
		if err != nil {
			return err
		}

		platformAdmin := model.User{
			Email: "admin@demo.com", Password: pw, DisplayName: "平台管理员", Status: 1, IsPlatform: 1,
		}
		if err := r.User.Create(&platformAdmin); err != nil {
			return err
		}
		for _, tid := range []uint64{tenantFashion.ID, tenantDigital.ID} {
			if err := r.User.AddMember(tid, platformAdmin.ID); err != nil {
				return err
			}
		}

		fashionUser := model.User{
			Email: "fashion@demo.com", Password: pw, DisplayName: "服装运营", Status: 1,
		}
		if err := r.User.Create(&fashionUser); err != nil {
			return err
		}
		if err := r.User.AddMember(tenantFashion.ID, fashionUser.ID); err != nil {
			return err
		}
		ownerRole, _ := findRole(r, tenantFashion.ID, "tenant_owner")
		if ownerRole != nil {
			_ = r.User.SetUserRoles(tenantFashion.ID, fashionUser.ID, []uint64{ownerRole.ID})
		}

		digitalUser := model.User{
			Email: "digital@demo.com", Password: pw, DisplayName: "3C 运营", Status: 1,
		}
		if err := r.User.Create(&digitalUser); err != nil {
			return err
		}
		if err := r.User.AddMember(tenantDigital.ID, digitalUser.ID); err != nil {
			return err
		}
		ownerRole2, _ := findRole(r, tenantDigital.ID, "tenant_owner")
		if ownerRole2 != nil {
			_ = r.User.SetUserRoles(tenantDigital.ID, digitalUser.ID, []uint64{ownerRole2.ID})
		}

		appURL := productCoreURL
		if appURL == "" {
			appURL = "http://localhost:5173"
		}
		app := model.Application{
			Code: "productcore", Name: "商品管理中心",
			Description: "多平台电商商品底库（PIM），管理 SPU/SKU、分类、品牌与铺货",
			Icon: "Goods", URL: appURL,
			Sort: 100, Enabled: 1, RequiredPerm: "product:read",
		}
		if err := r.App.Upsert(&app); err != nil {
			return err
		}
		sURL := supplyCoreURL
		if sURL == "" {
			sURL = "http://localhost:5175"
		}
		supplyApp := model.Application{
			Code: "supplycore", Name: "供应链中心",
			Description: "供应商管理（VMS）与采购管理（PMS），维护 SKU 供货报价与采购跟单",
			Icon: "Van", URL: sURL,
			Sort: 90, Enabled: 1, RequiredPerm: "supply:read",
		}
		if err := r.App.Upsert(&supplyApp); err != nil {
			return err
		}
		asURL := afterSalesCoreURL
		if asURL == "" {
			asURL = "http://localhost:5176"
		}
		afterSalesApp := model.Application{
			Code: "aftersalescore", Name: "售后中心",
			Description: "退货包裹开箱视频录制、快递单号识别与问题凭证管理",
			Icon: "VideoCamera", URL: asURL,
			Sort: 80, Enabled: 1, RequiredPerm: "aftersales:read",
		}
		return r.App.Upsert(&afterSalesApp)
	})
}

func seedPermissions(r *repo.Repos) error {
	perms := []model.Permission{
		{Code: "product:read", Name: "查看商品", AppCode: "productcore"},
		{Code: "product:write", Name: "编辑商品", AppCode: "productcore"},
		{Code: "product:delete", Name: "删除商品", AppCode: "productcore"},
		{Code: "product:import", Name: "导入商品", AppCode: "productcore"},
		{Code: "product:export", Name: "导出商品", AppCode: "productcore"},
		{Code: "sku:manage", Name: "SKU 管理", AppCode: "productcore"},
		{Code: "brand:manage", Name: "品牌管理", AppCode: "productcore"},
		{Code: "category:manage", Name: "分类管理", AppCode: "productcore"},
		{Code: "group:manage", Name: "分组管理", AppCode: "productcore"},
		{Code: "platform:manage", Name: "渠道店铺管理", AppCode: "productcore"},
		{Code: "listing:manage", Name: "铺货管理", AppCode: "productcore"},
		{Code: "supply:read", Name: "查看供应链", AppCode: "supplycore"},
		{Code: "supply:write", Name: "编辑供应链", AppCode: "supplycore"},
		{Code: "aftersales:read", Name: "查看售后", AppCode: "aftersalescore"},
		{Code: "aftersales:write", Name: "编辑售后", AppCode: "aftersalescore"},
		{Code: "store:read", Name: "查看门店", AppCode: "storecore"},
		{Code: "store:write", Name: "编辑门店", AppCode: "storecore"},
		{Code: "storesync:read", Name: "查看电商店铺同步", AppCode: "storesyncagent"},
		{Code: "storesync:write", Name: "编辑电商店铺同步", AppCode: "storesyncagent"},
		{Code: "tenant:admin", Name: "租户用户管理", AppCode: "usercore"},
	}
	return r.Role.EnsurePermissions(perms)
}

func seedBuiltinRoles(r *repo.Repos, tenantID uint64) error {
	builtins := []struct {
		code, name string
		perms      []string
	}{
		{"tenant_owner", "租户管理员", []string{
			"product:read", "product:write", "product:delete", "product:import", "product:export",
			"sku:manage", "brand:manage", "category:manage", "group:manage",
			"platform:manage", "listing:manage", "supply:read", "supply:write",
			"aftersales:read", "aftersales:write", "store:read", "store:write",
			"storesync:read", "storesync:write", "tenant:admin",
		}},
		{"tenant_operator", "运营人员", []string{
			"product:read", "product:write", "product:import", "product:export",
			"sku:manage", "brand:manage", "category:manage", "group:manage",
			"platform:manage", "listing:manage", "supply:read", "supply:write",
			"aftersales:read", "aftersales:write", "store:read", "store:write",
			"storesync:read", "storesync:write",
		}},
		{"tenant_viewer", "只读用户", []string{"product:read", "supply:read", "aftersales:read", "store:read", "storesync:read"}},
	}
	for _, b := range builtins {
		role := &model.Role{TenantID: tenantID, Code: b.code, Name: b.name, IsBuiltin: 1}
		if err := r.Role.Create(role); err != nil {
			return err
		}
		if err := r.Role.SetPermissions(role.ID, b.perms); err != nil {
			return err
		}
	}
	return nil
}

func findRole(r *repo.Repos, tenantID uint64, code string) (*model.Role, error) {
	roles, err := r.Role.ListByTenant(tenantID)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].Code == code {
			return &roles[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
