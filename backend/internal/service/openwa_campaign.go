package service

import (
	"context"
	"fmt"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== CAMPAIGN BRIDGE ==========

const (
	CampaignBatchSize = 50
	MaxFailureRate    = 0.20 // 20% failure rate triggers pause
)

type CampaignBridge struct {
	cfg          *config.Config
	openwa       *OpenWAService
	redis        *infrastructure.RedisClient
	logger       *infrastructure.Logger
	repos        *repository.Repositories
	queue        *SendQueue
	rateLimiter  *MessageRateLimiter
	workerPool   *WorkerPool
	sessionMgr   *SessionManager
}

func NewCampaignBridge(cfg *config.Config, openwa *OpenWAService, redis *infrastructure.RedisClient, logger *infrastructure.Logger, repos *repository.Repositories, queue *SendQueue, rateLimiter *MessageRateLimiter, workerPool *WorkerPool, sessionMgr *SessionManager) *CampaignBridge {
	return &CampaignBridge{
		cfg:         cfg,
		openwa:      openwa,
		redis:       redis,
		logger:      logger,
		repos:       repos,
		queue:       queue,
		rateLimiter: rateLimiter,
		workerPool:  workerPool,
		sessionMgr:  sessionMgr,
	}
}

// BroadcastRequest defines a campaign broadcast
type BroadcastRequest struct {
	CampaignID string   `json:"campaign_id"`
	UserID     string   `json:"user_id"`
	SessionID  string   `json:"session_id"`
	Message    string   `json:"message"`
	Recipients []string `json:"recipients"` // phone numbers
	TemplateID string   `json:"template_id,omitempty"`
	Variables  map[string]string `json:"variables,omitempty"`
}

// ExecuteCampaign runs a campaign broadcast
func (cb *CampaignBridge) ExecuteCampaign(ctx context.Context, req *BroadcastRequest) error {
	if len(req.Recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	cb.logger.Info("Starting campaign broadcast", "campaignID", req.CampaignID, "recipients", len(req.Recipients))

	// Ensure worker exists for the session
	cb.workerPool.EnsureWorker(req.SessionID)

	// Process in batches
	totalRecipients := len(req.Recipients)
	var sent, failed int

	for i := 0; i < totalRecipients; i += CampaignBatchSize {
		end := i + CampaignBatchSize
		if end > totalRecipients {
			end = totalRecipients
		}

		batch := req.Recipients[i:end]

		// Check failure rate before proceeding
		if i > 0 {
			currentRate := float64(failed) / float64(sent+failed)
			if currentRate > MaxFailureRate {
				cb.logger.Error("Campaign paused: failure rate exceeded threshold",
					"campaignID", req.CampaignID, "failureRate", currentRate,
					"sent", sent, "failed", failed)
				cb.updateCampaignStatus(ctx, req.CampaignID, "paused")
				return fmt.Errorf("campaign paused: failure rate %.2f exceeds %.2f threshold", currentRate, MaxFailureRate)
			}
		}

		// Enqueue batch
		for _, phone := range batch {
			// Check opt-out before sending
			cleanedPhone := CleanPhoneNumber(phone)
			optedOut, err := cb.repos.CampaignRecipient.IsOptedOut(ctx, req.UserID, cleanedPhone)
			if err != nil {
				cb.logger.Warn("Failed to check opt-out status", "phone", phone, "error", err)
			}
			if optedOut {
				cb.logger.Info("Skipping opted-out recipient", "phone", phone, "campaignID", req.CampaignID)
				cb.createRecipientRecord(ctx, req.CampaignID, req.UserID, phone)
				if err := cb.repos.CampaignRecipient.MarkOptedOut(ctx, req.UserID, cleanedPhone); err != nil {
					cb.logger.Warn("Failed to mark opted-out recipient", "phone", phone, "error", err)
				}
				continue
			}

			chatID := FormatChatID(phone)
			entry := &QueueEntry{
				ID:        fmt.Sprintf("camp_%s_%s_%d", req.CampaignID, phone, time.Now().UnixNano()),
				SessionID: req.SessionID,
				UserID:    req.UserID,
				ChatID:    chatID,
				MsgType:   MsgTypeText,
				Content:   req.Message,
				Priority:  PriorityBulk,
			}

			if req.TemplateID != "" {
				entry.MsgType = MsgTypeTemplate
				entry.Content = fmt.Sprintf(`{"template":%q,"variables":%v}`, req.TemplateID, req.Variables)
			}

			if err := cb.queue.Enqueue(entry); err != nil {
				cb.logger.Error("Failed to enqueue campaign message", "phone", phone, "error", err)
				failed++
				continue
			}
			sent++

			// Create campaign recipient record
			cb.createRecipientRecord(ctx, req.CampaignID, req.UserID, phone)
		}

		// Spread batches across time to respect rate limits
		if end < totalRecipients {
			time.Sleep(2 * time.Second)
		}
	}

	cb.logger.Info("Campaign broadcast completed", "campaignID", req.CampaignID, "sent", sent, "failed", failed)
	return nil
}

func (cb *CampaignBridge) updateCampaignStatus(ctx context.Context, campaignID, status string) {
	if err := cb.repos.Campaign.UpdateStatus(ctx, campaignID, status); err != nil {
		cb.logger.Error("Failed to update campaign status", "campaignID", campaignID, "status", status, "error", err)
	}
}

func (cb *CampaignBridge) createRecipientRecord(ctx context.Context, campaignID, userID, phone string) {
	recipient := &domain.CampaignRecipient{
		CampaignID: campaignID,
		UserID:     userID,
		Phone:      phone,
		Status:     "pending",
	}
	if err := cb.repos.CampaignRecipient.Create(ctx, recipient); err != nil {
		cb.logger.Warn("Failed to create recipient record", "phone", phone, "error", err)
	}
}

// GetCampaignAnalytics returns analytics for a campaign
func (cb *CampaignBridge) GetCampaignAnalytics(ctx context.Context, campaignID string) (map[string]interface{}, error) {
	recipients, err := cb.repos.CampaignRecipient.ListByCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	var total, sent, delivered, read, failed, blocked, optedOut int
	for i := range recipients {
		total++
		switch recipients[i].Status {
		case "sent":
			sent++
		case "delivered":
			delivered++
			sent++
		case "read":
			read++
			delivered++
			sent++
		case "failed":
			failed++
		case "blocked":
			blocked++
		case "opted_out":
			optedOut++
		default:
			// pending
		}
	}

	deliveryRate := 0.0
	if total > 0 {
		deliveryRate = float64(delivered) / float64(total) * 100
	}
	readRate := 0.0
	if delivered > 0 {
		readRate = float64(read) / float64(delivered) * 100
	}
	return map[string]interface{}{
		"total":        total,
		"sent":         sent,
		"delivered":    delivered,
		"read":         read,
		"failed":       failed,
		"blocked":      blocked,
		"opted_out":    optedOut,
		"delivery_rate": deliveryRate,
		"read_rate":    readRate,
	}, nil
}

// ProcessOptOut handles an opt-out message (STOP)
func (cb *CampaignBridge) ProcessOptOut(ctx context.Context, userID, phone string) error {
	phone = CleanPhoneNumber(phone)

	// Mark all pending campaign recipients as opted_out
	return cb.repos.CampaignRecipient.MarkOptedOut(ctx, userID, phone)
}

// IsOptedOut checks if a user has opted out
func (cb *CampaignBridge) IsOptedOut(ctx context.Context, userID, phone string) (bool, error) {
	phone = CleanPhoneNumber(phone)
	return cb.repos.CampaignRecipient.IsOptedOut(ctx, userID, phone)
}
