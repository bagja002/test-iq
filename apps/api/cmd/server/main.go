package main

import (
	"log"

	"test-iq-ku/apps/api/internal/config"
	"test-iq-ku/apps/api/internal/database"
	"test-iq-ku/apps/api/internal/router"
)

func main() {
	loadedEnvFiles, err := config.LoadEnvFiles()
	if err != nil {
		log.Fatalf("env loading failed: %v", err)
	}
	if len(loadedEnvFiles) > 0 {
		log.Printf("loaded env files: %v", loadedEnvFiles)
	}

	cfg := config.Load()
	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN is required")
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if cfg.AutoRunMigrations {
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	}

	if cfg.SeedDemoData {
		if err := database.SeedDemoData(db); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}

	app := router.New(cfg, db)
	log.Printf("API listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
