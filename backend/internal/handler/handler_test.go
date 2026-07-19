package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("requestID", "test-req-id")
		c.Next()
	})
	return r
}

func TestRegister_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing all fields",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "missing password",
			body:       map[string]string{"email": "test@example.com", "first_name": "John", "last_name": "Doe"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "invalid email",
			body:       map[string]string{"email": "not-an-email", "password": "12345678", "first_name": "John", "last_name": "Doe"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "password too short",
			body:       map[string]string{"email": "test@example.com", "password": "123", "first_name": "John", "last_name": "Doe"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "missing first name",
			body:       map[string]string{"email": "test@example.com", "password": "12345678", "last_name": "Doe"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter()
			router.POST("/auth/register", func(c *gin.Context) {
				var req struct {
					Email     string `json:"email" binding:"required,email"`
					Password  string `json:"password" binding:"required,min=8"`
					FirstName string `json:"first_name" binding:"required"`
					LastName  string `json:"last_name" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					utils.RespondValidationError(c, err.Error())
					return
				}
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp utils.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", resp.Code, tt.wantCode)
			}
			if resp.RequestID != "test-req-id" {
				t.Errorf("request_id = %q, want %q", resp.RequestID, "test-req-id")
			}
		})
	}
}

func TestLogin_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "missing email and password",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email format",
			body:       map[string]string{"email": "bad", "password": "pass"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			body:       map[string]string{"email": "a@b.com"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter()
			router.POST("/auth/login", func(c *gin.Context) {
				var req struct {
					Email    string `json:"email" binding:"required,email"`
					Password string `json:"password" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					utils.RespondValidationError(c, err.Error())
					return
				}
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestVerifyEmail_ValidationErrors(t *testing.T) {
	router := setupRouter()
	router.POST("/auth/verify-email", func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
			Code  string `json:"code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing code", map[string]string{"email": "a@b.com"}, http.StatusBadRequest},
		{"invalid email", map[string]string{"email": "bad", "code": "123"}, http.StatusBadRequest},
		{"empty body", map[string]string{}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/verify-email", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestChangePassword_ValidationErrors(t *testing.T) {
	router := setupRouter()
	router.POST("/auth/change-password", func(c *gin.Context) {
		var req struct {
			CurrentPassword string `json:"current_password" binding:"required"`
			NewPassword     string `json:"new_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing both", map[string]string{}, http.StatusBadRequest},
		{"new password too short", map[string]string{"current_password": "old", "new_password": "short"}, http.StatusBadRequest},
		{"missing current", map[string]string{"new_password": "longpassword"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestResetPassword_ValidationErrors(t *testing.T) {
	router := setupRouter()
	router.POST("/auth/reset-password", func(c *gin.Context) {
		var req struct {
			Token       string `json:"token" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing token", map[string]string{"new_password": "longpassword"}, http.StatusBadRequest},
		{"password too short", map[string]string{"token": "abc", "new_password": "short"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestErrorResponseFormat(t *testing.T) {
	router := setupRouter()
	router.GET("/test-error", func(c *gin.Context) {
		utils.RespondNotFound(c, "User")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-error", nil)
	router.ServeHTTP(w, req)

	var resp utils.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.Success != false {
		t.Error("Success should be false")
	}
	if resp.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", resp.Code)
	}
	if resp.Error != "User not found" {
		t.Errorf("Error = %q, want 'User not found'", resp.Error)
	}
	if resp.Retryable != false {
		t.Error("NOT_FOUND should not be retryable")
	}
}

func TestRateLimitResponse(t *testing.T) {
	router := setupRouter()
	router.GET("/test-rate", func(c *gin.Context) {
		utils.RespondRateLimit(c, 30)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-rate", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}

	var resp utils.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Retryable != true {
		t.Error("RATE_LIMITED should be retryable")
	}
	if resp.Code != "RATE_LIMITED" {
		t.Errorf("Code = %q, want RATE_LIMITED", resp.Code)
	}
}

func TestDirectChat_Validation(t *testing.T) {
	router := setupRouter()
	router.POST("/chat/direct", func(c *gin.Context) {
		var req struct {
			CustomerName string `json:"customer_name"`
			Message      string `json:"message" binding:"required"`
			Channel      string `json:"channel" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing message and channel", map[string]string{}, http.StatusBadRequest},
		{"missing channel", map[string]string{"message": "hi"}, http.StatusBadRequest},
		{"missing message", map[string]string{"channel": "web"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/chat/direct", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestTrainingCreateCategory_Validation(t *testing.T) {
	router := setupRouter()
	router.POST("/training/categories", func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/training/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTrainingCreateQAPair_Validation(t *testing.T) {
	router := setupRouter()
	router.POST("/training/qa", func(c *gin.Context) {
		var req struct {
			CategoryID string `json:"category_id" binding:"required"`
			Question   string `json:"question" binding:"required"`
			Answer     string `json:"answer" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"all missing", map[string]string{}, http.StatusBadRequest},
		{"missing question", map[string]string{"category_id": "cat1", "answer": "ans"}, http.StatusBadRequest},
		{"missing answer", map[string]string{"category_id": "cat1", "question": "q"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/training/qa", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestBatchTrainUnknown_Validation(t *testing.T) {
	router := setupRouter()
	router.POST("/training/batch-train", func(c *gin.Context) {
		var req struct {
			IDs        []string `json:"ids" binding:"required,min=1"`
			Answer     string   `json:"answer" binding:"required"`
			CategoryID string   `json:"category_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{"empty ids", map[string]interface{}{"ids": []string{}, "answer": "a", "category_id": "c"}, http.StatusBadRequest},
		{"missing ids", map[string]interface{}{"answer": "a", "category_id": "c"}, http.StatusBadRequest},
		{"nil ids", map[string]interface{}{"ids": nil, "answer": "a", "category_id": "c"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/training/batch-train", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestContentNegotiation(t *testing.T) {
	router := setupRouter()
	router.POST("/test", func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondValidationError(c, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Gin's ShouldBindJSON parses JSON regardless of Content-Type header,
	// which is standard behavior. Test that it still works.
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Gin parses JSON regardless of Content-Type, got %d", w.Code)
	}

	// Test with truly invalid JSON body
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte("not json")))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("invalid JSON body should return 400, got %d", w2.Code)
	}
}
