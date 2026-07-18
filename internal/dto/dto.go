package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	TenantID uint64 `json:"tenantId"`
}

type LoginResponse struct {
	AccessToken  string           `json:"accessToken"`
	ExpiresAt    int64            `json:"expiresAt"`
	RefreshToken string           `json:"refreshToken,omitempty"`
	User         UserProfileDTO   `json:"user"`
	Tenant       TenantBriefDTO   `json:"tenant"`
	Permissions  []string         `json:"permissions"`
	Tenants      []TenantBriefDTO `json:"tenants,omitempty"`
}

type UserProfileDTO struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	IsPlatform  bool   `json:"isPlatform"`
}

type TenantBriefDTO struct {
	ID        uint64 `json:"id"`
	CompanyID uint64 `json:"companyId"`
	Name      string `json:"name"`
	Code      string `json:"code"`
}

type MeResponse struct {
	User        UserProfileDTO   `json:"user"`
	Tenant      TenantBriefDTO   `json:"tenant"`
	Permissions []string         `json:"permissions"`
	Tenants     []TenantBriefDTO `json:"tenants"`
}

type AppDTO struct {
	ID          uint64 `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	URL         string `json:"url"`
	Sort        int    `json:"sort"`
}

// SaveAppOrderRequest 应用中心拖拽排序；appIds 为展示顺序（靠前的优先）
type SaveAppOrderRequest struct {
	AppIDs []uint64 `json:"appIds" binding:"required,min=1"`
}

type CreateUserRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=6"`
	DisplayName string   `json:"displayName" binding:"required"`
	Phone       string   `json:"phone"`
	RoleIDs     []uint64 `json:"roleIds"`
}

type UpdateUserRequest struct {
	DisplayName string   `json:"displayName"`
	Phone       string   `json:"phone"`
	Status      *int8    `json:"status"`
	RoleIDs     []uint64 `json:"roleIds"`
	Password    string   `json:"password"`
}

type UserDTO struct {
	ID          uint64     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	Phone       string     `json:"phone"`
	Status      int8       `json:"status"`
	IsPlatform  bool       `json:"isPlatform"`
	Roles       []RoleDTO  `json:"roles"`
	CreatedAt   string     `json:"createdAt"`
}

type RoleDTO struct {
	ID          uint64   `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsBuiltin   bool     `json:"isBuiltin"`
	Permissions []string `json:"permissions"`
}

type CreateRoleRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CreateTenantRequest struct {
	CompanyID uint64 `json:"companyId"`
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Remark    string `json:"remark"`
}

type UpdateTenantRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status *int8  `json:"status"`
	Remark string `json:"remark"`
}

type TenantDTO struct {
	ID          uint64 `json:"id"`
	CompanyID   uint64 `json:"companyId"`
	CompanyName string `json:"companyName,omitempty"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Status      int8   `json:"status"`
	Remark      string `json:"remark"`
}

type CompanyDTO struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Status      int8   `json:"status"`
	Remark      string `json:"remark"`
	TenantCount int64  `json:"tenantCount"`
}

type CreateCompanyRequest struct {
	Name   string `json:"name" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Remark string `json:"remark"`
}

type UpdateCompanyRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status *int8  `json:"status"`
	Remark string `json:"remark"`
}

type PermissionDTO struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	AppCode string `json:"appCode"`
}

type PageQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
}

type SwitchTenantRequest struct {
	TenantID uint64 `json:"tenantId" binding:"required"`
}
