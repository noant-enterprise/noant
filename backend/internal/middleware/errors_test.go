package middleware

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "noant/internal/errors"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- ClassifyError tests ---

func TestClassifyError_InvalidCredentials(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrInvalidCredentials)
	assertClassification(t, status, code, msg, 401, "INVALID_CREDENTIALS", "Invalid email or password")
}

func TestClassifyError_AccountLocked(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrAccountLocked)
	assertClassification(t, status, code, msg, 429, "ACCOUNT_LOCKED", "Account temporarily locked due to too many failed attempts")
}

func TestClassifyError_EmailNotVerified(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrEmailNotVerified)
	assertClassification(t, status, code, msg, 403, "EMAIL_NOT_VERIFIED", "Email verification required")
}

func TestClassifyError_InvalidVerification(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrInvalidVerification)
	assertClassification(t, status, code, msg, 400, "INVALID_VERIFICATION", "Invalid verification code")
}

func TestClassifyError_TooManyVerifications(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrTooManyVerifications)
	assertClassification(t, status, code, msg, 429, "TOO_MANY_VERIFICATIONS", "Too many verification attempts")
}

func TestClassifyError_EmailAlreadyVerified(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrEmailAlreadyVerified)
	assertClassification(t, status, code, msg, 409, "EMAIL_ALREADY_VERIFIED", "Email is already verified")
}

func TestClassifyError_NotFound(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrNotFound)
	assertClassification(t, status, code, msg, 404, "NOT_FOUND", "Resource not found")
}

func TestClassifyError_UnknownQuestion(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrUnknownQuestion)
	assertClassification(t, status, code, msg, 404, "NOT_FOUND", "Resource not found")
}

func TestClassifyError_Campaign(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrCampaign)
	assertClassification(t, status, code, msg, 404, "NOT_FOUND", "Campaign not found or access denied")
}

func TestClassifyError_InsufficientCredit(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrInsufficientCredit)
	assertClassification(t, status, code, msg, 402, "INSUFFICIENT_CREDIT", "Insufficient credits")
}

func TestClassifyError_CreditExpired(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrCreditExpired)
	assertClassification(t, status, code, msg, 402, "CREDIT_EXPIRED", "Credit balance has expired")
}

func TestClassifyError_CircuitBreakerOpen(t *testing.T) {
	status, code, msg := ClassifyError(apperrors.ErrCircuitBreakerOpen)
	assertClassification(t, status, code, msg, 503, "SERVICE_UNAVAILABLE", "AI service temporarily unavailable")
}

func TestClassifyError_SqlErrNoRows(t *testing.T) {
	status, code, msg := ClassifyError(sql.ErrNoRows)
	assertClassification(t, status, code, msg, 404, "NOT_FOUND", "Resource not found")
}

func TestClassifyError_UnknownError(t *testing.T) {
	status, code, msg := ClassifyError(errors.New("something broke"))
	assertClassification(t, status, code, msg, 500, "INTERNAL_ERROR", "An unexpected error occurred")
}

func TestClassifyError_NilError(t *testing.T) {
	status, code, msg := ClassifyError(nil)
	assertClassification(t, status, code, msg, 500, "INTERNAL_ERROR", "An unexpected error occurred")
}

// --- RespondError tests ---

func TestRespondError_ClientError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)

	RespondError(c, apperrors.ErrInvalidCredentials)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected code INVALID_CREDENTIALS, got %s", resp.Code)
	}
	if resp.Message != "Invalid email or password" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
	if resp.Details != "" {
		t.Errorf("expected empty details, got %s", resp.Details)
	}
}

func TestRespondError_ServerError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)
	c.Set("requestID", "test-req-123")

	RespondError(c, errors.New("db connection lost"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", resp.Code)
	}
}

func TestRespondError_JSONFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)

	RespondError(c, apperrors.ErrNotFound)

	contentType := w.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("expected Content-Type to start with application/json, got %s", contentType)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["code"]; !ok {
		t.Error("response missing 'code' field")
	}
	if _, ok := resp["message"]; !ok {
		t.Error("response missing 'message' field")
	}
}

// --- Panic recovery tests ---

func TestStandardizedResponseMiddleware_PanicRecovery(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/panic", nil)

	router := gin.New()
	router.Use(StandardizedResponseMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("something terrible happened")
	})

	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", resp.Code)
	}
}

func TestStandardizedResponseMiddleware_NoPanic(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/ok", nil)

	router := gin.New()
	router.Use(StandardizedResponseMiddleware())
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestStandardizedResponseMiddleware_AlreadyWritten(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/partial", nil)

	router := gin.New()
	router.Use(StandardizedResponseMiddleware())
	router.GET("/partial", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"sent": true})
		panic("after write")
	})

	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 (response already written), got %d", w.Code)
	}
}

// --- helpers ---

func assertClassification(t *testing.T, status int, code, msg string, wantStatus int, wantCode, wantMsg string) {
	t.Helper()
	if status != wantStatus {
		t.Errorf("status: got %d, want %d", status, wantStatus)
	}
	if code != wantCode {
		t.Errorf("code: got %s, want %s", code, wantCode)
	}
	if msg != wantMsg {
		t.Errorf("message: got %q, want %q", msg, wantMsg)
	}
}
