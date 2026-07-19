package service

import (
	"testing"
	"time"
)

func TestValidatePasswordStrength_Valid(t *testing.T) {
	validPasswords := []string{
		"StrongP@ss1",
		"Ab1!defgh",
		"MyP@ssw0rd!",
		"Test123!@#",
		"aB3$fghiJK",
	}

	for _, pw := range validPasswords {
		err := validatePasswordStrength(pw)
		if err != nil {
			t.Errorf("validatePasswordStrength(%q) unexpected error: %v", pw, err)
		}
	}
}

func TestValidatePasswordStrength_AllChecksPass(t *testing.T) {
	// "Ab1!" passes all regex checks (has uppercase, lowercase, digit, special)
	err := validatePasswordStrength("Ab1!")
	if err != nil {
		t.Errorf("Ab1! should pass validation: %v", err)
	}
}

func TestValidatePasswordStrength_NoUppercase(t *testing.T) {
	err := validatePasswordStrength("lowercase1!")
	if err == nil {
		t.Error("expected error for password without uppercase")
	}
}

func TestValidatePasswordStrength_NoLowercase(t *testing.T) {
	err := validatePasswordStrength("UPPERCASE1!")
	if err == nil {
		t.Error("expected error for password without lowercase")
	}
}

func TestValidatePasswordStrength_NoDigit(t *testing.T) {
	err := validatePasswordStrength("NoDigits!abc")
	if err == nil {
		t.Error("expected error for password without digits")
	}
}

func TestValidatePasswordStrength_NoSpecial(t *testing.T) {
	err := validatePasswordStrength("NoSpecial123")
	if err == nil {
		t.Error("expected error for password without special characters")
	}
}

func TestGenerateVerificationCode_Length(t *testing.T) {
	code := generateVerificationCode()
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %d digits: %s", len(code), code)
	}
}

func TestGenerateVerificationCode_DigitsOnly(t *testing.T) {
	code := generateVerificationCode()
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("expected only digits, got %c in %s", c, code)
		}
	}
}

func TestGenerateVerificationCode_Unique(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := generateVerificationCode()
		if codes[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		codes[code] = true
	}
}

func TestParseAIMetadata_AllTags(t *testing.T) {
	content := "[SENTIMENT:positive]\n[LANGUAGE:en]\n[SUGGESTIONS:Show products|What are prices?]\nHello! How can I help?"
	clean, sentiment, language, suggestions := parseAIMetadata(content)
	if sentiment != "positive" {
		t.Errorf("expected positive sentiment, got %s", sentiment)
	}
	if language != "en" {
		t.Errorf("expected en language, got %s", language)
	}
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(suggestions))
	}
	if clean != "Hello! How can I help?" {
		t.Errorf("expected clean text 'Hello! How can I help?', got %q", clean)
	}
}

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	if !cb.Allow() {
		t.Error("closed circuit breaker should allow requests")
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.Allow() {
		t.Error("open circuit breaker should block requests")
	}
}

func TestCircuitBreaker_SuccessResetsToClosed(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	cb.lastFailure = cb.lastFailure.Add(-61 * time.Second)
	cb.Allow()
	cb.RecordSuccess()
	if !cb.Allow() {
		t.Error("circuit breaker should be closed after success in half-open")
	}
}

func TestGetSetFloat64(t *testing.T) {
	m := map[string]interface{}{
		"count": 42,
		"rate":  3.14,
		"name":  "test",
	}

	if v := getInt(m, "count"); v != 42 {
		t.Errorf("getInt expected 42, got %d", v)
	}
	if v := getInt(m, "missing"); v != 0 {
		t.Errorf("getInt expected 0 for missing key, got %d", v)
	}
	if v := getFloat64(m, "rate"); v != 3.14 {
		t.Errorf("getFloat64 expected 3.14, got %f", v)
	}
}
