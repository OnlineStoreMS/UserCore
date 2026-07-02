package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"usercore/internal/config"
	"usercore/internal/database"
	"usercore/internal/router"
	"usercore/internal/seed"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	absConfig, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load(absConfig)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	seed.EnsureApps(db, cfg.Apps)
	seed.Run(db, cfg.Apps.ProductCoreURL, cfg.Apps.SupplyCoreURL, cfg.Apps.AfterSalesCoreURL)
	seed.SyncPlatformMembers(db)

	engine := router.Setup(db, cfg)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("UserCore API listening on http://localhost%s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
