package service

import (
	"context"
	"fmt"
	"strings"

	"noant/internal/domain"
)

func (b *AIBrain) searchKnowledgeBase(ctx context.Context, userID, query string, limit int) []domain.QAPair {
	// Try semantic search first (embeddings)
	if b.embeddings != nil {
		results, err := b.embeddings.SemanticSearchQAPairs(ctx, userID, query, limit, 0.65)
		if err == nil && len(results) > 0 {
			b.logger.Info("Semantic search found matches", "query", query, "results", len(results), "topScore", results[0].Score)
			qas := make([]domain.QAPair, len(results))
			for i, r := range results {
				qas[i] = domain.QAPair{
					ID:       r.ID,
					Question: r.Question,
					Answer:   r.Answer,
				}
			}
			return qas
		}
		if err != nil {
			b.logger.Warn("Semantic search failed, falling back to keyword", "error", err)
		}
	}

	// Fallback: keyword search (SQL LIKE)
	results, err := b.repos.QAPair.Search(ctx, userID, query)
	if err != nil {
		b.logger.Error("Failed to search Q&A pairs", "error", err)
		return nil
	}

	seen := make(map[string]struct{})
	merged := make([]domain.QAPair, 0, limit)
	for i := range results {
		if _, ok := seen[results[i].ID]; ok {
			continue
		}
		seen[results[i].ID] = struct{}{}
		merged = append(merged, results[i])
		if len(merged) >= limit {
			return merged
		}
	}

	// Word-by-word fallback
	for _, word := range strings.Fields(strings.ToLower(query)) {
		word = strings.Trim(word, "?!.,;:")
		if len(word) < 4 {
			continue
		}
		more, err := b.repos.QAPair.Search(ctx, userID, word)
		if err != nil {
			continue
		}
		for i := range more {
			if _, ok := seen[more[i].ID]; ok {
				continue
			}
			seen[more[i].ID] = struct{}{}
			merged = append(merged, more[i])
			if len(merged) >= limit {
				return merged
			}
		}
	}

	return merged
}

func (b *AIBrain) searchInventoryContext(ctx context.Context, userID, query string, limit int) []domain.InventoryItem {
	// Try semantic search first
	if b.embeddings != nil {
		results, err := b.embeddings.SemanticSearchInventory(ctx, userID, query, limit, 0.6)
		if err == nil && len(results) > 0 {
			var items []domain.InventoryItem
			for _, r := range results {
				item, err := b.repos.Inventory.GetByID(ctx, r.ID, userID)
				if err == nil && item != nil {
					items = append(items, *item)
				}
			}
			if len(items) > 0 {
				return items
			}
		}
	}

	// Fallback: keyword search
	items, err := b.repos.Inventory.Search(ctx, userID, query)
	if err != nil {
		b.logger.Error("Failed to search inventory", "error", err)
		return nil
	}

	seen := make(map[string]struct{})
	merged := make([]domain.InventoryItem, 0, limit)
	for i := range items {
		if _, ok := seen[items[i].ID]; ok {
			continue
		}
		seen[items[i].ID] = struct{}{}
		merged = append(merged, items[i])
		if len(merged) >= limit {
			return merged
		}
	}

	for _, word := range strings.Fields(strings.ToLower(query)) {
		word = strings.Trim(word, "?!.,;:")
		if len(word) < 4 {
			continue
		}
		more, err := b.repos.Inventory.Search(ctx, userID, word)
		if err != nil {
			continue
		}
		for i := range more {
			if _, ok := seen[more[i].ID]; ok {
				continue
			}
			seen[more[i].ID] = struct{}{}
			merged = append(merged, more[i])
			if len(merged) >= limit {
				return merged
			}
		}
	}

	return merged
}

// qaWordOverlap checks whether meaningful words in the query appear in the QA question.
// Returns 0.0-1.0 fraction of query words matched. Short queries (1-2 words) pass automatically.
func qaWordOverlap(query, question string) float64 {
	qWords := strings.Fields(strings.ToLower(query))
	if len(qWords) <= 2 {
		return 1.0 // short queries are direct intent signals
	}
	var meaningful []string
	for _, w := range qWords {
		w = strings.Trim(w, "?!.,;:")
		if len(w) < 3 {
			continue
		}
		meaningful = append(meaningful, w)
	}
	if len(meaningful) == 0 {
		return 1.0
	}
	qLower := strings.ToLower(question)
	matched := 0
	for _, w := range meaningful {
		if strings.Contains(qLower, w) {
			matched++
		}
	}
	return float64(matched) / float64(len(meaningful))
}

// findSimilarForUnknown suggests similar Q&As when escalating an unknown question
func (b *AIBrain) findSimilarForUnknown(ctx context.Context, userID, query string) []domain.QAPair {
	if b.embeddings == nil {
		return nil
	}
	similar, _ := b.embeddings.FindSimilarQA(ctx, userID, query, 3)
	return similar
}

// intentCategoryFallback uses LLM to classify a query into one of the user's categories,
// then returns the best QA pair from that category. Returns nil if no category matches.
func (b *AIBrain) intentCategoryFallback(ctx context.Context, userID, userQuery string) *domain.QAPair {
	categories, err := b.repos.Category.List(ctx, userID)
	if err != nil || len(categories) == 0 {
		return nil
	}

	catNames := make([]string, len(categories))
	catMap := make(map[string]domain.Category)
	for i, cat := range categories {
		catNames[i] = cat.Name
		catMap[strings.ToLower(strings.TrimSpace(cat.Name))] = cat
	}

	prompt := []MessageTurn{
		{Role: "system", Content: fmt.Sprintf(`You are a strict query classifier. Your only job is to output the name of exactly one category from the list below, or the word NONE.

Customer question: "%s"

Categories:
%s

Rules:
- The question MUST logically belong to the category based on what the customer is asking about
- Do NOT guess. If no category clearly matches, output NONE
- Output ONLY the exact category name or NONE
- No punctuation, no explanation, no extra characters`, userQuery, strings.Join(catNames, ", "))},
		{Role: "user", Content: userQuery},
	}

	response, _, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil {
		return nil
	}

	classified := strings.TrimSpace(response)
	classified = strings.TrimRight(classified, ".,;:!? \n\t")

	cat, ok := catMap[strings.ToLower(classified)]
	if !ok {
		return nil
	}

	qas, err := b.repos.QAPair.ListByCategoryAndUser(ctx, cat.ID, userID)
	if err != nil || len(qas) == 0 {
		return nil
	}

	return &qas[0]
}

// semanticFallback searches QA pairs with a lowered embedding threshold (0.4) to catch
// queries that are semantically related but didn't match the main search threshold.
func (b *AIBrain) semanticFallback(ctx context.Context, userID, userQuery string) *domain.QAPair {
	if b.embeddings == nil {
		return nil
	}
	results, err := b.embeddings.SemanticSearchQAPairs(ctx, userID, userQuery, 1, 0.4)
	if err != nil || len(results) == 0 {
		return nil
	}
	qa, err := b.repos.QAPair.GetByID(ctx, results[0].ID)
	if err != nil || qa == nil {
		return nil
	}
	return qa
}
