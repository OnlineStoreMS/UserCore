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
		{Code: "product:read", Name: "查看商品", AppCode: "productcore"},
		{Code: "product:write", Name: "编辑商品", AppCode: "productcore"},
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
		{Code: "material:read", Name: "查看素材中心", AppCode: "materialcore"},
		{Code: "material:write", Name: "编辑素材中心", AppCode: "materialcore"},
		{Code: "catalog:read", Name: "查看产品中心", AppCode: "catalogcore"},
		{Code: "catalog:write", Name: "编辑产品中心", AppCode: "catalogcore"},
		{Code: "quote:read", Name: "查看报价中心", AppCode: "quotecore"},
		{Code: "quote:write", Name: "编辑报价中心", AppCode: "quotecore"},
		{Code: "todo:read", Name: "查看待办中心", AppCode: "todocenter"},
		{Code: "todo:write", Name: "编辑待办中心", AppCode: "todocenter"},
		{Code: "self:read", Name: "查看自营中心", AppCode: "selfcore"},
		{Code: "self:write", Name: "编辑自营中心", AppCode: "selfcore"},
		{Code: "backup:read", Name: "查看备份恢复", AppCode: "osmsbackup"},
		{Code: "backup:write", Name: "执行备份恢复", AppCode: "osmsbackup"},
	}
	if err := r.Role.EnsurePermissions(perms); err != nil {
		log.Printf("ensure app permissions: %v", err)
		return
	}

	defs := []model.Application{
		{
			Code: "productcore", Name: "商品管理中心",
			Description: "多平台电商商品底库（PIM），管理 SPU/SKU、分类、品牌与铺货",
			Icon: "Goods", URL: defaultURL(apps.ProductCoreURL, "http://localhost:5173"),
			Sort: 100, Enabled: 1, RequiredPerm: "product:read",
		},
		{
			Code: "ordercore", Name: "订单中心",
			Description: "OSMS 订单中心：汇聚电商/门店/手工/小程序订单，自营·代发·采购发货分配与物流回传",
			Icon: "Tickets", URL: defaultURL(apps.OrderCoreURL, "http://localhost:5182"),
			Sort: 95, Enabled: 1, RequiredPerm: "order:read",
		},
		{
			Code: "supplycore", Name: "供应链中心",
			Description: "供应商管理（VMS）与采购管理（PMS），维护 SKU 供货报价与采购跟单",
			Icon: "ShoppingCart", URL: defaultURL(apps.SupplyCoreURL, "http://localhost:5175"),
			Sort: 90, Enabled: 1, RequiredPerm: "supply:read",
		},
		{
			Code: "selfcore", Name: "自营中心",
			Description: "自营订单统计，本地发货回传与仓储扣库；并管理分销商、批发价与分销订单",
			Icon: "Sell", URL: defaultURL(apps.SelfCoreURL, "http://localhost:5187"),
			Sort: 92, Enabled: 1, RequiredPerm: "self:read",
		},
		{
			Code: "aftersalescore", Name: "售后中心",
			Description: "退货包裹开箱视频录制、快递单号识别与问题凭证管理",
			Icon: "Headset", URL: defaultURL(apps.AfterSalesCoreURL, "http://localhost:5176"),
			Sort: 80, Enabled: 1, RequiredPerm: "aftersales:read",
		},
		{
			Code: "mallcore", Name: "私域商城",
			Description: "微信小程序私域商城：上架管理、小程序订单、门店服务预约与支付配置",
			Icon: "ShoppingBag", URL: defaultURL(apps.MallCoreURL, "http://localhost:5184"),
			Sort: 75, Enabled: 1, RequiredPerm: "mall:read",
		},
		{
			Code: "storecore", Name: "门店管理",
			Description: "OSMS 门店管理：收银台、销售订单、服务工单、库存、采购、监控",
			Icon: "Shop", URL: defaultURL(apps.StoreCoreURL, "http://localhost:5179"),
			Sort: 70, Enabled: 1, RequiredPerm: "store:read",
		},
		{
			Code: "storesyncagent", Name: "电商店铺同步",
			Description: "电商店铺订单与售后同步：快递助手对接、订单拉取、售后提醒与退换货管理",
			Icon: "Connection", URL: defaultURL(apps.StoreSyncAgentURL, "http://localhost:5178"),
			Sort: 65, Enabled: 1, RequiredPerm: "storesync:read",
		},
		{
			Code: "shippingcore", Name: "发货中心",
			Description: "电子面单与顺丰对接：发货人、承运商、运单创建与面单打印",
			Icon: "Van", URL: defaultURL(apps.ShippingCoreURL, "http://localhost:5181"),
			Sort: 62, Enabled: 1, RequiredPerm: "shipping:read",
		},
		{
			Code: "warehousecore", Name: "仓储中心",
			Description: "OSMS 仓储中心：仓配商品、仓库货位、库存账、盘点、调拨与其他出入库",
			Icon: "Box", URL: defaultURL(apps.WarehouseCoreURL, "http://localhost:5180"),
			Sort: 60, Enabled: 1, RequiredPerm: "warehouse:read",
		},
		{
			Code: "customercore", Name: "客户中心",
			Description: "平台客户底库：手机号建档、收货地址与全渠道身份绑定",
			Icon: "UserFilled", URL: defaultURL(apps.CustomerCoreURL, "http://localhost:5183"),
			Sort: 55, Enabled: 1, RequiredPerm: "customer:read",
		},
		{
			Code: "materialcore", Name: "素材中心",
			Description: "OSMS 素材中心：商品宣传图片/视频库，自定义分类，询盘快发与手机扫码上传",
			Icon: "Picture", URL: defaultURL(apps.MaterialCoreURL, "http://localhost:5185"),
			Sort: 85, Enabled: 1, RequiredPerm: "material:read",
		},
		{
			Code: "catalogcore", Name: "产品中心",
			Description: "品牌产品知识库：系列、子规格与说明书；资料主要来自公开页面爬取，暂不与商品中心联动",
			Icon: "Collection", URL: defaultURL(apps.CatalogCoreURL, "http://localhost:5188"),
			Sort: 86, Enabled: 1, RequiredPerm: "catalog:read",
		},
		{
			Code: "quotecore", Name: "报价中心",
			Description: "客户选品与手填明细，生成可复制图片 / 下载 PDF 的专业报价单",
			Icon: "Document", URL: defaultURL(apps.QuoteCoreURL, "http://localhost:5189"),
			Sort: 87, Enabled: 1, RequiredPerm: "quote:read",
		},
		{
			Code: "todocenter", Name: "待办中心",
			Description: "统一记录电商、发货、售后、门店等业务待办，支持图片笔记与手机扫码上传",
			Icon: "List", URL: defaultURL(apps.TodoCenterURL, "http://localhost:5186"),
			Sort: 88, Enabled: 1, RequiredPerm: "todo:read",
		},
		{
			Code: "opsmobile", Name: "手机端应用中心",
			Description: "轻量手机端：手工建单、自营订单、发货待发/已发查询",
			Icon: "Iphone", URL: defaultURL(apps.OpsMobileURL, "http://localhost:5190"),
			Sort: 200, Enabled: 1, RequiredPerm: "order:read",
		},
		{
			Code: "osmsbackup", Name: "备份恢复",
			Description: "OSMS 平台数据备份与灾难恢复：PostgreSQL、MinIO、Redis 归档下载与上传恢复",
			Icon: "FolderOpened", URL: defaultURL(apps.OsmsBackupURL, "http://localhost:5191"),
			Sort: 5, Enabled: 1, RequiredPerm: "backup:read",
		},
	}

	for i := range defs {
		app := defs[i]
		if err := r.App.Upsert(&app); err != nil {
			log.Printf("ensure app %s: %v", app.Code, err)
			return
		}
	}
	log.Println("apps ensured: productcore, ordercore, supplycore, selfcore, aftersalescore, mallcore, storecore, storesyncagent, shippingcore, warehousecore, customercore, materialcore, catalogcore, quotecore, todocenter, opsmobile, osmsbackup")
}

func defaultURL(cfg, fallback string) string {
	if cfg != "" {
		return cfg
	}
	return fallback
}
