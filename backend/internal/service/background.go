package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"noant/internal/infrastructure"
)

type WorkerTaskStatus string

const (
	WorkerTaskPending   WorkerTaskStatus = "pending"
	WorkerTaskRunning   WorkerTaskStatus = "running"
	WorkerTaskCompleted WorkerTaskStatus = "completed"
	WorkerTaskFailed    WorkerTaskStatus = "failed"
)

type WorkerTask struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Status     WorkerTaskStatus  `json:"status"`
	Payload    map[string]interface{} `json:"payload"`
	Result     interface{}       `json:"result,omitempty"`
	Error      string            `json:"error,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	StartedAt  *time.Time        `json:"started_at,omitempty"`
	DoneAt     *time.Time        `json:"done_at,omitempty"`
}

type BackgroundWorker struct {
	mu            sync.RWMutex
	tasks         map[string]*WorkerTask
	taskHistory   []*WorkerTask
	maxHistory    int
	workerCount   int
	taskCh        chan func()
	stopCh        chan struct{}
	wg            sync.WaitGroup
	logger        *infrastructure.Logger
	dbManager     *DBManagerService
}

func NewBackgroundWorker(logger *infrastructure.Logger, dbManager *DBManagerService, workerCount int) *BackgroundWorker {
	if workerCount <= 0 {
		workerCount = 3
	}
	bw := &BackgroundWorker{
		tasks:       make(map[string]*WorkerTask),
		taskHistory: make([]*WorkerTask, 0),
		maxHistory:  100,
		workerCount: workerCount,
		taskCh:      make(chan func(), 100),
		stopCh:      make(chan struct{}),
		logger:      logger,
		dbManager:   dbManager,
	}
	for i := 0; i < workerCount; i++ {
		bw.wg.Add(1)
		go bw.worker(i)
	}
	return bw
}

func (bw *BackgroundWorker) worker(id int) {
	defer bw.wg.Done()
	bw.logger.Info("Background worker started", "worker_id", id)
	for {
		select {
		case <-bw.stopCh:
			bw.logger.Info("Background worker stopped", "worker_id", id)
			return
		case taskFn := <-bw.taskCh:
			taskFn()
		}
	}
}

func (bw *BackgroundWorker) Shutdown() {
	bw.logger.Info("Shutting down background worker")
	close(bw.stopCh)
	bw.wg.Wait()
	bw.logger.Info("Background worker shut down")
}

func generateTaskID() string {
	return fmt.Sprintf("bg_task_%d", time.Now().UnixNano())
}

func (bw *BackgroundWorker) SubmitTask(name string, payload map[string]interface{}) string {
	bw.mu.Lock()
	task := &WorkerTask{
		ID:        generateTaskID(),
		Name:      name,
		Status:    WorkerTaskPending,
		Payload:   payload,
		CreatedAt: time.Now(),
	}
	bw.tasks[task.ID] = task
	bw.mu.Unlock()

	bw.logger.Info("Task submitted", "task_id", task.ID, "name", name)

	bw.taskCh <- func() {
		bw.executeTask(task)
	}

	return task.ID
}

func (bw *BackgroundWorker) executeTask(task *WorkerTask) {
	bw.mu.Lock()
	task.Status = WorkerTaskRunning
	now := time.Now()
	task.StartedAt = &now
	bw.mu.Unlock()

	bw.logger.Info("Executing task", "task_id", task.ID, "name", task.Name)

	switch task.Name {
	case "db_cleanup_all":
		config := extractCleanupConfig(task.Payload)
		report := bw.dbManager.RunAllCleanups(context.Background(), config)
		bw.mu.Lock()
		task.Result = report
		task.Status = WorkerTaskCompleted
		bw.mu.Unlock()

	case "db_cleanup_task":
		taskName, _ := task.Payload["task_name"].(string)
		if taskName == "" {
			taskName = "all"
		}
		config := extractCleanupConfig(task.Payload)
		result := bw.dbManager.RunTask(context.Background(), taskName, config)
		bw.mu.Lock()
		if result.Status == "failed" {
			task.Status = WorkerTaskFailed
			task.Error = result.Error
		} else {
			task.Status = WorkerTaskCompleted
		}
		task.Result = result
		bw.mu.Unlock()

	default:
		bw.mu.Lock()
		task.Status = WorkerTaskFailed
		task.Error = fmt.Sprintf("unknown task: %s", task.Name)
		bw.mu.Unlock()
	}

	bw.mu.Lock()
	doneAt := time.Now()
	task.DoneAt = &doneAt
	bw.taskHistory = append(bw.taskHistory, task)
	if len(bw.taskHistory) > bw.maxHistory {
		bw.taskHistory = bw.taskHistory[len(bw.taskHistory)-bw.maxHistory:]
	}
	bw.mu.Unlock()

	bw.logger.Info("Task completed", "task_id", task.ID, "name", task.Name, "status", task.Status)
}

func extractCleanupConfig(payload map[string]interface{}) *CleanupConfig {
	cfg := DefaultCleanupConfig()
	if payload == nil {
		return cfg
	}
	if v, ok := payload["old_conversations_days"].(float64); ok {
		cfg.OldConversationsDays = int(v)
	}
	if v, ok := payload["abandoned_conversations_days"].(float64); ok {
		cfg.AbandonedConversationsDays = int(v)
	}
	if v, ok := payload["unknown_questions_days"].(float64); ok {
		cfg.UnknownQuestionsDays = int(v)
	}
	if v, ok := payload["handoffs_days"].(float64); ok {
		cfg.HandoffsDays = int(v)
	}
	if v, ok := payload["audit_logs_days"].(float64); ok {
		cfg.AuditLogsDays = int(v)
	}
	if v, ok := payload["notifications_days"].(float64); ok {
		cfg.NotificationsDays = int(v)
	}
	if v, ok := payload["inactive_integration_days"].(float64); ok {
		cfg.InactiveIntegrationDays = int(v)
	}
	if v, ok := payload["expired_trial_days"].(float64); ok {
		cfg.ExpiredTrialDays = int(v)
	}
	if v, ok := payload["credit_purchases_days"].(float64); ok {
		cfg.CreditPurchasesDays = int(v)
	}
	if v, ok := payload["completed_campaigns_days"].(float64); ok {
		cfg.CompletedCampaignsDays = int(v)
	}
	return cfg
}

func (bw *BackgroundWorker) GetTask(taskID string) *WorkerTask {
	bw.mu.RLock()
	defer bw.mu.RUnlock()
	task, exists := bw.tasks[taskID]
	if !exists {
		for _, t := range bw.taskHistory {
			if t.ID == taskID {
				return t
			}
		}
		return nil
	}
	cpy := *task
	return &cpy
}

func (bw *BackgroundWorker) ListTasks(limit int) []*WorkerTask {
	bw.mu.RLock()
	defer bw.mu.RUnlock()
	if limit <= 0 || limit > len(bw.taskHistory) {
		limit = len(bw.taskHistory)
	}
	result := make([]*WorkerTask, limit)
	for i := 0; i < limit; i++ {
		idx := len(bw.taskHistory) - 1 - i
		if idx < 0 {
			break
		}
		cpy := *bw.taskHistory[idx]
		result[i] = &cpy
	}
	return result
}

func (bw *BackgroundWorker) Stats() map[string]interface{} {
	bw.mu.RLock()
	defer bw.mu.RUnlock()
	stats := map[string]interface{}{
		"active_tasks": 0,
		"total_history": len(bw.taskHistory),
		"queued":       len(bw.taskCh),
	}
	for _, t := range bw.tasks {
		if t.Status == WorkerTaskRunning || t.Status == WorkerTaskPending {
			stats["active_tasks"] = stats["active_tasks"].(int) + 1
		}
	}
	return stats
}
