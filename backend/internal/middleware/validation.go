package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// FieldRule defines validation rules for a single JSON field.
type FieldRule struct {
	Required  bool
	MinLen    int
	MaxLen    int
	Email     bool
	Type      string // "string", "number"
}

// ValidateJSON returns middleware that parses the request body as JSON and
// validates it against the provided field rules. On failure it responds with
// a structured ErrorResponse matching the project convention.
func ValidateJSON(fields map[string]FieldRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_BODY",
				Message: "Failed to read request body",
			})
			c.Abort()
			return
		}
		// Restore body so downstream handlers can still read it.
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) == 0 {
			for name, rule := range fields {
				if rule.Required {
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Code:    "VALIDATION_ERROR",
						Message: fmt.Sprintf("Field '%s' is required", name),
					})
					c.Abort()
					return
				}
			}
			c.Next()
			return
		}

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_JSON",
				Message: "Request body is not valid JSON",
			})
			c.Abort()
			return
		}

		if errs := ValidateFields(data, fields); len(errs) > 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: strings.Join(errs, "; "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// emailRegex is a simple but sufficient check for most cases.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateFields checks data against the given rules and returns any error messages.
func ValidateFields(data map[string]interface{}, fields map[string]FieldRule) []string {
	var errs []string

	for name, rule := range fields {
		val, exists := data[name]

		if rule.Required && (!exists || val == nil) {
			errs = append(errs, fmt.Sprintf("Field '%s' is required", name))
			continue
		}

		if !exists || val == nil {
			continue
		}

		strVal, isString := val.(string)

		switch {
		case rule.Type == "string" && !isString:
			errs = append(errs, fmt.Sprintf("Field '%s' must be a string", name))
			continue
		case rule.Type == "number":
			switch v := val.(type) {
			case float64:
				// OK – JSON numbers decode as float64
				_ = v
			case json.Number:
				// Already OK if using a decoder with UseNumber
			default:
				errs = append(errs, fmt.Sprintf("Field '%s' must be a number", name))
				continue
			}
		}

		if isString {
			if rule.MinLen > 0 && len(strVal) < rule.MinLen {
				errs = append(errs, fmt.Sprintf("Field '%s' must be at least %d characters", name, rule.MinLen))
			}
			if rule.MaxLen > 0 && len(strVal) > rule.MaxLen {
				errs = append(errs, fmt.Sprintf("Field '%s' must be at most %d characters", name, rule.MaxLen))
			}
			if rule.Email && !emailRegex.MatchString(strVal) {
				errs = append(errs, fmt.Sprintf("Field '%s' must be a valid email address", name))
			}
		}
	}

	return errs
}

// --- Pre-built validation middlewares for key endpoints ---

// ValidateRegister validates the register request body.
func ValidateRegister() gin.HandlerFunc {
	return ValidateJSON(map[string]FieldRule{
		"email":       {Required: true, Email: true},
		"password":    {Required: true, MinLen: 8},
		"first_name":  {Required: true, MaxLen: 100},
		"last_name":   {Required: true, MaxLen: 100},
		"company_name": {MaxLen: 200},
	})
}

// ValidateLogin validates the login request body.
func ValidateLogin() gin.HandlerFunc {
	return ValidateJSON(map[string]FieldRule{
		"email":    {Required: true, Email: true},
		"password": {Required: true},
	})
}

// ValidateCreateQAPair validates the create QA pair request body.
func ValidateCreateQAPair() gin.HandlerFunc {
	return ValidateJSON(map[string]FieldRule{
		"category_id": {Required: true},
		"question":    {Required: true, MaxLen: 5000},
		"answer":      {Required: true, MaxLen: 10000},
	})
}

// ValidateDirectChat validates the direct chat request body.
func ValidateDirectChat() gin.HandlerFunc {
	return ValidateJSON(map[string]FieldRule{
		"message": {Required: true, MaxLen: 10000},
		"channel": {Required: true},
	})
}

// ValidateSendMessage validates the send message request body.
func ValidateSendMessage() gin.HandlerFunc {
	return ValidateJSON(map[string]FieldRule{
		"content": {Required: true, MaxLen: 10000},
	})
}
