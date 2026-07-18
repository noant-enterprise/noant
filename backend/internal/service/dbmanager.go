package service

import (
	"context"
	"fmt"
	"time"

	"noant/internal/infrastructure"
	"noant/internal/repository"
)

type CleanupResult struct {
	Task             string `json:"task"`
	RowsAffected     int64  `json:"rows_affected"`
	DurationMs       int64  `json:"duration_ms"`
	Error            string `json:"error,omitempty"`
}

type CleanupReport struct {
	StartedAt  time.Time       `json:"started_at"`
	DurationMs int64           `json:"total_duration_ms"`
	Results    []CleanupResult `json:"results"`
	HasErrors  bool            `json:"has_errors"`
}

type DBManagerService struct {
	repos  *repository.Repositories
	logger *infrastructure.Logger
}

func NewDBManagerService(repos *repository.Repositories, logger *infrastructure.Logger) *DBManagerService {
	return &DBManagerService{repos: repos, logger: logger}
}

func (s *DBManagerService) measure(fn func() (int64, error)) (int64, error) {
	start := time.Now()
	affected, err := fn()
	elapsed := time.Since(start)
	if err != nil {
		s.logger.Error("Cleanup task failed", "error", err, "duration", elapsed)
	}
	return affected, err
}

func (s *DBManagerService) recordResult(results *[]CleanupResult, name string, affected int64, err error) {
	r := CleanupResult{
		Task:         name,
		RowsAffected: affected,
		DurationMs:   time.Now().UnixMilli(),
	}
	if err != nil {
		r.Error = err.Error()
	}
	*results = append(*results, r)
}

func (s *DBManagerService) CleanupOldResolvedConversations(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Conversation.CleanupOldResolved(ctx, days)
	})
}

func (s *DBManagerService) CleanupAbandonedConversations(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Conversation.CleanupAbandoned(ctx, days)
	})
}

func (s *DBManagerService) CleanupOrphanedMessages(ctx context.Context) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Message.CleanupOrphaned(ctx)
	})
}

func (s *DBManagerService) CleanupStaleUnknownQuestions(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.UnknownQ.CleanupStale(ctx, days)
	})
}

func (s *DBManagerService) CleanupExpiredHandoffs(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Handoff.CleanupExpired(ctx, days)
	})
}

func (s *DBManagerService) CleanupOldAuditLogs(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Audit.CleanupOld(ctx, days)
	})
}

func (s *DBManagerService) CleanupOldNotifications(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Notification.CleanupOld(ctx, days)
	})
}

func (s *DBManagerService) CleanupStaleInactiveIntegrations(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Integration.CleanupStaleInactive(ctx, days)
	})
}

func (s *DBManagerService) CleanupExpiredTrials(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.User.CleanupExpiredTrials(ctx, days)
	})
}

func (s *DBManagerService) CleanupExpiredCredits(ctx context.Context) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Credit.CleanupExpired(ctx)
	})
}

func (s *DBManagerService) CleanupStaleCreditPurchases(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Credit.CleanupStalePurchases(ctx, days)
	})
}

func (s *DBManagerService) CleanupCompletedCampaigns(ctx context.Context, days int) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.Campaign.CleanupCompleted(ctx, days)
	})
}

func (s *DBManagerService) CleanupExpiredMediaMessages(ctx context.Context) (int64, error) {
	return s.measure(func() (int64, error) {
		return s.repos.MediaMessage.CleanupExpired(ctx)
	})
}

func (s *DBManagerService) RunAllCleanups(ctx context.Context, config *CleanupConfig) *CleanupReport {
	startedAt := time.Now()
	report := &CleanupReport{
		StartedAt: startedAt,
		Results:   []CleanupResult{},
	}

	tasks := []struct {
		name string
		fn   func() (int64, error)
	}{
		{"old_resolved_conversations", func() (int64, error) { return s.CleanupOldResolvedConversations(ctx, config.OldConversationsDays) }},
		{"abandoned_conversations", func() (int64, error) { return s.CleanupAbandonedConversations(ctx, config.AbandonedConversationsDays) }},
		{"orphaned_messages", func() (int64, error) { return s.CleanupOrphanedMessages(ctx) }},
		{"stale_unknown_questions", func() (int64, error) { return s.CleanupStaleUnknownQuestions(ctx, config.UnknownQuestionsDays) }},
		{"expired_handoffs", func() (int64, error) { return s.CleanupExpiredHandoffs(ctx, config.HandoffsDays) }},
		{"old_audit_logs", func() (int64, error) { return s.CleanupOldAuditLogs(ctx, config.AuditLogsDays) }},
		{"old_notifications", func() (int64, error) { return s.CleanupOldNotifications(ctx, config.NotificationsDays) }},
		{"stale_inactive_integrations", func() (int64, error) { return s.CleanupStaleInactiveIntegrations(ctx, config.InactiveIntegrationDays) }},
		{"expired_trials", func() (int64, error) { return s.CleanupExpiredTrials(ctx, config.ExpiredTrialDays) }},
		{"expired_credits", func() (int64, error) { return s.CleanupExpiredCredits(ctx) }},
		{"stale_credit_purchases", func() (int64, error) { return s.CleanupStaleCreditPurchases(ctx, config.CreditPurchasesDays) }},
		{"completed_campaigns", func() (int64, error) { return s.CleanupCompletedCampaigns(ctx, config.CompletedCampaignsDays) }},
		{"expired_media_messages", func() (int64, error) { return s.CleanupExpiredMediaMessages(ctx) }},
	}

	for _, t := range tasks {
		start := time.Now()
		affected, err := t.fn()
		r := CleanupResult{
			Task:         t.name,
			RowsAffected: affected,
			DurationMs:   time.Since(start).Milliseconds(),
		}
		if err != nil {
			r.Error = err.Error()
			report.HasErrors = true
		}
		report.Results = append(report.Results, r)
	}

	report.DurationMs = time.Since(startedAt).Milliseconds()
	return report
}

type CleanupConfig struct {
	OldConversationsDays     int
	AbandonedConversationsDays int
	UnknownQuestionsDays     int
	HandoffsDays             int
	AuditLogsDays            int
	NotificationsDays        int
	InactiveIntegrationDays  int
	ExpiredTrialDays         int
	CreditPurchasesDays      int
	CompletedCampaignsDays   int
}

func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		OldConversationsDays:       30,
		AbandonedConversationsDays: 7,
		UnknownQuestionsDays:       30,
		HandoffsDays:               7,
		AuditLogsDays:              90,
		NotificationsDays:          30,
		InactiveIntegrationDays:    90,
		ExpiredTrialDays:           7,
		CreditPurchasesDays:        365,
		CompletedCampaignsDays:     90,
	}
}

type TaskResult struct {
	TaskID    string      `json:"task_id"`
	TaskName  string      `json:"task_name"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	StartedAt time.Time   `json:"started_at"`
	DoneAt    time.Time   `json:"done_at,omitempty"`
}

func (s *DBManagerService) RunTask(ctx context.Context, taskName string, config *CleanupConfig) *TaskResult {
	result := &TaskResult{
		TaskID:    fmt.Sprintf("task_%d", time.Now().UnixNano()),
		TaskName:  taskName,
		Status:    "running",
		StartedAt: time.Now(),
	}

	var affected int64
	var err error

	switch taskName {
	case "all":
		r := s.RunAllCleanups(ctx, config)
		result.Result = r
		result.DoneAt = time.Now()
		result.Status = "completed"
		return result
	case "old_resolved_conversations":
		affected, err = s.CleanupOldResolvedConversations(ctx, config.OldConversationsDays)
	case "abandoned_conversations":
		affected, err = s.CleanupAbandonedConversations(ctx, config.AbandonedConversationsDays)
	case "orphaned_messages":
		affected, err = s.CleanupOrphanedMessages(ctx)
	case "stale_unknown_questions":
		affected, err = s.CleanupStaleUnknownQuestions(ctx, config.UnknownQuestionsDays)
	case "expired_handoffs":
		affected, err = s.CleanupExpiredHandoffs(ctx, config.HandoffsDays)
	case "old_audit_logs":
		affected, err = s.CleanupOldAuditLogs(ctx, config.AuditLogsDays)
	case "old_notifications":
		affected, err = s.CleanupOldNotifications(ctx, config.NotificationsDays)
	case "stale_inactive_integrations":
		affected, err = s.CleanupStaleInactiveIntegrations(ctx, config.InactiveIntegrationDays)
	case "expired_trials":
		affected, err = s.CleanupExpiredTrials(ctx, config.ExpiredTrialDays)
	case "expired_credits":
		affected, err = s.CleanupExpiredCredits(ctx)
	case "stale_credit_purchases":
		affected, err = s.CleanupStaleCreditPurchases(ctx, config.CreditPurchasesDays)
	case "completed_campaigns":
		affected, err = s.CleanupCompletedCampaigns(ctx, config.CompletedCampaignsDays)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unknown task: %s", taskName)
		result.DoneAt = time.Now()
		return result
	}

	result.DoneAt = time.Now()
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
	} else {
		result.Status = "completed"
		result.Result = map[string]interface{}{"rows_affected": affected}
	}
	return result
}
