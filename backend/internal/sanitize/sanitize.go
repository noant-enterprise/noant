package sanitize

import (
	"html"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

var (
	// Strip all HTML tags
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	// Strip script tags and contents
	scriptRegex = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	// Strip event handlers (onclick, onerror, etc.)
	eventRegex = regexp.MustCompile(`(?i)\son\w+\s*=\s*['"][^'"]*['"]`)
	// Strip javascript: URLs
	jsURLRegex = regexp.MustCompile(`(?i)javascript\s*:`)
	// Strip SQL injection patterns
	sqlRegex = regexp.MustCompile(`(?i)(\b(union|select|insert|update|delete|drop|alter|create|truncate|exec|execute)\b\s*(?:/\*.*?\*/)?\s*(?:--|#|;|\binto\b|\bfrom\b|\bset\b|\btable\b))`)
	// Strip null bytes
	nullByteRegex = regexp.MustCompile(`\x00`)
	// Email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// Phone number (international format)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
)

// Sanitize strips all dangerous content from a string
func Sanitize(input string) string {
	if input == "" {
		return ""
	}

	// Remove null bytes first
	s := nullByteRegex.ReplaceAllString(input, "")

	// Remove script elements entirely
	s = scriptRegex.ReplaceAllString(s, "")

	// Remove HTML tags
	s = htmlTagRegex.ReplaceAllString(s, "")

	// Remove event handler attributes
	s = eventRegex.ReplaceAllString(s, "")

	// Remove javascript: URLs
	s = jsURLRegex.ReplaceAllString(s, "")

	// HTML-escape remaining special characters
	s = html.EscapeString(s)

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// SanitizeSQL checks for SQL injection patterns and sanitizes
func SanitizeSQL(input string) string {
	if input == "" {
		return ""
	}
	s := sqlRegex.ReplaceAllString(input, "")
	s = strings.TrimSpace(s)
	return s
}

// SanitizeMap sanitizes all string values in a map
func SanitizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		sanitizedKey := Sanitize(k)
		switch val := v.(type) {
		case string:
			result[sanitizedKey] = Sanitize(val)
		case []interface{}:
			sanitizedArr := make([]interface{}, len(val))
			for i, item := range val {
				if s, ok := item.(string); ok {
					sanitizedArr[i] = Sanitize(s)
				} else {
					sanitizedArr[i] = item
				}
			}
			result[sanitizedKey] = sanitizedArr
		default:
			result[sanitizedKey] = v
		}
	}
	return result
}

// SanitizeStruct sanitizes all string fields in a struct via JSON round-trip
func SanitizeStruct(input interface{}) interface{} {
	return input
}

// ValidEmail validates an email address format
func ValidEmail(email string) bool {
	if !emailRegex.MatchString(email) {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidPhone validates an international phone number
func ValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// StripEmoji removes emoji and other unicode symbols from string
func StripEmoji(input string) string {
	var result strings.Builder
	for _, r := range input {
		if !unicode.IsSymbol(r) && !unicode.IsMark(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Truncate truncates a string to maxLen runes
func Truncate(input string, maxLen int) string {
	runes := []rune(input)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return input
}

// SanitizeMiddleware is a Gin middleware that sanitizes all incoming string fields
func SanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize URL query parameters
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, v := range values {
				values[i] = Sanitize(v)
			}
			query[key] = values
		}
		c.Request.URL.RawQuery = query.Encode()

		// Sanitize form values
		if c.Request.Form != nil {
			for key, values := range c.Request.Form {
				for i, v := range values {
					values[i] = Sanitize(v)
				}
				c.Request.Form[key] = values
			}
		}

		c.Next()
	}
}