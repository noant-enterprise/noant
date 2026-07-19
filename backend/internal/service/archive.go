package service

import (
	"context"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== ARCHIVE SERVICE ==========

type ArchiveService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewArchiveService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *ArchiveService {
	return &ArchiveService{cfg: cfg, repos: repos, redis: redis, logger: logger}
}

func (s *ArchiveService) ListFolders(ctx context.Context, userID, folderType string) ([]domain.ArchiveFolder, error) {
	return s.repos.Archive.ListFolders(ctx, userID, folderType)
}

func (s *ArchiveService) CreateFolder(ctx context.Context, userID, name, folderType, color string) (*domain.ArchiveFolder, error) {
	folder := &domain.ArchiveFolder{
		UserID: userID,
		Name:   name,
		Type:   folderType,
		Color:  color,
	}
	if err := s.repos.Archive.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *ArchiveService) DeleteFolder(ctx context.Context, id string) error {
	return nil
}

func (s *ArchiveService) MoveChat(ctx context.Context, userID, conversationID, folderID string) error {
	return s.repos.Archive.MoveChat(ctx, conversationID, userID, folderID)
}

func (s *ArchiveService) RemoveFromArchive(ctx context.Context, userID, conversationID string) error {
	return s.repos.Archive.MoveChat(ctx, conversationID, userID, "")
}

func (s *ArchiveService) GetStatus(ctx context.Context, userID string) (map[string]interface{}, error) {
	folders, _ := s.repos.Archive.ListFolders(ctx, userID, "")
	return map[string]interface{}{
		"folders":     len(folders),
		"total_items": 0,
	}, nil
}
