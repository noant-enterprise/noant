package utils

import (
	"strings"
	"testing"
)

func BenchmarkSanitizeString(b *testing.B) {
	input := "   hello world with spaces and \x00 null bytes \x01\x02   "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeString(input)
	}
}

func BenchmarkSanitizeString_Clean(b *testing.B) {
	input := "clean input"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeString(input)
	}
}

func BenchmarkSanitizeXSS(b *testing.B) {
	input := "hello <script>alert(1)</script> <img src=x onerror=alert(1)> world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeXSS(input)
	}
}

func BenchmarkSanitizeXSS_NoXSS(b *testing.B) {
	input := "just a normal string with no HTML tags at all"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeXSS(input)
	}
}

func BenchmarkSanitizeName(b *testing.B) {
	input := strings.Repeat("a", 300) + "\n\t"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeName(input)
	}
}

func BenchmarkHTMLEscape(b *testing.B) {
	input := `<script>alert("xss")</script>&"hello"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HTMLEscape(input)
	}
}

func BenchmarkHTMLEscape_NoSpecialChars(b *testing.B) {
	input := "plain text without special characters"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HTMLEscape(input)
	}
}

func BenchmarkIsValidEmail_Valid(b *testing.B) {
	email := "user@example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidEmail(email)
	}
}

func BenchmarkIsValidEmail_Invalid(b *testing.B) {
	email := "not-an-email"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidEmail(email)
	}
}

func BenchmarkIsValidPhone(b *testing.B) {
	phone := "+12345678901"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidPhone(phone)
	}
}

func BenchmarkIsValidUUID(b *testing.B) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidUUID(id)
	}
}

func BenchmarkIsValidUUID_Invalid(b *testing.B) {
	id := "not-a-uuid"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidUUID(id)
	}
}

func BenchmarkSanitizeStruct(b *testing.B) {
	type nested struct {
		Value string
	}
	type testStruct struct {
		Name    string
		Email   string
		Detail  nested
		Tags    []string
		Data    map[string]string
	}

	input := testStruct{
		Name:   "   <script>alert(1)</script>John Doe   ",
		Email:  "   john@example.com\x00   ",
		Detail: nested{Value: "   <b>hello</b>   "},
		Tags:   []string{"   <script>x</script>tag1   "},
		Data:   map[string]string{"k": "   <img onerror=alert(1)>val   "},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cp := input
		SanitizeStruct(&cp)
	}
}

var testKey = "0123456789abcdef0123456789abcdef"

func BenchmarkEncrypter_Encrypt(b *testing.B) {
	enc, err := NewEncrypter(testKey)
	if err != nil {
		b.Fatal(err)
	}
	plaintext := "this is a secret message that needs encryption"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Encrypt(plaintext)
	}
}

func BenchmarkEncrypter_Decrypt(b *testing.B) {
	enc, err := NewEncrypter(testKey)
	if err != nil {
		b.Fatal(err)
	}
	ciphertext, _ := enc.Encrypt("this is a secret message that needs encryption")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Decrypt(ciphertext)
	}
}

func BenchmarkEncrypter_Roundtrip(b *testing.B) {
	enc, err := NewEncrypter(testKey)
	if err != nil {
		b.Fatal(err)
	}
	plaintext := "roundtrip benchmark payload"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ct, _ := enc.Encrypt(plaintext)
		enc.Decrypt(ct)
	}
}
