package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wtr-lab-clone/backend/internal/config"
	"wtr-lab-clone/backend/internal/middleware"
	"wtr-lab-clone/backend/internal/model"
	"wtr-lab-clone/backend/internal/router"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func migrateDB(db *gorm.DB) {
	if db.Migrator().HasColumn(&model.User{}, "is_admin") {
		slog.Info("migrating is_admin to role")
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'member'")
		db.Exec("UPDATE users SET role = 'admin' WHERE is_admin = TRUE")
		db.Exec("UPDATE users SET role = 'member' WHERE is_admin = FALSE OR is_admin IS NULL")
		db.Migrator().DropColumn(&model.User{}, "is_admin")
		slog.Info("is_admin migration complete")
	}

	if !db.Migrator().HasColumn(&model.Review{}, "edit_count") {
		slog.Info("migrating reviews: adding edit_count, parent_id, new index")
		db.Exec("ALTER TABLE reviews ADD COLUMN IF NOT EXISTS edit_count INTEGER DEFAULT 0")
		db.Exec("ALTER TABLE reviews ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES reviews(id)")
		db.Migrator().DropIndex(&model.Review{}, "idx_user_novel")
		db.Migrator().CreateIndex(&model.Review{}, "idx_user_novel_parent")
		db.Exec("ALTER TABLE reviews DROP CONSTRAINT IF EXISTS chk_reviews_rating")
		db.Exec("ALTER TABLE reviews ADD CONSTRAINT chk_reviews_rating CHECK (rating >= 0 AND rating <= 5)")
		slog.Info("review migration complete")
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
		&model.Share{},
		&model.TicketConfig{},
		&model.PasswordResetToken{},
		&model.TokenBlacklist{},
		&model.Notification{},
	); err != nil {
		slog.Error("failed to migrate", "error", err)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_created_at ON novels(created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_views ON novels(views DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_title_trgm ON novels USING GIN (title gin_trgm_ops)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_author ON novels(author)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chapters_novel_id_number ON chapters(novel_id, number)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reading_history_user_novel ON reading_histories(user_id, novel_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novel_follows_user_novel ON novel_follows(user_id, novel_id)")

	seedTicketConfigs(db)

	slog.Info("database migration completed")
}

func seedTicketConfigs(db *gorm.DB) {
	defaults := model.DefaultTicketConfigs()
	for _, cfg := range defaults {
		var count int64
		db.Model(&model.TicketConfig{}).Where("key = ?", cfg.Key).Count(&count)
		if count == 0 {
			db.Create(&cfg)
			slog.Info("seeded ticket config", "key", cfg.Key, "value", cfg.Value)
		}
	}
}

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	slog.Info("connection pool configured", "max_open", 25, "max_idle", 10, "max_lifetime", "30m")

	migrateDB(db)

	r := router.Setup(db, cfg.JWTSecret, cfg.FrontendURL, cfg.CookieSecure)

	middleware.StartBlacklistCleanup(db)

	port := cfg.ServerPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited")
}
