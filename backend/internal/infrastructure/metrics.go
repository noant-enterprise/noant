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
)
