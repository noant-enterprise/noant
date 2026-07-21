package service

import (
	"net/http"
	"sync"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== AI BRAIN SERVICE ==========

type CircuitBreaker struct {
	failures    int
	lastFailure time.Time
	state       string // closed, open, half-open
	mutex       sync.RWMutex
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	switch cb.state {
	case "open":
		if time.Since(cb.lastFailure) > 60*time.Second {
			cb.state = "half-open"
			cb.failures = 0
			return true
		}
		return false
	case "half-open":
		return true
	default: // closed
		return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.failures = 0
	cb.state = "closed"
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= circuitBreakerThreshold {
		cb.state = "open"
	}
}

type AIBrain struct {
	cfg         *config.Config
	repos       *repository.Repositories
	redis       *infrastructure.RedisClient
	logger      *infrastructure.Logger
	keyIndex    int
	keyMutex    sync.RWMutex
	cb          *CircuitBreaker
	broadcastFn func(convID string, msgType string, data interface{})
	embeddings  *EmbeddingService
	planSvc     *PlanService
	httpClient  *http.Client
}

// NewAIBrain creates the core AI orchestration engine. It manages intent classification,
// 3-tier semantic search (threshold 0.65 → category fallback → threshold 0.4),
// Groq-powered response humanization, sentiment analysis, and circuit breaker
// protection against API failures. The broadcastFn callback pushes real-time
// events to the WebSocket hub.
func NewAIBrain(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, broadcastFn func(convID string, msgType string, data interface{})) *AIBrain {
	transport := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	return &AIBrain{
		cfg:         cfg,
		repos:       repos,
		redis:       redis,
		logger:      logger,
		keyIndex:    0,
		cb:          &CircuitBreaker{state: "closed"},
		broadcastFn: broadcastFn,
		embeddings:  NewEmbeddingService(cfg, repos, redis, logger),
		planSvc:     NewPlanService(cfg, repos, redis, logger, NewCreditService(cfg, repos, redis, logger)),
		httpClient:  &http.Client{Transport: transport, Timeout: 30 * time.Second},
	}
}

type AIResponse struct {
	Content     string            `json:"content"`
	Confidence  float64           `json:"confidence"`
	Escalate    bool              `json:"escalate"`
	Source      string            `json:"source,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	MatchedQA   *string           `json:"matched_qa,omitempty"`
	Sentiment   string            `json:"sentiment,omitempty"`   // positive, negative, neutral, frustrated
	Language    string            `json:"language,omitempty"`    // en, yo, ha, ig, pcm
	Suggestions []string          `json:"suggestions,omitempty"` // quick action chips
	Summary     string            `json:"summary,omitempty"`     // conversation summary (set on handoff)
}
