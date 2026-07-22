package service

import (
	"context"
	"testing"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

func newTestTrainingService() *TrainingService {
	repos := &repository.Repositories{
		Category: repository.NewMockCategoryRepo(),
		QAPair:   repository.NewMockQAPairRepo(),
		UnknownQ: repository.NewMockUnknownQuestionRepo(),
	}
	return &TrainingService{
		cfg:    &config.Config{},
		repos:  repos,
		logger: infrastructure.NewNullLogger(),
	}
}

func TestCreateCategory_Success(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	cat, err := svc.CreateCategory(ctx, "user-1", "Support", "Customer support questions", "#ff0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat.Name != "Support" {
		t.Errorf("expected name 'Support', got '%s'", cat.Name)
	}
	if cat.Description != "Customer support questions" {
		t.Errorf("expected description 'Customer support questions', got '%s'", cat.Description)
	}
	if cat.Color != "#ff0000" {
		t.Errorf("expected color '#ff0000', got '%s'", cat.Color)
	}
	if cat.UserID != "user-1" {
		t.Errorf("expected user_id 'user-1', got '%s'", cat.UserID)
	}
	if cat.ID == "" {
		t.Error("expected non-empty category ID")
	}
}

func TestCreateCategory_EmptyName(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	cat, err := svc.CreateCategory(ctx, "user-1", "", "Empty name", "#000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat.Name != "" {
		t.Errorf("expected empty name, got '%s'", cat.Name)
	}
}

func TestListCategories(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	cats, err := svc.ListCategories(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 0 {
		t.Errorf("expected 0 categories initially, got %d", len(cats))
	}

	_, _ = svc.CreateCategory(ctx, "user-1", "Sales", "Sales questions", "#aaa")
	_, _ = svc.CreateCategory(ctx, "user-1", "Support", "Support questions", "#bbb")

	cats, err = svc.ListCategories(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("expected 2 categories after creation, got %d", len(cats))
	}

	names := map[string]bool{}
	for _, c := range cats {
		names[c.Name] = true
	}
	if !names["Sales"] || !names["Support"] {
		t.Errorf("expected Sales and Support categories, got %v", names)
	}
}

func TestBulkImport(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	qaPairs := []domain.QAPair{
		{Question: "What is X?", Answer: "Answer X"},
		{Question: "How to Y?", Answer: "Do Y"},
		{Question: "Where is Z?", Answer: "Z is there"},
	}

	err := svc.BulkImport(ctx, "user-1", "cat-1", qaPairs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range qaPairs {
		if qaPairs[i].UserID != "user-1" {
			t.Errorf("pair %d: expected user_id 'user-1', got '%s'", i, qaPairs[i].UserID)
		}
		if qaPairs[i].CategoryID != "cat-1" {
			t.Errorf("pair %d: expected category_id 'cat-1', got '%s'", i, qaPairs[i].CategoryID)
		}
		if !qaPairs[i].IsActive {
			t.Errorf("pair %d: expected IsActive true", i)
		}
	}

	pairs, err := svc.repos.QAPair.ListByOrg(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}
	if len(pairs) != 3 {
		t.Errorf("expected 3 QA pairs stored, got %d", len(pairs))
	}
}

func TestUploadCSV_Success(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	csvData := []byte("category,question,answer\nGeneral,What is X?,Answer Y\nGeneral,How to Z?,Do A\n")

	count, err := svc.UploadCSV(ctx, "user-1", "cat-1", csvData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	cats, err := svc.ListCategories(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error listing cats: %v", err)
	}
	if len(cats) != 1 {
		t.Fatalf("expected 1 auto-created category, got %d", len(cats))
	}
	if cats[0].Name != "General" {
		t.Errorf("expected category name 'General', got '%s'", cats[0].Name)
	}

	pairs, err := svc.repos.QAPair.ListByOrg(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error listing pairs: %v", err)
	}
	if len(pairs) != 2 {
		t.Errorf("expected 2 QA pairs from CSV, got %d", len(pairs))
	}
}

func TestUploadCSV_TooShort(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	csvData := []byte("category,question,answer\n")

	_, err := svc.UploadCSV(ctx, "user-1", "cat-1", csvData)
	if err == nil {
		t.Fatal("expected error for CSV with only headers")
	}
	if err.Error() != "CSV must have at least a header and one data row" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestUploadCSV_InvalidFormat(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	csvData := []byte("cat,q\nonly,two\n")

	count, err := svc.UploadCSV(ctx, "user-1", "cat-1", csvData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 (1 data row processed), got %d", count)
	}

	pairs, err := svc.repos.QAPair.ListByOrg(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 QA pairs for malformed CSV, got %d", len(pairs))
	}
}

func TestClearUnknownQuestions(t *testing.T) {
	svc := newTestTrainingService()
	ctx := context.Background()

	uqs := []domain.UnknownQuestion{
		{UserID: "user-1", Question: "Q1", Status: "pending"},
		{UserID: "user-1", Question: "Q2", Status: "pending"},
		{UserID: "user-2", Question: "Q3", Status: "pending"},
	}
	for i := range uqs {
		_ = svc.repos.UnknownQ.Create(ctx, &uqs[i])
	}

	err := svc.ClearUnknownQuestions(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count1, _ := svc.repos.UnknownQ.CountByFilter(ctx, "user-1", "")
	if count1 != 0 {
		t.Errorf("expected 0 questions for user-1 after clear, got %d", count1)
	}

	count2, _ := svc.repos.UnknownQ.CountByFilter(ctx, "user-2", "")
	if count2 != 1 {
		t.Errorf("expected 1 question for user-2 (not cleared), got %d", count2)
	}
}
