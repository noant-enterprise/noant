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

		action := describeAction(method, c.FullPath())
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

var routeActions = map[string]string{
	"/api/v1/chats/direct-chat":                       "Started a new conversation",
	"/api/v1/chats/clear":                              "Cleared chat history",
	"/api/v1/chats/conversations/:id/stream":           "Sent a message",
	"/api/v1/integrations/connect":                     "Connected an integration",
	"/api/v1/integrations/disconnect/:channel":         "Disconnected an integration",
	"/api/v1/training/unknown-questions/:id/train":     "Trained an unknown question",
	"/api/v1/training/unknown-questions/:id/ignore":    "Ignored an unknown question",
	"/api/v1/training/unknown-questions/:id/dismiss":   "Dismissed an unknown question",
	"/api/v1/training/unknown-questions/batch-ignore":  "Ignored unknown questions in bulk",
	"/api/v1/training/unknown-questions/batch-train":   "Trained unknown questions in bulk",
	"/api/v1/training/categories":                      "Created a new category",
	"/api/v1/training/categories/:id":                  "Updated a category",
	"/api/v1/qa-pairs":                                 "Created a new Q&A pair",
	"/api/v1/qa-pairs/:id":                             "Updated a Q&A pair",
	"/api/v1/integrations/whatsapp/send-template":      "Sent a WhatsApp template message",
	"/api/v1/campaigns":                                "Created a campaign",
	"/api/v1/campaigns/:id/send":                       "Sent a campaign",
	"/api/v1/campaigns/:id/pause":                      "Paused a campaign",
	"/api/v1/campaigns/:id/resume":                     "Resumed a campaign",
	"/api/v1/campaigns/:id/cancel":                     "Canceled a campaign",
	"/api/v1/handoffs":                                 "Created a handoff",
	"/api/v1/handoffs/:id/accept":                      "Accepted a handoff",
	"/api/v1/handoffs/:id/decline":                     "Declined a handoff",
	"/api/v1/handoffs/:id/resolve":                     "Resolved a handoff",
	"/api/v1/archive/folders":                          "Created an archive folder",
	"/api/v1/settings/api-keys":                        "Generated a new API key",
	"/api/v1/settings/api-keys/:id/revoke":             "Revoked an API key",
	"/api/v1/team/invite":                              "Invited a team member",
	"/api/v1/team/:id/role":                            "Updated a team member's role",
	"/api/v1/team/:id/remove":                          "Removed a team member",
	"/api/v1/integrations/openwa/connect":              "Connected WhatsApp",
	"/api/v1/integrations/openwa/disconnect":           "Disconnected WhatsApp",
	"/api/v1/integrations/openwa/send":                 "Sent a WhatsApp message",
	"/api/v1/integrations/openwa/send-media":           "Sent a WhatsApp media message",
	"/api/v1/integrations/openwa/send-bulk":            "Sent bulk WhatsApp messages",
	"/api/v1/integrations/openwa/sessions/:id/restart": "Restarted a WhatsApp session",
	"/api/v1/settings/profile":                         "Updated profile settings",
	"/api/v1/settings/password":                        "Changed password",
	"/api/v1/settings/notifications":                   "Updated notification preferences",
	"/api/v1/settings/2fa/enable":                      "Enabled two-factor authentication",
	"/api/v1/settings/2fa/disable":                     "Disabled two-factor authentication",
	"/api/v1/settings/delete-account":                  "Deleted account",
	"/api/v1/settings/export-data":                     "Exported account data",
	"/api/v1/settings/privacy":                         "Updated privacy settings",
	"/api/v1/notifications/:id/read":                   "Marked a notification as read",
	"/api/v1/notifications/read-all":                   "Marked all notifications as read",
	"/api/v1/integrations/openwa/verify":               "Verified WhatsApp connection",
	"/api/v1/integrations/openwa/status":               "Checked WhatsApp status",
	"/api/v1/widget/config":                            "Updated widget configuration",
	"/api/v1/notifications/widget":                     "Viewed widget configuration",
	"/api/v1/archive/folders/:id":                      "Updated an archive folder",
	"/api/v1/archive/conversations/:id/move":           "Moved a conversation to archive",
	"/api/v1/archive/conversations/:id/restore":        "Restored a conversation from archive",
}

func describeAction(method, path string) string {
	if desc, ok := routeActions[path]; ok {
		return desc
	}
	return method + " " + path
}
