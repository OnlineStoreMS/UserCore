package repo

import (
	"usercore/internal/model"

	"gorm.io/gorm"
)

type Repos struct {
	User    *UserRepo
	Tenant  *TenantRepo
	Role    *RoleRepo
	App     *AppRepo
	Company *CompanyRepo
}

func New(db *gorm.DB) *Repos {
	return &Repos{
		User:    NewUserRepo(db),
		Tenant:  NewTenantRepo(db),
		Role:    NewRoleRepo(db),
		App:     NewAppRepo(db),
		Company: NewCompanyRepo(db),
	}
}

type CompanyRepo struct{ db *gorm.DB }

func NewCompanyRepo(db *gorm.DB) *CompanyRepo { return &CompanyRepo{db: db} }

func (r *CompanyRepo) GetByID(id uint64) (*model.Company, error) {
	var c model.Company
	err := r.db.First(&c, id).Error
	return &c, err
}

func (r *CompanyRepo) Create(c *model.Company) error { return r.db.Create(c).Error }

func (r *CompanyRepo) Save(c *model.Company) error { return r.db.Save(c).Error }

func (r *CompanyRepo) List(keyword string, page, pageSize int) ([]model.Company, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	tx := r.db.Model(&model.Company{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		tx = tx.Where("name LIKE ? OR code LIKE ?", kw, kw)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Company
	offset := (page - 1) * pageSize
	err := tx.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *CompanyRepo) CountTenants(companyID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.Tenant{}).Where("company_id = ?", companyID).Count(&n).Error
	return n, err
}

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) DB() *gorm.DB { return r.db }

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *UserRepo) GetByID(id uint64) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	return &u, err
}

func (r *UserRepo) Create(u *model.User) error { return r.db.Create(u).Error }

func (r *UserRepo) Update(u *model.User) error { return r.db.Save(u).Error }

func (r *UserRepo) ListByTenant(tenantID uint64, keyword string, page, pageSize int) ([]model.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	tx := r.db.Model(&model.User{}).
		Joins("JOIN tenant_members tm ON tm.user_id = users.id AND tm.deleted_at IS NULL").
		Where("tm.tenant_id = ? AND tm.status = 1", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		tx = tx.Where("users.email LIKE ? OR users.display_name LIKE ?", kw, kw)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.User
	offset := (page - 1) * pageSize
	err := tx.Order("users.id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *UserRepo) ListTenantsForUser(userID uint64) ([]model.Tenant, error) {
	var tenants []model.Tenant
	err := r.db.Model(&model.Tenant{}).
		Joins("JOIN tenant_members tm ON tm.tenant_id = tenants.id AND tm.deleted_at IS NULL").
		Where("tm.user_id = ? AND tm.status = 1 AND tenants.status = 1", userID).
		Order("tenants.id ASC").
		Find(&tenants).Error
	return tenants, err
}

func (r *UserRepo) IsMember(userID, tenantID uint64) (bool, error) {
	var count int64
	err := r.db.Model(&model.TenantMember{}).
		Where("user_id = ? AND tenant_id = ? AND status = 1", userID, tenantID).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepo) AddMember(tenantID, userID uint64) error {
	m := model.TenantMember{TenantID: tenantID, UserID: userID, Status: 1}
	return r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Assign(model.TenantMember{Status: 1}).
		FirstOrCreate(&m).Error
}

func (r *UserRepo) RemoveMember(tenantID, userID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
			Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
			Delete(&model.TenantMember{}).Error
	})
}

func (r *UserRepo) SetUserRoles(tenantID, userID uint64, roleIDs []uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
			Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&model.UserRole{TenantID: tenantID, UserID: userID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepo) GetRoleIDs(tenantID, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&model.UserRole{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Pluck("role_id", &ids).Error
	return ids, err
}

func (r *UserRepo) PermissionsForUser(tenantID, userID uint64, isPlatform bool) ([]string, error) {
	if isPlatform {
		var codes []string
		err := r.db.Model(&model.Permission{}).Pluck("code", &codes).Error
		return codes, err
	}
	var codes []string
	err := r.db.Table("permissions p").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.tenant_id = ? AND ur.user_id = ?", tenantID, userID).
		Distinct().
		Pluck("p.code", &codes).Error
	return codes, err
}

type TenantRepo struct{ db *gorm.DB }

func NewTenantRepo(db *gorm.DB) *TenantRepo { return &TenantRepo{db: db} }

func (r *TenantRepo) GetByID(id uint64) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.First(&t, id).Error
	return &t, err
}

func (r *TenantRepo) ListByCompany(companyID uint64) ([]model.Tenant, error) {
	var list []model.Tenant
	err := r.db.Where("company_id = ? AND status = 1", companyID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *TenantRepo) Create(t *model.Tenant) error { return r.db.Create(t).Error }

func (r *TenantRepo) Save(t *model.Tenant) error { return r.db.Save(t).Error }

func (r *TenantRepo) ListAll(keyword string, page, pageSize int) ([]model.Tenant, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	tx := r.db.Model(&model.Tenant{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		tx = tx.Where("name LIKE ? OR code LIKE ?", kw, kw)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Tenant
	offset := (page - 1) * pageSize
	err := tx.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type RoleRepo struct{ db *gorm.DB }

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }

func (r *RoleRepo) ListByTenant(tenantID uint64) ([]model.Role, error) {
	var list []model.Role
	err := r.db.Where("tenant_id = ?", tenantID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *RoleRepo) GetByID(id uint64) (*model.Role, error) {
	var role model.Role
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *RoleRepo) Create(role *model.Role) error { return r.db.Create(role).Error }

func (r *RoleRepo) Update(role *model.Role) error { return r.db.Save(role).Error }

func (r *RoleRepo) Delete(id uint64) error { return r.db.Delete(&model.Role{}, id).Error }

func (r *RoleRepo) SetPermissions(roleID uint64, permCodes []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		if len(permCodes) == 0 {
			return nil
		}
		var perms []model.Permission
		if err := tx.Where("code IN ?", permCodes).Find(&perms).Error; err != nil {
			return err
		}
		for _, p := range perms {
			if err := tx.Create(&model.RolePermission{RoleID: roleID, PermissionID: p.ID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RoleRepo) GetPermissionCodes(roleID uint64) ([]string, error) {
	var codes []string
	err := r.db.Table("permissions p").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Where("rp.role_id = ?", roleID).
		Pluck("p.code", &codes).Error
	return codes, err
}

func (r *RoleRepo) ListPermissions() ([]model.Permission, error) {
	var list []model.Permission
	err := r.db.Order("app_code, code").Find(&list).Error
	return list, err
}

func (r *RoleRepo) EnsurePermissions(perms []model.Permission) error {
	for _, p := range perms {
		if err := r.db.Where("code = ?", p.Code).Assign(p).FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

type AppRepo struct{ db *gorm.DB }

func NewAppRepo(db *gorm.DB) *AppRepo { return &AppRepo{db: db} }

func (r *AppRepo) ListEnabled() ([]model.Application, error) {
	var list []model.Application
	err := r.db.Where("enabled = 1").Order("sort DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *AppRepo) Upsert(app *model.Application) error {
	return r.db.Where("code = ?", app.Code).Assign(app).FirstOrCreate(app).Error
}

func (r *AppRepo) ListForTenant(tenantID uint64) ([]model.Application, error) {
	var grants int64
	r.db.Model(&model.AppTenantGrant{}).Where("tenant_id = ?", tenantID).Count(&grants)
	if grants == 0 {
		return r.ListEnabled()
	}
	var list []model.Application
	err := r.db.Model(&model.Application{}).
		Joins("JOIN app_tenant_grants g ON g.app_id = applications.id").
		Where("g.tenant_id = ? AND applications.enabled = 1", tenantID).
		Order("applications.sort DESC, applications.id ASC").
		Find(&list).Error
	return list, err
}
