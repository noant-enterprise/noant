package middleware

import (
	"github.com/gin-gonic/gin"
	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
)

// SentryContextMiddleware attaches user, org, and breadcrumb data to the Sentry scope
// for every request. This runs AFTER AuthMiddleware so userID/orgID are in the context.
func SentryContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		hub := sentrygin.GetHubFromContext(c)
		if hub == nil {
			c.Next()
			return
		}

		hub.ConfigureScope(func(scope *sentry.Scope) {
			if userID, ok := c.Get("userID"); ok {
				if uid, ok := userID.(string); ok {
					scope.SetUser(sentry.User{ID: uid})
				}
			}
			if orgID, ok := c.Get("orgID"); ok {
				if oid, ok := orgID.(string); ok {
					scope.SetTag("org_id", oid)
				}
			}
			if role, ok := c.Get("userRole"); ok {
				if r, ok := role.(string); ok {
					scope.SetTag("role", r)
				}
			}

			scope.SetTag("method", c.Request.Method)
			scope.SetTag("path", c.Request.URL.Path)
			scope.SetTag("query", c.Request.URL.RawQuery)
			scope.SetTag("user_agent", c.Request.UserAgent())
			scope.SetTag("remote_addr", c.ClientIP())

			if reqID, ok := c.Get("requestID"); ok {
				if rid, ok := reqID.(string); ok {
					scope.SetTag("request_id", rid)
				}
			}
		})

		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "http",
			Message:  c.Request.Method + " " + c.Request.URL.Path,
			Level:    sentry.LevelInfo,
			Data: map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			},
		}, nil)

		c.Next()

		statusCode := c.Writer.Status()
		level := sentry.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			level = sentry.LevelWarning
		} else if statusCode >= 500 {
			level = sentry.LevelError
		}

		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "http",
			Message:  "Response",
			Level:    level,
			Data: map[string]interface{}{
				"status_code": statusCode,
			},
		}, nil)
	}
}

// CaptureMessage is a helper to send a message to Sentry with current request context
func CaptureMessage(c *gin.Context, msg string, level sentry.Level) {
	hub := sentrygin.GetHubFromContext(c)
	if hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			hub.CaptureMessage(msg)
		})
	} else {
		sentry.CaptureMessage(msg)
	}
}

// CaptureException is a helper to send an error to Sentry with current request context
func CaptureException(c *gin.Context, err error) {
	hub := sentrygin.GetHubFromContext(c)
	if hub != nil {
		hub.CaptureException(err)
	} else {
		sentry.CaptureException(err)
	}
}

// AddBreadcrumb adds a custom breadcrumb to the current request scope
func AddBreadcrumb(c *gin.Context, category, message string, data map[string]interface{}) {
	hub := sentrygin.GetHubFromContext(c)
	if hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: category,
			Message:  message,
			Level:    sentry.LevelInfo,
			Data:     data,
		}, nil)
	}
}
