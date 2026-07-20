package service

import (
	"context"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== INVENTORY SERVICE ==========

type InventoryService struct {
	cfg        *config.Config
	repos      *repository.Repositories
	redis      *infrastructure.RedisClient
	logger     *infrastructure.Logger
	embeddings *EmbeddingService
}

func NewInventoryService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, embeddings *EmbeddingService) *InventoryService {
	return &InventoryService{cfg: cfg, repos: repos, redis: redis, logger: logger, embeddings: embeddings}
}

func (s *InventoryService) Create(ctx context.Context, userID string, item *domain.InventoryItem) error {
	item.UserID = userID
	if item.Type == "" {
		item.Type = "product"
	}
	item.IsActive = true
	if err := s.repos.Inventory.Create(ctx, item); err != nil {
		return err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(userID)
	}
	return nil
}

func (s *InventoryService) List(ctx context.Context, userID, itemType string) ([]domain.InventoryItem, error) {
	return s.repos.Inventory.List(ctx, userID, itemType, false)
}

func (s *InventoryService) GetByID(ctx context.Context, id, userID string) (*domain.InventoryItem, error) {
	return s.repos.Inventory.GetByID(ctx, id, userID)
}

func (s *InventoryService) Update(ctx context.Context, item *domain.InventoryItem) error {
	if err := s.repos.Inventory.Update(ctx, item); err != nil {
		return err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(item.UserID)
	}
	return nil
}

func (s *InventoryService) Delete(ctx context.Context, id, userID string) error {
	if err := s.repos.Inventory.Delete(ctx, id, userID); err != nil {
		return err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(userID)
	}
	return nil
}

func (s *InventoryService) Search(ctx context.Context, userID, query string) ([]domain.InventoryItem, error) {
	return s.repos.Inventory.Search(ctx, userID, query)
}
