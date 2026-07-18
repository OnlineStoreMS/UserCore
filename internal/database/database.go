package database

import (
	"fmt"
	"os"
	"path/filepath"

	"usercore/internal/config"
	"usercore/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.PostgresDSN)
	case "sqlite":
		if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o755); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
	return gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Company{},
		&model.Tenant{},
		&model.User{},
		&model.TenantMember{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.Application{},
		&model.UserAppOrder{},
		&model.AppTenantGrant{},
	)
}
