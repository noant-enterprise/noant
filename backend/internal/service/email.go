package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

type SMTPSettings struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	SkipVerify bool
}

type EmailService struct {
	cfg    *config.Config
	logger *infrastructure.Logger
	resend *ResendService
}

func NewEmailService(cfg *config.Config, logger *infrastructure.Logger) *EmailService {
	svc := &EmailService{cfg: cfg, logger: logger}
	if cfg.ResendAPIKey != "" {
		svc.resend = NewResendService(cfg.ResendAPIKey, cfg.ResendFrom, cfg.APIURL)
	}
	return svc
}

func smtpSettingsFromConfig(cfg *config.Config, overrides map[string]interface{}) SMTPSettings {
	settings := SMTPSettings{
		Host:       cfg.SMTPHost,
		Port:       cfg.SMTPPort,
		Username:   cfg.SMTPUsername,
		Password:   cfg.SMTPPassword,
		From:       cfg.SMTPFrom,
		SkipVerify: cfg.SMTPSkipVerify,
	}

	if overrides != nil {
		if v, ok := overrides["smtp_host"].(string); ok && strings.TrimSpace(v) != "" {
			settings.Host = v
		}
		if v, ok := overrides["smtp_port"].(float64); ok && v > 0 {
			settings.Port = int(v)
		}
		if v, ok := overrides["smtp_port"].(string); ok {
			if p, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && p > 0 {
				settings.Port = p
			}
		}
		if v, ok := overrides["smtp_username"].(string); ok && strings.TrimSpace(v) != "" {
			settings.Username = v
		}
		if v, ok := overrides["smtp_password"].(string); ok && strings.TrimSpace(v) != "" {
			settings.Password = v
		}
		if v, ok := overrides["smtp_from"].(string); ok && strings.TrimSpace(v) != "" {
			settings.From = v
		}
		if v, ok := overrides["smtp_skip_verify"].(bool); ok {
			settings.SkipVerify = v
		}
		if v, ok := overrides["smtp_skip_verify"].(string); ok {
			settings.SkipVerify = strings.EqualFold(strings.TrimSpace(v), "true")
		}
	}

	if settings.From == "" {
		settings.From = settings.Username
	}
	if settings.Port == 0 {
		settings.Port = 587
	}
	return settings
}

func sendSMTPMessage(ctx context.Context, settings SMTPSettings, toEmail, subject, bodyHTML string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(settings.Host) == "" {
		return "", fmt.Errorf("smtp host is required")
	}
	if strings.TrimSpace(settings.Username) == "" || strings.TrimSpace(settings.Password) == "" {
		return "", fmt.Errorf("smtp username and password are required")
	}
	if strings.TrimSpace(settings.From) == "" {
		return "", fmt.Errorf("smtp from address is required")
	}
	if strings.TrimSpace(toEmail) == "" {
		return "", fmt.Errorf("recipient email is required")
	}

	addr := net.JoinHostPort(settings.Host, strconv.Itoa(settings.Port))
	client, err := smtp.Dial(addr)
	if err != nil {
		return "", err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         settings.Host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: settings.SkipVerify,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return "", err
		}
	}

	auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
	if err := client.Auth(auth); err != nil {
		return "", err
	}
	if err := client.Mail(settings.From); err != nil {
		return "", err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return "", err
	}

	writer, err := client.Data()
	if err != nil {
		return "", err
	}

	msg := bytes.Buffer{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", settings.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(bodyHTML)

	if _, err := writer.Write(msg.Bytes()); err != nil {
		_ = writer.Close()
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	if err := client.Quit(); err != nil {
		return "", err
	}

	return fmt.Sprintf("smtp_%d", time.Now().UnixNano()), nil
}

func (s *EmailService) SendPasswordReset(ctx context.Context, toEmail, resetToken string) (string, error) {
	// Try Resend first if configured
	if s.resend != nil {
		id, err := s.resend.SendPasswordReset(ctx, toEmail, resetToken)
		if err == nil {
			return id, nil
		}
		s.logger.Warn("Resend failed, falling back to SMTP", "error", err)
	}

	// Fall back to SMTP
	baseURL := s.cfg.APIURL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, url.QueryEscape(resetToken))
	safeURL := html.EscapeString(resetURL)
	safeEmail := html.EscapeString(toEmail)

	body := fmt.Sprintf(`
<html>
<body style="font-family:Arial,sans-serif;background:#f6f8fb;padding:24px;">
  <div style="max-width:600px;margin:0 auto;background:#fff;border-radius:12px;padding:32px;border:1px solid #e5e7eb;">
    <h2 style="margin:0 0 16px;">Reset your NOANT password</h2>
    <p style="color:#374151;line-height:1.6;">We received a request to reset the password for <strong>%s</strong>.</p>
    <p><a href="%s" style="display:inline-block;background:#2563eb;color:#fff;text-decoration:none;padding:12px 20px;border-radius:8px;">Reset Password</a></p>
    <p style="color:#6b7280;font-size:13px;">This link expires in 1 hour.</p>
  </div>
</body>
</html>`, safeEmail, safeURL)

	return sendSMTPMessage(ctx, smtpSettingsFromConfig(s.cfg, nil), toEmail, "Reset your NOANT password", body)
}

func (s *EmailService) SendNotificationEmail(ctx context.Context, toEmail, subject, bodyText string) (string, error) {
	// Try Resend first if configured
	if s.resend != nil {
		id, err := s.resend.SendNotificationEmail(ctx, toEmail, subject, bodyText)
		if err == nil {
			return id, nil
		}
		s.logger.Warn("Resend failed, falling back to SMTP", "error", err)
	}

	// Fall back to SMTP
	body := fmt.Sprintf("<html><body style=\"font-family:Arial,sans-serif;white-space:pre-wrap;color:#111827;\">%s</body></html>", html.EscapeString(bodyText))
	return sendSMTPMessage(ctx, smtpSettingsFromConfig(s.cfg, nil), toEmail, subject, body)
}
