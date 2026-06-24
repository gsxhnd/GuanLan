package task

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

// Executor 执行任务的具体逻辑。
type Executor interface {
	Run(ctx context.Context, task data.DataSyncTask) error
}

// DataSyncExecutor Phase 2 占位执行器：标记成功并更新个股状态。
type DataSyncExecutor struct {
	Store *data.Store
}

func (e *DataSyncExecutor) Run(ctx context.Context, task data.DataSyncTask) error {
	now := time.Now().UTC()
	version := fmt.Sprintf("v%s", now.Format("20060102"))

	if err := e.Store.UpsertStockDataStatus(ctx, data.StockDataStatus{
		StockCode:    task.TargetObject,
		StockName:    task.TargetObject,
		Market:       data.MarketA,
		Completeness: 95.0,
		LastUpdate:   &now,
		SyncStatus:   data.StockStatusReady,
	}); err != nil {
		return err
	}

	return e.Store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusSuccess, nil, &version)
}

// Scheduler 轮询 pending 任务并执行。
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
	close(s.stopCh)
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
		log.Printf("scheduler claim task: %v", err)
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

// ScheduledSync 为活跃股票池创建定时 data_sync 任务。
func ScheduledSync(ctx context.Context, store *data.Store) error {
	codes, err := store.ListPendingDataSyncTargets(ctx)
	if err != nil {
		return err
	}
	for _, code := range codes {
		_, err := store.CreateStockSyncTask(ctx, code, data.TriggerScheduled)
		if err != nil {
			log.Printf("scheduled sync %s: %v", code, err)
		}
	}
	return nil
}

// StartScheduledTicker 启动定时触发器。
func StartScheduledTicker(ctx context.Context, store *data.Store, interval time.Duration) func() {
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				if err := ScheduledSync(ctx, store); err != nil {
					log.Printf("scheduled sync: %v", err)
				}
			}
		}
	}()

	return func() { close(done) }
}
