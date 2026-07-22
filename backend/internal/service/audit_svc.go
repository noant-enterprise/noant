package service

import (
	"context"

	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== AUDIT SERVICE ==========

type AuditService struct {
	repos  *repository.Repositories
	logger *infrastructure.Logger
}

func NewAuditService(repos *repository.Repositories, logger *infrastructure.Logger) *AuditService {
	return &AuditService{repos: repos, logger: logger}
}

func (s *AuditService) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	return s.repos.Audit.ListByUser(ctx, userID, limit)
}

func (s *AuditService) ListWithFilters(ctx context.Context, filter *repository.AuditFilter) (*repository.AuditListResult, error) {
	return s.repos.Audit.ListWithFilters(ctx, filter)
}
