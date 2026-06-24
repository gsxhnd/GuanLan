package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func scanTask(row interface {
	Scan(dest ...any) error
}) (DataSyncTask, error) {
	var task DataSyncTask
	var startedAt, endedAt sql.NullTime
	var failureReason, logRef, dataVersion sql.NullString
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
		&dataVersion,
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
	if dataVersion.Valid {
		v := dataVersion.String
		task.DataVersion = &v
	}
	return task, nil
}

// CreateTask 创建统一任务记录。
func (s *Store) CreateTask(ctx context.Context, taskType TaskType, target string, trigger TriggerMethod, retryCount int) (DataSyncTask, error) {
	if target == "" {
		return DataSyncTask{}, fmt.Errorf("target_object is required")
	}
	if trigger == "" {
		trigger = TriggerManual
	}

	logRef := fmt.Sprintf("logs/tasks/%s.log", uuid.NewString()[:8])
	task := DataSyncTask{
		TaskID:        uuid.NewString(),
		TaskType:      taskType,
		TargetObject:  target,
		TriggerMethod: trigger,
		Status:        TaskStatusPending,
		CreatedAt:     time.Now().UTC(),
		RetryCount:    retryCount,
		LogRef:        &logRef,
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks (
			task_id, task_type, target_object, trigger_method, status,
			created_at, retry_count, log_ref
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, task.TaskID, task.TaskType, task.TargetObject, task.TriggerMethod,
		task.Status, task.CreatedAt, task.RetryCount, task.LogRef)
	if err != nil {
		return DataSyncTask{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}

// CreateStockSyncTask 为指定股票创建日频同步任务。
func (s *Store) CreateStockSyncTask(ctx context.Context, stockCode string, trigger TriggerMethod) (DataSyncTask, error) {
	if stockCode == "" {
		return DataSyncTask{}, fmt.Errorf("stock_code is required")
	}

	task, err := s.CreateTask(ctx, TaskTypeDataSync, stockCode, trigger, 0)
	if err != nil {
		return DataSyncTask{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE stock_data_status SET sync_status = ? WHERE stock_code = ?
	`, StockStatusSyncing, stockCode)
	if err != nil {
		return DataSyncTask{}, fmt.Errorf("mark stock syncing: %w", err)
	}
	return task, nil
}

// GetTask 按 ID 查询任务。
func (s *Store) GetTask(ctx context.Context, taskID string) (DataSyncTask, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT task_id, task_type, target_object, trigger_method, status,
		       created_at, started_at, ended_at, retry_count, failure_reason, log_ref, data_version
		FROM tasks WHERE task_id = ?
	`, taskID)
	task, err := scanTask(row)
	if err == sql.ErrNoRows {
		return DataSyncTask{}, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return DataSyncTask{}, fmt.Errorf("get task: %w", err)
	}
	return task, nil
}

// ListTasks 返回任务列表。
func (s *Store) ListTasks(ctx context.Context, taskType *TaskType, limit int) ([]DataSyncTask, error) {
	if limit <= 0 {
		limit = 50
	}

	var (
		args  []any
		where string
	)
	if taskType != nil {
		where = "WHERE task_type = ?"
		args = append(args, *taskType)
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT task_id, task_type, target_object, trigger_method, status,
		       created_at, started_at, ended_at, retry_count, failure_reason, log_ref, data_version
		FROM tasks %s
		ORDER BY created_at DESC
		LIMIT ?
	`, where)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var out []DataSyncTask
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		out = append(out, task)
	}
	return out, rows.Err()
}

// ListDataSyncTasks 返回数据同步任务记录。
func (s *Store) ListDataSyncTasks(ctx context.Context, limit int) ([]DataSyncTask, error) {
	t := TaskTypeDataSync
	return s.ListTasks(ctx, &t, limit)
}

// RetryTask 基于失败任务创建重试任务。
func (s *Store) RetryTask(ctx context.Context, taskID string) (DataSyncTask, error) {
	prev, err := s.GetTask(ctx, taskID)
	if err != nil {
		return DataSyncTask{}, err
	}
	if prev.Status != TaskStatusFailed {
		return DataSyncTask{}, fmt.Errorf("task %s is not failed", taskID)
	}

	return s.CreateTask(ctx, prev.TaskType, prev.TargetObject, TriggerRetry, prev.RetryCount+1)
}

// ClaimPendingTask 领取一个 pending 任务并标记为 running。
func (s *Store) ClaimPendingTask(ctx context.Context) (DataSyncTask, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DataSyncTask{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `
		SELECT task_id, task_type, target_object, trigger_method, status,
		       created_at, started_at, ended_at, retry_count, failure_reason, log_ref, data_version
		FROM tasks
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT 1
	`, TaskStatusPending)

	task, err := scanTask(row)
	if err == sql.ErrNoRows {
		return DataSyncTask{}, false, nil
	}
	if err != nil {
		return DataSyncTask{}, false, fmt.Errorf("claim pending task: %w", err)
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
		UPDATE tasks SET status = ?, started_at = ? WHERE task_id = ?
	`, TaskStatusRunning, now, task.TaskID)
	if err != nil {
		return DataSyncTask{}, false, fmt.Errorf("mark task running: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return DataSyncTask{}, false, err
	}

	task.Status = TaskStatusRunning
	task.StartedAt = &now
	return task, true, nil
}

// UpdateTaskStatus 更新任务状态与时间戳。
func (s *Store) UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus, failureReason *string, dataVersion *string) error {
	now := time.Now().UTC()
	var endedAt any
	if status == TaskStatusSuccess || status == TaskStatusFailed || status == TaskStatusPartialSuccess || status == TaskStatusCancelled {
		endedAt = now
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE tasks
		SET status = ?,
		    failure_reason = ?,
		    data_version = COALESCE(?, data_version),
		    started_at = COALESCE(started_at, ?),
		    ended_at = COALESCE(ended_at, ?)
		WHERE task_id = ?
	`, status, failureReason, dataVersion, now, endedAt, taskID)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	return nil
}

// UpdateDataSyncTaskStatus 兼容旧调用。
func (s *Store) UpdateDataSyncTaskStatus(ctx context.Context, taskID string, status TaskStatus, failureReason *string) error {
	return s.UpdateTaskStatus(ctx, taskID, status, failureReason, nil)
}

// ListPendingDataSyncTargets 返回数据底座股票池中待同步代码。
func (s *Store) ListPendingDataSyncTargets(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.yfinance_symbol
		FROM stock_pool p
		LEFT JOIN stock_data_status s ON s.stock_code = p.yfinance_symbol
		WHERE p.is_active = TRUE AND p.sync_daily = TRUE
		  AND (s.sync_status IS NULL OR s.sync_status IN (?, ?))
	`, StockStatusMissing, StockStatusSyncing)
	if err != nil {
		return nil, fmt.Errorf("list pending sync targets: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

// StockHasReadyData 判断 DuckDB 是否已有就绪日频数据。
func (s *Store) StockHasReadyData(ctx context.Context, stockCode string) (bool, error) {
	var status string
	err := s.db.QueryRowContext(ctx, `
		SELECT sync_status FROM stock_data_status WHERE stock_code = ?
	`, stockCode).Scan(&status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stock has ready data: %w", err)
	}
	return status == string(StockStatusReady), nil
}

func encodeTags(tags []string) (any, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func decodeTags(raw sql.NullString) ([]string, error) {
	if !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return nil, nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw.String), &tags); err != nil {
		return nil, err
	}
	return tags, nil
}
