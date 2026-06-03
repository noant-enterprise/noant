package middleware

import (
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

// SanitizeMiddleware is a Gin middleware that sanitizes all incoming URL query and form parameters.
func SanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize URL query parameters
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, v := range values {
				values[i] = utils.SanitizeXSS(v)
			}
			query[key] = values
		}
		c.Request.URL.RawQuery = query.Encode()

		// Sanitize form values
		if c.Request.Form != nil {
			for key, values := range c.Request.Form {
				for i, v := range values {
					values[i] = utils.SanitizeXSS(v)
				}
				c.Request.Form[key] = values
			}
		}

		c.Next()
	}
}
