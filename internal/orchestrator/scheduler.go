package orchestrator

import (
	"context"
	"log"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

// Executor runs a claimed task.
type Executor interface {
	Run(ctx context.Context, task data.DataSyncTask) error
}

// Scheduler polls pending tasks and dispatches them to an Executor.
type Scheduler struct {
	Store    *data.Store
	Executor Executor
	Interval time.Duration
	stopCh   chan struct{}
}

func NewScheduler(store *data.Store, exec Executor, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &Scheduler{
		Store:    store,
		Executor: exec,
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go s.loop(ctx)
}

func (s *Scheduler) Stop() {
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	task, ok, err := s.Store.ClaimPendingTask(ctx)
	if err != nil {
		log.Printf("orchestrator claim task: %v", err)
		return
	}
	if !ok {
		return
	}

	if err := s.Executor.Run(ctx, task); err != nil {
		reason := err.Error()
		_ = s.Store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusFailed, &reason, nil)
		log.Printf("task %s failed: %v", task.TaskID, err)
	}
}
