package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== SETTINGS SERVICE ==========

type SettingsService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
	email  *EmailService
}

func NewSettingsService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, email *EmailService) *SettingsService {
	return &SettingsService{cfg: cfg, repos: repos, redis: redis, logger: logger, email: email}
}

func (s *SettingsService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return s.repos.User.GetByID(ctx, userID)
}

func (s *SettingsService) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	firstName, _ := updates["first_name"].(string)
	lastName, _ := updates["last_name"].(string)
	companyName, _ := updates["company_name"].(string)
	phone, _ := updates["phone"].(string)
	return s.repos.User.UpdateProfile(ctx, userID, firstName, lastName, companyName, phone)
}

func (s *SettingsService) GetNotifPrefs(ctx context.Context, userID string) (*repository.NotifPrefs, error) {
	return s.repos.User.GetNotifPrefs(ctx, userID)
}

func (s *SettingsService) UpdateNotifPrefs(ctx context.Context, userID string, prefs *repository.NotifPrefs) error {
	return s.repos.User.UpdateNotifPrefs(ctx, userID, prefs)
}

func (s *SettingsService) DeleteAccount(ctx context.Context, userID string) error {
	return s.repos.User.Delete(ctx, userID)
}

func (s *SettingsService) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	return s.repos.User.ExportUserData(ctx, userID)
}

func (s *SettingsService) ListAPIKeys(ctx context.Context, userID string) ([]domain.APIKey, error) {
	return s.repos.APIKey.ListByOrg(ctx, userID)
}

func (s *SettingsService) CreateAPIKey(ctx context.Context, userID, name string) (*domain.APIKey, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("generate API key: %w", err)
	}
	apiKey := &domain.APIKey{
		UserID:   userID,
		Name:     name,
		Key:      "noant_" + hex.EncodeToString(keyBytes),
		IsActive: true,
	}
	if err := s.repos.APIKey.Create(ctx, apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *SettingsService) RevokeAPIKey(ctx context.Context, userID, id string) error {
	return s.repos.APIKey.Revoke(ctx, id, userID)
}

func (s *SettingsService) ListTeam(ctx context.Context, orgID string) ([]domain.TeamMember, error) {
	return s.repos.Team.ListByOrg(ctx, orgID)
}

func (s *SettingsService) InviteTeamMember(ctx context.Context, orgID, email, role string) (*domain.TeamMember, error) {
	member := &domain.TeamMember{
		Email:    email,
		Role:     role,
		IsActive: false,
	}
	if err := s.repos.Team.Create(ctx, orgID, member); err != nil {
		return nil, err
	}

	// Send invite email
	if s.email != nil {
		org, _ := s.repos.Org.GetByID(ctx, orgID)
		ownerName := "Your team"
		if org != nil {
			owner, _ := s.repos.User.GetByID(ctx, org.OwnerID)
			if owner != nil {
				ownerName = owner.FirstName
			}
		}
		subject := fmt.Sprintf("%s invited you to join NOANT", ownerName)
		body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>You're invited to join NOANT</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0a;font-family:'Inter',system-ui,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#0a0a0a;padding:40px 20px;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:linear-gradient(145deg,#111111,#1a1a1a);border:1px solid #1e1e2e;border-radius:16px;overflow:hidden;max-width:600px;width:100%%;">
        <tr>
          <td style="background:linear-gradient(135deg,#1e3a5f 0%%,#0f2444 100%%);padding:40px 48px 32px;text-align:center;border-bottom:1px solid #1e2d4a;">
            <div style="display:inline-block;background:rgba(59,130,246,0.15);border:1px solid rgba(59,130,246,0.3);border-radius:12px;padding:10px 20px;margin-bottom:20px;">
              <span style="color:#3b82f6;font-size:22px;font-weight:800;letter-spacing:-0.5px;">NOANT</span>
              <span style="color:#64748b;font-size:11px;font-weight:500;margin-left:8px;text-transform:uppercase;letter-spacing:2px;">AI Support</span>
            </div>
            <h1 style="color:#f1f5f9;font-size:26px;font-weight:700;margin:0;line-height:1.3;">You're invited!</h1>
            <p style="color:#64748b;font-size:14px;margin:10px 0 0;">Join your team on NOANT AI Support</p>
          </td>
        </tr>
        <tr>
          <td style="padding:40px 48px;">
            <p style="color:#94a3b8;font-size:15px;line-height:1.7;margin:0 0 28px;">
              Hi there,<br/><br/>
              <strong style="color:#e2e8f0;">%s</strong> has invited you to join their NOANT team as a <strong style="color:#3b82f6;">%s</strong>.
            </p>
            <div style="text-align:center;margin:36px 0;">
              <a href="%s/team" style="display:inline-block;background:linear-gradient(135deg,#3b82f6,#2563eb);color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;padding:14px 36px;border-radius:10px;letter-spacing:0.3px;box-shadow:0 4px 20px rgba(59,130,246,0.4);">
                Accept Invitation
              </a>
            </div>
            <p style="color:#475569;font-size:13px;line-height:1.6;margin:28px 0 0;padding-top:24px;border-top:1px solid #1e293b;">
              If you don't have a NOANT account yet, you'll be prompted to create one.<br/><br/>
              If you believe you received this invitation by mistake, you can safely ignore this email.
            </p>
          </td>
        </tr>
        <tr>
          <td style="background:#0d0d0d;border-top:1px solid #1e1e2e;padding:24px 48px;text-align:center;">
            <p style="color:#334155;font-size:12px;margin:0;">© 2026 NOANT. All rights reserved.</p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, html.EscapeString(ownerName), html.EscapeString(role), s.cfg.AppURL)
		if _, err := s.email.SendHTMLEmail(ctx, email, subject, body); err != nil {
			s.logger.Warn("Failed to send team invite email", "error", err, "email", email)
		}
	}

	return member, nil
}

func (s *SettingsService) RemoveTeamMember(ctx context.Context, id string) error {
	return s.repos.Team.Delete(ctx, id)
}
