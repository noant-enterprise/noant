package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== EMBEDDING SERVICE ==========

type EmbeddingService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
	// In-memory embedding cache: userID -> []cachedEmbedding
	cache   map[string][]cachedEmbedding
	cacheMu sync.RWMutex
}

type cachedEmbedding struct {
	ID         string
	Text       string
	Embedding  []float32
	CreatedAt  time.Time
}

func NewEmbeddingService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *EmbeddingService {
	return &EmbeddingService{
		cfg:    cfg,
		repos:  repos,
		redis:  redis,
		logger: logger,
		cache:  make(map[string][]cachedEmbedding),
	}
}

// GenerateEmbedding calls Groq embedding API to convert text to vector
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	apiKey := s.getNextEmbeddingKey()
	if apiKey == "" {
		return nil, fmt.Errorf("no Groq API keys configured for embeddings")
	}

	payload := map[string]interface{}{
		"model": "text-embedding-3-small",
		"input": text,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/embeddings", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error: %s", string(body)[:min(200, len(body))])
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return result.Data[0].Embedding, nil
}

// GenerateEmbeddings batch generates embeddings for multiple texts
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	apiKey := s.getNextEmbeddingKey()
	if apiKey == "" {
		return nil, fmt.Errorf("no Groq API keys configured")
	}

	payload := map[string]interface{}{
		"model": "text-embedding-3-small",
		"input": texts,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/embeddings", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("batch embedding error: %s", string(body)[:min(200, len(body))])
	}

	var result struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Sort by index and extract embeddings
	sort.Slice(result.Data, func(i, j int) bool { return result.Data[i].Index < result.Data[j].Index })
	output := make([][]float32, len(texts))
	for _, d := range result.Data {
		if d.Index < len(output) {
			output[d.Index] = d.Embedding
		}
	}
	return output, nil
}

func (s *EmbeddingService) getNextEmbeddingKey() string {
	if len(s.cfg.GroqAPIKeys) == 0 {
		return ""
	}
	// Use a simple rotation for embedding keys
	idx := int(time.Now().UnixNano()) % len(s.cfg.GroqAPIKeys)
	return s.cfg.GroqAPIKeys[idx]
}

// ========== SEMANTIC SEARCH ==========

type SemanticResult struct {
	ID       string
	Score    float32
	Question string
	Answer   string
	Name     string
}

// CosineSimilarity computes cosine similarity between two vectors
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// SemanticSearchQAPairs searches Q&A pairs using embeddings + cosine similarity
func (s *EmbeddingService) SemanticSearchQAPairs(ctx context.Context, userID, query string, limit int, threshold float32) ([]SemanticResult, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		s.logger.Warn("Failed to generate query embedding, falling back to keyword search", "error", err)
		return nil, err
	}

	// Load all Q&A embeddings for this user (cached in memory)
	embeddings := s.getOrCreateCache(ctx, userID)
	if len(embeddings) == 0 {
		return nil, nil
	}

	// Compute similarity scores
	var results []SemanticResult
	for _, e := range embeddings {
		score := CosineSimilarity(queryEmbedding, e.Embedding)
		if score >= threshold {
			results = append(results, SemanticResult{
				ID:       e.ID,
				Score:    score,
				Question: e.Text,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	// Fetch full Q&A data
	for i, r := range results {
		qa, err := s.repos.QAPair.GetByID(ctx, r.ID)
		if err == nil && qa != nil {
			results[i].Question = qa.Question
			results[i].Answer = qa.Answer
		}
	}

	return results, nil
}

// SemanticSearchInventory searches inventory using embeddings
func (s *EmbeddingService) SemanticSearchInventory(ctx context.Context, userID, query string, limit int, threshold float32) ([]SemanticResult, error) {
	queryEmbedding, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	embeddings := s.getInventoryCache(ctx, userID)
	if len(embeddings) == 0 {
		return nil, nil
	}

	var results []SemanticResult
	for _, e := range embeddings {
		score := CosineSimilarity(queryEmbedding, e.Embedding)
		if score >= threshold {
			results = append(results, SemanticResult{
				ID:    e.ID,
				Score: score,
				Name:  e.Text,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// getOrCreateCache loads Q&A embeddings from DB or creates them
func (s *EmbeddingService) getOrCreateCache(ctx context.Context, userID string) []cachedEmbedding {
	s.cacheMu.RLock()
	if cached, ok := s.cache[userID]; ok && len(cached) > 0 {
		// Check if cache is less than 1 hour old
		if time.Since(cached[0].CreatedAt) < 1*time.Hour {
			s.cacheMu.RUnlock()
			return cached
		}
	}
	s.cacheMu.RUnlock()

	// Load from DB
	qas, err := s.repos.QAPair.ListByOrg(ctx, userID, "")
	if err != nil || len(qas) == 0 {
		return nil
	}

	// Batch generate embeddings
	texts := make([]string, len(qas))
	for i := range qas {
		texts[i] = qas[i].Question
		if len(qas[i].Answer) > 100 {
			texts[i] += " " + qas[i].Answer[:100]
		}
	}

	embeddings, err := s.GenerateEmbeddings(ctx, texts)
	if err != nil {
		s.logger.Warn("Failed to batch generate embeddings", "error", err)
		return nil
	}

	// Build cache
	now := time.Now()
	cached := make([]cachedEmbedding, 0, len(qas))
	for i := range qas {
		if i < len(embeddings) && embeddings[i] != nil {
			cached = append(cached, cachedEmbedding{
				ID:        qas[i].ID,
				Text:      qas[i].Question,
				Embedding: embeddings[i],
				CreatedAt: now,
			})
		}
	}

	s.cacheMu.Lock()
	s.cache[userID] = cached
	s.cacheMu.Unlock()

	return cached
}

func (s *EmbeddingService) getInventoryCache(ctx context.Context, userID string) []cachedEmbedding {
	s.cacheMu.RLock()
	key := userID + ":inventory"
	if cached, ok := s.cache[key]; ok && len(cached) > 0 {
		if time.Since(cached[0].CreatedAt) < 1*time.Hour {
			s.cacheMu.RUnlock()
			return cached
		}
	}
	s.cacheMu.RUnlock()

	items, err := s.repos.Inventory.List(ctx, userID, "", true)
	if err != nil || len(items) == 0 {
		return nil
	}

	texts := make([]string, len(items))
	for i := range items {
		texts[i] = items[i].Name
		if items[i].Description != "" {
			texts[i] += " " + items[i].Description
		}
	}

	embeddings, err := s.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return nil
	}

	now := time.Now()
	cached := make([]cachedEmbedding, 0, len(items))
	for i := range items {
		if i < len(embeddings) && embeddings[i] != nil {
			cached = append(cached, cachedEmbedding{
				ID:        items[i].ID,
				Text:      items[i].Name,
				Embedding: embeddings[i],
				CreatedAt: now,
			})
		}
	}

	s.cacheMu.Lock()
	s.cache[key] = cached
	s.cacheMu.Unlock()

	return cached
}

// InvalidateCache clears cached embeddings for a user (call when Q&A or inventory changes)
func (s *EmbeddingService) InvalidateCache(userID string) {
	s.cacheMu.Lock()
	delete(s.cache, userID)
	delete(s.cache, userID+":inventory")
	s.cacheMu.Unlock()
}

// FindSimilarQA finds the most similar Q&A pairs for a given query (used for unknown question suggestions)
func (s *EmbeddingService) FindSimilarQA(ctx context.Context, userID, query string, limit int) (qas []domain.QAPair, bestScore float32) {
	results, err := s.SemanticSearchQAPairs(ctx, userID, query, limit, 0.5)
	if err != nil || len(results) == 0 {
		return nil, 0
	}

	qas = make([]domain.QAPair, len(results))
	for i, r := range results {
		qas[i] = domain.QAPair{
			ID:       r.ID,
			Question: r.Question,
			Answer:   r.Answer,
		}
	}
	return qas, results[0].Score
}
