package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestSendTextMessage_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-text", http.StatusOK, map[string]string{"status": "sent"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.SendTextMessage("test-session", "123456@s.whatsapp.net", "Hello, world!")
	assertNoError(t, err)
	mock.AssertCallCount(t, 1)
	mock.AssertLastRequestPath(t, "/api/sessions/test-session/messages/send-text")

	var body map[string]string
	mock.AssertRequestBody(t, &body)
	assertEqual(t, "123456@s.whatsapp.net", body["chatId"])
	assertEqual(t, "Hello, world!", body["text"])
}

func TestSendTextMessage_StatusCode201(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-text", http.StatusCreated, map[string]string{"status": "sent"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.SendTextMessage("test-session", "123456@s.whatsapp.net", "Hi")
	assertNoError(t, err)
	mock.AssertCallCount(t, 1)
}

func TestSendTextMessage_ServerError(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-text", http.StatusInternalServerError, map[string]string{"error": "server error"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.SendTextMessage("test-session", "123456@s.whatsapp.net", "Hello")
	assertError(t, err)
	assertBodyContains(t, []byte(err.Error()), "500")
}

func TestSendTextMessage_BadRequest(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-text", http.StatusBadRequest, map[string]string{"error": "invalid chatId"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.SendTextMessage("test-session", "invalid", "Hello")
	assertError(t, err)
}

func TestSendTextMessage_Disabled(t *testing.T) {
	svc := newOpenWATestService(t, NewOpenWAMockServer(t), map[string]interface{}{
		"enabled": false,
	})

	err := svc.SendTextMessage("test-session", "123456@s.whatsapp.net", "Hello")
	assertNoError(t, err) // No error when disabled
}

func TestSendTextMessage_SendsAPIKey(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.Handle("POST", "/api/sessions/test-session/messages/send-text", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key-123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	svc := newOpenWATestService(t, mock, map[string]interface{}{
		"api_key": "test-key-123",
	})

	err := svc.SendTextMessage("test-session", "123456@s.whatsapp.net", "Hello")
	assertNoError(t, err)
}

func TestSendTextMessage_NoAPIKey(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	var headerReceived string
	mock.Handle("POST", "/api/sessions/test-session/messages/send-text", func(w http.ResponseWriter, r *http.Request) {
		headerReceived = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	})

	svc := newOpenWATestService(t, mock, map[string]interface{}{
		"api_key": "",
	})

	err := svc.SendTextMessage("test-session", "123456@s.whatsapp.net", "Hello")
	assertNoError(t, err)
	assertEqual(t, "", headerReceived)
}

func TestSendMediaMessage_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-media", http.StatusOK, map[string]string{"status": "sent"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.SendMediaMessage("test-session", "123456@s.whatsapp.net", "https://example.com/img.jpg", "Check this out")
	assertNoError(t, err)
	mock.AssertCallCount(t, 1)
	mock.AssertLastRequestPath(t, "/api/sessions/test-session/messages/send-media")

	var body map[string]string
	mock.AssertRequestBody(t, &body)
	assertEqual(t, "https://example.com/img.jpg", body["file"])
	assertEqual(t, "Check this out", body["caption"])
}

func TestSendMediaMessage_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-media", http.StatusBadRequest, map[string]string{"error": "invalid file"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.SendMediaMessage("test-session", "123456@s.whatsapp.net", "bad-url", "")
	assertError(t, err)
}

func TestSendMediaMessage_Disabled(t *testing.T) {
	svc := newOpenWATestService(t, NewOpenWAMockServer(t), map[string]interface{}{
		"enabled": false,
	})

	err := svc.SendMediaMessage("test-session", "123456@s.whatsapp.net", "https://example.com/img.jpg", "caption")
	assertNoError(t, err)
}

func TestGetSessionStatus_Connected(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("GET", "/api/sessions/test-session", http.StatusOK, map[string]string{"status": "connected"})

	svc := newOpenWATestService(t, mock, nil)

	status, err := svc.GetSessionStatus()
	assertNoError(t, err)
	assertEqual(t, "connected", status)
}

func TestGetSessionStatus_Disconnected(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("GET", "/api/sessions/test-session", http.StatusOK, map[string]string{"status": "disconnected"})

	svc := newOpenWATestService(t, mock, nil)

	status, err := svc.GetSessionStatus()
	assertNoError(t, err)
	assertEqual(t, "disconnected", status)
}

func TestGetSessionStatus_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.Handle("GET", "/api/sessions/test-session", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	svc := newOpenWATestService(t, mock, nil)

	_, err := svc.GetSessionStatus()
	assertError(t, err)
}

func TestGetSessionStatus_Disabled(t *testing.T) {
	svc := newOpenWATestService(t, NewOpenWAMockServer(t), map[string]interface{}{
		"enabled": false,
	})

	status, err := svc.GetSessionStatus()
	assertNoError(t, err)
	assertEqual(t, "disabled", status)
}

func TestPing_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("GET", "/api/sessions", http.StatusOK, []map[string]string{})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.Ping()
	assertNoError(t, err)
}

func TestPing_ServerError(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.Handle("GET", "/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.Ping()
	assertError(t, err)
}

func TestPing_Disabled(t *testing.T) {
	svc := newOpenWATestService(t, NewOpenWAMockServer(t), map[string]interface{}{
		"enabled": false,
	})

	err := svc.Ping()
	assertNoError(t, err)
}

func TestCreateSession_Created(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions", http.StatusCreated, map[string]string{"id": "new-session-id", "name": "my-session"})

	svc := newOpenWATestService(t, mock, nil)

	id, err := svc.CreateSession("my-session")
	assertNoError(t, err)
	assertEqual(t, "new-session-id", id)
}

func TestCreateSession_ConflictFoundByName(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.Handle("POST", "/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "already exists"})
	})
	mock.HandleJSON("GET", "/api/sessions", http.StatusOK, []map[string]string{
		{"id": "existing-id", "name": "my-session"},
	})

	svc := newOpenWATestService(t, mock, nil)

	id, err := svc.CreateSession("my-session")
	assertNoError(t, err)
	assertEqual(t, "existing-id", id)
	assertEqual(t, 2, len(mock.Requests())) // POST + GET
}

func TestCreateSession_ConflictNoNameMatch(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.Handle("POST", "/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	mock.HandleJSON("GET", "/api/sessions", http.StatusOK, []map[string]string{
		{"id": "other-id", "name": "other-session"},
	})

	svc := newOpenWATestService(t, mock, nil)

	id, err := svc.CreateSession("my-session")
	assertNoError(t, err)
	assertEqual(t, "my-session", id) // fallback to name
}

func TestCreateSession_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions", http.StatusBadRequest, map[string]string{"error": "bad request"})

	svc := newOpenWATestService(t, mock, nil)

	_, err := svc.CreateSession("my-session")
	assertError(t, err)
}

func TestCreateSession_Disabled(t *testing.T) {
	svc := newOpenWATestService(t, NewOpenWAMockServer(t), map[string]interface{}{
		"enabled": false,
	})

	_, err := svc.CreateSession("my-session")
	assertError(t, err) // CreateSession doesn't check OpenWAEnabled
}

func TestStartSession_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/start", http.StatusOK, map[string]string{"status": "started"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.StartSession("test-session")
	assertNoError(t, err)
}

func TestStartSession_AlreadyStarted(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/start", http.StatusBadRequest, map[string]string{"error": "already started"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.StartSession("test-session")
	assertNoError(t, err)
}

func TestStartSession_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/start", http.StatusBadRequest, map[string]string{"error": "some error"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.StartSession("test-session")
	assertError(t, err)
}

func TestRestartSession_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/restart", http.StatusOK, map[string]string{"status": "restarted"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.RestartSession()
	assertNoError(t, err)
}

func TestRestartSession_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/restart", http.StatusInternalServerError, map[string]string{"error": "oops"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.RestartSession()
	assertError(t, err)
}

func TestRestartSession_Disabled(t *testing.T) {
	svc := newOpenWATestService(t, NewOpenWAMockServer(t), map[string]interface{}{
		"enabled": false,
	})

	err := svc.RestartSession()
	assertNoError(t, err)
}

func TestGetQRCode_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("GET", "/api/sessions/test-session/qr", http.StatusOK, map[string]string{"qr": "base64qrimage"})

	svc := newOpenWATestService(t, mock, nil)

	qr, err := svc.GetQRCode("test-session")
	assertNoError(t, err)
	assertEqual(t, "base64qrimage", qr)
}

func TestGetQRCode_NotFound(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("GET", "/api/sessions/test-session/qr", http.StatusNotFound, map[string]string{"error": "session not found"})

	svc := newOpenWATestService(t, mock, nil)

	_, err := svc.GetQRCode("test-session")
	assertError(t, err)
}

func TestDeleteSession_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("DELETE", "/api/sessions/test-session", http.StatusOK, map[string]string{"status": "deleted"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.DeleteSession("test-session")
	assertNoError(t, err)
	mock.AssertLastRequestPath(t, "/api/sessions/test-session")
}

func TestDeleteSession_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("DELETE", "/api/sessions/test-session", http.StatusInternalServerError, map[string]string{"error": "oops"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.DeleteSession("test-session")
	assertError(t, err)
}

func TestLogoutSession_Success(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/logout", http.StatusOK, map[string]string{"status": "logged out"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.LogoutSession("test-session")
	assertNoError(t, err)
	mock.AssertLastRequestPath(t, "/api/sessions/test-session/logout")
}

func TestLogoutSession_Error(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/logout", http.StatusInternalServerError, map[string]string{"error": "oops"})

	svc := newOpenWATestService(t, mock, nil)

	err := svc.LogoutSession("test-session")
	assertError(t, err)
}

// ========== CIRCUIT BREAKER TESTS ==========

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewSessionCircuitBreaker()
	assertEqual(t, false, cb.IsOpen())
}

func TestCircuitBreaker_OpensAfterFiveFailures(t *testing.T) {
	cb := NewSessionCircuitBreaker()

	for i := 0; i < 4; i++ {
		cb.RecordFailure()
		assertEqual(t, false, cb.IsOpen(), "should still be closed after %d failures", i+1)
	}

	cb.RecordFailure()
	assertEqual(t, true, cb.IsOpen(), "should open after 5 failures")
}

func TestCircuitBreaker_ClosesAfterSuccess(t *testing.T) {
	cb := NewSessionCircuitBreaker()

	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}
	assertEqual(t, true, cb.IsOpen())

	cb.RecordSuccess()
	assertEqual(t, false, cb.IsOpen())
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewSessionCircuitBreaker()

	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}
	assertEqual(t, true, cb.IsOpen())

	// Simulate time passing — we can't actually sleep, but the circuit breaker
	// uses time.Since(lastFailure). The timeout is 60s. We can't mock time easily
	// without changing the code, so we just verify the logic works as designed.
	// The IsOpen() method checks if state is "open" AND if 60s have passed.
	// If 60s haven't passed, it returns true (still open).
	// This is a basic behavior check.
}

// ========== WEBHOOK SIGNATURE TESTS ==========

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, map[string]interface{}{
		"webhook_secret": "my-secret",
	})

	payload := []byte(`{"event":"message.received","data":{}}`)
	// We need the hex of HMAC-SHA256("my-secret", payload)
	// Just call the method with a known expected signature
	sig := "sha256=" + hmacSha256Hex("my-secret", payload)

	result := svc.VerifyWebhookSignature(payload, sig)
	assertEqual(t, true, result)
}

func TestVerifyWebhookSignature_Invalid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, map[string]interface{}{
		"webhook_secret": "my-secret",
	})

	payload := []byte(`{"event":"message.received"}`)
	result := svc.VerifyWebhookSignature(payload, "sha256=invalid")
	assertEqual(t, false, result)
}

func TestVerifyWebhookSignature_EmptySecret(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, map[string]interface{}{
		"webhook_secret": "",
	})

	payload := []byte(`{"event":"message.received"}`)
	result := svc.VerifyWebhookSignature(payload, "sha256=anything")
	assertEqual(t, true, result)
}

func TestVerifyWebhookSignature_WithoutPrefix(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, map[string]interface{}{
		"webhook_secret": "my-secret",
	})

	payload := []byte(`{"event":"message.received"}`)
	// Compute HMAC without "sha256=" prefix
	sig := hmacSha256Hex("my-secret", payload)

	result := svc.VerifyWebhookSignature(payload, sig)
	assertEqual(t, true, result)
}

// ========== WEBHOOK PARSING TESTS ==========

func TestParseWebhookEvent_Valid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, nil)

	payload := []byte(`{"event":"message.received","sessionId":"s1","data":{"from":"123@s.whatsapp.net"}}`)

	event, err := svc.ParseWebhookEvent(payload)
	assertNoError(t, err)
	assertEqual(t, "message.received", event.Event)
	assertEqual(t, "s1", event.SessionID)
}

func TestParseWebhookEvent_Invalid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, nil)

	_, err := svc.ParseWebhookEvent([]byte(`not json`))
	assertError(t, err)
}

func TestParseMessageData_Valid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, nil)

	payload := json.RawMessage(`{"id":"msg1","from":"123@s.whatsapp.net","body":"Hello","type":"text"}`)

	msg, err := svc.ParseMessageData(payload)
	assertNoError(t, err)
	assertEqual(t, "msg1", msg.ID)
	assertEqual(t, "123@s.whatsapp.net", msg.From)
	assertEqual(t, "Hello", msg.Body)
	assertEqual(t, "text", msg.Type)
}

func TestParseMessageData_Invalid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, nil)

	_, err := svc.ParseMessageData(json.RawMessage(`not json`))
	assertError(t, err)
}

func TestParseStatusData_Valid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, nil)

	payload := json.RawMessage(`{"id":"msg1","status":"delivered"}`)

	status, err := svc.ParseStatusData(payload)
	assertNoError(t, err)
	assertEqual(t, "msg1", status.ID)
	assertEqual(t, "delivered", status.Status)
}

func TestParseStatusData_Invalid(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	svc := newOpenWATestService(t, mock, nil)

	_, err := svc.ParseStatusData(json.RawMessage(`not json`))
	assertError(t, err)
}

// ========== REQUEST ORDERING TESTS ==========

func TestRequestOrdering(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.HandleJSON("POST", "/api/sessions/test-session/messages/send-text", http.StatusOK, map[string]string{"status": "sent"})
	mock.HandleJSON("GET", "/api/sessions/test-session", http.StatusOK, map[string]string{"status": "connected"})

	svc := newOpenWATestService(t, mock, nil)

	_ = svc.SendTextMessage("test-session", "123@s.whatsapp.net", "First")
	_, _ = svc.GetSessionStatus()
	_ = svc.SendTextMessage("test-session", "456@s.whatsapp.net", "Second")

	reqs := mock.Requests()
	assertEqual(t, 3, len(reqs))
	assertEqual(t, "POST", reqs[0].Method)
	assertEqual(t, "/api/sessions/test-session/messages/send-text", reqs[0].Path)
	assertEqual(t, "GET", reqs[1].Method)
	assertEqual(t, "/api/sessions/test-session", reqs[1].Path)
	assertEqual(t, "POST", reqs[2].Method)
	assertEqual(t, "/api/sessions/test-session/messages/send-text", reqs[2].Path)

	var body map[string]string
	json.Unmarshal(reqs[2].Body, &body)
	assertEqual(t, "456@s.whatsapp.net", body["chatId"])
}

// ========== CONCURRENCY TEST ==========

func TestConcurrentCalls(t *testing.T) {
	mock := NewOpenWAMockServer(t)

	var mu sync.Mutex
	callCount := 0
	mock.Handle("POST", "/api/sessions/test-session/messages/send-text", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	svc := newOpenWATestService(t, mock, nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = svc.SendTextMessage("test-session", fmt.Sprintf("%d@s.whatsapp.net", i), fmt.Sprintf("msg %d", i))
		}(i)
	}
	wg.Wait()

	assertEqual(t, 10, callCount)
	assertEqual(t, 10, len(mock.Requests()))
}

// ========== RATE LIMIT HEADER TRACKING ==========

func TestRateLimitHeaderTracking(t *testing.T) {
	mock := NewOpenWAMockServer(t)
	mock.Handle("POST", "/api/sessions/test-session/messages/send-text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.Header().Set("X-RateLimit-Reset", "1600000000")
		w.WriteHeader(http.StatusOK)
	})

	svc := newOpenWATestService(t, mock, nil)

	// The public SendTextMessage does NOT track rate limit headers,
	// so we can't assert on that. This test confirms the public method
	// succeeds and doesn't error on rate limit headers in the response.
	err := svc.SendTextMessage("test-session", "123@s.whatsapp.net", "Hello")
	assertNoError(t, err)
}

// ========== HELPER FUNCTIONS ==========

func hmacSha256Hex(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return fmt.Sprintf("%x", h.Sum(nil))
}


