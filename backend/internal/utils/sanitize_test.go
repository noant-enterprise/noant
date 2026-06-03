package utils

import (
	"reflect"
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Trims whitespace",
			input:    "   hello world   ",
			expected: "hello world",
		},
		{
			name:     "Removes null bytes",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "Removes control characters except newlines/tabs",
			input:    "hello\x01\x02\nworld\t!",
			expected: "hello\nworld\t!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SanitizeString(tt.input)
			if actual != tt.expected {
				t.Errorf("SanitizeString(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestSanitizeXSS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Strips script tags and contents",
			input:    "hello <script>alert(1)</script>world",
			expected: "hello world",
		},
		{
			name:     "Strips HTML tags",
			input:    "hello <b>world</b>",
			expected: "hello world",
		},
		{
			name:     "Strips event handlers",
			input:    `<img src="x" onerror="alert(1)" onload='doIt()'>`,
			expected: "",
		},
		{
			name:     "Strips javascript URLs",
			input:    `<a href="javascript:alert(1)">click me</a>`,
			expected: "click me",
		},
		{
			name:     "Keeps normal symbols like quote and apostrophe",
			input:    "Let's check \"this\" & that.",
			expected: "Let's check \"this\" & that.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SanitizeXSS(tt.input)
			if actual != tt.expected {
				t.Errorf("SanitizeXSS(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	longName := make([]byte, 300)
	for i := range longName {
		longName[i] = 'a'
	}
	expectedLongName := string(longName[:255])

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Truncates long name",
			input:    string(longName),
			expected: expectedLongName,
		},
		{
			name:     "Removes newlines and tabs",
			input:    "John\nDoe\tSmith",
			expected: "John Doe Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SanitizeName(tt.input)
			if actual != tt.expected {
				t.Errorf("SanitizeName(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestValidationRegexes(t *testing.T) {
	t.Run("UUID validation", func(t *testing.T) {
		validUUID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
		invalidUUID := "f47ac10b-58cc-4372-a567-0e02b2c3d47g"
		if !IsValidUUID(validUUID) {
			t.Errorf("Expected IsValidUUID(%q) to be true", validUUID)
		}
		if IsValidUUID(invalidUUID) {
			t.Errorf("Expected IsValidUUID(%q) to be false", invalidUUID)
		}
	})

	t.Run("Email validation", func(t *testing.T) {
		validEmail := "test@example.com"
		invalidEmail := "testexample.com"
		if !IsValidEmail(validEmail) {
			t.Errorf("Expected IsValidEmail(%q) to be true", validEmail)
		}
		if IsValidEmail(invalidEmail) {
			t.Errorf("Expected IsValidEmail(%q) to be false", invalidEmail)
		}
	})

	t.Run("Phone validation", func(t *testing.T) {
		validPhone := "+12345678901"
		invalidPhone := "123"
		if !IsValidPhone(validPhone) {
			t.Errorf("Expected IsValidPhone(%q) to be true", validPhone)
		}
		if IsValidPhone(invalidPhone) {
			t.Errorf("Expected IsValidPhone(%q) to be false", invalidPhone)
		}
	})
}

type NestedTest struct {
	Details string
}

type TestStruct struct {
	Name            string
	Email           string
	Password        string
	SecretToken     string
	AccessKey       string
	IgnoredField    string `sanitize:"skip"`
	Nested          NestedTest
	NestedPtr       *NestedTest
	SliceOfStrings  []string
	MapOfStrings    map[string]string
	MapOfInterfaces map[string]interface{}
}

func TestSanitizeStruct(t *testing.T) {
	input := TestStruct{
		Name:         "   John <script>alert(1)</script>Doe   ",
		Email:        "   john.doe@example.com\x00   ",
		Password:     "   my<script>password   ", // Should be ignored
		SecretToken:  "   my<script>token   ",    // Should be ignored
		AccessKey:    "   my<script>key   ",      // Should be ignored
		IgnoredField: "   ignored <script>alert(1)</script>   ",
		Nested: NestedTest{
			Details: "   nested <script>alert(1)</script>   ",
		},
		NestedPtr: &NestedTest{
			Details: "   nested ptr <script>alert(1)</script>   ",
		},
		SliceOfStrings: []string{
			"   slice element <script>alert(1)</script>   ",
		},
		MapOfStrings: map[string]string{
			"key1": "   map value <script>alert(1)</script>   ",
		},
		MapOfInterfaces: map[string]interface{}{
			"key2": "   interface value <script>alert(1)</script>   ",
		},
	}

	SanitizeStruct(&input)

	expected := TestStruct{
		Name:         "John Doe",
		Email:        "john.doe@example.com",
		Password:     "   my<script>password   ", // Kept original
		SecretToken:  "   my<script>token   ",    // Kept original
		AccessKey:    "   my<script>key   ",      // Kept original
		IgnoredField: "   ignored <script>alert(1)</script>   ",
		Nested: NestedTest{
			Details: "nested",
		},
		NestedPtr: &NestedTest{
			Details: "nested ptr",
		},
		SliceOfStrings: []string{
			"slice element",
		},
		MapOfStrings: map[string]string{
			"key1": "map value",
		},
		MapOfInterfaces: map[string]interface{}{
			"key2": "interface value",
		},
	}

	if !reflect.DeepEqual(input, expected) {
		t.Errorf("SanitizeStruct result mismatch.\nActual: %+v\nExpected: %+v", input, expected)
	}
}
