package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"

	"github.com/gin-gonic/gin"
)

const maxAuditBodySize = 1024 * 64 // 64KB max body capture

// AuditLog is a helper to write audit entries from handlers (auth events, etc.)
func AuditLog(ctx context.Context, repo *repository.AuditRepository, userID, action, resourceType string, resourceID *string, details map[string]interface{}, ip, ua string) {
	log := &domain.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    &ip,
		UserAgent:    &ua,
	}
	go func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = repo.Create(c, log)
	}()
}

// AuditMiddleware logs every authenticated action to the audit log
func AuditMiddleware(auditRepo *repository.AuditRepository, logger *infrastructure.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit mutating operations (POST, PUT, DELETE, PATCH)
		method := c.Request.Method
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			c.Next()
			return
		}

		// Capture request details before handler runs
		userID, _ := c.Get("userID")
		userIDStr := ""
		if userID != nil {
			if s, ok := userID.(string); ok {
				userIDStr = s
			}
		}

		action := method + " " + c.FullPath()
		resourceType := c.FullPath()
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("channel")
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		// Capture request body for mutation details (limited to 64KB)
		var bodyBytes []byte
		if c.Request.Body != nil && method != "GET" {
			bodyBytes, _ = io.ReadAll(io.LimitReader(c.Request.Body, maxAuditBodySize))
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Run the handler
		c.Next()

		// Log both successful and failed operations
		if userIDStr != "" {
			status := c.Writer.Status()
			details := map[string]interface{}{
				"status": status,
			}

			// Capture error response for failed operations
			if status >= 400 {
				// Try to extract error from response body
				var errResp map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
					if msg, ok := errResp["error"]; ok {
						details["error"] = msg
					}
				}
				details["success"] = false
			} else {
				details["success"] = true
			}

			// Capture sanitized request body (exclude passwords/tokens)
			if len(bodyBytes) > 0 {
				var bodyMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
					// Redact sensitive fields
					for _, sensitive := range []string{"password", "new_password", "current_password", "token", "secret", "api_key"} {
						if _, exists := bodyMap[sensitive]; exists {
							bodyMap[sensitive] = "[REDACTED]"
						}
					}
					details["request"] = bodyMap
				}
			}

			log := &domain.AuditLog{
				UserID:       userIDStr,
				Action:       action,
				ResourceType: resourceType,
				ResourceID:   &resourceID,
				Details:      details,
				IPAddress:    &ip,
				UserAgent:    &ua,
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Audit log goroutine panic recovered", "panic", r)
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := auditRepo.Create(ctx, log); err != nil {
					logger.Warn("Failed to write audit log", "error", err)
				}
			}()
		}
	}
}
