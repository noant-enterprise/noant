package main

import (
	"context"
	"log/slog"
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

	"github.com/SherClockHolmes/webpush-go"
	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	logger := infrastructure.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger.Logger)
	logger.Info("Starting Noant Enterprise Platform v2.0")

	// Initialize Sentry error monitoring
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 0.2, // 20% of transactions for tracing
			Environment:      cfg.NodeEnv,
			Release:          "noant@2.0.0",
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				// Scrub sensitive fields before sending
				if event.Request != nil {
					delete(event.Request.Headers, "Authorization")
					delete(event.Request.Headers, "X-API-Key")
				}
				return event
			},
		}); err != nil {
			logger.Warn("Sentry initialization failed", "error", err)
		} else {
			defer sentry.Flush(2 * time.Second)
			logger.Info("Sentry error monitoring enabled", "environment", cfg.NodeEnv)
		}
	}

	db, err := infrastructure.NewTiDBConnection(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to TiDB", "error", err)
	}
	defer func() { _ = db.Close() }()

	// Auto-generate VAPID keys if not configured
	if cfg.VAPIDPublicKey == "" || cfg.VAPIDPrivateKey == "" {
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			logger.Fatal("Failed to generate VAPID keys", "error", err)
		}
		cfg.VAPIDPrivateKey = privateKey
		cfg.VAPIDPublicKey = publicKey
		logger.Warn("VAPID keys auto-generated. Set these in .env for persistence:",
			"VAPID_PUBLIC_KEY", publicKey,
			"VAPID_PRIVATE_KEY", privateKey,
		)
	}

	// Apply database migrations
	if err := infrastructure.RunMigrations(db, "./migrations"); err != nil {
		logger.Fatal("Failed to apply migrations", "error", err)
	}
	logger.Info("Database migrations applied successfully")

	// Redis is optional — works offline with in-memory fallbacks
	redisClient, err := infrastructure.NewRedisClient(cfg)
	if err != nil {
		logger.Warn("Redis unavailable — running in offline mode", "error", err)
	} else {
		logger.Info("Redis connected")
		defer func() { _ = redisClient.Close() }()
	}

	repos := repository.NewRepositories(db, redisClient)

	auditRepo := repository.NewAuditRepository(db, redisClient)

	// CORS: allow common local dev origins + any configured origins
	corsOrigins := cfg.CORSOrigins
	if cfg.NodeEnv != "production" {
		corsOrigins = append(corsOrigins,
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
			"http://localhost:3002",
			"http://127.0.0.1:3002",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		)
	}

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

	// Inject Redis into OpenWA subsystems (Redis may be nil)
	services.OpenWA.InjectDependencies(redisClient)

	// Start the session health monitor for OpenWA
	services.OpenWA.StartSessionManager()

	// Initialize layers: Cache, JobQueue
	cacheStore := infrastructure.NewCache(cfg, redisClient)
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
		_, _ = services.DBManager.CleanupOldResolvedConversations(ctx, dbCleanupCfg.OldConversationsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_abandoned", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupAbandonedConversations(ctx, dbCleanupCfg.AbandonedConversationsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_orphaned_msgs", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupOrphanedMessages(ctx)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_unknown_questions", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupStaleUnknownQuestions(ctx, dbCleanupCfg.UnknownQuestionsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_handoffs", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupExpiredHandoffs(ctx, dbCleanupCfg.HandoffsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_audit_logs", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupOldAuditLogs(ctx, dbCleanupCfg.AuditLogsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_notifications", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupOldNotifications(ctx, dbCleanupCfg.NotificationsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_integrations", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupStaleInactiveIntegrations(ctx, dbCleanupCfg.InactiveIntegrationDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_trials", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupExpiredTrials(ctx, dbCleanupCfg.ExpiredTrialDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_credits", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupExpiredCredits(ctx)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_credit_purchases", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupStaleCreditPurchases(ctx, dbCleanupCfg.CreditPurchasesDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_campaigns", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupCompletedCampaigns(ctx, dbCleanupCfg.CompletedCampaignsDays)
		return nil
	})
	jobQueue.RegisterHandler("db_cleanup_expired_media", func(ctx context.Context, job *infrastructure.Job) error {
		_, _ = services.DBManager.CleanupExpiredMediaMessages(ctx)
		return nil
	})
	jobQueue.RegisterHandler("openwa_media_cleanup", func(ctx context.Context, job *infrastructure.Job) error {
		if mh := services.OpenWA.GetMediaHandler(); mh != nil {
			removed, err := mh.CleanupExpiredMedia()
			if err != nil {
				logger.Error("OpenWA media cleanup failed", "error", err)
				return err
			}
			logger.Info("OpenWA media cleanup completed", "files_removed", removed)
		}
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
		for i := range integrations {
			sessionID, ok := integrations[i].Config["session_id"].(string)
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
	jobQueue.ScheduleRecurring("openwa_media_cleanup", map[string]interface{}{}, 1*time.Hour)
	jobQueue.ScheduleRecurring("db_cleanup_expired_media", map[string]interface{}{}, 30*time.Minute)

	// Pass wsHub to handlers
	handlers := handler.NewHandlers(cfg, services, repos, auditRepo, logger, wsHub)
	healthHandler := handler.NewHealthHandler(db, redisClient, cfg.GroqAPIKeys, logger)

	startHealthChecks(services.Integration, logger)

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
	router.Use(middleware.BodyLimitMiddleware())

	router.Use(middleware.CSRFMiddleware(corsOrigins))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(infrastructure.PrometheusMiddleware())
	if cfg.SentryDSN != "" {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
		router.Use(middleware.SentryContextMiddleware())
	}

	router.GET("/metrics", middleware.RequireAdminMiddleware(), gin.WrapH(promhttp.Handler()))
	router.GET("/health", healthHandler.Check)
	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/readyz", healthHandler.Check)

	// WebSocket endpoint — secured with JWT auth and origin validation
	router.GET("/ws", middleware.WebSocketAuth(cfg.JWTSecret, redisClient), wsHub.HandleWebSocket)

	api := router.Group("/api/v1")
	{
		// Auth mutation endpoints: strict rate limiting (10 req/min per IP)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(redisClient, 10, time.Minute))
		auth.POST("/register", middleware.ValidateRegister(), handlers.Auth.Register)
		auth.POST("/login", middleware.ValidateLogin(), handlers.Auth.Login)
		auth.POST("/logout", handlers.Auth.Logout)
		auth.POST("/change-password", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.Auth.ChangePassword)
		auth.POST("/forgot-password", handlers.Auth.ForgotPassword)
		auth.POST("/reset-password", handlers.Auth.ResetPassword)
		auth.POST("/verify", handlers.Auth.VerifyEmail)
		auth.POST("/resend-verification", handlers.Auth.ResendVerification)
		if cfg.NodeEnv != "production" {
			auth.POST("/dev/verify", handlers.Auth.DevVerify)
		}

		// Session check endpoints: relaxed rate limiting (120 req/min per IP)
		// These are called automatically on every page load / token expiry.
		authSession := api.Group("/auth")
		authSession.Use(middleware.RateLimitMiddleware(redisClient, 120, time.Minute))
		authSession.POST("/refresh", handlers.Auth.RefreshToken)
		authSession.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.Auth.Me)

		chats := api.Group("/chats")
		chats.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		chats.Use(middleware.RateLimitByUserMiddleware(redisClient, 500, time.Minute))
		chats.Use(middleware.AuditMiddleware(auditRepo, logger))
		chats.POST("/direct-chat", middleware.ValidateDirectChat(), handlers.Chat.DirectChat)
		chats.GET("/conversations", handlers.Chat.ListConversations)
		chats.GET("/conversations/:id", handlers.Chat.GetConversation)
		chats.POST("/conversations/:id/messages", middleware.ValidateSendMessage(), handlers.Chat.SendMessage)
		chats.POST("/conversations/:id/stream", handlers.Chat.StreamMessage)
		chats.PUT("/conversations/:id/takeover", handlers.Chat.HumanTakeover)
		chats.POST("/conversations/:id/escalate", handlers.Chat.Escalate)
		chats.POST("/conversations/:id/rate", handlers.Chat.RateConversation)
		chats.DELETE("/clear", handlers.Chat.ClearChats)

		training := api.Group("/training")
		training.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		training.Use(middleware.RateLimitByUserMiddleware(redisClient, 300, time.Minute))
		training.Use(middleware.AuditMiddleware(auditRepo, logger))
		training.GET("/search", handlers.Training.SearchQAPairs)
		training.GET("/categories", handlers.Training.ListCategories)
		training.POST("/categories", handlers.Training.CreateCategory)
		training.DELETE("/categories/:id", handlers.Training.DeleteCategory)
		training.GET("/categories/:id/qa", handlers.Training.ListQAPairs)
		training.POST("/qa", middleware.ValidateCreateQAPair(), handlers.Training.CreateQAPair)
		training.PUT("/qa/:id", handlers.Training.UpdateQAPair)
		training.DELETE("/qa/:id", handlers.Training.DeleteQAPair)
		training.POST("/bulk-qa", handlers.Training.BulkImport)
		training.GET("/unknown-questions", handlers.Training.ListUnknownQuestions)
		training.POST("/unknown-questions/:id/train", handlers.Training.TrainUnknown)
		training.POST("/unknown-questions/:id/ignore", handlers.Training.IgnoreUnknown)
		training.POST("/unknown-questions/batch-train", handlers.Training.BatchTrainUnknown)
		training.POST("/unknown-questions/batch-ignore", handlers.Training.BatchIgnoreUnknown)
		training.DELETE("/unknown-questions/clear", handlers.Training.ClearUnknown)
		training.POST("/csv-upload", handlers.Training.UploadCSV)

		analytics := api.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		analytics.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		analytics.Use(middleware.AuditMiddleware(auditRepo, logger))
		analytics.GET("/overview", handlers.Analytics.Overview)
		analytics.GET("/channels", handlers.Analytics.ChannelDistribution)
		analytics.GET("/insights", handlers.Analytics.Insights)
		analytics.GET("/trends", handlers.Analytics.Trends)
		analytics.GET("/satisfaction", handlers.Analytics.Satisfaction)
		analytics.GET("/unknown-questions", handlers.Analytics.UnknownQuestions)
		analytics.GET("/popular-questions", handlers.Analytics.PopularQuestions)
		analytics.GET("/messages-trend", handlers.Analytics.MessagesTrend)
		analytics.GET("/uptime", handlers.Analytics.Uptime)

		integrations := api.Group("/integrations")
		integrations.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		integrations.Use(middleware.RateLimitByUserMiddleware(redisClient, 300, time.Minute))
		integrations.Use(middleware.AuditMiddleware(auditRepo, logger))
		integrations.GET("/list", handlers.Integration.List)
		integrations.POST("/connect", handlers.Integration.Connect)
		integrations.POST("/disconnect/:channel", handlers.Integration.Disconnect)
		integrations.POST("/test/:channel", handlers.Integration.Test)

		settings := api.Group("/settings")
		settings.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		settings.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		settings.Use(middleware.AuditMiddleware(auditRepo, logger))
		settings.GET("/profile", handlers.Settings.GetProfile)
		settings.PUT("/profile", handlers.Settings.UpdateProfile)
		settings.GET("/api-keys", handlers.Settings.ListAPIKeys)
		settings.POST("/api-keys", handlers.Settings.CreateAPIKey)
		settings.DELETE("/api-keys/:id", handlers.Settings.RevokeAPIKey)
		settings.GET("/team", handlers.Settings.ListTeam)
		settings.POST("/team/invite", handlers.Settings.InviteTeamMember)
		settings.DELETE("/team/:id", handlers.Settings.RemoveTeamMember)
		settings.GET("/audit-logs", handlers.Audit.ListLogs)
		settings.GET("/audit-logs/search", handlers.Audit.SearchLogs)
		settings.GET("/notifications", handlers.Settings.GetNotifPrefs)
		settings.PUT("/notifications", handlers.Settings.UpdateNotifPrefs)
		settings.DELETE("/account", handlers.Settings.DeleteAccount)
		settings.GET("/account/export", handlers.Settings.ExportData)

		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		notifications.GET("", handlers.Notification.List)
		notifications.GET("/unread-count", handlers.Notification.UnreadCount)
		notifications.POST("/:id/read", handlers.Notification.MarkRead)
		notifications.POST("/read-all", handlers.Notification.MarkAllRead)

		widget := api.Group("/widget")
		configGroup := widget.Group("/config")
		configGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		configGroup.GET("", handlers.Widget.Get)
		configGroup.POST("", handlers.Widget.Upsert)
		widget.GET("/public/config", handlers.Widget.GetPublic)
		widget.POST("/public/chat", middleware.RateLimitMiddleware(redisClient, 30, time.Minute), handlers.Widget.PublicChat)

		archive := api.Group("/archive")
		archive.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		archive.Use(middleware.AuditMiddleware(auditRepo, logger))
		archive.GET("/folders", handlers.Archive.ListFolders)
		archive.POST("/folders", handlers.Archive.CreateFolder)
		archive.DELETE("/folders/:id", handlers.Archive.DeleteFolder)
		archive.POST("/move", handlers.Archive.MoveChat)
		archive.POST("/remove", handlers.Archive.RemoveFromArchive)
		archive.GET("/status", handlers.Archive.GetStatus)

		payments := api.Group("/payments")
		payments.GET("/plans", handlers.Payment.ListPlans)
		payments.POST("/subscribe", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), middleware.AuditMiddleware(auditRepo, logger), handlers.Payment.Subscribe)
		payments.POST("/webhook", middleware.RateLimitMiddleware(redisClient, 500, time.Minute), handlers.Payment.Webhook)
		payments.GET("/status", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.Payment.Status)

		inventory := api.Group("/inventory")
		inventory.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		inventory.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		inventory.Use(middleware.AuditMiddleware(auditRepo, logger))
		inventory.GET("", handlers.Inventory.List)
		inventory.POST("", handlers.Inventory.Create)
		inventory.GET("/search", handlers.Inventory.Search)
		inventory.GET("/:id", handlers.Inventory.GetByID)
		inventory.PUT("/:id", handlers.Inventory.Update)
		inventory.DELETE("/:id", handlers.Inventory.Delete)

		handoffs := api.Group("/handoffs")
		handoffs.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
		handoffs.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		handoffs.Use(middleware.AuditMiddleware(auditRepo, logger))
		handoffs.GET("", handlers.Handoff.List)
		handoffs.GET("/:id", handlers.Handoff.GetByID)
		handoffs.PUT("/status", handlers.Handoff.UpdateStatus)

		openwa := api.Group("/openwa")
		// Webhook endpoint — no auth (verified by HMAC signature), rate-limited
		openwa.POST("/webhook", middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), handlers.OpenWA.WhatsAppWebhook)
		// Session management — auth required
		openwa.GET("/status", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.OpenWA.GetSessionStatus)
		openwa.POST("/restart", middleware.AuthMiddleware(cfg.JWTSecret, redisClient), handlers.OpenWA.RestartSession)

		// Simplified WhatsApp channel endpoints
		telegram := api.Group("/telegram")
		telegram.POST("/webhook", middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), handlers.Telegram.Webhook)

	channels := api.Group("/channels")
	channels.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	channels.POST("/whatsapp/connect", handlers.OpenWA.ConnectWhatsApp)
	channels.GET("/whatsapp/status/:sessionId", handlers.OpenWA.GetWhatsAppStatus)
	channels.POST("/whatsapp/refresh/:sessionId", handlers.OpenWA.RefreshWhatsAppQR)
	channels.POST("/whatsapp/disconnect", handlers.OpenWA.DisconnectWhatsApp)
	channels.POST("/whatsapp/ping", handlers.OpenWA.PhonePing)
	channels.POST("/whatsapp/check", handlers.OpenWA.CheckNumber)
	channels.GET("/whatsapp/health", handlers.OpenWA.HealthCheck)
	channels.GET("/whatsapp/sessions/health", handlers.OpenWA.SessionHealthDashboard)

	// Credit endpoints (30 req/min per user)
	credits := api.Group("/credits")
	credits.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	credits.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	credits.GET("/balance", handlers.Credit.GetBalance)
	credits.GET("/limits", handlers.Credit.GetLimits)
	credits.POST("/purchase", handlers.Credit.PurchasePack)
	credits.GET("/history", handlers.Credit.GetHistory)

	// Campaign endpoints (30 req/min per user)
	campaigns := api.Group("/campaigns")
	campaigns.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	campaigns.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	campaigns.GET("", handlers.Campaign.List)
	campaigns.POST("", handlers.Campaign.Create)
	campaigns.DELETE("/:id", handlers.Campaign.Cancel)

	// DB Manager endpoints (admin/owner only)
	dbManager := api.Group("/db-manager")
	dbManager.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	dbManager.Use(middleware.RequireAdminMiddleware())
	dbManager.GET("/tasks", handlers.DBManager.ListCleanupTasks)
	dbManager.GET("/config", handlers.DBManager.GetCleanupConfig)
	dbManager.POST("/run-all", handlers.DBManager.RunAllCleanups)
	dbManager.POST("/run", handlers.DBManager.RunCleanupTask)

	// Background Worker endpoints (admin/owner only)
	background := api.Group("/background")
	background.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	background.Use(middleware.RequireAdminMiddleware())
	background.POST("/submit", handlers.Background.SubmitTask)
	background.GET("/tasks", handlers.Background.ListTasks)
	background.GET("/tasks/:id", handlers.Background.GetTaskStatus)
	background.GET("/stats", handlers.Background.WorkerStats)
	}

	// WhatsApp Template endpoints (30 req/min per user)
	templates := api.Group("/templates")
	templates.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	templates.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	templates.GET("", handlers.Template.List)
	templates.POST("", handlers.Template.Create)
	templates.GET("/:id", handlers.Template.GetByID)
	templates.PUT("/:id", handlers.Template.Update)
	templates.DELETE("/:id", handlers.Template.Delete)
	templates.POST("/:id/submit", handlers.Template.SubmitForApproval)
	templates.POST("/send", handlers.Template.Send)
	templates.GET("/common", handlers.Template.GetCommon)

	// WhatsApp Campaign endpoints (30 req/min per user)
	whatsappCampaign := api.Group("/whatsapp/campaigns")
	whatsappCampaign.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	whatsappCampaign.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	whatsappCampaign.POST("/broadcast", handlers.OpenWA.BroadcastCampaign)
	whatsappCampaign.GET("/:campaignID/analytics", handlers.OpenWA.CampaignAnalytics)

	// WhatsApp Media endpoints
	media := api.Group("/chats/conversations")
	media.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	media.POST("/:id/media", handlers.OpenWA.UploadMedia)
	media.GET("/:id/media", handlers.OpenWA.ListMedia)
	media.GET("/media/:mediaID", handlers.OpenWA.GetMedia)
	media.GET("/media/:mediaID/thumbnail", handlers.OpenWA.GetMediaThumbnail)

	// WhatsApp Queue & Session management endpoints
	whatsappAdmin := api.Group("/whatsapp/admin")
	whatsappAdmin.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	whatsappAdmin.GET("/queue/stats", handlers.OpenWA.QueueStats)
	whatsappAdmin.GET("/sessions", handlers.OpenWA.ListManagedSessions)
	whatsappAdmin.GET("/sessions/:sessionID/metrics", handlers.OpenWA.SessionMetrics)
	whatsappAdmin.POST("/sessions/:sessionID/reconnect", handlers.OpenWA.ForceReconnect)

	// Interactive message endpoints
	interactive := api.Group("/whatsapp/interactive")
	interactive.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	interactive.POST("/list", handlers.OpenWA.SendListMessage)
	interactive.POST("/buttons", handlers.OpenWA.SendButtonsMessage)

	// Onboarding Assistant endpoints
	assistant := api.Group("/assistant")
	assistant.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	assistant.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	assistant.POST("/chat", handlers.Assistant.Chat)

	// Onboarding wizard endpoints
	onboarding := api.Group("/onboarding")
	onboarding.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	onboarding.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	onboarding.GET("/status", handlers.Onboarding.GetStatus)
	onboarding.POST("/step", handlers.Onboarding.CompleteStep)
	onboarding.POST("/categories/auto-create", handlers.Onboarding.AutoCreateCategories)
	onboarding.GET("/industry-templates", handlers.Onboarding.GetIndustryTemplates)

	// Push notification subscription endpoints
	pushSub := api.Group("/push")
	pushSub.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	pushSub.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
	pushSub.POST("/subscribe", handlers.Push.Subscribe)
	pushSub.POST("/unsubscribe", handlers.Push.Unsubscribe)

	// Admin panel endpoints
	adminRoutes := api.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret, redisClient))
	adminRoutes.Use(middleware.RequireAdmin())
	adminRoutes.GET("/overview", handlers.Admin.Overview)
	adminRoutes.GET("/users", handlers.Admin.Users)
	adminRoutes.GET("/users/:id", handlers.Admin.User)
	adminRoutes.GET("/system/health", handlers.Admin.SystemHealth)
	adminRoutes.GET("/analytics", handlers.Admin.Analytics)
	adminRoutes.GET("/revenue", handlers.Admin.Revenue)
	adminRoutes.GET("/ai/health", handlers.Admin.AIHealth)
	adminRoutes.GET("/alerts", handlers.Admin.Alerts)
	adminRoutes.GET("/activity", handlers.Admin.RecentActivity)
	adminRoutes.GET("/audit-logs", handlers.Admin.AuditLogs)
	adminRoutes.GET("/knowledge-base", handlers.Admin.KnowledgeBase)
	adminRoutes.POST("/knowledge-base/train", handlers.Admin.TrainKnowledge)

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
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
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

	// Gracefully stop background systems
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown OpenWA with queue drain (Fix 1: graceful shutdown)
	services.OpenWA.Shutdown(shutdownCtx)
	services.Background.Shutdown()
	jobQueue.Shutdown()

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
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in health check goroutine", "recover", r)
			}
		}()
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
