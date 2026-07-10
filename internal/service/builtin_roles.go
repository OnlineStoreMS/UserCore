package service

import (
	"errors"

	"usercore/internal/model"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

const PlatformAdminRoleCode = "platform_admin"

// PlatformAdminRolePermissions 平台管理员在本租户内的权限（与租户管理员一致，并含 platform:admin 标识）。
func PlatformAdminRolePermissions() []string {
	return []string{
		"product:read", "product:write", "product:delete", "product:import", "product:export",
		"sku:manage", "brand:manage", "category:manage", "group:manage",
		"platform:manage", "listing:manage",
		"supply:read", "supply:write",
		"aftersales:read", "aftersales:write",
		"store:read", "store:write",
		"storesync:read", "storesync:write",
		"tenant:admin", "platform:admin",
	}
}

func SeedBuiltinRoles(repos *repo.Repos, tenantID uint64) error {
	builtins := []struct {
		code, name, desc string
		perms            []string
	}{
		{
			PlatformAdminRoleCode, "平台管理员", "平台超管在本租户的全权限",
			PlatformAdminRolePermissions(),
		},
		{
			"tenant_owner", "租户管理员", "",
			[]string{
				"product:read", "product:write", "product:delete", "product:import", "product:export",
				"sku:manage", "brand:manage", "category:manage", "group:manage",
				"platform:manage", "listing:manage",
				"supply:read", "supply:write",
				"aftersales:read", "aftersales:write",
				"store:read", "store:write",
				"storesync:read", "storesync:write",
				"tenant:admin",
			},
		},
		{
			"tenant_operator", "运营人员", "",
			[]string{
				"product:read", "product:write", "product:import", "product:export",
				"sku:manage", "brand:manage", "category:manage", "group:manage",
				"platform:manage", "listing:manage",
				"supply:read", "supply:write",
				"aftersales:read", "aftersales:write",
				"store:read", "store:write",
				"storesync:read", "storesync:write",
			},
		},
		{
			"tenant_viewer", "只读用户", "",
			[]string{"product:read", "supply:read", "aftersales:read", "store:read", "storesync:read"},
		},
	}
	for _, b := range builtins {
		role := &model.Role{
			TenantID: tenantID, Code: b.code, Name: b.name,
			Description: b.desc, IsBuiltin: 1,
		}
		if err := repos.Role.Create(role); err != nil {
			return err
		}
		if err := repos.Role.SetPermissions(role.ID, b.perms); err != nil {
			return err
		}
	}
	return nil
}

func EnsurePlatformAdminRole(repos *repo.Repos, tenantID uint64) (*model.Role, error) {
	role, err := repos.Role.GetByTenantAndCode(tenantID, PlatformAdminRoleCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role = &model.Role{
			TenantID:    tenantID,
			Code:        PlatformAdminRoleCode,
			Name:        "平台管理员",
			Description: "平台超管在本租户的全权限",
			IsBuiltin:   1,
		}
		if err := repos.Role.Create(role); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if err := repos.Role.SetPermissions(role.ID, PlatformAdminRolePermissions()); err != nil {
		return nil, err
	}
	return role, nil
}

func AssignPlatformAdminRole(repos *repo.Repos, tenantID, userID uint64) error {
	role, err := EnsurePlatformAdminRole(repos, tenantID)
	if err != nil {
		return err
	}
	ids, err := repos.User.GetRoleIDs(tenantID, userID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == role.ID {
			return nil
		}
	}
	ids = append(ids, role.ID)
	return repos.User.SetUserRoles(tenantID, userID, ids)
}
