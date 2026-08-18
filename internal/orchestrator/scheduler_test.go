package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/gsxhnd/guanlan/internal/biz"
	"github.com/gsxhnd/guanlan/internal/data"
	"github.com/gsxhnd/guanlan/internal/orchestrator"
)

type stubExecutor struct {
	store *data.Store
}

func (e *stubExecutor) Run(ctx context.Context, task data.DataSyncTask) error {
	now := time.Now().UTC()
	version := "v-test"
	_ = e.store.UpsertStockDataStatus(ctx, data.StockDataStatus{
		StockCode:  task.TargetObject,
		StockName:  task.TargetObject,
		Market:     data.MarketUS,
		LastUpdate: &now,
		SyncStatus: data.StockStatusReady,
	})
	return e.store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusSuccess, nil, &version)
}

func TestSchedulerRunsPendingTask(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := orchestrator.NewScheduler(store, &stubExecutor{store: store}, 50*time.Millisecond)
	sched.Start(ctx)
	defer sched.Stop()

	bizSvc := biz.New(store)
	taskRec, err := bizSvc.Task.CreateStockSyncTask(ctx, "AAPL", data.TriggerManual)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		got, err := store.GetTask(ctx, taskRec.TaskID)
		if err != nil {
			t.Fatalf("get task: %v", err)
		}
		if got.Status == data.TaskStatusSuccess {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("task did not complete in time")
}
