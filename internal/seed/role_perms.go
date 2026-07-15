package seed

import (
	"log"

	"usercore/internal/repo"

	"gorm.io/gorm"
)

// builtinRoleExtraPerms 新应用上线时，为已有内置角色补权限（幂等，每次启动执行）。
var builtinRoleExtraPerms = map[string][]string{
	"tenant_owner":    {"store:read", "store:write", "storesync:read", "storesync:write", "warehouse:read", "warehouse:write"},
	"tenant_operator": {"store:read", "store:write", "storesync:read", "storesync:write", "warehouse:read", "warehouse:write"},
	"tenant_viewer":   {"store:read", "storesync:read", "warehouse:read"},
	"platform_admin":  {"warehouse:read", "warehouse:write"},
}

// EnsureBuiltinRolePermissions merges new app permissions into existing builtin roles.
func EnsureBuiltinRolePermissions(db *gorm.DB) {
	r := repo.New(db)
	tenants, err := r.Tenant.ListActive()
	if err != nil {
		log.Printf("ensure builtin role perms: list tenants: %v", err)
		return
	}
	for _, tenant := range tenants {
		roles, err := r.Role.ListByTenant(tenant.ID)
		if err != nil {
			log.Printf("ensure builtin role perms: tenant %d: %v", tenant.ID, err)
			continue
		}
		for _, role := range roles {
			if role.IsBuiltin != 1 {
				continue
			}
			extra, ok := builtinRoleExtraPerms[role.Code]
			if !ok {
				continue
			}
			existing, err := r.Role.GetPermissionCodes(role.ID)
			if err != nil {
				log.Printf("ensure builtin role perms: role %d: %v", role.ID, err)
				continue
			}
			merged := mergePermCodes(existing, extra)
			if len(merged) == len(existing) {
				continue
			}
			if err := r.Role.SetPermissions(role.ID, merged); err != nil {
				log.Printf("ensure builtin role perms: set role %d: %v", role.ID, err)
			}
		}
	}
}

func mergePermCodes(base, extra []string) []string {
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))
	for _, c := range base {
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	for _, c := range extra {
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}
