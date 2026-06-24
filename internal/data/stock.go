package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

func scanStockListItem(row interface {
	Scan(dest ...any) error
}) (StockListItem, error) {
	var item StockListItem
	var trainingIndex, missingJSON sql.NullString
	var dataStart, dataEnd, lastUpdate sql.NullTime
	if err := row.Scan(
		&item.StockCode,
		&item.StockName,
		&item.Market,
		&trainingIndex,
		&dataStart,
		&dataEnd,
		&item.Completeness,
		&missingJSON,
		&lastUpdate,
		&item.SyncStatus,
		&item.Open,
		&item.High,
		&item.Low,
		&item.Close,
		&item.Volume,
		&item.Change,
	); err != nil {
		return StockListItem{}, err
	}
	if trainingIndex.Valid {
		v := trainingIndex.String
		item.TrainingIndexCode = &v
	}
	if dataStart.Valid {
		t := dataStart.Time
		item.DataStartDate = &t
	}
	if dataEnd.Valid {
		t := dataEnd.Time
		item.DataEndDate = &t
	}
	if lastUpdate.Valid {
		t := lastUpdate.Time
		item.LastUpdate = &t
	}
	if missingJSON.Valid && missingJSON.String != "" {
		if err := json.Unmarshal([]byte(missingJSON.String), &item.MissingRanges); err != nil {
			return StockListItem{}, fmt.Errorf("decode missing_ranges: %w", err)
		}
	}
	return item, nil
}

// ListStocks 返回个股数据状态列表，附带最近一根日 K（GET /api/data/stocks）。
func (s *Store) ListStocks(ctx context.Context, filter ListStocksFilter) ([]StockListItem, error) {
	var (
		args  []any
		where []string
	)

	if filter.Market != nil {
		where = append(where, "p.market = ?")
		args = append(args, *filter.Market)
	}
	if filter.Status != nil {
		where = append(where, "COALESCE(s.sync_status, 'missing') = ?")
		args = append(args, *filter.Status)
	}
	if q := strings.TrimSpace(filter.Search); q != "" {
		where = append(where, "(LOWER(p.yfinance_symbol) LIKE ? OR LOWER(p.original_code) LIKE ? OR LOWER(p.stock_name) LIKE ?)")
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like, like)
	}

	where = append(where, "p.is_active = TRUE")
	whereSQL := "WHERE " + strings.Join(where, " AND ")

	orderBy := "p.yfinance_symbol"
	switch filter.Sort {
	case "name":
		orderBy = "p.stock_name"
	case "change":
		orderBy = "change_pct DESC"
	case "volume":
		orderBy = "latest.volume DESC NULLS LAST"
	}

	query := fmt.Sprintf(`
		WITH latest AS (
			SELECT
				stock_code,
				open,
				high,
				low,
				close,
				volume,
				trade_date,
				LAG(close) OVER (PARTITION BY stock_code ORDER BY trade_date) AS prev_close
			FROM daily_bars
		),
		latest_bar AS (
			SELECT * FROM latest
			QUALIFY ROW_NUMBER() OVER (PARTITION BY stock_code ORDER BY trade_date DESC) = 1
		)
		SELECT
			p.yfinance_symbol,
			COALESCE(NULLIF(s.stock_name, ''), p.stock_name),
			p.market,
			s.training_index_code,
			s.data_start_date,
			s.data_end_date,
			COALESCE(s.completeness, 0),
			s.missing_ranges,
			s.last_update,
			COALESCE(s.sync_status, 'missing'),
			COALESCE(latest.open, 0),
			COALESCE(latest.high, 0),
			COALESCE(latest.low, 0),
			COALESCE(latest.close, 0),
			COALESCE(latest.volume, 0),
			CASE
				WHEN latest.prev_close IS NULL OR latest.prev_close = 0 THEN 0
				ELSE (latest.close - latest.prev_close) / latest.prev_close * 100
			END AS change_pct
		FROM stock_pool p
		LEFT JOIN stock_data_status s ON s.stock_code = p.yfinance_symbol
		LEFT JOIN latest_bar latest ON latest.stock_code = p.yfinance_symbol
		%s
		ORDER BY %s
	`, whereSQL, orderBy)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list stocks: %w", err)
	}
	defer rows.Close()

	var out []StockListItem
	for rows.Next() {
		item, err := scanStockListItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan stock list item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanDailyBar(row interface {
	Scan(dest ...any) error
}) (DailyBar, error) {
	var bar DailyBar
	var amount, adjFactor sql.NullFloat64
	if err := row.Scan(
		&bar.StockCode,
		&bar.Market,
		&bar.TradeDate,
		&bar.Open,
		&bar.High,
		&bar.Low,
		&bar.Close,
		&bar.Volume,
		&amount,
		&adjFactor,
		&bar.Source,
		&bar.DataVersion,
	); err != nil {
		return DailyBar{}, err
	}
	if amount.Valid {
		v := amount.Float64
		bar.Amount = &v
	}
	if adjFactor.Valid {
		v := adjFactor.Float64
		bar.AdjFactor = &v
	}
	return bar, nil
}

// ListDailyBars 返回个股日频行情（GET /api/data/stocks/{stock_code}/daily-bars）。
func (s *Store) ListDailyBars(ctx context.Context, params ListDailyBarsParams) ([]DailyBar, error) {
	if params.StockCode == "" {
		return nil, fmt.Errorf("stock_code is required")
	}

	var (
		args  = []any{params.StockCode}
		where = []string{"stock_code = ?"}
	)
	if params.StartDate != nil {
		where = append(where, "trade_date >= ?")
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		where = append(where, "trade_date <= ?")
		args = append(args, *params.EndDate)
	}

	limitSQL := ""
	if params.Limit > 0 {
		limitSQL = " LIMIT ?"
		args = append(args, params.Limit)
	}

	query := fmt.Sprintf(`
		SELECT stock_code, market, trade_date, open, high, low, close, volume,
		       amount, adj_factor, source, data_version
		FROM daily_bars
		WHERE %s
		ORDER BY trade_date ASC
		%s
	`, strings.Join(where, " AND "), limitSQL)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list daily bars: %w", err)
	}
	defer rows.Close()

	var out []DailyBar
	for rows.Next() {
		bar, err := scanDailyBar(rows)
		if err != nil {
			return nil, fmt.Errorf("scan daily bar: %w", err)
		}
		out = append(out, bar)
	}
	return out, rows.Err()
}

// UpsertStockDataStatus 写入或更新个股数据状态。
func (s *Store) UpsertStockDataStatus(ctx context.Context, item StockDataStatus) error {
	var missingJSON any
	if len(item.MissingRanges) > 0 {
		b, err := json.Marshal(item.MissingRanges)
		if err != nil {
			return fmt.Errorf("encode missing_ranges: %w", err)
		}
		missingJSON = string(b)
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO stock_data_status (
			stock_code, stock_name, market, training_index_code,
			data_start_date, data_end_date, completeness, missing_ranges,
			last_update, sync_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (stock_code) DO UPDATE SET
			stock_name = excluded.stock_name,
			market = excluded.market,
			training_index_code = excluded.training_index_code,
			data_start_date = excluded.data_start_date,
			data_end_date = excluded.data_end_date,
			completeness = excluded.completeness,
			missing_ranges = excluded.missing_ranges,
			last_update = excluded.last_update,
			sync_status = excluded.sync_status
	`, item.StockCode, item.StockName, item.Market, item.TrainingIndexCode,
		item.DataStartDate, item.DataEndDate, item.Completeness, missingJSON,
		item.LastUpdate, item.SyncStatus)
	if err != nil {
		return fmt.Errorf("upsert stock data status: %w", err)
	}
	return nil
}

// UpsertDailyBar 写入或更新单根日 K。
func (s *Store) UpsertDailyBar(ctx context.Context, bar DailyBar) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO daily_bars (
			stock_code, market, trade_date, open, high, low, close, volume,
			amount, adj_factor, source, data_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (stock_code, trade_date) DO UPDATE SET
			market = excluded.market,
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume,
			amount = excluded.amount,
			adj_factor = excluded.adj_factor,
			source = excluded.source,
			data_version = excluded.data_version
	`, bar.StockCode, bar.Market, bar.TradeDate, bar.Open, bar.High, bar.Low, bar.Close,
		bar.Volume, bar.Amount, bar.AdjFactor, bar.Source, bar.DataVersion)
	if err != nil {
		return fmt.Errorf("upsert daily bar: %w", err)
	}
	return nil
}
