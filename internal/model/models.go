package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Company 公司（一个公司可拥有多个租户）
type Company struct {
	BaseModel
	Name   string `gorm:"size:128;not null" json:"name"`
	Code   string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Status int8   `gorm:"default:1" json:"status"` // 1启用 0禁用
	Remark string `gorm:"size:512" json:"remark"`
}

func (Company) TableName() string { return "companies" }

// Tenant 租户（业务隔离单元，归属某公司）
type Tenant struct {
	BaseModel
	CompanyID uint64 `gorm:"index;not null" json:"companyId"`
	Name      string `gorm:"size:128;not null" json:"name"`
	Code      string `gorm:"size:64;not null" json:"code"`
	Status    int8   `gorm:"default:1" json:"status"`
	Remark    string `gorm:"size:512" json:"remark"`
}

func (Tenant) TableName() string { return "tenants" }

// User 用户账号
type User struct {
	BaseModel
	Email       string `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Password    string `gorm:"size:128;not null" json:"-"`
	DisplayName string `gorm:"size:64;not null" json:"displayName"`
	Phone       string `gorm:"size:32" json:"phone"`
	Status      int8   `gorm:"default:1" json:"status"`
	IsPlatform  int8   `gorm:"default:0" json:"isPlatform"` // 平台超管
}

func (User) TableName() string { return "users" }

// TenantMember 用户在某租户下的成员关系
type TenantMember struct {
	BaseModel
	TenantID uint64 `gorm:"uniqueIndex:idx_tenant_user;not null" json:"tenantId"`
	UserID   uint64 `gorm:"uniqueIndex:idx_tenant_user;not null" json:"userId"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (TenantMember) TableName() string { return "tenant_members" }

// Role 角色（租户级；platform 角色 tenant_id=0）
type Role struct {
	BaseModel
	TenantID    uint64 `gorm:"index;not null;default:0" json:"tenantId"`
	Code        string `gorm:"size:64;not null" json:"code"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Description string `gorm:"size:256" json:"description"`
	IsBuiltin   int8   `gorm:"default:0" json:"isBuiltin"`
}

func (Role) TableName() string { return "roles" }

// Permission 权限点（全局字典）
type Permission struct {
	BaseModel
	Code        string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string `gorm:"size:64;not null" json:"name"`
	AppCode     string `gorm:"size:32;index" json:"appCode"`
	Description string `gorm:"size:256" json:"description"`
}

func (Permission) TableName() string { return "permissions" }

// RolePermission 角色-权限
type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey"`
	PermissionID uint64 `gorm:"primaryKey"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// UserRole 用户在租户下的角色
type UserRole struct {
	BaseModel
	TenantID uint64 `gorm:"uniqueIndex:idx_user_role;not null" json:"tenantId"`
	UserID   uint64 `gorm:"uniqueIndex:idx_user_role;not null" json:"userId"`
	RoleID   uint64 `gorm:"uniqueIndex:idx_user_role;not null" json:"roleId"`
}

func (UserRole) TableName() string { return "user_roles" }

// Application 应用中心注册的应用
type Application struct {
	BaseModel
	Code        string `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Description string `gorm:"size:256" json:"description"`
	Icon        string `gorm:"size:512" json:"icon"`
	URL         string `gorm:"size:512;not null" json:"url"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Enabled     int8   `gorm:"default:1" json:"enabled"`
	RequiredPerm string `gorm:"size:64" json:"requiredPerm"` // 进入应用所需权限
}

func (Application) TableName() string { return "applications" }

// UserAppOrder 用户在应用中心的自定义排序（每人独立，不影响默认 Application.sort）
type UserAppOrder struct {
	UserID uint64 `gorm:"primaryKey" json:"userId"`
	AppID  uint64 `gorm:"primaryKey" json:"appId"`
	Sort   int    `gorm:"not null;default:0" json:"sort"`
}

func (UserAppOrder) TableName() string { return "user_app_orders" }

// AppTenantGrant 租户可使用的应用（可选，空表表示全部启用应用可用）
type AppTenantGrant struct {
	BaseModel
	TenantID uint64 `gorm:"uniqueIndex:idx_app_tenant;not null" json:"tenantId"`
	AppID    uint64 `gorm:"uniqueIndex:idx_app_tenant;not null" json:"appId"`
}

func (AppTenantGrant) TableName() string { return "app_tenant_grants" }
