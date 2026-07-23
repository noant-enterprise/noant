package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"noant/config"
	apperrors "noant/internal/errors"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/utils"
)

// ========== TRAINING SERVICE ==========

type TrainingService struct {
	cfg        *config.Config
	repos      *repository.Repositories
	redis      *infrastructure.RedisClient
	logger     *infrastructure.Logger
	embeddings *EmbeddingService
}

// NewTrainingService creates a TrainingService for managing QA pairs, categories,
// unknown questions, and CSV bulk imports. The embeddings parameter handles vector
// generation for semantic search indexing.
func NewTrainingService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, embeddings *EmbeddingService) *TrainingService {
	return &TrainingService{cfg: cfg, repos: repos, redis: redis, logger: logger, embeddings: embeddings}
}

func (s *TrainingService) ClearUnknownQuestions(ctx context.Context, userID string) error {
	return s.repos.UnknownQ.Clear(ctx, userID)
}

func (s *TrainingService) CreateCategory(ctx context.Context, userID, name, description, color string) (*domain.Category, error) {
	cat := &domain.Category{
		UserID:      userID,
		OrgID:       userID,
		Name:        name,
		Description: description,
		Color:       color,
	}
	if err := s.repos.Category.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *TrainingService) ListCategories(ctx context.Context, userID string) ([]domain.Category, error) {
	return s.repos.Category.List(ctx, userID)
}

func (s *TrainingService) BulkImport(ctx context.Context, userID, categoryID string, qaPairs []domain.QAPair) error {
	for i := range qaPairs {
		qaPairs[i].UserID = userID
		qaPairs[i].OrgID = userID
		qaPairs[i].CategoryID = categoryID
		qaPairs[i].IsActive = true
	}
	return s.repos.QAPair.BulkCreate(ctx, qaPairs)
}

func (s *TrainingService) UploadCSV(ctx context.Context, userID, categoryID string, csvData []byte) (int, error) {
	reader := csv.NewReader(bytes.NewReader(csvData))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to parse CSV: %w", err)
	}
	if len(records) < 2 {
		return 0, fmt.Errorf("CSV must have at least a header and one data row")
	}

	categoryMap := make(map[string]string)
	var qaPairs []domain.QAPair

	for i, record := range records[1:] {
		if len(record) < 3 {
			s.logger.Warn("Skipping invalid CSV row", "row", i+2)
			continue
		}
		categoryName := utils.SanitizeName(record[0])
		question := utils.SanitizeXSS(record[1])
		answer := utils.SanitizeXSS(strings.Join(record[2:], ","))

		if categoryName == "" || question == "" || answer == "" {
			s.logger.Warn("Skipping empty CSV row", "row", i+2)
			continue
		}

		catID, exists := categoryMap[categoryName]
		if !exists {
			existing, _ := s.repos.Category.GetByName(ctx, userID, categoryName)
			if existing != nil {
				catID = existing.ID
			} else {
			cat := &domain.Category{
				UserID:      userID,
				OrgID:       userID,
				Name:        categoryName,
				Description: "Auto-imported from CSV",
				Color:       "#3b82f6",
			}
				if err := s.repos.Category.Create(ctx, cat); err != nil {
					s.logger.Warn("Failed to create category", "name", categoryName, "error", err)
					continue
				}
				catID = cat.ID
			}
			categoryMap[categoryName] = catID
		}
		existingQA, err := s.repos.QAPair.GetByQuestion(ctx, userID, question)
		if err == nil && existingQA != nil {
			existingQA.Answer = answer
			existingQA.CategoryID = catID
			if err := s.repos.QAPair.Update(ctx, existingQA); err != nil {
				s.logger.Warn("Failed to update existing QAPair", "question", question, "error", err)
			}
		} else {
			qaPairs = append(qaPairs, domain.QAPair{
				UserID:     userID,
				OrgID:      userID,
				CategoryID: catID,
				Category:   categoryName,
				Question:   question,
				Answer:     answer,
				IsActive:   true,
			})
		}
	}

	if len(qaPairs) > 0 {
		err = s.repos.QAPair.BulkCreate(ctx, qaPairs)
		if err != nil {
			return 0, err
		}
		if s.embeddings != nil {
			s.embeddings.InvalidateCache(userID)
		}
	}
	// Return the total number of processed records (updates + inserts)
	return len(records) - 1, nil
}

func (s *TrainingService) ListUnknownQuestions(ctx context.Context, userID, status string, limit, offset int) ([]domain.UnknownQuestion, error) {
	return s.repos.UnknownQ.List(ctx, userID, status, limit, offset)
}

func (s *TrainingService) CountUnknownQuestions(ctx context.Context, userID, status string) (int, error) {
	return s.repos.UnknownQ.CountByFilter(ctx, userID, status)
}

func (s *TrainingService) BatchTrainUnknown(ctx context.Context, userID, answer, categoryID string, ids []string) error {
	return s.repos.UnknownQ.BatchTrain(ctx, userID, answer, categoryID, ids)
}

func (s *TrainingService) BatchIgnoreUnknown(ctx context.Context, userID string, ids []string) error {
	return s.repos.UnknownQ.BatchIgnore(ctx, userID, ids)
}

func (s *TrainingService) TrainUnknown(ctx context.Context, userID, id, answer, categoryID string) error {
	target, err := s.repos.UnknownQ.GetByIDAndOrg(ctx, id, userID)
	if err != nil {
		return err
	}
	if target == nil {
		return apperrors.ErrUnknownQuestion
	}
	qa := &domain.QAPair{
		UserID:     target.UserID,
		OrgID:      userID,
		CategoryID: categoryID,
		Question:   target.Question,
		Answer:     answer,
		IsActive:   true,
	}
	if err := s.repos.QAPair.Create(ctx, qa); err != nil {
		return err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(userID)
	}
	return s.repos.UnknownQ.UpdateStatus(ctx, id, userID, "trained", &answer, &categoryID)
}

func (s *TrainingService) IgnoreUnknown(ctx context.Context, userID, id string) error {
	if err := s.repos.UnknownQ.UpdateStatus(ctx, id, userID, "ignored", nil, nil); err != nil {
		return err
	}
	return nil
}

func (s *TrainingService) ListQAPairs(ctx context.Context, userID, categoryID string) ([]domain.QAPair, error) {
	return s.repos.QAPair.ListByCategoryAndOrg(ctx, categoryID, userID)
}

func (s *TrainingService) CreateQAPair(ctx context.Context, userID, categoryID, question, answer string) (*domain.QAPair, error) {
	qa := &domain.QAPair{
		UserID:     userID,
		OrgID:      userID,
		CategoryID: categoryID,
		Question:   question,
		Answer:     answer,
		IsActive:   true,
	}
	if err := s.repos.QAPair.Create(ctx, qa); err != nil {
		return nil, err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(userID)
	}
	return qa, nil
}

func (s *TrainingService) UpdateQAPair(ctx context.Context, userID, qaID, categoryID, question, answer string) error {
	qa := &domain.QAPair{
		ID:         qaID,
		UserID:     userID,
		OrgID:      userID,
		CategoryID: categoryID,
		Question:   question,
		Answer:     answer,
		IsActive:   true,
	}
	if err := s.repos.QAPair.Update(ctx, qa); err != nil {
		return err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(userID)
	}
	return nil
}

func (s *TrainingService) DeleteQAPair(ctx context.Context, userID, qaID string) error {
	if err := s.repos.QAPair.Delete(ctx, qaID, userID); err != nil {
		return err
	}
	if s.embeddings != nil {
		s.embeddings.InvalidateCache(userID)
	}
	return nil
}

func (s *TrainingService) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	return s.repos.Category.Delete(ctx, categoryID, userID)
}

func (s *TrainingService) SearchQAPairs(ctx context.Context, userID, query string) ([]domain.QAPair, error) {
	return s.repos.QAPair.Search(ctx, userID, query)
}
