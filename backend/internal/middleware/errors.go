package middleware

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	apperrors "noant/internal/errors"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard JSON error payload returned by all API endpoints.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ClassifyError maps an error to an HTTP status code, a machine-readable error
// code, and a user-facing message.
func ClassifyError(err error) (statusCode int, errorCode, userMessage string) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password"
	case errors.Is(err, apperrors.ErrAccountLocked):
		return http.StatusTooManyRequests, "ACCOUNT_LOCKED", "Account temporarily locked due to too many failed attempts"
	case errors.Is(err, apperrors.ErrEmailNotVerified):
		return http.StatusForbidden, "EMAIL_NOT_VERIFIED", "Email verification required"
	case errors.Is(err, apperrors.ErrInvalidVerification):
		return http.StatusBadRequest, "INVALID_VERIFICATION", "Invalid verification code"
	case errors.Is(err, apperrors.ErrTooManyVerifications):
		return http.StatusTooManyRequests, "TOO_MANY_VERIFICATIONS", "Too many verification attempts"
	case errors.Is(err, apperrors.ErrEmailAlreadyVerified):
		return http.StatusConflict, "EMAIL_ALREADY_VERIFIED", "Email is already verified"
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "Resource not found"
	case errors.Is(err, apperrors.ErrUnknownQuestion):
		return http.StatusNotFound, "NOT_FOUND", "Resource not found"
	case errors.Is(err, apperrors.ErrCampaign):
		return http.StatusNotFound, "NOT_FOUND", "Campaign not found or access denied"
	case errors.Is(err, apperrors.ErrInsufficientCredit):
		return http.StatusPaymentRequired, "INSUFFICIENT_CREDIT", "Insufficient credits"
	case errors.Is(err, apperrors.ErrCreditExpired):
		return http.StatusPaymentRequired, "CREDIT_EXPIRED", "Credit balance has expired"
	case errors.Is(err, apperrors.ErrCircuitBreakerOpen):
		return http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service temporarily unavailable"
	case errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound, "NOT_FOUND", "Resource not found"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred"
	}
}

// RespondError writes a standardized JSON error response to the client.
// For 5xx status codes the error is logged with the request ID (if available).
func RespondError(c *gin.Context, err error) {
	statusCode, code, message := ClassifyError(err)
	resp := ErrorResponse{
		Code:    code,
		Message: message,
	}

	if statusCode >= 500 {
		reqID, _ := c.Get("requestID")
		slog.Error("internal server error",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", statusCode,
			"code", code,
			"request_id", reqID,
			"error", err,
		)
	}

	c.JSON(statusCode, resp)
}

// StandardizedResponseMiddleware recovers from panics and returns a consistent
// 500 error response. Wrap top-level router groups with this to prevent
// unhandled panics from crashing the process or leaking stack traces.
func StandardizedResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID, _ := c.Get("requestID")
				slog.Error("panic recovered",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"request_id", reqID,
					"panic", r,
				)

				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Code:    "INTERNAL_ERROR",
						Message: "An unexpected error occurred",
					})
				}
			}
		}()
		c.Next()
	}
}
