package middleware

import (
	"context"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"

	"github.com/gin-gonic/gin"
)

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
			userIDStr = userID.(string)
		}

		action := method + " " + c.FullPath()
		resourceType := c.FullPath()
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("channel")
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		// Run the handler
		c.Next()

				// Only log if handler succeeded (status < 400) and userID exists
		if c.Writer.Status() < 400 && userIDStr != "" {
			log := &domain.AuditLog{
				UserID:       userIDStr,
				Action:       action,
				ResourceType: resourceType,
				ResourceID:   &resourceID,
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
