package task_test

import (
	"context"
	"testing"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
	"github.com/gsxhnd/guanlan/internal/task"
)

func TestSchedulerRunsPendingTask(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := task.NewScheduler(store, &task.DataSyncExecutor{Store: store}, 50*time.Millisecond)
	sched.Start(ctx)
	defer sched.Stop()

	taskRec, err := store.CreateStockSyncTask(ctx, "AAPL", data.TriggerManual)
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
