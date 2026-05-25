package handler

import (
	"net/http"
	"time"

	"database/sql"
	"strings"
	"noant/internal/infrastructure"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db        *sql.DB
	redis     *infrastructure.RedisClient
	groqKeys  []string
	logger    *infrastructure.Logger
}

func NewHealthHandler(db *sql.DB, redis *infrastructure.RedisClient, groqKeys []string, logger *infrastructure.Logger) *HealthHandler {
	return &HealthHandler{db: db, redis: redis, groqKeys: groqKeys, logger: logger}
}

func (h *HealthHandler) Check(c *gin.Context) {
	checks := map[string]string{
		"api":     "healthy",
		"version": "2.0.0",
	}

	// Check DB
	if h.db != nil {
		if err := h.db.PingContext(c.Request.Context()); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
		} else {
			checks["database"] = "healthy"
		}
	} else {
		checks["database"] = "not_configured"
	}

	// Check Redis (optional - app works with in-memory fallback)
	if h.redis != nil {
		if err := h.redis.Ping(c.Request.Context()); err != nil {
			checks["redis"] = "degraded: " + err.Error()
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	// Check Groq API key (lightweight - validate format, not call API)
	if len(h.groqKeys) > 0 && h.groqKeys[0] != "" {
		// Check first key looks valid (starts with gsk_)
		validKey := false
		for _, key := range h.groqKeys {
			if len(key) > 20 && key[:4] == "gsk_" {
				validKey = true
				break
			}
		}
		if validKey {
			checks["groq"] = "healthy"
		} else {
			checks["groq"] = "unhealthy: invalid key format"
		}
	} else {
		checks["groq"] = "unhealthy: no API keys configured"
	}

	overall := "healthy"
	criticalFailed := false
	for k, v := range checks {
		if k == "version" {
			continue
		}
		if v != "healthy" && v != "not_configured" && !strings.HasPrefix(v, "degraded:") {
			// Only DB and API are critical; Redis/Groq degradation is tolerated
			if k == "database" || k == "api" {
				criticalFailed = true
			}
			overall = "degraded"
		}
	}

	status := http.StatusOK
	if criticalFailed {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":    overall,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    checks,
	})
}
