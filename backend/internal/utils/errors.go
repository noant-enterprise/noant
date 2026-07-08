package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Retryable bool      `json:"retryable"`
}

func RespondError(c *gin.Context, status int, code string, message string, retryable bool) {
	requestID, _ := c.Get("requestID")
	if requestID == nil {
		requestID = "unknown"
	}

	c.JSON(status, ErrorResponse{
		Success:   false,
		Error:     message,
		Code:      code,
		RequestID: requestID.(string),
		Timestamp: time.Now().UTC(),
		Retryable: retryable,
	})
}

func RespondValidationError(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", message, false)
}

func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, false)
}

func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, "FORBIDDEN", message, false)
}

func RespondNotFound(c *gin.Context, resource string) {
	RespondError(c, http.StatusNotFound, "NOT_FOUND", resource+" not found", false)
}

func RespondConflict(c *gin.Context, message string) {
	RespondError(c, http.StatusConflict, "CONFLICT", message, false)
}

func RespondInternalError(c *gin.Context, _ string) {
	RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred. Please try again later.", true)
}

func RespondRateLimit(c *gin.Context, retryAfter float64) {
	requestID, _ := c.Get("requestID")
	if requestID == nil {
		requestID = "unknown"
	}

	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Success:   false,
		Error:     "Rate limit exceeded. Please slow down.",
		Code:      "RATE_LIMITED",
		RequestID: requestID.(string),
		Timestamp: time.Now().UTC(),
		Retryable: true,
	})
}
