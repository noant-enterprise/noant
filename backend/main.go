package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"noant/config"
	"noant/internal/handler"
	"noant/internal/infrastructure"
	"noant/internal/middleware"
	"noant/internal/repository"
	"noant/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
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

	// CORS: allow Vite dev server + any configured origins
	corsOrigins := append(cfg.CORSOrigins, "http://localhost:3000", "http://127.0.0.1:3000")

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
	emailSvc := service.NewResendService(cfg.ResendAPIKey)
	_ = polarSvc

	// Wire email service and broadcaster into the service layer
	services := service.NewServices(cfg, repos, redisClient, logger, emailSvc, broadcastFn)

	// Pass wsHub to handlers
	handlers := handler.NewHandlers(services, logger, wsHub)
	healthHandler := handler.NewHealthHandler(db, redisClient, cfg.GroqAPIKeys, logger)

	// Start background health checks
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
	router.GET("/ws", middleware.WebSocketAuth(cfg.JWTSecret), wsHub.HandleWebSocket)

	api := router.Group("/api/v1")
	{
		// Auth endpoints: strict rate limiting (10 req/min per IP)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(redisClient, 10, time.Minute))
		{
			auth.POST("/register", handlers.Auth.Register)
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/refresh", handlers.Auth.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(cfg.JWTSecret), handlers.Auth.Logout)
			auth.POST("/change-password", middleware.AuthMiddleware(cfg.JWTSecret), handlers.Auth.ChangePassword)
			auth.POST("/forgot-password", handlers.Auth.ForgotPassword)
			auth.POST("/reset-password", handlers.Auth.ResetPassword)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), handlers.Auth.Me)
		}

		chats := api.Group("/chats")
		chats.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		chats.Use(middleware.RateLimitByUserMiddleware(redisClient, 500, time.Minute))
		chats.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			chats.POST("/direct-chat", handlers.Chat.DirectChat)
			chats.GET("/conversations", handlers.Chat.ListConversations)
			chats.GET("/conversations/:id", handlers.Chat.GetConversation)
			chats.POST("/conversations/:id/messages", handlers.Chat.SendMessage)
			chats.PUT("/conversations/:id/takeover", handlers.Chat.HumanTakeover)
			chats.POST("/conversations/:id/escalate", handlers.Chat.Escalate)
		}

		training := api.Group("/training")
		training.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		training.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
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
			training.POST("/csv-upload", handlers.Training.UploadCSV)
		}

		analytics := api.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		analytics.Use(middleware.RateLimitByUserMiddleware(redisClient, 60, time.Minute))
		analytics.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			analytics.GET("/overview", handlers.Analytics.Overview)
			analytics.GET("/channels", handlers.Analytics.ChannelDistribution)
			analytics.GET("/insights", handlers.Analytics.Insights)
			analytics.GET("/trends", handlers.Analytics.Trends)
		}

		integrations := api.Group("/integrations")
		integrations.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		integrations.Use(middleware.RateLimitByUserMiddleware(redisClient, 30, time.Minute))
		integrations.Use(middleware.AuditMiddleware(auditRepo, logger))
		{
			integrations.GET("/list", handlers.Integration.List)
			integrations.POST("/connect", handlers.Integration.Connect)
			integrations.POST("/disconnect/:channel", handlers.Integration.Disconnect)
			integrations.POST("/test/:channel", handlers.Integration.Test)
		}

		settings := api.Group("/settings")
		settings.Use(middleware.AuthMiddleware(cfg.JWTSecret))
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
		notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			notifications.GET("", handlers.Notification.List)
			notifications.GET("/unread-count", handlers.Notification.UnreadCount)
			notifications.POST("/:id/read", handlers.Notification.MarkRead)
			notifications.POST("/read-all", handlers.Notification.MarkAllRead)
		}

		widget := api.Group("/widget")
		{
			configGroup := widget.Group("/config")
			configGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				configGroup.GET("", handlers.Widget.Get)
				configGroup.POST("", handlers.Widget.Upsert)
			}
			widget.GET("/public/config", handlers.Widget.GetPublic)
			widget.POST("/public/chat", handlers.Widget.PublicChat)
		}

		archive := api.Group("/archive")
		archive.Use(middleware.AuthMiddleware(cfg.JWTSecret))
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
			payments.POST("/subscribe", middleware.AuthMiddleware(cfg.JWTSecret), middleware.AuditMiddleware(auditRepo, logger), handlers.Payment.Subscribe)
			payments.POST("/webhook", handlers.Payment.Webhook)
			payments.GET("/status", middleware.AuthMiddleware(cfg.JWTSecret), handlers.Payment.Status)
		}
	}

	// API-only mode — frontend served by Vite dev server on :3000

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