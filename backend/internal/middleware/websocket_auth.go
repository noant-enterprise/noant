package middleware

import (
	"net/http"
	"time"

	"noant/internal/infrastructure"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func WebSocketAuth(jwtSecret string, redis *infrastructure.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := GetAccessTokenFromRequest(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "websocket token required"})
			c.Abort()
			return
		}

		if redis != nil {
			if blacklisted, err := redis.Exists(c.Request.Context(), "blacklist:"+tokenString); err == nil && blacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
				c.Abort()
				return
			}
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || token == nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid websocket token"})
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
		c.Next()
	}
}
