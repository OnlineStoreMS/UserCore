package seed

import (
	"log"

	"usercore/internal/config"
	"usercore/internal/model"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

// EnsureApps upserts application registry and permissions (safe on every startup).
func EnsureApps(db *gorm.DB, apps config.AppsConfig) {
	r := repo.New(db)
	perms := []model.Permission{
		{Code: "supply:read", Name: "查看供应链", AppCode: "supplycore"},
		{Code: "supply:write", Name: "编辑供应链", AppCode: "supplycore"},
		{Code: "aftersales:read", Name: "查看售后", AppCode: "aftersalescore"},
		{Code: "aftersales:write", Name: "编辑售后", AppCode: "aftersalescore"},
		{Code: "store:read", Name: "查看门店", AppCode: "storecore"},
		{Code: "store:write", Name: "编辑门店", AppCode: "storecore"},
	}
	if err := r.Role.EnsurePermissions(perms); err != nil {
		log.Printf("ensure app permissions: %v", err)
		return
	}

	supplyURL := apps.SupplyCoreURL
	if supplyURL == "" {
		supplyURL = "http://localhost:5175"
	}
	supplyApp := model.Application{
		Code: "supplycore", Name: "供应链中心",
		Description: "供应商管理（VMS）与采购管理（PMS），维护 SKU 供货报价与采购跟单",
		Icon: "Van", URL: supplyURL,
		Sort: 90, Enabled: 1, RequiredPerm: "supply:read",
	}
	if err := r.App.Upsert(&supplyApp); err != nil {
		log.Printf("ensure supplycore app: %v", err)
		return
	}

	afterSalesURL := apps.AfterSalesCoreURL
	if afterSalesURL == "" {
		afterSalesURL = "http://localhost:5176"
	}
	afterSalesApp := model.Application{
		Code: "aftersalescore", Name: "售后中心",
		Description: "退货包裹开箱视频录制、快递单号识别与问题凭证管理",
		Icon: "VideoCamera", URL: afterSalesURL,
		Sort: 80, Enabled: 1, RequiredPerm: "aftersales:read",
	}
	if err := r.App.Upsert(&afterSalesApp); err != nil {
		log.Printf("ensure aftersalescore app: %v", err)
		return
	}

	storeURL := apps.StoreCoreURL
	if storeURL == "" {
		storeURL = "http://localhost:5179"
	}
	storeApp := model.Application{
		Code: "storecore", Name: "门店管理",
		Description: "OSMS 门店管理：收银台、销售订单、服务工单、库存、采购、监控",
		Icon: "Shop", URL: storeURL,
		Sort: 70, Enabled: 1, RequiredPerm: "store:read",
	}
	if err := r.App.Upsert(&storeApp); err != nil {
		log.Printf("ensure storecore app: %v", err)
		return
	}
	log.Println("apps ensured: supplycore, aftersalescore, storecore")
}
