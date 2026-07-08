package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CSRFMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = c.Request.Header.Get("Referer")
		}

		if origin == "" || origin == "null" {
			c.Next()
			return
		}

		for _, allowed := range allowedOrigins {
			if strings.HasPrefix(origin, allowed) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF check failed"})
	}
}
