package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"strconv"

	"noant/config"
	"noant/internal/handler"
	"noant/internal/infrastructure"
	"noant/internal/middleware"
	"noant/internal/repository"
	"noant/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testEnv struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	mysqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.0",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "testpass",
				"MYSQL_DATABASE":      "noant_test",
				"MYSQL_USER":          "noant",
				"MYSQL_PASSWORD":      "testpass",
			},
			WaitingFor: wait.ForLog("ready for connections").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start MySQL container: %v", err)
	}
	t.Cleanup(func() { _ = mysqlContainer.Terminate(ctx) })

	mysqlPort, err := mysqlContainer.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("failed to get MySQL port: %v", err)
	}

	dsn := fmt.Sprintf("noant:testpass@tcp(127.0.0.1:%s)/noant_test?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true", mysqlPort.Port())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open MySQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err := infrastructure.RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}
	t.Cleanup(func() { _ = redisContainer.Terminate(ctx) })

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get Redis port: %v", err)
	}

	redisPortInt, err := strconv.Atoi(redisPort.Port())
	if err != nil {
		t.Fatalf("failed to parse Redis port: %v", err)
	}

	cfg := &config.Config{
		RedisHost:     "127.0.0.1",
		RedisPort:     redisPortInt,
		RedisPassword: "",
	}

	redisClient, err := infrastructure.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })

	return &testEnv{db: db, redis: redisClient}
}

func newTestRouter(t *testing.T, env *testEnv, cfg *config.Config) *gin.Engine {
	t.Helper()
	logger := infrastructure.NewNullLogger()

	repos := repository.NewRepositories(env.db, env.redis)
	auditRepo := repository.NewAuditRepository(env.db, env.redis)

	services := service.NewServices(cfg, repos, env.redis, logger, nil, nil, nil)
	handlers := handler.NewHandlers(cfg, services, repos, auditRepo, logger, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-1")
		c.Set("userEmail", "test@example.com")
		c.Set("userRole", "owner")
		c.Next()
	})

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

		chats := api.Group("/chats")
		chats.POST("/direct-chat", handlers.Chat.DirectChat)
		chats.GET("/conversations", handlers.Chat.ListConversations)

		training := api.Group("/training")
		training.GET("/categories", handlers.Training.ListCategories)
		training.POST("/categories", handlers.Training.CreateCategory)
		training.POST("/qa", handlers.Training.CreateQAPair)
		training.GET("/unknown-questions", handlers.Training.ListUnknownQuestions)

		settings := api.Group("/settings")
		settings.GET("/profile", handlers.Settings.GetProfile)
		settings.GET("/api-keys", handlers.Settings.ListAPIKeys)
		settings.POST("/api-keys", handlers.Settings.CreateAPIKey)

		inventory := api.Group("/inventory")
		inventory.GET("", handlers.Inventory.List)
		inventory.POST("", handlers.Inventory.Create)

		analytics := api.Group("/analytics")
		analytics.GET("/overview", handlers.Analytics.Overview)

		credits := api.Group("/credits")
		credits.GET("/balance", handlers.Credit.GetBalance)

		handoffs := api.Group("/handoffs")
		handoffs.GET("", handlers.Handoff.List)

		campaigns := api.Group("/campaigns")
		campaigns.GET("", handlers.Campaign.List)
		campaigns.POST("", handlers.Campaign.Create)

		_ = auditRepo
	}

	return router
}

func doRequest(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestHealthEndpoint(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/health", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "ok") {
		t.Errorf("expected 'ok' in response, got: %s", w.Body.String())
	}
}

func TestTrainingCategoriesCRUD(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "POST", "/api/v1/training/categories", `{"name":"Shipping","description":"Shipping questions"}`)
	if w.Code != 200 && w.Code != 201 {
		t.Fatalf("CreateCategory: expected 200/201, got %d: %s", w.Code, w.Body.String())
	}

	w = doRequest(t, router, "GET", "/api/v1/training/categories", "")
	if w.Code != 200 {
		t.Fatalf("ListCategories: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Shipping") {
		t.Errorf("expected categories to contain 'Shipping', got: %s", w.Body.String())
	}
}

func TestInventoryCRUD(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "POST", "/api/v1/inventory", `{"name":"Premium Widget","price":2999,"type":"product","description":"Our best widget"}`)
	if w.Code != 200 && w.Code != 201 {
		t.Fatalf("CreateInventory: expected 200/201, got %d: %s", w.Code, w.Body.String())
	}

	w = doRequest(t, router, "GET", "/api/v1/inventory", "")
	if w.Code != 200 {
		t.Fatalf("ListInventory: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsProfile(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/settings/profile", "")
	if w.Code != 200 {
		t.Fatalf("GetProfile: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatsListConversations(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/chats/conversations", "")
	if w.Code != 200 {
		t.Fatalf("ListConversations: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "[") {
		t.Errorf("expected JSON array, got: %s", w.Body.String())
	}
}

func TestCampaignsList(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/campaigns", "")
	if w.Code != 200 {
		t.Fatalf("ListCampaigns: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandoffsList(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/handoffs", "")
	if w.Code != 200 {
		t.Fatalf("ListHandoffs: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAnalyticsOverview(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/analytics/overview", "")
	if w.Code != 200 {
		t.Fatalf("AnalyticsOverview: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreditBalance(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/credits/balance", "")
	if w.Code != 200 {
		t.Fatalf("CreditBalance: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAPICreateAndList(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "POST", "/api/v1/settings/api-keys", `{"name":"test-key"}`)
	if w.Code != 200 && w.Code != 201 {
		t.Fatalf("CreateAPIKey: expected 200/201, got %d: %s", w.Code, w.Body.String())
	}

	w = doRequest(t, router, "GET", "/api/v1/settings/api-keys", "")
	if w.Code != 200 {
		t.Fatalf("ListAPIKeys: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTrainingUnknownQuestions(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "GET", "/api/v1/training/unknown-questions", "")
	if w.Code != 200 {
		t.Fatalf("ListUnknownQuestions: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEndToEndTrainingWorkflow(t *testing.T) {
	env := setupTestEnv(t)
	cfg := &config.Config{}
	router := newTestRouter(t, env, cfg)

	w := doRequest(t, router, "POST", "/api/v1/training/categories", `{"name":"Returns","description":"Return policy questions"}`)
	if w.Code != 200 && w.Code != 201 {
		t.Fatalf("CreateCategory: expected 200/201, got %d: %s", w.Code, w.Body.String())
	}

	w = doRequest(t, router, "POST", "/api/v1/training/qa", `{"category_id":"dummy","question":"How do I return an item?","answer":"You can return within 30 days"}`)
	if w.Code >= 500 {
		t.Fatalf("CreateQAPair: expected non-5xx, got %d: %s", w.Code, w.Body.String())
	}

	w = doRequest(t, router, "GET", "/api/v1/training/categories", "")
	if w.Code != 200 {
		t.Fatalf("ListCategories after workflow: expected 200, got %d", w.Code)
	}
}
