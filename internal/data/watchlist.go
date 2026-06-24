package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func scanWatchlistItem(row interface {
	Scan(dest ...any) error
}) (WatchlistItem, error) {
	var item WatchlistItem
	var tagsJSON sql.NullString
	var notes sql.NullString
	var removedAt, lastActionAt sql.NullTime
	var lastAction sql.NullString
	var syncStatus sql.NullString
	var completeness sql.NullFloat64

	if err := row.Scan(
		&item.StockCode,
		&item.Market,
		&tagsJSON,
		&item.Priority,
		&notes,
		&item.IsActive,
		&item.AddedAt,
		&removedAt,
		&item.Source,
		&lastAction,
		&lastActionAt,
		&syncStatus,
		&completeness,
	); err != nil {
		return WatchlistItem{}, err
	}

	tags, err := decodeTags(tagsJSON)
	if err != nil {
		return WatchlistItem{}, err
	}
	item.Tags = tags
	if notes.Valid {
		item.Notes = notes.String
	}
	if removedAt.Valid {
		t := removedAt.Time
		item.RemovedAt = &t
	}
	if lastAction.Valid {
		item.LastAction = lastAction.String
	}
	if lastActionAt.Valid {
		t := lastActionAt.Time
		item.LastActionAt = &t
	}
	if syncStatus.Valid {
		item.SyncStatus = syncStatus.String
	}
	if completeness.Valid {
		item.Completeness = completeness.Float64
	}
	return item, nil
}

// ListWatchlistItems 返回股票池列表。
func (s *Store) ListWatchlistItems(ctx context.Context, activeOnly bool) ([]WatchlistItem, error) {
	where := ""
	if activeOnly {
		where = "WHERE w.is_active = TRUE"
	}

	query := fmt.Sprintf(`
		SELECT
			w.stock_code, w.market, w.tags, w.priority, w.notes, w.is_active,
			w.added_at, w.removed_at, w.source, w.last_action, w.last_action_at,
			COALESCE(s.sync_status, 'missing'),
			COALESCE(s.completeness, 0)
		FROM watchlist_items w
		LEFT JOIN stock_data_status s ON s.stock_code = w.stock_code
		%s
		ORDER BY w.priority DESC, w.added_at DESC
	`, where)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list watchlist: %w", err)
	}
	defer rows.Close()

	var out []WatchlistItem
	for rows.Next() {
		item, err := scanWatchlistItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan watchlist item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// AddWatchlistItem 添加股票到池。
func (s *Store) AddWatchlistItem(ctx context.Context, item WatchlistItem) (WatchlistItem, error) {
	code := strings.TrimSpace(strings.ToUpper(item.StockCode))
	if code == "" {
		return WatchlistItem{}, fmt.Errorf("stock_code is required")
	}
	if item.Market == "" {
		return WatchlistItem{}, fmt.Errorf("market is required")
	}

	now := time.Now().UTC()
	if item.AddedAt.IsZero() {
		item.AddedAt = now
	}
	if item.Source == "" {
		item.Source = "manual"
	}
	item.StockCode = code
	item.LastAction = "add"
	item.LastActionAt = &now

	tagsJSON, err := encodeTags(item.Tags)
	if err != nil {
		return WatchlistItem{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO watchlist_items (
			stock_code, market, tags, priority, notes, is_active,
			added_at, removed_at, source, last_action, last_action_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?)
		ON CONFLICT (stock_code) DO UPDATE SET
			market = excluded.market,
			tags = excluded.tags,
			priority = excluded.priority,
			notes = excluded.notes,
			is_active = excluded.is_active,
			removed_at = NULL,
			source = excluded.source,
			last_action = excluded.last_action,
			last_action_at = excluded.last_action_at
	`, item.StockCode, item.Market, tagsJSON, item.Priority, nullString(item.Notes),
		item.IsActive, item.AddedAt, item.Source, item.LastAction, item.LastActionAt)
	if err != nil {
		return WatchlistItem{}, fmt.Errorf("add watchlist item: %w", err)
	}

	// 确保 stock_data_status 行存在
	_ = s.UpsertStockDataStatus(ctx, StockDataStatus{
		StockCode: code,
		StockName: code,
		Market:    item.Market,
		SyncStatus: StockStatusMissing,
	})

	items, err := s.ListWatchlistItems(ctx, false)
	if err != nil {
		return WatchlistItem{}, err
	}
	for _, w := range items {
		if w.StockCode == code {
			return w, nil
		}
	}
	return item, nil
}

// RemoveWatchlistItem 删除或停用股票池条目。
func (s *Store) RemoveWatchlistItem(ctx context.Context, stockCode string, hardDelete bool) (WatchlistItem, error) {
	code := strings.TrimSpace(strings.ToUpper(stockCode))
	if code == "" {
		return WatchlistItem{}, fmt.Errorf("stock_code is required")
	}

	now := time.Now().UTC()
	if hardDelete {
		_, err := s.db.ExecContext(ctx, `DELETE FROM watchlist_items WHERE stock_code = ?`, code)
		if err != nil {
			return WatchlistItem{}, fmt.Errorf("delete watchlist item: %w", err)
		}
		return WatchlistItem{StockCode: code, LastAction: "remove", LastActionAt: &now}, nil
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE watchlist_items
		SET is_active = FALSE,
		    removed_at = ?,
		    last_action = 'disable',
		    last_action_at = ?
		WHERE stock_code = ?
	`, now, now, code)
	if err != nil {
		return WatchlistItem{}, fmt.Errorf("disable watchlist item: %w", err)
	}

	items, err := s.ListWatchlistItems(ctx, false)
	if err != nil {
		return WatchlistItem{}, err
	}
	for _, w := range items {
		if w.StockCode == code {
			return w, nil
		}
	}
	return WatchlistItem{StockCode: code, IsActive: false, LastAction: "disable", LastActionAt: &now}, nil
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
