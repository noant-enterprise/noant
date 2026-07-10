package middleware

import (
	"bytes"
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

		if c.Request.ContentLength == -1 || c.Request.ContentLength == 0 {
			// Chunked or unknown length — wrap with a limited reader and on-the-fly enforcement
			limited := io.LimitReader(c.Request.Body, MaxRequestBodySize+1)
			readBuf, err := io.ReadAll(limited)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
				return
			}
			if len(readBuf) > MaxRequestBodySize {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": "Request body too large",
				})
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(readBuf))
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(io.LimitReader(c.Request.Body, MaxRequestBodySize))

		c.Next()
	}
}
