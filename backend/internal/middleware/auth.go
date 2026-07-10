package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenCookieName  = "noant_access"
	refreshTokenCookieName = "noant_refresh"
)

func AuthMiddleware(jwtSecret string, redis *infrastructure.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := GetAccessTokenFromRequest(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		if redis != nil {
			if blacklisted, err := redis.Exists(c.Request.Context(), "blacklist:"+tokenString); err == nil && blacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
				c.Abort()
				return
			}
		} else if defaultMemoryBlacklist.Exists(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || token == nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		if tokenType, _ := claims["type"].(string); tokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
			c.Abort()
			return
		}

		exp, err := claims.GetExpirationTime()
		if err != nil || exp == nil || time.Now().After(exp.Time) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			c.Abort()
			return
		}

		userID, _ := claims["user_id"].(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("userEmail", claims["email"])
		c.Set("userRole", claims["role"])
		c.Set("tokenExpiry", exp.Time.Unix())

		c.Next()
	}
}

func GetAccessTokenFromRequest(c *gin.Context) string {
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	if cookie, err := c.Cookie(accessTokenCookieName); err == nil && cookie != "" {
		return cookie
	}

	return ""
}

func GetRefreshTokenFromRequest(c *gin.Context) string {
	if cookie, err := c.Cookie(refreshTokenCookieName); err == nil && cookie != "" {
		return cookie
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err == nil {
		return strings.TrimSpace(req.RefreshToken)
	}
	return ""
}

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	secure := isSecureRequest(c)

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(accessTokenCookieName, accessToken, int(accessTTL.Seconds()), "/", "", secure, true)
	c.SetCookie(refreshTokenCookieName, refreshToken, int(refreshTTL.Seconds()), "/", "", secure, true)
}

func ClearAuthCookies(c *gin.Context) {
	secure := isSecureRequest(c)

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(accessTokenCookieName, "", -1, "/", "", secure, true)
	c.SetCookie(refreshTokenCookieName, "", -1, "/", "", secure, true)
}

func isSecureRequest(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		return true
	}
	return false
}

func LoggerMiddleware(logger *infrastructure.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Get request ID from context (set by RequestIDMiddleware)
		reqID, _ := c.Get("requestID")
		reqIDStr, _ := reqID.(string)

		// Instrument Prometheus metrics
		infrastructure.RequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Inc()
		infrastructure.RequestDuration.WithLabelValues(method, path).Observe(latency.Seconds())

		if statusCode >= 500 {
			logger.Error("HTTP request", "method", method, "path", path, "status", statusCode, "latency_ms", latency.Milliseconds(), "ip", clientIP, "request_id", reqIDStr)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request", "method", method, "path", path, "status", statusCode, "latency_ms", latency.Milliseconds(), "ip", clientIP, "request_id", reqIDStr)
		} else {
			logger.Info("HTTP request", "method", method, "path", path, "status", statusCode, "latency_ms", latency.Milliseconds(), "ip", clientIP, "request_id", reqIDStr)
		}
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "0")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: blob:; connect-src 'self' api.groq.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self';")
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		c.Next()
	}
}

var (
	defaultMemoryRateLimiter = infrastructure.NewMemoryRateLimiter(5 * time.Minute)
	defaultMemoryBlacklist   = infrastructure.NewMemoryBlacklist()
)

func BlacklistAccessToken(token string) {
	if token != "" {
		defaultMemoryBlacklist.Add(token, 24*time.Hour)
	}
}

func RateLimitMiddleware(redis *infrastructure.RedisClient, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redis == nil {
			if !defaultMemoryRateLimiter.Allow(c.ClientIP()+":"+c.Request.URL.Path, requests, window) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate limit exceeded",
					"retry_after": window.Seconds(),
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		key := fmt.Sprintf("ratelimit:%s:%s", clientIP, path)

		allowed, remaining, err := redis.RateLimitWithRemaining(c.Request.Context(), key, requests, window)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitByUserMiddleware(redis *infrastructure.RedisClient, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redis == nil {
			userID, exists := c.Get("userID")
			if !exists {
				c.Next()
				return
			}
			if !defaultMemoryRateLimiter.Allow(fmt.Sprintf("user:%s:%s", userID, c.Request.URL.Path), requests, window) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate limit exceeded",
					"retry_after": window.Seconds(),
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		key := fmt.Sprintf("ratelimit:user:%s:%s", userID, path)

		allowed, remaining, err := redis.RateLimitWithRemaining(c.Request.Context(), key, requests, window)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func PlanEnforcementMiddleware(redis *infrastructure.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		c.Next()
	}
}

// TrialExpirationMiddleware checks if the user's trial has expired and blocks feature access.
// It reads the trial_expires_at from a database lookup via a provided userProvider function.
func TrialExpirationMiddleware(userProvider func(ctx context.Context, userID string) (*time.Time, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(string)
		if !ok || uid == "" {
			c.Next()
			return
		}

		trialExpiresAt, err := userProvider(c.Request.Context(), uid)
		if err != nil || trialExpiresAt == nil {
			c.Next()
			return
		}

		if time.Now().After(*trialExpiresAt) {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":         "Your free trial has expired. Please subscribe to continue using Noant.",
				"trial_expired": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		roleStr, ok := role.(string)
		if !ok || (roleStr != "owner" && roleStr != "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
		c.Next()
	}
}

// GetTrialExpiry is a helper to extract trial_expires_at from the user model
func GetTrialExpiry(user *domain.User) *time.Time {
	if user == nil {
		return nil
	}
	return user.TrialExpiresAt
}
