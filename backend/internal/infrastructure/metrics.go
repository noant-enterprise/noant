package infrastructure

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "noant_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	AICallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_ai_calls_total",
		Help: "Total AI API calls",
	}, []string{"model", "status"})

	AIDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "noant_ai_duration_seconds",
		Help:    "AI response duration",
		Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"model"})

	DBConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "noant_db_connections",
		Help: "Current database connections",
	})

	RedisConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "noant_redis_connections",
		Help: "Current Redis connections",
	})

	AISentimentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_ai_sentiment_total",
		Help: "AI response sentiment distribution",
	}, []string{"sentiment"})

	AILanguageTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_ai_language_total",
		Help: "AI response language distribution",
	}, []string{"language"})

	CSATScore = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "noant_csat_score",
		Help:    "CSAT rating distribution (1-5)",
		Buckets: []float64{1, 2, 3, 4, 5},
	})
)
