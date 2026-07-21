package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TracedDB wraps *sql.DB and records query metrics.
type TracedDB struct {
	db *sql.DB
}

// NewTracedDB creates a new TracedDB wrapper.
func NewTracedDB(db *sql.DB) *TracedDB {
	return &TracedDB{db: db}
}

// ExecContext wraps sql.DB.ExecContext with metrics recording.
func (t *TracedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := t.db.ExecContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	recordDBQuery("exec", err, duration)
	return result, err
}

// QueryContext wraps sql.DB.QueryContext with metrics recording.
func (t *TracedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.db.QueryContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	recordDBQuery("query", err, duration)
	return rows, err
}

// QueryRowContext wraps sql.DB.QueryRowContext with metrics recording.
func (t *TracedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := t.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	recordDBQuery("query_row", nil, duration)
	return row
}

// Ping wraps sql.DB.Ping with metrics recording.
func (t *TracedDB) Ping(ctx context.Context) error {
	start := time.Now()
	err := t.db.PingContext(ctx)
	duration := time.Since(start).Seconds()

	recordDBQuery("ping", err, duration)
	return err
}

// DB returns the underlying *sql.DB for operations not wrapped here.
func (t *TracedDB) DB() *sql.DB {
	return t.db
}

// Prometheus metrics for DB queries
var (
	DBQueryTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "noant_db_queries_total",
		Help: "Total database queries by operation and status",
	}, []string{"operation", "status"})

	DBQueryDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "noant_db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"operation"})
)

func init() {
	prometheus.MustRegister(DBQueryTotal, DBQueryDuration)
}

func recordDBQuery(operation string, err error, duration float64) {
	status := "success"
	if err != nil {
		status = "error"
	}
	DBQueryTotal.WithLabelValues(operation, status).Inc()
	DBQueryDuration.WithLabelValues(operation).Observe(duration)
}
