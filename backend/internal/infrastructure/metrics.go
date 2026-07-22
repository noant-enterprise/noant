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

	OpenWAMessagesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_openwa_messages_sent_total",
		Help: "Total OpenWA messages sent",
	}, []string{"type", "status"})

	OpenWAQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "noant_openwa_queue_depth",
		Help: "Current OpenWA send queue depth",
	})

	NoantGroqRateLimited = promauto.NewCounter(prometheus.CounterOpts{
		Name: "noant_groq_rate_limited",
		Help: "Total number of Groq API calls rate limited",
	})

	// Fix 10: Queue health metrics
	OpenWAQueueDepthGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "noant_openwa_queue_depth_by_session",
		Help: "OpenWA queue depth by session",
	}, []string{"session_id"})

	OpenWAMessagesInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "noant_openwa_messages_in_flight",
		Help: "Messages currently being processed by workers",
	})

	OpenWACircuitBreakerOpen = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "noant_openwa_circuit_breaker_open",
		Help: "Whether a session circuit breaker is open (1) or closed (0)",
	}, []string{"session_id"})

	// Fix 2: Delivery tracking
	OpenWADeliveryStatusTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_openwa_delivery_status_total",
		Help: "WhatsApp message delivery status updates",
	}, []string{"status"})
)
