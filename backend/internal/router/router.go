package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/handler"
	"wtr-lab-clone/backend/internal/middleware"
	"wtr-lab-clone/backend/internal/service"
	"wtr-lab-clone/backend/internal/ticket"
)

func Setup(db *gorm.DB, jwtSecret string, frontendURL string, cookieSecure bool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(frontendURL))
	r.Use(middleware.Logger())
	r.Use(middleware.MetricsMiddleware())
	r.Use(middleware.CSRF(cookieSecure))

	api := r.Group("/api/v1")

	ticketCfg := ticket.NewConfig(db)

	authSvc := service.NewAuthService(db, jwtSecret, cookieSecure)
	novelSvc := service.NewNovelService(db)
	reviewSvc := service.NewReviewService(db)
	ticketSvc := service.NewTicketService(db, ticketCfg)

	healthHandler := handler.NewHealthHandler(db)
	api.GET("/health", healthHandler.Check)
	api.GET("/metrics", middleware.MetricsHandler)

	novelHandler := handler.NewNovelHandler(db, ticketCfg, novelSvc)
	api.GET("/novels", novelHandler.List)
	api.GET("/novels/trending", novelHandler.Trending)
	api.GET("/novels/recommendations", novelHandler.Recommendations)
	api.GET("/novels/random", novelHandler.Random)
	api.GET("/novels/:id", novelHandler.Get)
	api.GET("/novels/:id/chapters", novelHandler.Chapters)
	api.GET("/novels/:id/chapters/:num", middleware.OptionalAuth(jwtSecret, db), novelHandler.GetChapterByNum)

	chapterHandler := handler.NewChapterHandler(db)
	api.GET("/chapters/:id", middleware.OptionalAuth(jwtSecret, db), chapterHandler.Get)

	authHandler := handler.NewAuthHandler(db, jwtSecret, cookieSecure, ticketCfg, authSvc)
	authLimiter := middleware.NewRateLimiter(10, 1*time.Minute)
	authGroup := api.Group("/auth")
	authGroup.Use(authLimiter.Middleware())
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
	}
	authMeGroup := api.Group("/auth", middleware.AuthRequired(jwtSecret, db))
	authMeGroup.GET("/me", authHandler.Me)

	rankingHandler := handler.NewRankingHandler(db)
	api.GET("/ranking/:period", rankingHandler.Get)

	updateHandler := handler.NewUpdateHandler(db)
	api.GET("/updates", updateHandler.Recent)

	searchHandler := handler.NewSearchHandler(db, novelSvc)
	api.GET("/search", searchHandler.Search)
	api.GET("/search/autocomplete", searchHandler.Autocomplete)

	genreHandler := handler.NewGenreHandler(db)
	api.GET("/genres", genreHandler.List)

	leaderboardHandler := handler.NewLeaderboardHandler(db)
	api.GET("/leaderboard", leaderboardHandler.Get)

	newsHandler := handler.NewNewsHandler(db)
	api.GET("/news", newsHandler.List)
	api.GET("/news/:id", newsHandler.Get)

	statsHandler := handler.NewStatsHandler(db)
	api.GET("/stats", statsHandler.Get)

	api.GET("/config/upgrade-costs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"edit_reset":     ticketCfg.Get("edit_reset_cost"),
			"gate_bypass":    ticketCfg.Get("gate_bypass_cost"),
			"replace_review": ticketCfg.Get("replace_review_cost"),
		})
	})

	authorHandler := handler.NewAuthorHandler(db)
	api.GET("/author/:name/novels", authorHandler.Novels)

	userHandler := handler.NewUserHandler(db)
	api.GET("/profile/:id", userHandler.GetProfile)

	reviewHandler := handler.NewReviewHandler(db, ticketCfg, reviewSvc, ticketSvc)
	api.GET("/novels/:id/reviews", reviewHandler.List)

	rewardHandler := handler.NewRewardHandler(db, ticketCfg, ticketSvc)

	readingHandler := handler.NewReadingHandler(db)

	protected := api.Group("")
	protected.Use(middleware.AuthRequired(jwtSecret, db))
	{
		voteHandler := handler.NewVoteHandler(db)
		protected.POST("/votes", voteHandler.Create)

		requestHandler := handler.NewRequestHandler(db)
		protected.POST("/requests", requestHandler.Create)
		protected.GET("/requests", requestHandler.List)

		libraryHandler := handler.NewLibraryHandler(db)
		protected.GET("/library", libraryHandler.Get)

		protected.POST("/novels/:id/reviews", reviewHandler.Create)
		protected.PUT("/novels/:id/reviews/:reviewId", reviewHandler.Update)
		protected.POST("/novels/:id/chapters/:num/read", readingHandler.TrackRead)
		protected.POST("/novels/:id/chapters/:num/xp", readingHandler.ClaimXP)
		protected.GET("/novels/:id/my-progress", readingHandler.Progress)

		shareHandler := handler.NewShareHandler(db)
		protected.POST("/novels/:id/share", shareHandler.Create)

		followHandler := handler.NewFollowHandler(db)
		protected.POST("/novels/:id/follow", followHandler.Create)
		protected.DELETE("/novels/:id/follow", followHandler.Delete)
		protected.GET("/novels/:id/follow", followHandler.Check)

		aiTranslateHandler := handler.NewAITranslateHandler(db)
		protected.GET("/user/ai-settings", aiTranslateHandler.GetSettings)
		protected.PUT("/user/ai-settings", aiTranslateHandler.UpdateSettings)
		protected.POST("/translate/ai", aiTranslateHandler.Translate)

		protected.POST("/rewards/daily", rewardHandler.ClaimDaily)
		protected.GET("/rewards/status", rewardHandler.Status)

		protected.PUT("/auth/password", authHandler.UpdatePassword)

		notificationHandler := handler.NewNotificationHandler(db)
		protected.GET("/notifications/unread-count", notificationHandler.UnreadCount)
		protected.GET("/notifications", notificationHandler.List)
		protected.PUT("/notifications/:id/read", notificationHandler.MarkRead)

		purchaseHandler := handler.NewPurchaseHandler(db)
		protected.POST("/tickets/purchase", purchaseHandler.Create)
	}

	adminChapterHandler := handler.NewAdminChapterHandler(db)

	writerGroup := protected.Group("")
	writerGroup.Use(middleware.RequireRole("writer", "admin"))
	{
		writerGroup.POST("/novels", novelHandler.Create)
		writerGroup.PUT("/novels/:id", novelHandler.Update)
		writerGroup.DELETE("/novels/:id", novelHandler.Delete)

		writerGroup.POST("/admin/novels/:id/chapters", adminChapterHandler.Create)
		writerGroup.PUT("/admin/novels/:id/chapters/:chapterID", adminChapterHandler.Update)
		writerGroup.GET("/admin/novels/:id/chapters", adminChapterHandler.List)
		writerGroup.GET("/admin/chapters/:id", adminChapterHandler.Get)

		mdImportHandler := handler.NewMdImportHandler(db)
		writerGroup.POST("/admin/novels/:id/chapters/import-md", mdImportHandler.Import)
	}

	adminGroup := protected.Group("")
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		requestHandler := handler.NewRequestHandler(db)
		adminGroup.PUT("/requests/:id", requestHandler.Review)

		importerHandler := handler.NewImporterHandler(db)
		adminGroup.POST("/novels/import", importerHandler.Import)

		scraperHandler := handler.NewScraperHandler(db)
		adminGroup.POST("/novels/scrape", scraperHandler.Scrape)
		adminGroup.POST("/novels/scrape/import", scraperHandler.Import)

		lncrawlHandler := handler.NewLncrawlHandler(db)
		adminGroup.POST("/novels/lncrawl", lncrawlHandler.Crawl)

		adminHandler := handler.NewAdminHandler(db)
		adminGroup.GET("/admin/users", adminHandler.ListUsers)
		adminGroup.GET("/admin/users/:id", adminHandler.GetUser)
		adminGroup.PUT("/admin/users/:id", adminHandler.UpdateUser)
		adminGroup.DELETE("/admin/users/:id", adminHandler.DeleteUser)
		adminGroup.POST("/admin/users/admin", adminHandler.CreateAdmin)
		adminGroup.GET("/admin/stats", adminHandler.GetStats)
		adminGroup.GET("/admin/reviews", adminHandler.ListReviews)
		adminGroup.DELETE("/admin/reviews/:id", adminHandler.DeleteReview)
		adminGroup.GET("/admin/requests", adminHandler.ListRequests)
		adminGroup.POST("/admin/news", adminHandler.CreateNews)
		adminGroup.PUT("/admin/news/:id", adminHandler.UpdateNews)
		adminGroup.DELETE("/admin/news/:id", adminHandler.DeleteNews)

		adminGroup.DELETE("/admin/chapters/:id", adminChapterHandler.Delete)
		adminGroup.POST("/admin/rewards/monthly", rewardHandler.DistributeMonthly)

		ticketCfgHandler := handler.NewTicketConfigHandler(ticketCfg)
		adminGroup.GET("/admin/config/tickets", ticketCfgHandler.List)
		adminGroup.PUT("/admin/config/tickets", ticketCfgHandler.Update)
	}

	translateHandler := handler.NewTranslateHandler()
	api.POST("/translate", translateHandler.Translate)

	importSearchLimiter := middleware.NewRateLimiter(30, 1*time.Minute)
	api.GET("/novels/import/search", importSearchLimiter.Middleware(), func(c *gin.Context) {
		impHandler := handler.NewImporterHandler(db)
		impHandler.Search(c)
	})

	return r
}
