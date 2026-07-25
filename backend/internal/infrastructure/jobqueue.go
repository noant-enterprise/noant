package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	sentry "github.com/getsentry/sentry-go"
)

// JobStatus represents the status of a background job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

// Job represents a unit of background work
type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Status     JobStatus              `json:"status"`
	Priority   int                    `json:"priority"`
	Retries    int                    `json:"retries"`
	MaxRetries int                    `json:"max_retries"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Error      string                 `json:"error,omitempty"`
	Result     string                 `json:"result,omitempty"`
}

// JobHandler processes a job
type JobHandler func(ctx context.Context, job *Job) error

// JobQueue manages background job processing with retries
type JobQueue struct {
	mu          sync.RWMutex
	jobs        map[string]*Job
	queue       []string
	handlers    map[string]JobHandler
	workerCount int
	workerWg    sync.WaitGroup
	stopCh      chan struct{}
	logger      *Logger
	redis       *RedisClient
}

// NewJobQueue creates a new background job queue
func NewJobQueue(logger *Logger, redis *RedisClient, workerCount int) *JobQueue {
	if workerCount <= 0 {
		workerCount = 5
	}
	jq := &JobQueue{
		jobs:        make(map[string]*Job),
		queue:       make([]string, 0),
		handlers:    make(map[string]JobHandler),
		workerCount: workerCount,
		stopCh:      make(chan struct{}),
		logger:      logger,
		redis:       redis,
	}
	for i := 0; i < workerCount; i++ {
		jq.workerWg.Add(1)
		go jq.worker(i)
	}
	return jq
}

// RegisterHandler registers a handler for a job type
func (jq *JobQueue) RegisterHandler(jobType string, handler JobHandler) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.handlers[jobType] = handler
}

// Enqueue adds a job to the queue
func (jq *JobQueue) Enqueue(jobType string, payload map[string]interface{}, opts ...JobOption) string {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job := &Job{
		ID:         generateJobID(),
		Type:       jobType,
		Payload:    payload,
		Status:     JobStatusPending,
		Priority:   0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	for _, opt := range opts {
		opt(job)
	}

	jq.jobs[job.ID] = job
	jq.queue = append(jq.queue, job.ID)
	jq.logger.Info("Job enqueued", "job_id", job.ID, "type", jobType, "priority", job.Priority)

	if jq.redis != nil {
		go func() {
			snapshot := *job
			data, _ := json.Marshal(snapshot)
			_ = jq.redis.Set(context.Background(), "job:"+job.ID, string(data), 24*time.Hour)
		}()
	}
	return job.ID
}

// JobOption configures a job
type JobOption func(*Job)

func WithPriority(p int) JobOption {
	return func(j *Job) { j.Priority = p }
}
func WithMaxRetries(n int) JobOption {
	return func(j *Job) { j.MaxRetries = n }
}

func (jq *JobQueue) worker(id int) {
	defer jq.workerWg.Done()
	for {
		select {
		case <-jq.stopCh:
			return
		default:
			job := jq.dequeue()
			if job == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			jq.processJob(job)
		}
	}
}

func (jq *JobQueue) dequeue() *Job {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	if len(jq.queue) == 0 {
		return nil
	}

	bestIdx := 0
	for i, id := range jq.queue {
		if jq.jobs[id].Priority > jq.jobs[jq.queue[bestIdx]].Priority {
			bestIdx = i
		}
	}
	jobID := jq.queue[bestIdx]
	jq.queue = append(jq.queue[:bestIdx], jq.queue[bestIdx+1:]...)
	job := jq.jobs[jobID]
	job.Status = JobStatusRunning
	job.UpdatedAt = time.Now()
	return job
}

func (jq *JobQueue) processJob(job *Job) {
	jq.mu.RLock()
	handler, exists := jq.handlers[job.Type]
	jq.mu.RUnlock()
	if !exists {
		jq.failJob(job, fmt.Sprintf("no handler for job type: %s", job.Type))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= job.MaxRetries; attempt++ {
		if attempt > 0 {
			job.Status = JobStatusRetrying
			job.UpdatedAt = time.Now()
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			time.Sleep(backoff)
		}
		err := handler(ctx, job)
		if err == nil {
			job.Status = JobStatusCompleted
			job.UpdatedAt = time.Now()
			job.Result = "success"
			jq.logger.Info("Job completed", "job_id", job.ID, "type", job.Type)
			if jq.redis != nil {
				go func() {
					snapshot := *job
					data, _ := json.Marshal(snapshot)
					_ = jq.redis.Set(context.Background(), "job:"+job.ID, string(data), 24*time.Hour)
				}()
			}
			return
		}
		lastErr = err
		jq.logger.Warn("Job attempt failed", "job_id", job.ID, "type", job.Type, "attempt", attempt, "error", err)
	}
	jq.failJob(job, lastErr.Error())
}

func (jq *JobQueue) failJob(job *Job, errMsg string) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	job.Status = JobStatusFailed
	job.Error = errMsg
	job.UpdatedAt = time.Now()
	job.Retries = job.MaxRetries
	jq.logger.Error("Job failed permanently", "job_id", job.ID, "type", job.Type, "error", errMsg)

	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "background_job",
		Message:  "job failed permanently",
		Level:    sentry.LevelError,
		Data: map[string]interface{}{
			"job_id": job.ID,
			"type":   job.Type,
			"error":  errMsg,
		},
	})

	if jq.redis != nil {
		snapshot := *job
		data, _ := json.Marshal(snapshot)
		_ = jq.redis.Set(context.Background(), "job:failed:"+job.ID, string(data), 7*24*time.Hour)
	}
}

func (jq *JobQueue) GetJobStatus(jobID string) *Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	job, exists := jq.jobs[jobID]
	if !exists {
		return nil
	}
	copied := *job
	return &copied
}

func (jq *JobQueue) Shutdown() {
	close(jq.stopCh)
	jq.workerWg.Wait()
	jq.logger.Info("Job queue shut down")
}

func (jq *JobQueue) Stats() map[string]int {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	stats := map[string]int{"queued": len(jq.queue), "total": len(jq.jobs), "running": 0, "completed": 0, "failed": 0, "retrying": 0}
	for _, job := range jq.jobs {
		switch job.Status {
		case JobStatusRunning:
			stats["running"]++
		case JobStatusCompleted:
			stats["completed"]++
		case JobStatusFailed:
			stats["failed"]++
		case JobStatusRetrying:
			stats["retrying"]++
		}
	}
	return stats
}

func generateJobID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "job_" + hex.EncodeToString(b)
}

// ScheduleRecurring schedules a recurring job
func (jq *JobQueue) ScheduleRecurring(jobType string, payload map[string]interface{}, interval time.Duration, opts ...JobOption) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				jq.logger.Error("Panic in recurring job scheduler", "job_type", jobType, "error", r)
			}
		}()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			jq.Enqueue(jobType, payload, opts...)
		}
	}()
}

// Predefined background job handlers
func HealthCheckHandler(integrationSvc interface{ Test(ctx context.Context, channel string, config map[string]interface{}) (bool, string) }) JobHandler {
	return func(ctx context.Context, job *Job) error {
		channels := []string{"telegram", "whatsapp", "facebook", "instagram"}
		for _, ch := range channels {
			ok, msg := integrationSvc.Test(ctx, ch, nil)
			if !ok {
				slog.Warn("channel health check failed",
					"channel", ch,
					"details", msg,
				)
			}
		}
		return nil
	}
}

func CacheCleanupHandler(cache *Cache) JobHandler {
	return func(ctx context.Context, job *Job) error {
		cache.Clear()
		return nil
	}
}

func NotificationBatchHandler(notifFn func(ctx context.Context, payload map[string]interface{}) error) JobHandler {
	return func(ctx context.Context, job *Job) error {
		return notifFn(ctx, job.Payload)
	}
}

func HandoffReminderHandler(handoffProcessor interface{ ProcessReminders(ctx context.Context) error }) JobHandler {
	return func(ctx context.Context, job *Job) error {
		return handoffProcessor.ProcessReminders(ctx)
	}
}

// CreditExpiryHandler checks for credits expiring in 3 days and sends notifications
func CreditExpiryHandler(creditSvc interface{ CheckAndNotifyExpiry(context.Context) error }) JobHandler {
	return func(ctx context.Context, job *Job) error {
		return creditSvc.CheckAndNotifyExpiry(ctx)
	}
}

// CampaignStartHandler processes campaigns that should start today
func CampaignStartHandler(campaignSvc interface{ ProcessStarting(context.Context) error }) JobHandler {
	return func(ctx context.Context, job *Job) error {
		return campaignSvc.ProcessStarting(ctx)
	}
}

// CampaignEndHandler processes campaigns that should end today
func CampaignEndHandler(campaignSvc interface{ ProcessEnding(context.Context) error }) JobHandler {
	return func(ctx context.Context, job *Job) error {
		return campaignSvc.ProcessEnding(ctx)
	}
}

// FreeWeeklyResetHandler resets the weekly free counter for free users
func FreeWeeklyResetHandler(planSvc interface{ GetFreeWeeklyUsage(context.Context, string) (int, error) }) JobHandler {
	return func(ctx context.Context, job *Job) error {
		return nil
	}
}

