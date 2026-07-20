package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"strings"
)

type sessionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Ping checks if OpenWA server is reachable
func (s *OpenWAService) Ping() error {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("OpenWA server error: %d", resp.StatusCode)
	}

	return nil
}

// findSessionByName lists all sessions and finds the one matching the name
func (s *OpenWAService) findSessionByName(name string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return "", err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var sessions []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return "", err
	}

	for _, sess := range sessions {
		if sess.Name == name {
			return sess.ID, nil
		}
	}
	return "", fmt.Errorf("session not found by name: %s", name)
}

// CreateSession creates a new OpenWA session
func (s *OpenWAService) CreateSession(sessionName string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	payload := map[string]string{"name": sessionName}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &result)

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		if result.ID != "" {
			return result.ID, nil
		}
		return sessionName, nil
	}

	if resp.StatusCode == http.StatusConflict {
		s.logger.Info("Session already exists, finding ID", "name", sessionName)
		id, err := s.findSessionByName(sessionName)
		if err == nil && id != "" {
			return id, nil
		}
		return sessionName, nil
	}

	return "", fmt.Errorf("create session failed: %d %s", resp.StatusCode, string(body))
}

// StartSession starts an OpenWA session (generates QR code)
func (s *OpenWAService) StartSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/start", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("POST", url, http.NoBody)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusBadRequest && bytes.Contains(bytes.ToLower(body), []byte("already started")) {
			s.logger.Info("OpenWA session already started", "sessionID", sessionID)
			return nil
		}
		return fmt.Errorf("start session failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetQRCode retrieves the QR code for a session
func (s *OpenWAService) GetQRCode(sessionID string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions/%s/qr", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return "", err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get QR: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	s.logger.Info("OpenWA QR response", "status", resp.StatusCode, "body", string(body)[:min(500, len(body))])

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			return "", nil
		}
		return "", fmt.Errorf("get QR failed: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		QR     string `json:"qr"`
		Image  string `json:"image"`
		QRCode string `json:"qrCode"`
		Data   string `json:"data"`
		Base64 string `json:"base64"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		trimmed := string(body)
		if len(trimmed) > 100 {
			return trimmed, nil
		}
		return "", fmt.Errorf("failed to parse QR response: %w", err)
	}

	var rawQR string
	switch {
	case result.Image != "":
		rawQR = result.Image
	case result.QR != "":
		rawQR = result.QR
	case result.QRCode != "":
		rawQR = result.QRCode
	case result.Data != "":
		rawQR = result.Data
	case result.Base64 != "":
		rawQR = result.Base64
	}

	if rawQR == "" {
		return "", nil
	}

	return s.OverlayLogo(rawQR), nil
}

// OverlayLogo adds a logo to the center of a QR code image
func (s *OpenWAService) OverlayLogo(qrBase64 string) string {
	qrData := qrBase64
	if len(qrData) > 22 && qrData[:22] == "data:image/png;base64," {
		qrData = qrData[22:]
	}

	qrBytes, err := base64.StdEncoding.DecodeString(qrData)
	if err != nil {
		return qrBase64
	}

	qrImg, _, err := image.Decode(bytes.NewReader(qrBytes))
	if err != nil {
		return qrBase64
	}

	qrBounds := qrImg.Bounds()
	qrWidth := qrBounds.Dx()
	logoSize := qrWidth / 4

	output := image.NewRGBA(qrBounds)
	draw.Draw(output, qrBounds, qrImg, qrBounds.Min, draw.Src)

	centerX := qrWidth / 2
	centerY := qrWidth / 2

	white := color.RGBA{255, 255, 255, 255}
	radius := logoSize / 2
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				px := centerX + x
				py := centerY + y
				if px >= qrBounds.Min.X && px < qrBounds.Max.X && py >= qrBounds.Min.Y && py < qrBounds.Max.Y {
					output.Set(px, py, white)
				}
			}
		}
	}

	black := color.RGBA{0, 0, 0, 255}
	innerRadius := logoSize / 3
	for y := -innerRadius; y <= innerRadius; y++ {
		for x := -innerRadius; x <= innerRadius; x++ {
			if x*x+y*y <= innerRadius*innerRadius {
				px := centerX + x
				py := centerY + y
				if px >= qrBounds.Min.X && px < qrBounds.Max.X && py >= qrBounds.Min.Y && py < qrBounds.Max.Y {
					output.Set(px, py, black)
				}
			}
		}
	}

	dotColor := color.RGBA{255, 255, 255, 255}
	dotSizes := []int{innerRadius / 5, innerRadius / 4, innerRadius / 3}
	dotOffsets := []int{-innerRadius / 3, 0, innerRadius / 3}

	for i, dotR := range dotSizes {
		dotX := centerX + dotOffsets[i]
		dotY := centerY
		for y := -dotR; y <= dotR; y++ {
			for x := -dotR; x <= dotR; x++ {
				if x*x+y*y <= dotR*dotR {
					px := dotX + x
					py := dotY + y
					if px >= qrBounds.Min.X && px < qrBounds.Max.X && py >= qrBounds.Min.Y && py < qrBounds.Max.Y {
						output.Set(px, py, dotColor)
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, output)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// GetSessionStatus checks if the WhatsApp session is connected
func (s *OpenWAService) GetSessionStatus() (string, error) {
	if !s.cfg.OpenWAEnabled {
		return "disabled", nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s",
		s.cfg.OpenWABaseURL, s.cfg.OpenWASessionID)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return "", err
	}

	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "disconnected", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "unknown", err
	}

	return result.Status, nil
}

// GetSessionStatusByID gets the status of a specific session
func (s *OpenWAService) GetSessionStatusByID(sessionID string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions/%s", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return "", err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "disconnected", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return "expired", nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "unknown", err
	}

	normalized := normalizeSessionStatus(result.Status)
	return normalized, nil
}

// normalizeSessionStatus maps all variants of a connected session status to "connected"
func normalizeSessionStatus(status string) string {
	lower := strings.ToLower(strings.TrimSpace(status))
	switch lower {
	case "connected", "authenticated", "ready", "open":
		return "connected"
	case "qr_read":
		return "connected"
	case "qr_ready", "scan_qr_code", "waitforlogin":
		return "qr_ready"
	case "starting", "initializing", "connecting":
		return "initializing"
	case "failed", "timeout", "error":
		return "failed"
	case "disconnected", "stopped", "inactive":
		return "disconnected"
	case "expired":
		return "expired"
	default:
		if status == "" {
			return "unknown"
		}
		return status
	}
}

// RestartSession reconnects a disconnected WhatsApp session
func (s *OpenWAService) RestartSession() error {
	if !s.cfg.OpenWAEnabled {
		return nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s/restart",
		s.cfg.OpenWABaseURL, s.cfg.OpenWASessionID)

	req, err := http.NewRequest("POST", url, http.NoBody)
	if err != nil {
		return err
	}

	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa restart failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// RestartSessionByID restarts a specific session
func (s *OpenWAService) RestartSessionByID(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/restart", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("POST", url, http.NoBody)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("restart failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteSession deletes an OpenWA session
func (s *OpenWAService) DeleteSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("DELETE", url, http.NoBody)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete session failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// LogoutSession logs out the active WhatsApp session to clear credentials
func (s *OpenWAService) LogoutSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/logout", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("POST", url, http.NoBody)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout session failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteAllSessions removes all existing sessions
func (s *OpenWAService) DeleteAllSessions() error {
	sessions, err := s.ListSessions()
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		if len(sess.Name) > 6 && sess.Name[:6] == "noant-" {
			if err := s.DeleteSession(sess.ID); err != nil {
				s.logger.Warn("Failed to delete session", "id", sess.ID, "error", err)
			}
		}
	}
	return nil
}

// ListSessions returns all sessions (exported)
func (s *OpenWAService) ListSessions() ([]sessionInfo, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var sessions []sessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}
