package middleware

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const MaxRequestBodySize = 1 << 20 // 1 MB

func BodyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		if c.Request.ContentLength > MaxRequestBodySize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
			})
			return
		}

		c.Request.Body = io.NopCloser(io.LimitReader(c.Request.Body, MaxRequestBodySize))

		c.Next()
	}
}
