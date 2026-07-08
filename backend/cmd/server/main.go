package main

import (
	"log"
	"os"

	"wtr-lab-clone/backend/internal/config"
	"wtr-lab-clone/backend/internal/model"
	"wtr-lab-clone/backend/internal/router"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func migrateDB(db *gorm.DB) {
	if db.Migrator().HasColumn(&model.User{}, "is_admin") {
		log.Println("migrating is_admin → role...")
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'member'")
		db.Exec("UPDATE users SET role = 'admin' WHERE is_admin = TRUE")
		db.Exec("UPDATE users SET role = 'member' WHERE is_admin = FALSE OR is_admin IS NULL")
		db.Migrator().DropColumn(&model.User{}, "is_admin")
		log.Println("is_admin migration complete")
	}

	if err := db.AutoMigrate(
		&model.Genre{},
		&model.Novel{},
		&model.NovelGenre{},
		&model.Chapter{},
		&model.User{},
		&model.Vote{},
		&model.Request{},
		&model.TicketTransaction{},
		&model.News{},
		&model.ReadingHistory{},
		&model.NovelFollow{},
		&model.Review{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	log.Println("database migration completed")
}

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	migrateDB(db)

	r := router.Setup(db, cfg.JWTSecret, cfg.FrontendURL, cfg.CookieSecure)

	port := cfg.ServerPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
