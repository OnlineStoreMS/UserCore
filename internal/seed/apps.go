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
		{Code: "customer:read", Name: "查看客户", AppCode: "customercore"},
		{Code: "customer:write", Name: "编辑客户", AppCode: "customercore"},
		{Code: "supply:read", Name: "查看供应链", AppCode: "supplycore"},
		{Code: "supply:write", Name: "编辑供应链", AppCode: "supplycore"},
		{Code: "aftersales:read", Name: "查看售后", AppCode: "aftersalescore"},
		{Code: "aftersales:write", Name: "编辑售后", AppCode: "aftersalescore"},
		{Code: "store:read", Name: "查看门店", AppCode: "storecore"},
		{Code: "store:write", Name: "编辑门店", AppCode: "storecore"},
		{Code: "storesync:read", Name: "查看电商店铺同步", AppCode: "storesyncagent"},
		{Code: "storesync:write", Name: "编辑电商店铺同步", AppCode: "storesyncagent"},
		{Code: "warehouse:read", Name: "查看仓储", AppCode: "warehousecore"},
		{Code: "warehouse:write", Name: "编辑仓储", AppCode: "warehousecore"},
		{Code: "shipping:read", Name: "查看发货", AppCode: "shippingcore"},
		{Code: "shipping:write", Name: "编辑发货", AppCode: "shippingcore"},
		{Code: "order:read", Name: "查看订单中心", AppCode: "ordercore"},
		{Code: "order:write", Name: "编辑订单中心", AppCode: "ordercore"},
		{Code: "mall:read", Name: "查看私域商城", AppCode: "mallcore"},
		{Code: "mall:write", Name: "编辑私域商城", AppCode: "mallcore"},
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

	storeSyncURL := apps.StoreSyncAgentURL
	if storeSyncURL == "" {
		storeSyncURL = "http://localhost:5178"
	}
	storeSyncApp := model.Application{
		Code: "storesyncagent", Name: "电商店铺同步",
		Description: "电商店铺订单与售后同步：快递助手对接、订单拉取、售后提醒与退换货管理",
		Icon: "Shop", URL: storeSyncURL,
		Sort: 65, Enabled: 1, RequiredPerm: "storesync:read",
	}
	if err := r.App.Upsert(&storeSyncApp); err != nil {
		log.Printf("ensure storesyncagent app: %v", err)
		return
	}


	customerURL := apps.CustomerCoreURL
	if customerURL == "" {
		customerURL = "http://localhost:5183"
	}
	customerApp := model.Application{
		Code: "customercore", Name: "客户中心",
		Description: "平台客户底库：手机号建档、收货地址与全渠道身份绑定",
		Icon: "User", URL: customerURL,
		Sort: 55, Enabled: 1, RequiredPerm: "customer:read",
	}
	if err := r.App.Upsert(&customerApp); err != nil {
		log.Printf("ensure customercore app: %v", err)
		return
	}

	warehouseURL := apps.WarehouseCoreURL
	if warehouseURL == "" {
		warehouseURL = "http://localhost:5180"
	}
	warehouseApp := model.Application{
		Code: "warehousecore", Name: "仓储中心",
		Description: "OSMS 仓储中心：仓配商品、仓库货位、库存账、盘点、调拨与其他出入库",
		Icon: "Box", URL: warehouseURL,
		Sort: 60, Enabled: 1, RequiredPerm: "warehouse:read",
	}
	if err := r.App.Upsert(&warehouseApp); err != nil {
		log.Printf("ensure warehousecore app: %v", err)
		return
	}

	shippingURL := apps.ShippingCoreURL
	if shippingURL == "" {
		shippingURL = "http://localhost:5181"
	}
	shippingApp := model.Application{
		Code: "shippingcore", Name: "发货中心",
		Description: "电子面单与顺丰对接：发货人、承运商、运单创建与面单打印",
		Icon: "Van", URL: shippingURL,
		Sort: 62, Enabled: 1, RequiredPerm: "shipping:read",
	}
	if err := r.App.Upsert(&shippingApp); err != nil {
		log.Printf("ensure shippingcore app: %v", err)
		return
	}

	orderURL := apps.OrderCoreURL
	if orderURL == "" {
		orderURL = "http://localhost:5182"
	}
	orderApp := model.Application{
		Code: "ordercore", Name: "订单中心",
		Description: "OSMS 订单中心：汇聚电商/门店/手工/小程序订单，自营·代发·采购发货分配与物流回传",
		Icon: "Tickets", URL: orderURL,
		Sort: 95, Enabled: 1, RequiredPerm: "order:read",
	}
	if err := r.App.Upsert(&orderApp); err != nil {
		log.Printf("ensure ordercore app: %v", err)
		return
	}

	mallURL := apps.MallCoreURL
	if mallURL == "" {
		mallURL = "http://localhost:5184"
	}
	mallApp := model.Application{
		Code: "mallcore", Name: "私域商城",
		Description: "微信小程序私域商城：上架管理、小程序订单、门店服务预约与支付配置",
		Icon: "ShoppingBag", URL: mallURL,
		Sort: 75, Enabled: 1, RequiredPerm: "mall:read",
	}
	if err := r.App.Upsert(&mallApp); err != nil {
		log.Printf("ensure mallcore app: %v", err)
		return
	}
	log.Println("apps ensured: supplycore, aftersalescore, storecore, storesyncagent, warehousecore, shippingcore, ordercore, customercore, mallcore")
}
