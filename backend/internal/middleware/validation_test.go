package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// helper: build a gin context with the given JSON body and the middleware under test.
func performRequest(body string, mw gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	mw(c)
	return w
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()
	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

// ---------------------------------------------------------------------------
// ValidateJSON generic tests
// ---------------------------------------------------------------------------

func TestValidateJSON_ValidBody(t *testing.T) {
	mw := ValidateJSON(map[string]FieldRule{
		"name": {Required: true},
	})
	called := false
	router := gin.New()
	router.POST("/", mw, func(c *gin.Context) { called = true })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"Alice"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, c.Request)

	if !called {
		t.Error("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestValidateJSON_MissingRequired(t *testing.T) {
	mw := ValidateJSON(map[string]FieldRule{
		"name": {Required: true},
	})
	w := performRequest(`{}`, mw)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	resp := decodeError(t, w)
	if resp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Code)
	}
}

func TestValidateJSON_InvalidJSON(t *testing.T) {
	mw := ValidateJSON(map[string]FieldRule{
		"name": {Required: true},
	})
	w := performRequest(`not json`, mw)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	resp := decodeError(t, w)
	if resp.Code != "INVALID_JSON" {
		t.Errorf("expected INVALID_JSON, got %s", resp.Code)
	}
}

func TestValidateJSON_EmptyBodyWithRequiredField(t *testing.T) {
	mw := ValidateJSON(map[string]FieldRule{
		"email": {Required: true},
	})
	w := performRequest(``, mw)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateJSON_EmptyBodyNoRequiredFields(t *testing.T) {
	mw := ValidateJSON(map[string]FieldRule{
		"tag": {MaxLen: 10},
	})
	called := false
	router := gin.New()
	router.POST("/", mw, func(c *gin.Context) { called = true })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(``))
	c.Request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, c.Request)

	if !called {
		t.Error("expected next handler to be called")
	}
}

// ---------------------------------------------------------------------------
// ValidateFields unit tests
// ---------------------------------------------------------------------------

func TestValidateFields_Email(t *testing.T) {
	rules := map[string]FieldRule{"email": {Required: true, Email: true}}
	tests := []struct {
		name string
		data map[string]interface{}
		want int // expected number of errors
	}{
		{"valid", map[string]interface{}{"email": "a@b.com"}, 0},
		{"missing", map[string]interface{}{}, 1},
		{"invalid no at", map[string]interface{}{"email": "ab.com"}, 1},
		{"invalid no tld", map[string]interface{}{"email": "a@"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateFields(tt.data, rules)
			if len(errs) != tt.want {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.want, errs)
			}
		})
	}
}

func TestValidateFields_MinLen(t *testing.T) {
	rules := map[string]FieldRule{"password": {Required: true, MinLen: 8}}
	errs := ValidateFields(map[string]interface{}{"password": "short"}, rules)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	errs = ValidateFields(map[string]interface{}{"password": "longenough"}, rules)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestValidateFields_MaxLen(t *testing.T) {
	rules := map[string]FieldRule{"message": {Required: true, MaxLen: 10}}
	errs := ValidateFields(map[string]interface{}{"message": "hello"}, rules)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
	errs = ValidateFields(map[string]interface{}{"message": "this is way too long message"}, rules)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateFields_TypeNumber(t *testing.T) {
	rules := map[string]FieldRule{"score": {Required: true, Type: "number"}}
	errs := ValidateFields(map[string]interface{}{"score": float64(3)}, rules)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
	errs = ValidateFields(map[string]interface{}{"score": "three"}, rules)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateFields_TypeString(t *testing.T) {
	rules := map[string]FieldRule{"name": {Required: true, Type: "string"}}
	errs := ValidateFields(map[string]interface{}{"name": "Alice"}, rules)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
	errs = ValidateFields(map[string]interface{}{"name": 123}, rules)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateFields_MultipleErrors(t *testing.T) {
	rules := map[string]FieldRule{
		"email":    {Required: true, Email: true},
		"password": {Required: true, MinLen: 8},
	}
	errs := ValidateFields(map[string]interface{}{}, rules)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// Endpoint-specific middleware tests
// ---------------------------------------------------------------------------

func TestValidateRegister_Success(t *testing.T) {
	w := performRequest(`{"email":"test@example.com","password":"password1","first_name":"A","last_name":"B"}`, ValidateRegister())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestValidateRegister_BadEmail(t *testing.T) {
	w := performRequest(`{"email":"notanemail","password":"password1","first_name":"A","last_name":"B"}`, ValidateRegister())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	resp := decodeError(t, w)
	if resp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Code)
	}
}

func TestValidateRegister_ShortPassword(t *testing.T) {
	w := performRequest(`{"email":"a@b.com","password":"short","first_name":"A","last_name":"B"}`, ValidateRegister())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateRegister_MissingFields(t *testing.T) {
	w := performRequest(`{}`, ValidateRegister())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateLogin_Success(t *testing.T) {
	w := performRequest(`{"email":"a@b.com","password":"pass"}`, ValidateLogin())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestValidateLogin_MissingEmail(t *testing.T) {
	w := performRequest(`{"password":"pass"}`, ValidateLogin())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateCreateQAPair_Success(t *testing.T) {
	w := performRequest(`{"category_id":"cat1","question":"q","answer":"a"}`, ValidateCreateQAPair())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestValidateCreateQAPair_MissingQuestion(t *testing.T) {
	w := performRequest(`{"category_id":"c","answer":"a"}`, ValidateCreateQAPair())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateDirectChat_Success(t *testing.T) {
	w := performRequest(`{"message":"hello","channel":"web"}`, ValidateDirectChat())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestValidateDirectChat_MissingMessage(t *testing.T) {
	w := performRequest(`{"channel":"web"}`, ValidateDirectChat())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateDirectChat_MessageTooLong(t *testing.T) {
	long := make([]byte, 10001)
	for i := range long {
		long[i] = 'a'
	}
	body, _ := json.Marshal(map[string]string{"message": string(long), "channel": "web"})
	w := performRequest(string(body), ValidateDirectChat())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateSendMessage_Success(t *testing.T) {
	w := performRequest(`{"content":"hi"}`, ValidateSendMessage())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestValidateSendMessage_EmptyContent(t *testing.T) {
	w := performRequest(`{"content":""}`, ValidateSendMessage())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (empty string is still a value), got %d", w.Code)
	}
}

func TestValidateSendMessage_MissingContent(t *testing.T) {
	w := performRequest(`{}`, ValidateSendMessage())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// ErrorResponse structure test
// ---------------------------------------------------------------------------

func TestErrorResponseStructure(t *testing.T) {
	w := performRequest(`{}`, ValidateLogin())
	resp := decodeError(t, w)
	if resp.Code == "" || resp.Message == "" {
		t.Error("expected non-empty code and message in ErrorResponse")
	}
}
