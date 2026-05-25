package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"time"
)

type ResendService struct {
	apiKey string
	client *http.Client
	from   string
}

func NewResendService(apiKey string) *ResendService {
	return &ResendService{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
		from:   "noreply@noant.com",
	}
}

type resendResponse struct {
	ID string `json:"id"`
}

func (s *ResendService) SendPasswordReset(ctx context.Context, toEmail, resetToken string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("resend API key not configured")
	}

	// Use localhost for development, production domain otherwise
	baseURL := "http://localhost:3000"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, url.QueryEscape(resetToken))
	safeURL := html.EscapeString(resetURL)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Reset your NOANT password</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0a;font-family:'Inter',system-ui,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#0a0a0a;padding:40px 20px;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:linear-gradient(145deg,#111111,#1a1a1a);border:1px solid #1e1e2e;border-radius:16px;overflow:hidden;max-width:600px;width:100%%;">
        <!-- Header -->
        <tr>
          <td style="background:linear-gradient(135deg,#1e3a5f 0%%,#0f2444 100%%);padding:40px 48px 32px;text-align:center;border-bottom:1px solid #1e2d4a;">
            <div style="display:inline-block;background:rgba(59,130,246,0.15);border:1px solid rgba(59,130,246,0.3);border-radius:12px;padding:10px 20px;margin-bottom:20px;">
              <span style="color:#3b82f6;font-size:22px;font-weight:800;letter-spacing:-0.5px;">NOANT</span>
              <span style="color:#64748b;font-size:11px;font-weight:500;margin-left:8px;text-transform:uppercase;letter-spacing:2px;">AI Support</span>
            </div>
            <h1 style="color:#f1f5f9;font-size:26px;font-weight:700;margin:0;line-height:1.3;">Reset your password</h1>
            <p style="color:#64748b;font-size:14px;margin:10px 0 0;">A password reset was requested for your account.</p>
          </td>
        </tr>
        <!-- Body -->
        <tr>
          <td style="padding:40px 48px;">
            <p style="color:#94a3b8;font-size:15px;line-height:1.7;margin:0 0 28px;">
              Hi there,<br/><br/>
              We received a request to reset the password for your NOANT account associated with <strong style="color:#e2e8f0;">%s</strong>.<br/><br/>
              Click the button below to choose a new password. This link is valid for <strong style="color:#3b82f6;">1 hour</strong>.
            </p>
            <div style="text-align:center;margin:36px 0;">
              <a href="%s" style="display:inline-block;background:linear-gradient(135deg,#3b82f6,#2563eb);color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;padding:14px 36px;border-radius:10px;letter-spacing:0.3px;box-shadow:0 4px 20px rgba(59,130,246,0.4);">
                Reset Password →
              </a>
            </div>
            <p style="color:#475569;font-size:13px;line-height:1.6;margin:28px 0 0;padding-top:24px;border-top:1px solid #1e293b;">
              If you didn't request a password reset, you can safely ignore this email — your password won't change.<br/><br/>
              For security, this link will expire in 1 hour and can only be used once.
            </p>
          </td>
        </tr>
        <!-- Footer -->
        <tr>
          <td style="background:#0d0d0d;border-top:1px solid #1e1e2e;padding:24px 48px;text-align:center;">
            <p style="color:#334155;font-size:12px;margin:0;">© 2025 NOANT. All rights reserved.</p>
            <p style="color:#1e293b;font-size:11px;margin:8px 0 0;">You're receiving this because a password reset was requested for your account.</p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, toEmail, safeURL)

	payload := map[string]interface{}{
		"from":    s.from,
		"to":      []string{toEmail},
		"subject": "Reset your NOANT password",
		"html":    htmlBody,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("resend API error: %s", resp.Status)
	}

	var result resendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode resend response: %w", err)
	}

	return result.ID, nil
}

func (s *ResendService) SendNotificationEmail(ctx context.Context, toEmail, subject, bodyText string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("resend API key not configured")
	}

	payload := map[string]interface{}{
		"from":    s.from,
		"to":      []string{toEmail},
		"subject": subject,
		"html":    fmt.Sprintf(`<p>%s</p>`, html.EscapeString(bodyText)),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("resend API error: %s", resp.Status)
	}

	var result resendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode resend response: %w", err)
	}

	return result.ID, nil
}