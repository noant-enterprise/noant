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

		// Skip CSRF for Bearer token auth — not vulnerable to CSRF
		if auth := c.Request.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = c.Request.Header.Get("Referer")
		}

		if origin == "" || origin == "null" {
			// Require Origin or Referer for all mutating requests
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF check failed: missing origin"})
			return
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed || origin == allowed+"/" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF check failed"})
	}
}
