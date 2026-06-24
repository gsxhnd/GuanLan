package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func scanIndexDataset(row interface {
	Scan(dest ...any) error
}) (IndexDataset, error) {
	var item IndexDataset
	var lastSync sql.NullTime
	if err := row.Scan(
		&item.IndexCode,
		&item.Market,
		&item.IndexName,
		&item.DataCompleteness,
		&lastSync,
		&item.SyncStatus,
	); err != nil {
		return IndexDataset{}, err
	}
	if lastSync.Valid {
		t := lastSync.Time
		item.LastSyncTime = &t
	}
	return item, nil
}

// ListIndexDatasets 返回预置训练指数列表（GET /api/data/indexes）。
func (s *Store) ListIndexDatasets(ctx context.Context) ([]IndexDataset, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT index_code, market, index_name, data_completeness, last_sync_time, sync_status
		FROM index_datasets
		ORDER BY index_code
	`)
	if err != nil {
		return nil, fmt.Errorf("list index datasets: %w", err)
	}
	defer rows.Close()

	var out []IndexDataset
	for rows.Next() {
		item, err := scanIndexDataset(rows)
		if err != nil {
			return nil, fmt.Errorf("scan index dataset: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanIndexConstituent(row interface {
	Scan(dest ...any) error
}) (IndexConstituent, error) {
	var item IndexConstituent
	var weight sql.NullFloat64
	if err := row.Scan(
		&item.IndexCode,
		&item.StockCode,
		&item.SnapDate,
		&weight,
		&item.IsActive,
	); err != nil {
		return IndexConstituent{}, err
	}
	if weight.Valid {
		v := weight.Float64
		item.Weight = &v
	}
	return item, nil
}

// ListIndexConstituents 返回指数成分股快照（GET /api/data/indexes/{index_code}/constituents）。
func (s *Store) ListIndexConstituents(ctx context.Context, indexCode string) ([]IndexConstituent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT index_code, stock_code, snap_date, weight, is_active
		FROM index_constituents
		WHERE index_code = ?
		ORDER BY stock_code
	`, indexCode)
	if err != nil {
		return nil, fmt.Errorf("list index constituents: %w", err)
	}
	defer rows.Close()

	var out []IndexConstituent
	for rows.Next() {
		item, err := scanIndexConstituent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan index constituent: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// UpsertIndexDataset 写入或更新指数数据集状态。
func (s *Store) UpsertIndexDataset(ctx context.Context, item IndexDataset) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO index_datasets (
			index_code, market, index_name, data_completeness, last_sync_time, sync_status
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (index_code) DO UPDATE SET
			market = excluded.market,
			index_name = excluded.index_name,
			data_completeness = excluded.data_completeness,
			last_sync_time = excluded.last_sync_time,
			sync_status = excluded.sync_status
	`, item.IndexCode, item.Market, item.IndexName, item.DataCompleteness, item.LastSyncTime, item.SyncStatus)
	if err != nil {
		return fmt.Errorf("upsert index dataset: %w", err)
	}
	return nil
}

// UpsertIndexConstituent 写入或更新成分股快照。
func (s *Store) UpsertIndexConstituent(ctx context.Context, item IndexConstituent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO index_constituents (index_code, stock_code, snap_date, weight, is_active)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (index_code, stock_code, snap_date) DO UPDATE SET
			weight = excluded.weight,
			is_active = excluded.is_active
	`, item.IndexCode, item.StockCode, item.SnapDate, item.Weight, item.IsActive)
	if err != nil {
		return fmt.Errorf("upsert index constituent: %w", err)
	}
	return nil
}

// TouchIndexSyncTime 更新指数最近同步时间。
func (s *Store) TouchIndexSyncTime(ctx context.Context, indexCode string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE index_datasets SET last_sync_time = ? WHERE index_code = ?
	`, at, indexCode)
	if err != nil {
		return fmt.Errorf("touch index sync time: %w", err)
	}
	return nil
}
