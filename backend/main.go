package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"noant/config"
	"noant/internal/handler"
	"noant/internal/infrastructure"
	"noant/internal/middleware"
	"noant/internal/repository"
	"noant/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	logger := infrastructure.NewLogger(cfg.LogLevel)
	logger.Info("Starting Noant Enterprise Platform v2.0")

	db, err := infrastructure.NewTiDBConnection(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to TiDB", "error", err)
	}
	defer db.Close()

	// Apply database migrations
	if err := infrastructure.RunMigrations(db, "./migrations"); err != nil {
		logger.Fatal("Failed to apply migrations", "error", err)
	}
	logger.Info("Database migrations applied successfully")

	// Repair inventory-related schema directly for older databases or skipped migrations.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS inventory_items (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		type ENUM('product','service','package') DEFAULT 'product',
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(15,2) NOT NULL,
		min_price DECIMAL(15,2),
		stock_quantity INT,
		image_url VARCHAR(500),
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_active (user_id, is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	if err != nil {
		logger.Fatal("Failed to ensure inventory_items table", "error", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS handoffs (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		conversation_id VARCHAR(36) NOT NULL,
		customer_name VARCHAR(100),
		customer_phone VARCHAR(50),
		customer_whatsapp VARCHAR(50),
		customer_location TEXT,
		product_name VARCHAR(255),
		original_price DECIMAL(15,2),
		agreed_price DECIMAL(15,2),
		quantity INT DEFAULT 1,
		status ENUM('pending','sold','lost','expired') DEFAULT 'pending',
		final_price DECIMAL(15,2),
		owner_notes TEXT,
		owner_notified_at TIMESTAMP,
		reminder_count INT DEFAULT 0,
		next_reminder_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_status (user_id, status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	if err != nil {
		logger.Fatal("Failed to ensure handoffs table", "error", err)
	}

	_, err = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS owner_whatsapp VARCHAR(50)`)
	if err != nil {
		logger.Warn("Failed to ensure owner_whatsapp column", "error", err)
	}

	_, err = db.Exec(`ALTER TABLE conversations ADD COLUMN IF NOT EXISTS customer_avatar VARCHAR(500)`)
	if err != nil {
		logger.Warn("Failed to ensure customer_avatar column", "error", err)
	}

	// Ensure audit_logs table exists (direct creation as fallback)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		action VARCHAR(255) NOT NULL,
		resource_type VARCHAR(100),
		resource_id VARCHAR(36),
		details JSON,
		ip_address VARCHAR(45),
		user_agent TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_audit_user (user_id),
		INDEX idx_audit_created (created_at),
		INDEX idx_audit_action (action)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	if err != nil {
		logger.Warn("Failed to create audit_logs table directly", "error", err)
	} else {
		logger.Info("Audit logs table verified")
	}

	// Redis is optional — works offline with in-memory fallbacks
	redisClient, err := infrastructure.NewRedisClient(cfg)
	if err != nil {
		logger.Warn("Redis unavailable — running in offline mode", "error", err)
	} else {
		logger.Info("Redis connected")
		defer redisClient.Close()
	}

	repos := repository.NewRepositories(db, redisClient)

	auditRepo := repository.NewAuditRepository(db, redisClient)

	// CORS: allow common local dev origins + any configured origins
	corsOrigins := append(cfg.CORSOrigins,
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	)

	// Create WebSocket hub with origin enforcement
	wsHub := handler.NewWebSocketHub(logger, corsOrigins)
	go wsHub.Run()

	broadcastFn := func(convID string, msgType string, data interface{}) {
		if wsHub != nil {
			wsHub.BroadcastMessage(handler.WebSocketMessage{
				ConversationID: convID,
				Type:           msgType,
				Data:           data,
			})
		}
	}

	polarSvc := service.NewPolarService(cfg)
	emailSvc := service.NewEmailService(cfg, logger)

	// Wire Polar payment service, email service, and broadcaster into the service layer
	services := service.NewServices(cfg, repos, redisClient, logger, emailSvc, polarSvc, broadcastFn)
	if err := services.Integration.SyncTelegramWebhooks(context.Background()); err != nil {
		logger.Warn("Failed to sync Telegram webhooks", "error", err)
	}

	// Initialize layers: Cache, Bottleneck, JobQueue
	cacheStore := infrastructure.NewCache(cfg, redisClient)
	bottleneck := infrastructure.NewBottleneck(
		infrastructure.WithMaxConcurrent(200),
		infrastructure.WithMaxQueue(1000),
	)
	jobQueue := infrastructure.NewJobQueue(logger, redisClient, 10)

	// Register background job handlers
	jobQueue.RegisterHandler("health_check", infrastructure.HealthCheckHandler(services.Integration))
	jobQueue.RegisterHandler("cache_cleanup", infrastructure.CacheCleanupHandler(cacheStore))
	jobQueue.RegisterHandler("handoff_reminder", infrastructure.HandoffReminderHandler(services.Handoff))
	jobQueue.RegisterHandler("check_credit_expiry", infrastructure.CreditExpiryHandler(services.Credit))
	jobQueue.RegisterHandler("process_campaigns_start", infrastructure.CampaignStartHandler(services.Campaign))
	jobQueue.RegisterHandler("process_campaigns_end", infrastructure.CampaignEndHandler(services.Campaign))
	jobQueue.RegisterHandler("free_weekly_reset", infrastructure.FreeWeeklyResetHandler(services.Plan))
	dbCleanupCfg := service.DefaultCleanupConfig()
	jobQueue.RegisterHandler("db_cleanup_all", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.RunAllCleanups(ctx, dbCleanupCfg)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_old_conversations", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupOldResolvedConversations(ctx, dbCleanupCfg.OldConversationsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_abandoned", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupAbandonedConversations(ctx, dbCleanupCfg.AbandonedConversationsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_orphaned_msgs", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupOrphanedMessages(ctx)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_unknown_questions", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupStaleUnknownQuestions(ctx, dbCleanupCfg.UnknownQuestionsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_handoffs", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupExpiredHandoffs(ctx, dbCleanupCfg.HandoffsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_audit_logs", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupOldAuditLogs(ctx, dbCleanupCfg.AuditLogsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_notifications", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupOldNotifications(ctx, dbCleanupCfg.NotificationsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_integrations", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupStaleInactiveIntegrations(ctx, dbCleanupCfg.InactiveIntegrationDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_trials", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupExpiredTrials(ctx, dbCleanupCfg.ExpiredTrialDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_credits", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupExpiredCredits(ctx)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_credit_purchases", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupStaleCreditPurchases(ctx, dbCleanupCfg.CreditPurchasesDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_campaigns", func(ctx context.Context, job *infrastructure.Job) error {
		services.DBManager.CleanupCompletedCampaigns(ctx, dbCleanupCfg.CompletedCampaignsDays)
		return nil
	})
	jobQueue.RegisterHandler("openwa_webhook_repair", func(ctx context.Context, job *infrastructure.Job) error {
		integrations, err := repos.Integration.ListByChannel(ctx, "whatsapp")
		if err != nil {
			logger.Error("Webhook repair: failed to list WhatsApp integrations", "error", err)
			return err
		}
		webhookURLs := []string{
			"http://host.docker.internal:8080/api/v1/openwa/webhook",
			"http://172.19.0.1:8080/api/v1/openwa/webhook",
			"http://localhost:8080/api/v1/openwa/webhook",
		}
		for _, integration := range integrations {
			sessionID, ok := integration.Config["session_id"].(string)
			if !ok || sessionID == "" {
				continue
			}
			for _, webhookURL := range webhookURLs {
				if err := services.OpenWA.ConfigureWebhook(sessionID, webhookURL, cfg.OpenWAWebhookSecret); err == nil {
					logger.Info("Webhook repair: configured", "sessionID", sessionID, "url", webhookURL)
					break
				}
			}
		}
		return nil
	})

	// Start recurring background jobs
	jobQueue.ScheduleRecurring("health_check", map[string]interface{}{}, 5*time.Minute)
	jobQueue.ScheduleRecurring("cache_cleanup", map[string]interface{}{}, 15*time.Minute)
	jobQueue.ScheduleRecurring("handoff_reminder", map[string]interface{}{}, 15*time.Minute)
	jobQueue.ScheduleRecurring("check_credit_expiry", map[string]interface{}{}, 24*time.Hour)
	jobQueue.ScheduleRecurring("process_campaigns_start", map[string]interface{}{}, 24*time.Hour)
	jobQueue.ScheduleRecurring("process_campaigns_end", map[string]interface{}{}, 24*time.Hour)
	jobQueue.ScheduleRecurring("free_weekly_reset", map[string]interface{}{}, 7*24*time.Hour)
	jobQueue.ScheduleRecurring("db_cleanup_all", map[string]interface{}{}, 6*time.Hour)
	jobQueue.ScheduleRecurring("db_cleanup_expired_credits", map[string]interface{}{}, 1*time.Hour)
	jobQueue.ScheduleRecurring("db_cleanup_orphaned_msgs", map[string]interface{}{}, 1*time.Hour)
	jobQueue.ScheduleRecurring("db_cleanup_expired_handoffs", map[string]interface{}{}, 30*time.Minute)
	jobQueue.ScheduleRecurring("openwa_webhook_repair", map[string]interface{}{}, 30*time.Minute)

	// Pass wsHub to handlers
	handlers := handler.NewHandlers(cfg, services, logger, wsHub)
	healthHandler := handler.NewHealthHandler(db, redisClient, cfg.GroqAPIKeys, logger)
	_ = cacheStore
	_ = bottleneck
	_ = jobQueue

	if cfg.NodeEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.MaxMultipartMemory = 8 << 20
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.SanitizeMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", healthHandler.Check)

	// WebSocket endpoint — secured with JWT auth and origin validation
	router.GET("/ws", middleware.WebSocketAuth(cfg.JWTSecret, redisClient), wsHub.HandleWebSocket)

	api := router.Group("/api/v1")
	{
		// Auth mutation endpoints: strict rate limiting (10 req/min per IP)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(redisClient, 10, time.Minute))
		{
			auth.POST("/register", handlers.Auth.Register)
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/logout", handlers.Auth.Logout)
			auth.POST("/change-password", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.Auth.ChangePassword)
			auth.POST("/forgot-password", handlers.Auth.ForgotPassword)
			auth.POST("/reset-password", handlers.Auth.ResetPassword)
		}

		// Session check endpoints: relaxed rate limiting (120 req/min per IP)
		// These are called automatically on every page load / token expiry.
		authSession := api.Group("/auth")
		authSession.Use(middleware.RateLimitMiddleware(redisClient, 120, time.Minute))
		{
			authSession.POST("/refresh", handlers.Auth.RefreshToken)
			authSession.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.Auth.Me)
		}

		chats := api.Group("/chats")
		chats.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		chats.Use(middleware.RateLimitByUserMiddleware(redisClient, 500, time.Minute))
		chats.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			chats.POST("/direct-chat", handlers.Chat.DirectChat)
			chats.GET("/conversations", handlers.Chat.ListConversations)
			chats.GET("/conversations/:id", handlers.Chat.GetConversation)
			chats.POST("/conversations/:id/messages", handlers.Chat.SendMessage)
			chats.PUT("/conversations/:id/takeover", handlers.Chat.HumanTakeover)
			chats.POST("/conversations/:id/escalate", handlers.Chat.Escalate)
			chats.DELETE("/clear", handlers.Chat.ClearChats)
		}

		training := api.Group("/training")
		training.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		training.Use(middleware.RateLimitByUserMiddleware(redisClient, 300, time.Minute))
		training.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			training.GET("/search", handlers.Training.SearchQAPairs)
			training.GET("/categories", handlers.Training.ListCategories)
			training.POST("/categories", handlers.Training.CreateCategory)
			training.DELETE("/categories/:id", handlers.Training.DeleteCategory)
			training.GET("/categories/:id/qa", handlers.Training.ListQAPairs)
			training.POST("/qa", handlers.Training.CreateQAPair)
			training.PUT("/qa/:id", handlers.Training.UpdateQAPair)
			training.DELETE("/qa/:id", handlers.Training.DeleteQAPair)
			training.POST("/bulk-qa", handlers.Training.BulkImport)
			training.GET("/unknown-questions", handlers.Training.ListUnknownQuestions)
			training.POST("/unknown-questions/:id/train", handlers.Training.TrainUnknown)
			training.POST("/unknown-questions/:id/ignore", handlers.Training.IgnoreUnknown)
			training.DELETE("/unknown-questions/clear", handlers.Training.ClearUnknown)
			training.POST("/csv-upload", handlers.Training.UploadCSV)
		}

		analytics := api.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		analytics.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		analytics.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			analytics.GET("/overview", handlers.Analytics.Overview)
			analytics.GET("/channels", handlers.Analytics.ChannelDistribution)
			analytics.GET("/insights", handlers.Analytics.Insights)
			analytics.GET("/trends", handlers.Analytics.Trends)
		}

		integrations := api.Group("/integrations")
		integrations.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		integrations.Use(middleware.RateLimitByUserMiddleware(redisClient, 300, time.Minute))
		integrations.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			integrations.GET("/list", handlers.Integration.List)
			integrations.POST("/connect", handlers.Integration.Connect)
			integrations.POST("/disconnect/:channel", handlers.Integration.Disconnect)
			integrations.POST("/test/:channel", handlers.Integration.Test)
		}

		settings := api.Group("/settings")
		settings.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		settings.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		settings.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			settings.GET("/profile", handlers.Settings.GetProfile)
			settings.PUT("/profile", handlers.Settings.UpdateProfile)
			settings.GET("/api-keys", handlers.Settings.ListAPIKeys)
			settings.POST("/api-keys", handlers.Settings.CreateAPIKey)
			settings.DELETE("/api-keys/:id", handlers.Settings.RevokeAPIKey)
			settings.GET("/team", handlers.Settings.ListTeam)
			settings.POST("/team/invite", handlers.Settings.InviteTeamMember)
			settings.DELETE("/team/:id", handlers.Settings.RemoveTeamMember)
			settings.GET("/audit-logs", handlers.Audit.ListLogs)
			settings.GET("/notifications", handlers.Settings.GetNotifPrefs)
			settings.PUT("/notifications", handlers.Settings.UpdateNotifPrefs)
			settings.DELETE("/account", handlers.Settings.DeleteAccount)
			settings.GET("/account/export", handlers.Settings.ExportData)
		}

		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		{
			notifications.GET("", handlers.Notification.List)
			notifications.GET("/unread-count", handlers.Notification.UnreadCount)
			notifications.POST("/:id/read", handlers.Notification.MarkRead)
			notifications.POST("/read-all", handlers.Notification.MarkAllRead)
		}

		widget := api.Group("/widget")
		{
			configGroup := widget.Group("/config")
			configGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
			{
				configGroup.GET("", handlers.Widget.Get)
				configGroup.POST("", handlers.Widget.Upsert)
			}
			widget.GET("/public/config", handlers.Widget.GetPublic)
			widget.POST("/public/chat", handlers.Widget.PublicChat)
		}

		archive := api.Group("/archive")
		archive.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		archive.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			archive.GET("/folders", handlers.Archive.ListFolders)
			archive.POST("/folders", handlers.Archive.CreateFolder)
			archive.DELETE("/folders/:id", handlers.Archive.DeleteFolder)
			archive.POST("/move", handlers.Archive.MoveChat)
			archive.POST("/remove", handlers.Archive.RemoveFromArchive)
			archive.GET("/status", handlers.Archive.GetStatus)
		}

		payments := api.Group("/payments")
		{
			payments.GET("/plans", handlers.Payment.ListPlans)
			payments.POST("/subscribe", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), middleware.AuditMiddleware(auditRepo, logger), handlers.Payment.Subscribe)
			payments.POST("/webhook", handlers.Payment.Webhook)
			payments.GET("/status", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.Payment.Status)
		}

		inventory := api.Group("/inventory")
		inventory.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		inventory.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		inventory.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			inventory.GET("", handlers.Inventory.List)
			inventory.POST("", handlers.Inventory.Create)
			inventory.GET("/search", handlers.Inventory.Search)
			inventory.GET("/:id", handlers.Inventory.GetByID)
			inventory.PUT("/:id", handlers.Inventory.Update)
			inventory.DELETE("/:id", handlers.Inventory.Delete)
		}

		handoffs := api.Group("/handoffs")
		handoffs.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		handoffs.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		handoffs.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			handoffs.GET("", handlers.Handoff.List)
			handoffs.GET("/:id", handlers.Handoff.GetByID)
			handoffs.PUT("/status", handlers.Handoff.UpdateStatus)
		}

		openwa := api.Group("/openwa")
		{
			// Webhook endpoint — no auth (verified by HMAC signature)
			openwa.POST("/webhook", handlers.OpenWA.WhatsAppWebhook)
			// Session management — auth required
			openwa.GET("/status", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.OpenWA.GetSessionStatus)
			openwa.POST("/restart", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.OpenWA.RestartSession)
		}

		// Simplified WhatsApp channel endpoints
		telegram := api.Group("/telegram")
		{
			telegram.POST("/webhook", handlers.Telegram.Webhook)
		}

	channels := api.Group("/channels")
	channels.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	{
		channels.POST("/whatsapp/connect", handlers.OpenWA.ConnectWhatsApp)
		channels.GET("/whatsapp/status/:sessionId", handlers.OpenWA.GetWhatsAppStatus)
		channels.POST("/whatsapp/refresh/:sessionId", handlers.OpenWA.RefreshWhatsAppQR)
		channels.POST("/whatsapp/disconnect", handlers.OpenWA.DisconnectWhatsApp)
		channels.POST("/whatsapp/ping", handlers.OpenWA.PhonePing)
		channels.POST("/whatsapp/check", handlers.OpenWA.CheckNumber)
		channels.GET("/whatsapp/health", handlers.OpenWA.HealthCheck)
	}

	// Credit endpoints (30 req/min per user)
	credits := api.Group("/credits")
	credits.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	credits.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	{
		credits.GET("/balance", handlers.Credit.GetBalance)
		credits.GET("/limits", handlers.Credit.GetLimits)
		credits.POST("/purchase", handlers.Credit.PurchasePack)
		credits.GET("/history", handlers.Credit.GetHistory)
	}

	// Campaign endpoints (30 req/min per user)
	campaigns := api.Group("/campaigns")
	campaigns.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	campaigns.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	{
		campaigns.GET("", handlers.Campaign.List)
		campaigns.POST("", handlers.Campaign.Create)
		campaigns.DELETE("/:id", handlers.Campaign.Cancel)
	}

	// DB Manager endpoints (admin/owner only)
	dbManager := api.Group("/db-manager")
	dbManager.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	{
		dbManager.GET("/tasks", handlers.DBManager.ListCleanupTasks)
		dbManager.GET("/config", handlers.DBManager.GetCleanupConfig)
		dbManager.POST("/run-all", handlers.DBManager.RunAllCleanups)
		dbManager.POST("/run", handlers.DBManager.RunCleanupTask)
	}

	// Background Worker endpoints (admin/owner only)
	background := api.Group("/background")
	background.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	{
		background.POST("/submit", handlers.Background.SubmitTask)
		background.GET("/tasks", handlers.Background.ListTasks)
		background.GET("/tasks/:id", handlers.Background.GetTaskStatus)
		background.GET("/stats", handlers.Background.WorkerStats)
	}
	}
	// Serve frontend static files if the static directory exists
	if _, err := os.Stat("./static"); err == nil {
		logger.Info("Serving static frontend files from ./static")
		
		// Serve assets directory directly
		router.Static("/assets", "./static/assets")

		// Serve other root-level static files or fallback to index.html for SPA routing
		fileServer := http.FileServer(http.Dir("./static"))
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			
			// Don't intercept API or WebSocket paths
			if len(path) >= 4 && (path[:4] == "/api" || path[:3] == "/ws") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
				return
			}
			
			// Clean the path to prevent traversal attacks
			cleanPath := filepath.Clean(path)
			if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "/../") || strings.Contains(cleanPath, "\\..\\") {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden path traversal attempt"})
				return
			}
			localPath := filepath.Join("static", cleanPath)
			
			// Check if file exists and is not a directory
			if fileInfo, err := os.Stat(localPath); err == nil && !fileInfo.IsDir() {
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
			
			// Fallback to index.html for React SPA
			c.File("./static/index.html")
		})
	} else {
		logger.Info("Static directory not found, running in API-only mode")
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", "error", err)
		}
	}()

	logger.Info("Server running", "port", cfg.Port, "env", cfg.NodeEnv, "redis", redisClient != nil)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited gracefully")
}

func startHealthChecks(integrationSvc *service.IntegrationService, logger *infrastructure.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	channels := []string{"telegram", "whatsapp", "facebook", "instagram"}

	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			logger.Info("Running periodic channel health checks")
			for _, ch := range channels {
				// Pass nil config — health checks use env-configured credentials
				ok, msg := integrationSvc.Test(ctx, ch, nil)
				if !ok {
					logger.Warn("Channel health check failed", "channel", ch, "details", msg)
				} else {
					logger.Info("Channel health check passed", "channel", ch)
				}
			}
			cancel()
		}
	}()
}
