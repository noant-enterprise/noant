package infrastructure

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "noant_active_connections",
		Help: "Number of active HTTP connections",
	})

	AIResponseTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_ai_response_total",
		Help: "Total AI responses",
	}, []string{"status"})

	GroqAPICallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_groq_api_calls_total",
		Help: "Total Groq API calls",
	}, []string{"status"})

	CreditUsageTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_credit_usage_total",
		Help: "Total credits used",
	}, []string{"operation"})
)

var activeConns int64

func recordConnectionStart() {
	atomic.AddInt64(&activeConns, 1)
	ActiveConnections.Set(float64(atomic.LoadInt64(&activeConns)))
}

func recordConnectionEnd() {
	atomic.AddInt64(&activeConns, -1)
	ActiveConnections.Set(float64(atomic.LoadInt64(&activeConns)))
}

// PrometheusMiddleware instruments all HTTP requests with request count, duration,
// and active connection tracking.
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		recordConnectionStart()
		defer recordConnectionEnd()

		start := time.Now()

		c.Next()

		elapsed := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		RequestsTotal.WithLabelValues(method, path, status).Inc()
		RequestDuration.WithLabelValues(method, path).Observe(elapsed)
	}
}

// RecordAIResponse increments the AI response counter with the given status.
func RecordAIResponse(success bool) {
	if success {
		AIResponseTotal.WithLabelValues("success").Inc()
	} else {
		AIResponseTotal.WithLabelValues("failure").Inc()
	}
}

// RecordGroqAPICall increments the Groq API call counter with the given status.
func RecordGroqAPICall(status string) {
	GroqAPICallsTotal.WithLabelValues(status).Inc()
}

// RecordCreditUsage increments the credit usage counter for the given operation.
func RecordCreditUsage(operation string) {
	CreditUsageTotal.WithLabelValues(operation).Inc()
}
