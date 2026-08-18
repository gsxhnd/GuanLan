package data

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// SnapshotMeta is a row in the snapshots table.
type SnapshotMeta struct {
	SnapshotID   string
	SnapshotPath string
	Description  string
	RowCount     int64
	CreatedAt    time.Time
}

// ExportSnapshot writes daily_bars to a Parquet file and records metadata.
func (s *Store) ExportSnapshot(ctx context.Context, outDir, description string) (SnapshotMeta, error) {
	if outDir == "" {
		outDir = "data/snapshots"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return SnapshotMeta{}, fmt.Errorf("mkdir snapshots: %w", err)
	}

	id := uuid.NewString()
	path := filepath.Join(outDir, fmt.Sprintf("%s.parquet", id))
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	// DuckDB COPY TO parquet; path must be SQL-safe (uuid-based).
	_, err = s.db.ExecContext(ctx, fmt.Sprintf(`
		COPY (SELECT * FROM daily_bars ORDER BY stock_code, trade_date)
		TO '%s' (FORMAT PARQUET)
	`, stringsReplace(abs)))
	if err != nil {
		return SnapshotMeta{}, fmt.Errorf("export parquet: %w", err)
	}

	var rowCount int64
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM daily_bars`).Scan(&rowCount)

	meta := SnapshotMeta{
		SnapshotID:   id,
		SnapshotPath: abs,
		Description:  description,
		RowCount:     rowCount,
		CreatedAt:    time.Now().UTC(),
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO snapshots (snapshot_id, snapshot_path, description, row_count, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, meta.SnapshotID, meta.SnapshotPath, meta.Description, meta.RowCount, meta.CreatedAt)
	if err != nil {
		return SnapshotMeta{}, fmt.Errorf("insert snapshot meta: %w", err)
	}
	return meta, nil
}

func stringsReplace(path string) string {
	// Escape single quotes for SQL literal.
	out := make([]byte, 0, len(path))
	for i := 0; i < len(path); i++ {
		if path[i] == '\'' {
			out = append(out, '\'', '\'')
		} else {
			out = append(out, path[i])
		}
	}
	return string(out)
}

// LatestCloses returns the most recent close price per stock_code.
func (s *Store) LatestCloses(ctx context.Context) (map[string]float64, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT stock_code, close FROM daily_bars
		QUALIFY ROW_NUMBER() OVER (PARTITION BY stock_code ORDER BY trade_date DESC) = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("latest closes: %w", err)
	}
	defer rows.Close()

	out := map[string]float64{}
	for rows.Next() {
		var code string
		var close float64
		if err := rows.Scan(&code, &close); err != nil {
			return nil, err
		}
		out[code] = close
	}
	return out, rows.Err()
}

// UpsertAssetSnapshot writes/updates a portfolio asset snapshot row.
func (s *Store) UpsertAssetSnapshot(ctx context.Context, snap AssetSnapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO portfolio_asset_snapshots (
			snapshot_date, cash_balance, holding_market_value, total_asset, source
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (snapshot_date) DO UPDATE SET
			cash_balance = excluded.cash_balance,
			holding_market_value = excluded.holding_market_value,
			total_asset = excluded.total_asset,
			source = excluded.source
	`, snap.SnapshotDate, snap.CashBalance, snap.HoldingMarketValue, snap.TotalAsset, snap.Source)
	if err != nil {
		return fmt.Errorf("upsert asset snapshot: %w", err)
	}
	return nil
}

// ListAssetSnapshots returns portfolio asset snapshots in date range.
func (s *Store) ListAssetSnapshots(ctx context.Context, startDate, endDate *time.Time) ([]AssetSnapshot, error) {
	var (
		args  []any
		where []string
	)
	if startDate != nil {
		where = append(where, "snapshot_date >= ?")
		args = append(args, *startDate)
	}
	if endDate != nil {
		where = append(where, "snapshot_date <= ?")
		args = append(args, *endDate)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + where[0]
		for i := 1; i < len(where); i++ {
			whereSQL += " AND " + where[i]
		}
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT snapshot_date, cash_balance, holding_market_value, total_asset, source
		FROM portfolio_asset_snapshots `+whereSQL+`
		ORDER BY snapshot_date ASC
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list asset snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []AssetSnapshot
	for rows.Next() {
		var snap AssetSnapshot
		if err := rows.Scan(&snap.SnapshotDate, &snap.CashBalance, &snap.HoldingMarketValue, &snap.TotalAsset, &snap.Source); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, rows.Err()
}

// MarkStockSyncing sets sync_status=syncing for a stock (no-op if row missing).
func (s *Store) MarkStockSyncing(ctx context.Context, stockCode string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE stock_data_status SET sync_status = ? WHERE stock_code = ?
	`, StockStatusSyncing, stockCode)
	if err != nil {
		return fmt.Errorf("mark stock syncing: %w", err)
	}
	return nil
}
