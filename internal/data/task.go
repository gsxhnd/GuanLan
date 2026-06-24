package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func scanDataSyncTask(row interface {
	Scan(dest ...any) error
}) (DataSyncTask, error) {
	var task DataSyncTask
	var startedAt, endedAt sql.NullTime
	var failureReason, logRef sql.NullString
	if err := row.Scan(
		&task.TaskID,
		&task.TaskType,
		&task.TargetObject,
		&task.TriggerMethod,
		&task.Status,
		&task.CreatedAt,
		&startedAt,
		&endedAt,
		&task.RetryCount,
		&failureReason,
		&logRef,
	); err != nil {
		return DataSyncTask{}, err
	}
	if startedAt.Valid {
		t := startedAt.Time
		task.StartedAt = &t
	}
	if endedAt.Valid {
		t := endedAt.Time
		task.EndedAt = &t
	}
	if failureReason.Valid {
		v := failureReason.String
		task.FailureReason = &v
	}
	if logRef.Valid {
		v := logRef.String
		task.LogRef = &v
	}
	return task, nil
}

// CreateStockSyncTask 为指定股票创建日频同步任务（POST /api/data/stocks/{stock_code}/sync）。
func (s *Store) CreateStockSyncTask(ctx context.Context, stockCode string, trigger TriggerMethod) (DataSyncTask, error) {
	if stockCode == "" {
		return DataSyncTask{}, fmt.Errorf("stock_code is required")
	}
	if trigger == "" {
		trigger = TriggerManual
	}

	task := DataSyncTask{
		TaskID:        uuid.NewString(),
		TaskType:      TaskTypeDataSync,
		TargetObject:  stockCode,
		TriggerMethod: trigger,
		Status:        TaskStatusPending,
		CreatedAt:     time.Now().UTC(),
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO data_sync_tasks (
			task_id, task_type, target_object, trigger_method, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`, task.TaskID, task.TaskType, task.TargetObject, task.TriggerMethod, task.Status, task.CreatedAt)
	if err != nil {
		return DataSyncTask{}, fmt.Errorf("create stock sync task: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE stock_data_status SET sync_status = ? WHERE stock_code = ?
	`, StockStatusSyncing, stockCode)
	if err != nil {
		return DataSyncTask{}, fmt.Errorf("mark stock syncing: %w", err)
	}

	return task, nil
}

// ListDataSyncTasks 返回数据同步任务记录（GET /api/data/tasks）。
func (s *Store) ListDataSyncTasks(ctx context.Context, limit int) ([]DataSyncTask, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT task_id, task_type, target_object, trigger_method, status,
		       created_at, started_at, ended_at, retry_count, failure_reason, log_ref
		FROM data_sync_tasks
		WHERE task_type = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, TaskTypeDataSync, limit)
	if err != nil {
		return nil, fmt.Errorf("list data sync tasks: %w", err)
	}
	defer rows.Close()

	var out []DataSyncTask
	for rows.Next() {
		task, err := scanDataSyncTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan data sync task: %w", err)
		}
		out = append(out, task)
	}
	return out, rows.Err()
}

// UpdateDataSyncTaskStatus 更新任务状态与时间戳。
func (s *Store) UpdateDataSyncTaskStatus(ctx context.Context, taskID string, status TaskStatus, failureReason *string) error {
	now := time.Now().UTC()
	var endedAt any
	if status == TaskStatusSuccess || status == TaskStatusFailed || status == TaskStatusPartialSuccess || status == TaskStatusCancelled {
		endedAt = now
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE data_sync_tasks
		SET status = ?,
		    failure_reason = ?,
		    started_at = COALESCE(started_at, ?),
		    ended_at = COALESCE(ended_at, ?)
		WHERE task_id = ?
	`, status, failureReason, now, endedAt, taskID)
	if err != nil {
		return fmt.Errorf("update data sync task status: %w", err)
	}
	return nil
}
