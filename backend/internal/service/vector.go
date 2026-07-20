package service

import (
    "context"
    "strings"

    "noant/internal/domain"
    "noant/internal/repository"
)

type VectorSearch struct {
    repos*repository.Repositories
}

func NewVectorSearch(repos*repository.Repositories) *VectorSearch {
    return &VectorSearch{repos: repos}
}

// Search uses keyword matching as fallback until vector DB is integrated
func (v *VectorSearch) Search(ctx context.Context, userID, query string, limit int) ([]domain.QAPair, error) {
    // TODO: Integrate Pinecone/Weaviate for semantic search
    results, err := v.repos.QAPair.Search(ctx, userID, query)
    if err != nil {
        return nil, err
    }
    
	// If few results, try a combined word search (single query, not N+1)
	if len(results) < limit {
		words := strings.Fields(query)
		var longWords []string
		for _, w := range words {
			if len(w) >= 4 {
				longWords = append(longWords, w)
			}
		}
		if len(longWords) > 0 {
			combined := strings.Join(longWords, " ")
			more, _ := v.repos.QAPair.Search(ctx, userID, combined)
			for i := range more {
				exists := false
				for j := range results {
					if results[j].ID == more[i].ID {
						exists = true
						break
					}
				}
				if !exists {
					results = append(results, more[i])
				}
			}
		}
	}
    
    if len(results) > limit {
        results = results[:limit]
    }
    
    return results, nil
}
