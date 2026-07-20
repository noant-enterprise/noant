package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type PushNotificationService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	logger *infrastructure.Logger
}

func NewPushNotificationService(cfg *config.Config, repos *repository.Repositories, logger *infrastructure.Logger) *PushNotificationService {
	return &PushNotificationService{cfg: cfg, repos: repos, logger: logger}
}

func (s *PushNotificationService) Subscribe(ctx context.Context, userID, endpoint, auth, p256dh, userAgent string) error {
	sub := &domain.PushSubscription{
		UserID:    userID,
		Endpoint:  endpoint,
		Auth:      auth,
		P256dh:    p256dh,
		UserAgent: userAgent,
	}
	return s.repos.PushSubscription.Create(ctx, sub)
}

func (s *PushNotificationService) Unsubscribe(ctx context.Context, userID, endpoint string) error {
	if endpoint == "" {
		return s.repos.PushSubscription.DeleteAllByUser(ctx, userID)
	}
	return s.repos.PushSubscription.Delete(ctx, userID, endpoint)
}

type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
}

func (s *PushNotificationService) SendToUser(ctx context.Context, userID, title, body, url string) error {
	if s.cfg.VAPIDPrivateKey == "" || s.cfg.VAPIDPublicKey == "" {
		return nil // VAPID not configured, skip silently
	}

	subs, err := s.repos.PushSubscription.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}

	payload, _ := json.Marshal(PushPayload{
		Title: title,
		Body:  body,
		URL:   url,
	})

	for _, sub := range subs {
		if err := s.sendPush(sub, payload); err != nil {
			s.logger.Error("Push send failed", "error", err, "userID", userID)
			// Remove invalid subscription (endpoint expired/blocked)
			if isEndpointGone(err) {
				_ = s.repos.PushSubscription.DeleteByID(ctx, sub.ID)
			}
		}
	}

	return nil
}

func (s *PushNotificationService) SendToUsers(ctx context.Context, userIDs []string, title, body, url string) error {
	if s.cfg.VAPIDPrivateKey == "" || s.cfg.VAPIDPublicKey == "" {
		return nil
	}

	subs, err := s.repos.PushSubscription.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}

	payload, _ := json.Marshal(PushPayload{
		Title: title,
		Body:  body,
		URL:   url,
	})

	for _, sub := range subs {
		if err := s.sendPush(sub, payload); err != nil {
			s.logger.Error("Push send failed", "error", err, "userID", sub.UserID)
			if isEndpointGone(err) {
				_ = s.repos.PushSubscription.DeleteByID(ctx, sub.ID)
			}
		}
	}

	return nil
}

func (s *PushNotificationService) sendPush(sub *domain.PushSubscription, payload []byte) error {
	sdkSub := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.Auth,
			P256dh: sub.P256dh,
		},
	}

	resp, err := webpush.SendNotificationWithContext(context.Background(), payload, sdkSub, &webpush.Options{
		Subscriber:      s.cfg.VAPIDSubject,
		VAPIDPublicKey:  s.cfg.VAPIDPublicKey,
		VAPIDPrivateKey: s.cfg.VAPIDPrivateKey,
		TTL:             86400,
		HTTPClient:      http.DefaultClient,
	})
	if err != nil {
		return err
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
	return nil
}

func isEndpointGone(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return pushContainsAny(errStr, "410", "gone", "Gone", "unsubscribed", "expired")
}

func pushContainsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
