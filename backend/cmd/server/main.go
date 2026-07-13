package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"readlab/backend/internal/config"
	"readlab/backend/internal/middleware"
	"readlab/backend/internal/model"
	"readlab/backend/internal/router"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func migrateDB(db *gorm.DB) {
	if db.Migrator().HasColumn(&model.User{}, "is_admin") {
		slog.Info("migrating is_admin to role")
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'member'")
		db.Exec("UPDATE users SET role = 'admin' WHERE is_admin = TRUE")
		db.Exec("UPDATE users SET role = 'member' WHERE is_admin = FALSE OR is_admin IS NULL")
		if err := db.Migrator().DropColumn(&model.User{}, "is_admin"); err != nil {
			slog.Warn("failed to drop is_admin column", "error", err)
		}
		slog.Info("is_admin migration complete")
	}

	if !db.Migrator().HasColumn(&model.Review{}, "edit_count") {
		slog.Info("migrating reviews: adding edit_count, parent_id, new index")
		db.Exec("ALTER TABLE reviews ADD COLUMN IF NOT EXISTS edit_count INTEGER DEFAULT 0")
		db.Exec("ALTER TABLE reviews ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES reviews(id)")
		if err := db.Migrator().DropIndex(&model.Review{}, "idx_user_novel"); err != nil {
			slog.Warn("failed to drop idx_user_novel", "error", err)
		}
		if err := db.Migrator().CreateIndex(&model.Review{}, "idx_user_novel_parent"); err != nil {
			slog.Warn("failed to create idx_user_novel_parent", "error", err)
		}
		db.Exec("ALTER TABLE reviews DROP CONSTRAINT IF EXISTS chk_reviews_rating")
		db.Exec("ALTER TABLE reviews ADD CONSTRAINT chk_reviews_rating CHECK (rating >= 0 AND rating <= 5)")
		slog.Info("review migration complete")
	}

	if err := db.AutoMigrate(
		&model.Genre{},
		&model.Tag{},
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
		&model.TicketUnit{},
	); err != nil {
		slog.Error("failed to migrate", "error", err)
	}

	migrateCoverURLs(db)

	db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_created_at ON novels(created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_views ON novels(views DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_title_trgm ON novels USING GIN (title gin_trgm_ops)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novels_author ON novels(author)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chapters_novel_id_number ON chapters(novel_id, number)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reading_history_user_novel ON reading_histories(user_id, novel_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_novel_follows_user_novel ON novel_follows(user_id, novel_id)")

	seedTicketConfigs(db)
	seedTicketUnits(db)
	migrateSpentToBanked(db)
	migrateAllToBank(db)

	slog.Info("database migration completed")
}

func migrateCoverURLs(db *gorm.DB) {
	var novels []model.Novel
	db.Where("cover_url LIKE '/home/%/.lncrawl/novels/%'").Find(&novels)
	if len(novels) == 0 {
		return
	}
	slog.Info("fixing lncrawl cover URLs", "count", len(novels))
	prefix := "/.lncrawl/novels/"
	for _, n := range novels {
		idx := strings.Index(n.CoverURL, prefix)
		if idx == -1 {
			continue
		}
		suffix := n.CoverURL[idx+len(prefix):]
		newURL := "/api/covers/" + suffix
		db.Model(&n).Update("cover_url", newURL)
	}
	slog.Info("lncrawl cover URL migration complete")
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

func seedTicketUnits(db *gorm.DB) {
	var count int64
	db.Model(&model.TicketUnit{}).Count(&count)
	if count > 0 {
		return
	}

	var users []model.User
	db.Where("tickets > 0").Find(&users)
	if len(users) == 0 {
		return
	}

	slog.Info("seeding ticket units from existing balances", "count", len(users))
	units := make([]model.TicketUnit, 0, len(users))
	for _, u := range users {
		units = append(units, model.TicketUnit{
			Serial: model.NewSerial(),
			UserID: u.ID,
			Amount: u.Tickets,
			Status: "active",
		})
	}
	db.Create(&units)
	slog.Info("ticket units seeded", "units", len(units))
}

func migrateSpentToBanked(db *gorm.DB) {
	var count int64
	db.Model(&model.TicketUnit{}).Where("status = 'spent'").Count(&count)
	if count == 0 {
		return
	}
	slog.Info("migrating spent ticket units to banked", "count", count)
	db.Model(&model.TicketUnit{}).Where("status = 'spent'").Update("status", "banked")
}

func migrateAllToBank(db *gorm.DB) {
	var flag int64
	db.Model(&model.TicketConfig{}).Where("key = 'bank_seeded'").Count(&flag)
	if flag > 0 {
		return
	}

	var marked int64
	db.Model(&model.TicketUnit{}).Where("status = 'active'").Count(&marked)
	if marked == 0 {
		db.Create(&model.TicketConfig{Key: "bank_seeded", Value: 1, Label: "Bank seed flag (1 = seeded)"})
		return
	}

	slog.Info("moving all active ticket units to bank", "count", marked)
	db.Model(&model.TicketUnit{}).Where("status = 'active'").Update("status", "banked")
	db.Exec("UPDATE users SET tickets = (SELECT COALESCE(SUM(amount), 0) FROM ticket_units WHERE user_id = users.id AND status = 'active')")
	db.Create(&model.TicketConfig{Key: "bank_seeded", Value: 1, Label: "Bank seed flag (1 = seeded)"})
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

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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
