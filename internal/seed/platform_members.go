package seed

import (
	"log"

	"usercore/internal/model"
	"usercore/internal/repo"
	"usercore/internal/service"

	"gorm.io/gorm"
)

func SyncPlatformMembers(db *gorm.DB) {
	repos := repo.New(db)
	perms := []model.Permission{
		{Code: "platform:admin", Name: "平台租户管理", AppCode: "usercore"},
	}
	if err := repos.Role.EnsurePermissions(perms); err != nil {
		log.Printf("ensure platform permissions: %v", err)
		return
	}
	if err := service.SyncAllPlatformMembers(repos); err != nil {
		log.Printf("sync platform tenant members failed: %v", err)
		return
	}
	log.Println("platform tenant members synced")
}
