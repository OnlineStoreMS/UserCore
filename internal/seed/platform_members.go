package seed

import (
	"log"

	"usercore/internal/repo"
	"usercore/internal/service"

	"gorm.io/gorm"
)

func SyncPlatformMembers(db *gorm.DB) {
	repos := repo.New(db)
	if err := service.SyncAllPlatformMembers(repos); err != nil {
		log.Printf("sync platform tenant members failed: %v", err)
		return
	}
	log.Println("platform tenant members synced")
}
