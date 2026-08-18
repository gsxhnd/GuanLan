package biz

import (
	"context"
	"fmt"

	"github.com/gsxhnd/guanlan/internal/data"
)

// CreateStockSyncTask creates a pending data_sync task and marks the stock syncing.
func (t *Task) CreateStockSyncTask(ctx context.Context, stockCode string, trigger data.TriggerMethod) (data.DataSyncTask, error) {
	if stockCode == "" {
		return data.DataSyncTask{}, fmt.Errorf("stock_code is required")
	}
	if trigger == "" {
		trigger = data.TriggerManual
	}
	task, err := t.Store.CreateTask(ctx, data.TaskTypeDataSync, stockCode, trigger, 0)
	if err != nil {
		return data.DataSyncTask{}, err
	}
	_ = t.Store.MarkStockSyncing(ctx, stockCode)
	return task, nil
}

// RetryTask recreates a failed task with incremented retry count.
func (t *Task) RetryTask(ctx context.Context, taskID string) (data.DataSyncTask, error) {
	prev, err := t.Store.GetTask(ctx, taskID)
	if err != nil {
		return data.DataSyncTask{}, err
	}
	if prev.Status != data.TaskStatusFailed {
		return data.DataSyncTask{}, fmt.Errorf("task %s is not failed", taskID)
	}
	task, err := t.Store.CreateTask(ctx, prev.TaskType, prev.TargetObject, data.TriggerRetry, prev.RetryCount+1)
	if err != nil {
		return data.DataSyncTask{}, err
	}
	if prev.TaskType == data.TaskTypeDataSync {
		_ = t.Store.MarkStockSyncing(ctx, prev.TargetObject)
	}
	return task, nil
}

// EnsureSyncIfNeeded creates a sync task when the stock has no ready data.
func (t *Task) EnsureSyncIfNeeded(ctx context.Context, stockCode string, trigger data.TriggerMethod) (created bool, task data.DataSyncTask, err error) {
	ready, err := t.Store.StockHasReadyData(ctx, stockCode)
	if err != nil {
		return false, data.DataSyncTask{}, err
	}
	if ready {
		return false, data.DataSyncTask{}, nil
	}
	task, err = t.CreateStockSyncTask(ctx, stockCode, trigger)
	if err != nil {
		return false, data.DataSyncTask{}, err
	}
	return true, task, nil
}

// EnqueueDailySync creates scheduled sync tasks for all daily-sync pool stocks.
func (t *Task) EnqueueDailySync(ctx context.Context) (int, error) {
	items, err := t.Store.ListStockPool(ctx, data.StockPoolListOptions{DailySyncOnly: true})
	if err != nil {
		return 0, err
	}
	n := 0
	for _, item := range items {
		if _, err := t.CreateStockSyncTask(ctx, item.YfinanceSymbol, data.TriggerScheduled); err != nil {
			continue
		}
		n++
	}
	return n, nil
}
