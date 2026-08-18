package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// QualityIssueRow is a single data_quality_issues insert.
type QualityIssueRow struct {
	IssueID     string
	StockCode   string
	TradeDate   *time.Time
	IssueType   string
	Severity    string
	Message     string
	DataVersion string
	CreatedAt   time.Time
}

// IngestDailyBarsResult summarizes a stock ingest.
type IngestDailyBarsResult struct {
	StockCode    string
	BarsWritten  int
	DataVersion  string
	Completeness float64
}

// IngestDailyBars writes cleaned bars (+ optional quality issues) and marks stock ready.
func (s *Store) IngestDailyBars(
	ctx context.Context,
	stockCode string,
	market Market,
	stockName string,
	bars []DailyBar,
	issues []QualityIssueRow,
	completeness float64,
	dataVersion string,
) (IngestDailyBarsResult, error) {
	stockCode = strings.ToUpper(strings.TrimSpace(stockCode))
	if stockCode == "" {
		return IngestDailyBarsResult{}, fmt.Errorf("stock_code is required")
	}
	if dataVersion == "" {
		dataVersion = fmt.Sprintf("v%s", time.Now().UTC().Format("20060102T150405"))
	}
	if market == "" {
		entry, err := ParsePoolSymbol(stockCode, "")
		if err == nil {
			market = entry.Market
		} else {
			market = MarketUS
		}
	}
	if stockName == "" {
		stockName = stockCode
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return IngestDailyBarsResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO data_versions (version_id, description, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT (version_id) DO NOTHING
	`, dataVersion, fmt.Sprintf("ingest %s", stockCode), now)

	var startDate, endDate *time.Time
	for _, bar := range bars {
		bar.StockCode = stockCode
		if bar.Market == "" {
			bar.Market = market
		}
		bar.DataVersion = dataVersion
		if bar.Source == "" {
			bar.Source = "yfinance"
		}
		_, err := tx.ExecContext(ctx, `
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
			return IngestDailyBarsResult{}, fmt.Errorf("ingest bar: %w", err)
		}
		d := bar.TradeDate
		if startDate == nil || d.Before(*startDate) {
			startDate = &d
		}
		if endDate == nil || d.After(*endDate) {
			endDate = &d
		}

		_, _ = tx.ExecContext(ctx, `
			INSERT INTO daily_bars_raw (
				stock_code, market, trade_date, open, high, low, close, volume,
				amount, source, ingested_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (stock_code, trade_date, source) DO UPDATE SET
				open = excluded.open,
				high = excluded.high,
				low = excluded.low,
				close = excluded.close,
				volume = excluded.volume,
				amount = excluded.amount,
				ingested_at = excluded.ingested_at
		`, bar.StockCode, bar.Market, bar.TradeDate, bar.Open, bar.High, bar.Low, bar.Close,
			bar.Volume, bar.Amount, bar.Source, now)
	}

	for _, issue := range issues {
		if issue.IssueID == "" {
			issue.IssueID = uuid.NewString()
		}
		if issue.CreatedAt.IsZero() {
			issue.CreatedAt = now
		}
		if issue.DataVersion == "" {
			issue.DataVersion = dataVersion
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO data_quality_issues (
				issue_id, stock_code, trade_date, issue_type, severity, message, data_version, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, issue.IssueID, issue.StockCode, issue.TradeDate, issue.IssueType, issue.Severity,
			issue.Message, issue.DataVersion, issue.CreatedAt)
		if err != nil {
			return IngestDailyBarsResult{}, fmt.Errorf("ingest quality issue: %w", err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_data_status (
			stock_code, stock_name, market, training_index_code,
			data_start_date, data_end_date, completeness, missing_ranges,
			last_update, sync_status
		) VALUES (?, ?, ?, NULL, ?, ?, ?, NULL, ?, ?)
		ON CONFLICT (stock_code) DO UPDATE SET
			stock_name = excluded.stock_name,
			market = excluded.market,
			data_start_date = COALESCE(excluded.data_start_date, stock_data_status.data_start_date),
			data_end_date = COALESCE(excluded.data_end_date, stock_data_status.data_end_date),
			completeness = excluded.completeness,
			last_update = excluded.last_update,
			sync_status = excluded.sync_status
	`, stockCode, stockName, market, startDate, endDate, completeness, now, StockStatusReady)
	if err != nil {
		return IngestDailyBarsResult{}, fmt.Errorf("upsert stock status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return IngestDailyBarsResult{}, err
	}
	return IngestDailyBarsResult{
		StockCode:    stockCode,
		BarsWritten:  len(bars),
		DataVersion:  dataVersion,
		Completeness: completeness,
	}, nil
}

// MarkStockSyncFailed sets stock_data_status to missing after a failed sync.
func (s *Store) MarkStockSyncFailed(ctx context.Context, stockCode string) error {
	return s.UpsertStockDataStatus(ctx, StockDataStatus{
		StockCode:  stockCode,
		StockName:  stockCode,
		Market:     inferMarketFromCode(stockCode),
		SyncStatus: StockStatusMissing,
	})
}

func inferMarketFromCode(code string) Market {
	entry, err := ParsePoolSymbol(code, "")
	if err != nil {
		return MarketUS
	}
	return entry.Market
}
