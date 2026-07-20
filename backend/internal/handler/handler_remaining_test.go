package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"noant/config"
	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/service"

	"github.com/gin-gonic/gin"
)

func setupHandoffHandler(t *testing.T) (*HandoffHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	creditSvc := service.NewCreditService(cfg, &repository.Repositories{Credit: repos.Credit, User: repos.User}, nil, logger)
	planSvc := service.NewPlanService(cfg, &repository.Repositories{Inventory: repos.Inventory, User: repos.User}, nil, logger, creditSvc)
	svc := service.NewHandoffService(cfg, &repository.Repositories{
		Handoff:      repos.Handoff,
		User:         repos.User,
		Notification: repos.Notification,
		Inventory:    repos.Inventory,
	}, nil, logger, nil, planSvc)
	return NewHandoffHandler(svc, logger), repos
}

func setupInventoryHandler(t *testing.T) (*InventoryHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	svc := service.NewInventoryService(cfg, &repository.Repositories{
		Inventory: repos.Inventory,
	}, nil, logger, nil)
	return NewInventoryHandler(svc, logger), repos
}

func setupCampaignHandler(t *testing.T) (*CampaignHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	creditSvc := service.NewCreditService(cfg, &repository.Repositories{Credit: repos.Credit, User: repos.User}, nil, logger)
	svc := service.NewCampaignService(cfg, &repository.Repositories{
		Campaign: repos.Campaign,
	}, nil, logger, creditSvc)
	return NewCampaignHandler(svc, logger), repos
}

func setupAuditHandler(t *testing.T) (*AuditHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	logger := infrastructure.NewNullLogger()
	svc := service.NewAuditService(&repository.Repositories{
		Audit: repos.Audit,
	}, logger)
	return NewAuditHandler(svc, logger), repos
}

func setupSettingsHandler(t *testing.T) (*SettingsHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	svc := service.NewSettingsService(cfg, &repository.Repositories{
		User:   repos.User,
		APIKey: repos.APIKey,
		Team:   repos.Team,
	}, nil, logger, nil)
	return NewSettingsHandler(svc, logger), repos
}

func setupCreditHandler(t *testing.T) (*CreditHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	creditSvc := service.NewCreditService(cfg, &repository.Repositories{
		Credit: repos.Credit,
		User:   repos.User,
	}, nil, logger)
	planSvc := service.NewPlanService(cfg, &repository.Repositories{
		Inventory: repos.Inventory,
		User:      repos.User,
	}, nil, logger, creditSvc)
	return NewCreditHandler(creditSvc, planSvc, logger), repos
}

// ==================== HANDOFF TESTS ====================

func TestHandoffListEndpoint(t *testing.T) {
	h, _ := setupHandoffHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/handoffs", h.List)

	req := httptest.NewRequest("GET", "/api/handoffs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "handoffs") {
		t.Errorf("expected handoffs in response: %s", w.Body.String())
	}
}

func TestHandoffListWithStatusFilter(t *testing.T) {
	h, _ := setupHandoffHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/handoffs", h.List)

	req := httptest.NewRequest("GET", "/api/handoffs?status=pending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandoffUpdateStatusMissingBody(t *testing.T) {
	h, _ := setupHandoffHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.PUT("/api/handoffs/status", h.UpdateStatus)

	req := httptest.NewRequest("PUT", "/api/handoffs/status", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ==================== INVENTORY TESTS ====================

func TestInventoryListEndpoint(t *testing.T) {
	h, _ := setupInventoryHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/inventory", h.List)

	req := httptest.NewRequest("GET", "/api/inventory", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "items") {
		t.Errorf("expected items in response: %s", w.Body.String())
	}
}

func TestInventoryCreateEndpoint(t *testing.T) {
	h, _ := setupInventoryHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/inventory", h.Create)

	body := `{"type":"product","name":"Widget","price":1000,"stock_quantity":50}`
	req := httptest.NewRequest("POST", "/api/inventory", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInventoryCreateMissingFields(t *testing.T) {
	h, _ := setupInventoryHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/inventory", h.Create)

	req := httptest.NewRequest("POST", "/api/inventory", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInventorySearchEndpoint(t *testing.T) {
	h, _ := setupInventoryHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/inventory/search", h.Search)

	req := httptest.NewRequest("GET", "/api/inventory/search?q=widget", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInventoryGetByIDNotFound(t *testing.T) {
	h, _ := setupInventoryHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/inventory/:id", h.GetByID)

	req := httptest.NewRequest("GET", "/api/inventory/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// ==================== CAMPAIGN TESTS ====================

func TestCampaignListEndpoint(t *testing.T) {
	h, _ := setupCampaignHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/campaigns", h.List)

	req := httptest.NewRequest("GET", "/api/campaigns", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "campaigns") {
		t.Errorf("expected campaigns in response: %s", w.Body.String())
	}
}

func TestCampaignCreateInvalidJSON(t *testing.T) {
	h, _ := setupCampaignHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/campaigns", h.Create)

	req := httptest.NewRequest("POST", "/api/campaigns", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCampaignCancelEmptyID(t *testing.T) {
	h, _ := setupCampaignHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.DELETE("/api/campaigns/:id", h.Cancel)

	req := httptest.NewRequest("DELETE", "/api/campaigns/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for empty path, got %d: %s", w.Code, w.Body.String())
	}
}

// ==================== AUDIT TESTS ====================

func TestAuditListEndpoint(t *testing.T) {
	h, _ := setupAuditHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/audit", h.ListLogs)

	req := httptest.NewRequest("GET", "/api/audit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "audit_logs") {
		t.Errorf("expected audit_logs in response: %s", w.Body.String())
	}
}

// ==================== SETTINGS TESTS ====================

func TestSettingsGetProfileEndpoint(t *testing.T) {
	h, mock := setupSettingsHandler(t)

	seedUser(t, mock.User, "profile@example.com", "StrongP@ss1", true)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-profile@example.com")
		c.Next()
	})
	r.GET("/api/settings/profile", h.GetProfile)

	req := httptest.NewRequest("GET", "/api/settings/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsListAPIKeysEndpoint(t *testing.T) {
	h, _ := setupSettingsHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/settings/api-keys", h.ListAPIKeys)

	req := httptest.NewRequest("GET", "/api/settings/api-keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "api_keys") {
		t.Errorf("expected api_keys in response: %s", w.Body.String())
	}
}

func TestSettingsCreateAPIKeyEndpoint(t *testing.T) {
	h, _ := setupSettingsHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/settings/api-keys", h.CreateAPIKey)

	body := `{"name":"Test Key"}`
	req := httptest.NewRequest("POST", "/api/settings/api-keys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsCreateAPIKeyMissingName(t *testing.T) {
	h, _ := setupSettingsHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/settings/api-keys", h.CreateAPIKey)

	req := httptest.NewRequest("POST", "/api/settings/api-keys", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsListTeamEndpoint(t *testing.T) {
	h, _ := setupSettingsHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/settings/team", h.ListTeam)

	req := httptest.NewRequest("GET", "/api/settings/team", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsInviteTeamMemberInvalidEmail(t *testing.T) {
	h, _ := setupSettingsHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/settings/team/invite", h.InviteTeamMember)

	req := httptest.NewRequest("POST", "/api/settings/team/invite", strings.NewReader(`{"email":"bad","role":"agent"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ==================== CREDIT TESTS ====================

func TestCreditGetBalanceEndpoint(t *testing.T) {
	h, _ := setupCreditHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/credits/balance", h.GetBalance)

	req := httptest.NewRequest("GET", "/api/credits/balance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "balance") {
		t.Errorf("expected balance in response: %s", w.Body.String())
	}
}

func TestCreditGetLimitsEndpoint(t *testing.T) {
	h, mock := setupCreditHandler(t)

	seedUser(t, mock.User, "limits@example.com", "StrongP@ss1", true)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-limits@example.com")
		c.Next()
	})
	r.GET("/api/credits/limits", h.GetLimits)

	req := httptest.NewRequest("GET", "/api/credits/limits", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreditPurchasePackMissingBody(t *testing.T) {
	h, _ := setupCreditHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/credits/purchase", h.PurchasePack)

	req := httptest.NewRequest("POST", "/api/credits/purchase", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreditGetHistoryEndpoint(t *testing.T) {
	h, _ := setupCreditHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/credits/history", h.GetHistory)

	req := httptest.NewRequest("GET", "/api/credits/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "history") {
		t.Errorf("expected history in response: %s", w.Body.String())
	}
}
