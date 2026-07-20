package utils

import (
	"html"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Regex patterns for XSS sanitization
var (
	scriptRegex   = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	htmlTagRegex  = regexp.MustCompile(`<[^>]*>`)
	eventRegex    = regexp.MustCompile(`(?i)\son\w+\s*=\s*['"][^'"]*['"]`)
	jsURLRegex    = regexp.MustCompile(`(?i)javascript\s*:`)
	nullByteRegex = regexp.MustCompile(`\x00`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
)

// SanitizeString cleans normal text inputs (removes null bytes, control chars, trims).
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = nullByteRegex.ReplaceAllString(s, "")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\t' || r == '\r' || r >= 0x20 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SanitizeXSS strips dangerous HTML, script tags, event handlers, and javascript links.
func SanitizeXSS(s string) string {
	if s == "" {
		return ""
	}
	s = SanitizeString(s)
	s = scriptRegex.ReplaceAllString(s, "")
	s = htmlTagRegex.ReplaceAllString(s, "")
	s = eventRegex.ReplaceAllString(s, "")
	s = jsURLRegex.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// HTMLEscape encodes HTML characters (only use when rendering, or as an extra guard).
func HTMLEscape(s string) string {
	return html.EscapeString(s)
}

// SanitizeName cleans names/titles by limiting length and removing newlines/tabs.
func SanitizeName(s string) string {
	s = SanitizeXSS(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if utf8.RuneCountInString(s) > 255 {
		runes := []rune(s)
		s = string(runes[:255])
	}
	return s
}

// IsValidEmail validates an email address.
func IsValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsValidPhone validates a phone number.
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsValidUUID checks if a string is a valid UUID.
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func IsValidUUID(id string) bool {
	return uuidRegex.MatchString(id)
}

// SanitizeStruct recursively traverses any struct and sanitizes string fields.
// It skips fields with "password", "secret", "token", "key" in their name (case-insensitive) or fields tagged with `sanitize:"skip"`.
func SanitizeStruct(v interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return
	}
	sanitizeValue(val.Elem(), "")
}

func sanitizeValue(val reflect.Value, fieldName string) {
	if !val.IsValid() {
		return
	}

	lowerFieldName := strings.ToLower(fieldName)
	// Skip sensitive fields
	if strings.Contains(lowerFieldName, "password") ||
		strings.Contains(lowerFieldName, "secret") ||
		strings.Contains(lowerFieldName, "token") ||
		strings.Contains(lowerFieldName, "key") {
		return
	}

	switch val.Kind() {
	case reflect.String:
		if val.CanSet() {
			str := val.String()
			var sanitized string
			if strings.Contains(lowerFieldName, "email") {
				sanitized = SanitizeString(str)
			} else {
				sanitized = SanitizeXSS(str)
			}
			val.SetString(sanitized)
		}

	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			structField := val.Type().Field(i)
			// Skip unexported fields
			if structField.PkgPath != "" {
				continue
			}
			if structField.Tag.Get("sanitize") == "skip" {
				continue
			}
			sanitizeValue(field, structField.Name)
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			sanitizeValue(val.Index(i), fieldName)
		}

	case reflect.Map:
		if val.IsNil() {
			return
		}
		for _, key := range val.MapKeys() {
			mapVal := val.MapIndex(key)
			switch mapVal.Kind() {
			case reflect.String:
				str := mapVal.String()
				var sanitized string
				if strings.Contains(lowerFieldName, "email") {
					sanitized = SanitizeString(str)
				} else {
					sanitized = SanitizeXSS(str)
				}
				val.SetMapIndex(key, reflect.ValueOf(sanitized))
			case reflect.Interface:
				elem := mapVal.Elem()
				if elem.Kind() == reflect.String {
					str := elem.String()
					var sanitized string
					if strings.Contains(lowerFieldName, "email") {
						sanitized = SanitizeString(str)
					} else {
						sanitized = SanitizeXSS(str)
					}
					val.SetMapIndex(key, reflect.ValueOf(sanitized))
				} else if elem.Kind() == reflect.Map || elem.Kind() == reflect.Slice || elem.Kind() == reflect.Struct || elem.Kind() == reflect.Pointer {
					sanitizeValue(elem, fieldName)
				}
			default:
				sanitizeValue(mapVal, fieldName)
			}
		}

	case reflect.Pointer:
		if !val.IsNil() {
			sanitizeValue(val.Elem(), fieldName)
		}

	case reflect.Interface:
		if !val.IsNil() {
			sanitizeValue(val.Elem(), fieldName)
		}
	}
}
