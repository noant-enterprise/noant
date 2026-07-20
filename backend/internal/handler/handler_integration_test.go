package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthHandler(t *testing.T) (*AuthHandler, *repository.MockUserRepo) {
	t.Helper()
	mock := repository.NewMockUserRepo()
	cfg := &config.Config{JWTSecret: "test-secret-123"}
	logger := infrastructure.NewNullLogger()
	svc := service.NewAuthService(cfg, mock, nil, logger, nil)
	return NewAuthHandler(svc, logger), mock
}

func setupTrainingHandler(t *testing.T) (*TrainingHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	reposDB := &repository.Repositories{
		Category: repos.Category,
		QAPair:   repos.QAPair,
		UnknownQ: repos.UnknownQ,
	}
	svc := service.NewTrainingService(cfg, reposDB, nil, logger, nil)
	return NewTrainingHandler(svc, logger), repos
}

func setupChatHandler(t *testing.T) (*ChatHandler, *repository.MockRepositories) {
	t.Helper()
	repos := repository.NewMockRepositories()
	cfg := &config.Config{}
	logger := infrastructure.NewNullLogger()
	reposDB := &repository.Repositories{
		Conversation: repos.Conversation,
		Message:      repos.Message,
		Integration:  repos.Integration,
		Handoff:      repos.Handoff,
	}
	svc := service.NewChatService(cfg, reposDB, nil, nil, logger, nil, nil)
	return NewChatHandler(svc, logger, nil), repos
}

func withUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", "test-user-1")
		c.Next()
	}
}

func seedUser(t *testing.T, mock *repository.MockUserRepo, email, password string, verified bool) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		t.Fatalf("seedUser hash: %v", err)
	}
	user := &domain.User{
		ID:        "user-" + email,
		Email:     email,
		Password:  string(hash),
		FirstName: "Test",
		LastName:  "User",
		Role:      "owner",
		IsVerified: verified,
		IsActive:   true,
	}
	if err := mock.Create(context.Background(), user); err != nil {
		t.Fatalf("seedUser: %v", err)
	}
}

func TestAuthRegisterEndpoint(t *testing.T) {
	h, _ := setupAuthHandler(t)
	r := gin.New()
	r.POST("/api/auth/register", h.Register)

	body := `{"email":"test@example.com","password":"StrongP@ss1","first_name":"John","last_name":"Doe"}`
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthRegisterInvalidJSON(t *testing.T) {
	h, _ := setupAuthHandler(t)
	r := gin.New()
	r.POST("/api/auth/register", h.Register)

	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthRegisterMissingFields(t *testing.T) {
	h, _ := setupAuthHandler(t)
	r := gin.New()
	r.POST("/api/auth/register", h.Register)

	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthLoginEndpoint(t *testing.T) {
	h, mock := setupAuthHandler(t)
	seedUser(t, mock, "test@example.com", "StrongP@ss1", true)

	r := gin.New()
	r.POST("/api/auth/login", h.Login)

	body := `{"email":"test@example.com","password":"StrongP@ss1"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthLoginWrongPassword(t *testing.T) {
	h, mock := setupAuthHandler(t)
	seedUser(t, mock, "test@example.com", "StrongP@ss1", true)

	r := gin.New()
	r.POST("/api/auth/login", h.Login)

	body := `{"email":"test@example.com","password":"WrongPassword1"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthLoginNotVerified(t *testing.T) {
	h, mock := setupAuthHandler(t)
	seedUser(t, mock, "unverified@example.com", "StrongP@ss1", false)

	r := gin.New()
	r.POST("/api/auth/login", h.Login)

	body := `{"email":"unverified@example.com","password":"StrongP@ss1"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthLoginNonExistentUser(t *testing.T) {
	h, _ := setupAuthHandler(t)

	r := gin.New()
	r.POST("/api/auth/login", h.Login)

	body := `{"email":"nobody@example.com","password":"StrongP@ss1"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTrainingListCategoriesEndpoint(t *testing.T) {
	h, _ := setupTrainingHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/training/categories", h.ListCategories)

	req := httptest.NewRequest("GET", "/api/training/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "categories") {
		t.Errorf("expected 'categories' in response: %s", w.Body.String())
	}
}

func TestTrainingCreateCategoryEndpoint(t *testing.T) {
	h, _ := setupTrainingHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/training/categories", h.CreateCategory)

	body := `{"name":"General","description":"General QA","color":"#FF0000"}`
	req := httptest.NewRequest("POST", "/api/training/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTrainingCreateCategoryMissingName(t *testing.T) {
	h, _ := setupTrainingHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/training/categories", h.CreateCategory)

	req := httptest.NewRequest("POST", "/api/training/categories", strings.NewReader(`{"description":"no name"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatListConversationsEndpoint(t *testing.T) {
	h, _ := setupChatHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.GET("/api/conversations", h.ListConversations)

	req := httptest.NewRequest("GET", "/api/conversations?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatDirectChatMissingChannel(t *testing.T) {
	h, _ := setupChatHandler(t)
	r := gin.New()
	r.Use(withUserID())
	r.POST("/api/chat/direct", h.DirectChat)

	req := httptest.NewRequest("POST", "/api/chat/direct", strings.NewReader(`{"message":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthVerifyEmailEndpoint(t *testing.T) {
	h, mock := setupAuthHandler(t)
	seedUser(t, mock, "verify@example.com", "StrongP@ss1", false)

	r := gin.New()
	r.POST("/api/auth/verify-email", h.VerifyEmail)

	body := `{"email":"verify@example.com","code":"123456"}`
	req := httptest.NewRequest("POST", "/api/auth/verify-email", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Will fail with invalid code but exercises the handler path
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid code, got %d: %s", w.Code, w.Body.String())
	}
}
