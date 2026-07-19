package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type noLenReader struct {
	data []byte
	pos  int
}

func (r *noLenReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *noLenReader) Close() error { return nil }

func TestBodyLimitMiddleware_SkipsSafeMethods(t *testing.T) {
	rr := testBodyLimitRequest(t, "GET", "/health", "", true)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for GET, got %d", rr.Code)
	}
}

func TestBodyLimitMiddleware_RejectsLargeRequest(t *testing.T) {
	body := strings.Repeat("a", MaxRequestBodySize+1)
	rr := testBodyLimitRequest(t, "POST", "/api", body, true)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for oversized request, got %d", rr.Code)
	}
}

func TestBodyLimitMiddleware_AllowsSmallRequest(t *testing.T) {
	body := `{"hello": "world"}`
	rr := testBodyLimitRequest(t, "POST", "/api", body, true)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for small request, got %d", rr.Code)
	}
}

func TestBodyLimitMiddleware_RejectsOversizedStreamingBody(t *testing.T) {
	body := strings.Repeat("a", MaxRequestBodySize+100)
	rr := testBodyLimitStreamRequest(t, "POST", "/api", body)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for oversized streaming body, got %d", rr.Code)
	}
}

func TestBodyLimitMiddleware_AllowsSmallStreamingBody(t *testing.T) {
	body := strings.Repeat("a", 100)
	rr := testBodyLimitStreamRequest(t, "POST", "/api", body)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for small streaming body, got %d", rr.Code)
	}
}

func testBodyLimitRequest(t *testing.T, method, path, body string, setContentLength bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimitMiddleware())
	r.Any("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.Any("/api", func(c *gin.Context) {
		data, _ := io.ReadAll(c.Request.Body)
		_ = data
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if setContentLength && body != "" {
		req.ContentLength = int64(len(body))
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func testBodyLimitStreamRequest(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimitMiddleware())
	r.Any("/api", func(c *gin.Context) {
		data, _ := io.ReadAll(c.Request.Body)
		_ = data
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, &noLenReader{data: []byte(body)})
	// ContentLength defaults to -1 (unknown), simulating chunked encoding

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}
