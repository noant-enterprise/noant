package middleware

import (
	"fmt"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"noant/internal/infrastructure"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

	
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			c.Abort()
			return
		}
	}
	
	c.Set("userID", claims["user_id"])
		c.Set("userEmail", claims["email"])
		c.Set("userRole", claims["role"])
		c.Set("tokenExpiry", claims["exp"])

		c.Next()
	}
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

		if statusCode >= 500 {
			logger.Error("HTTP request", "method", method, "path", path, "status", statusCode, "latency_ms", latency.Milliseconds(), "ip", clientIP)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request", "method", method, "path", path, "status", statusCode, "latency_ms", latency.Milliseconds(), "ip", clientIP)
		} else {
			logger.Info("HTTP request", "method", method, "path", path, "status", statusCode, "latency_ms", latency.Milliseconds(), "ip", clientIP)
		}	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' fonts.googleapis.com cdn.jsdelivr.net; font-src 'self' fonts.gstatic.com; img-src 'self' data: blob:; connect-src 'self' api.groq.com;")
		c.Next()
	}
}


func RateLimitMiddleware(redis *infrastructure.RedisClient, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redis == nil {
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
			fmt.Printf("[RATE LIMIT HIT] IP: %s, Endpoint: %s, Time: %s\n", 
				clientIP, path, time.Now().Format(time.RFC3339))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded",
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
			fmt.Printf("[RATE LIMIT HIT] UserID: %s, Endpoint: %s, Time: %s\n", 
				userID, path, time.Now().Format(time.RFC3339))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
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
	rand.Read(b)
	return hex.EncodeToString(b)
}

func PlanEnforcementMiddleware(redis *infrastructure.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		// Check if user has exceeded monthly AI quota
		// In production, fetch plan limits from DB/cache
		// For now, allow all authenticated requests
		c.Next()
	}
}