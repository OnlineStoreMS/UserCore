package service

import "usercore/internal/repo"

func EnsurePlatformUsersInTenant(repos *repo.Repos, tenantID uint64) error {
	users, err := repos.User.ListPlatformUsers()
	if err != nil {
		return err
	}
	for i := range users {
		if err := repos.User.AddMember(tenantID, users[i].ID); err != nil {
			return err
		}
	}
	return nil
}

func SyncAllPlatformMembers(repos *repo.Repos) error {
	users, err := repos.User.ListPlatformUsers()
	if err != nil {
		return err
	}
	tenants, err := repos.Tenant.ListActive()
	if err != nil {
		return err
	}
	for i := range tenants {
		for j := range users {
			if err := repos.User.AddMember(tenants[i].ID, users[j].ID); err != nil {
				return err
			}
		}
	}
	return nil
}
