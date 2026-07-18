package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"noant/config"
	"noant/internal/infrastructure"
)

// OpenWARequest records a single request received by the mock server.
type OpenWARequest struct {
	Method string
	Path   string
	Body   []byte
	Header http.Header
}

// OpenWAMockServer is a pre-configured mock OpenWA HTTP API.
// It records all requests in order and lets you stub responses per route.
type OpenWAMockServer struct {
	t      *testing.T
	srv    *httptest.Server
	mu     sync.Mutex
	routes map[string]func(w http.ResponseWriter, r *http.Request)
	reqs   []OpenWARequest
}

// NewOpenWAMockServer creates a mock server that returns 404 for any
// unregistered route. Use Handle to register route handlers.
func NewOpenWAMockServer(t *testing.T) *OpenWAMockServer {
	t.Helper()
	m := &OpenWAMockServer{
		t:      t,
		routes: make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
	m.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.reqs = append(m.reqs, OpenWARequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   body,
			Header: r.Header.Clone(),
		})
		m.mu.Unlock()

		key := r.Method + " " + r.URL.Path
		handler, ok := m.routes[key]
		if !ok {
			// Try more specific match first, then fallback to prefix pattern
			handler, ok = m.routes["*"]
			if !ok {
				http.NotFound(w, r)
				return
			}
		}
		handler(w, r)
	}))
	t.Cleanup(m.srv.Close)
	return m
}

// Handle registers a response handler for a specific method+path.
// Path is the exact path (e.g. "/api/sessions/foo/messages/send-text").
// If the code changes the path, the handler won't match and the test will
// get a 404, making path changes visible.
func (m *OpenWAMockServer) Handle(method, path string, handler func(w http.ResponseWriter, r *http.Request)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes[method+" "+path] = handler
}

// HandleFunc is a convenience that sets a JSON 200 response.
func (m *OpenWAMockServer) HandleJSON(method, path string, statusCode int, response interface{}) {
	m.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	})
}

// URL returns the base URL of the mock server.
func (m *OpenWAMockServer) URL() string { return m.srv.URL }

// Requests returns a copy of all received requests (in order).
func (m *OpenWAMockServer) Requests() []OpenWARequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]OpenWARequest, len(m.reqs))
	copy(out, m.reqs)
	return out
}

// LastRequest returns the last received request, or nil.
func (m *OpenWAMockServer) LastRequest() *OpenWARequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.reqs) == 0 {
		return nil
	}
	r := m.reqs[len(m.reqs)-1]
	return &r
}

// AssertCallCount fails if the total request count doesn't match.
func (m *OpenWAMockServer) AssertCallCount(t *testing.T, expected int) {
	t.Helper()
	if got := len(m.reqs); got != expected {
		t.Errorf("expected %d OpenWA calls, got %d", expected, got)
	}
}

// AssertLastRequestPath fails if the last request's path doesn't match.
func (m *OpenWAMockServer) AssertLastRequestPath(t *testing.T, expectedPath string) {
	t.Helper()
	last := m.LastRequest()
	if last == nil {
		t.Fatalf("no requests received, expected path %q", expectedPath)
	}
	if last.Path != expectedPath {
		t.Errorf("expected last request path %q, got %q", expectedPath, last.Path)
	}
}

// AssertRequestBody decodes the last request's body into dest and fails
// if decoding fails.
func (m *OpenWAMockServer) AssertRequestBody(t *testing.T, dest interface{}) {
	t.Helper()
	last := m.LastRequest()
	if last == nil {
		t.Fatalf("no requests received")
	}
	if err := json.Unmarshal(last.Body, dest); err != nil {
		t.Fatalf("failed to decode last request body: %v\nbody: %s", err, string(last.Body))
	}
}

// Common response factories

// OpenWASuccess returns a handler that responds with a simple success JSON body.
func OpenWASuccess(msg string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": msg})
	}
}

// OpenAIError returns a handler that responds with the given status code and error body.
func OpenAIError(statusCode int, msg string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
	}
}

// newOpenWATestService creates an OpenWAService wired to a mock server.
// Caller must keep the mock alive for the duration of the test.
func newOpenWATestService(t *testing.T, mock *OpenWAMockServer, cfgOverrides map[string]interface{}) *OpenWAService {
	t.Helper()
	cfg := &config.Config{
		OpenWAEnabled:              true,
		OpenWABaseURL:              mock.URL(),
		OpenWAApiKey:               "",
		OpenWASessionID:            "test-session",
		OpenWARateLimitText:        20,
		OpenWARateLimitMedia:       10,
		OpenWARateLimitTemplate:    30,
		OpenWARateLimitBurst:       5,
		OpenWAQueueDepth:           100,
		OpenWAMediaDir:             t.TempDir(),
		OpenWAMediaRetention:       0,
		OpenWASessionHealthInterval: 0,
		OpenWAMaxReconnectAttempts:  1,
		OpenWAConnPoolSize:         1,
		OpenWAConnTimeout:          0,
		OpenWAReqTimeout:           0,
	}

	// Apply overrides
	if v, ok := cfgOverrides["enabled"].(bool); ok {
		cfg.OpenWAEnabled = v
	}
	if v, ok := cfgOverrides["api_key"].(string); ok {
		cfg.OpenWAApiKey = v
	}
	if v, ok := cfgOverrides["session_id"].(string); ok {
		cfg.OpenWASessionID = v
	}
	if v, ok := cfgOverrides["webhook_secret"].(string); ok {
		cfg.OpenWAWebhookSecret = v
	}

	svc := NewOpenWAService(cfg, infrastructure.NewNullLogger())
	// Point the httpClient at the mock server.
	svc.httpClient = mock.srv.Client()
	return svc
}

// assertNoError is a small helper so we don't need testify.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func assertEqual[T comparable](t *testing.T, want, got T, msgAndArgs ...interface{}) {
	t.Helper()
	if want != got {
		msg := fmt.Sprintf("expected %v, got %v", want, got)
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if ok {
				msg = fmt.Sprintf(format, msgAndArgs[1:]...) + ": " + msg
			}
		}
		t.Fatal(msg)
	}
}

func assertBodyContains(t *testing.T, body []byte, substr string) {
	t.Helper()
	if !strings.Contains(string(body), substr) {
		t.Fatalf("expected body to contain %q, got: %s", substr, string(body))
	}
}

// MustMarshalJSON marshals v to JSON, failing the test on error.
func MustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return b
}
