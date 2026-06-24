package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func scanStockPoolEntry(row interface {
	Scan(dest ...any) error
}) (StockPoolEntry, error) {
	var item StockPoolEntry
	var exchange sql.NullString
	if err := row.Scan(
		&item.YfinanceSymbol,
		&item.OriginalCode,
		&item.Market,
		&exchange,
		&item.StockName,
		&item.Currency,
		&item.Source,
		&item.IsActive,
		&item.SyncDaily,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return StockPoolEntry{}, err
	}
	if exchange.Valid {
		item.Exchange = exchange.String
	}
	return item, nil
}

// UpsertStockPoolEntry 写入或更新数据底座股票池。
func (s *Store) UpsertStockPoolEntry(ctx context.Context, item StockPoolEntry) error {
	yf := strings.TrimSpace(item.YfinanceSymbol)
	if yf == "" || item.OriginalCode == "" {
		return fmt.Errorf("yfinance_symbol and original_code are required")
	}
	if item.Source == "" {
		item.Source = PoolSourceAPIManual
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	item.YfinanceSymbol = yf

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO stock_pool (
			yfinance_symbol, original_code, market, exchange, stock_name, currency,
			source, is_active, sync_daily, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (yfinance_symbol) DO UPDATE SET
			original_code = excluded.original_code,
			market = excluded.market,
			stock_name = COALESCE(NULLIF(excluded.stock_name, ''), stock_pool.stock_name),
			exchange = excluded.exchange,
			currency = excluded.currency,
			source = CASE
				WHEN stock_pool.source = 'api_manual' THEN stock_pool.source
				ELSE excluded.source
			END,
			is_active = excluded.is_active,
			sync_daily = excluded.sync_daily,
			updated_at = excluded.updated_at
	`, item.YfinanceSymbol, item.OriginalCode, item.Market,
		nullString(item.Exchange), item.StockName, item.Currency,
		item.Source, item.IsActive, item.SyncDaily, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert stock pool: %w", err)
	}
	return nil
}

// GetStockPoolEntry 按 yfinance 符号查询股票池条目。
func (s *Store) GetStockPoolEntry(ctx context.Context, yfinanceSymbol string) (StockPoolEntry, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT yfinance_symbol, original_code, market, exchange, stock_name, currency,
		       source, is_active, sync_daily, created_at, updated_at
		FROM stock_pool WHERE yfinance_symbol = ?
	`, strings.TrimSpace(yfinanceSymbol))
	item, err := scanStockPoolEntry(row)
	if err == sql.ErrNoRows {
		return StockPoolEntry{}, false, nil
	}
	if err != nil {
		return StockPoolEntry{}, false, err
	}
	return item, true, nil
}

// StockPoolListOptions 股票池列表筛选。
type StockPoolListOptions struct {
	Source        string
	DailySyncOnly bool
}

// ListStockPool 返回数据底座股票池列表。
func (s *Store) ListStockPool(ctx context.Context, opts StockPoolListOptions) ([]StockPoolEntry, error) {
	var clauses []string
	var args []any
	if opts.DailySyncOnly {
		clauses = append(clauses, "is_active = TRUE", "sync_daily = TRUE")
	}
	if source := strings.TrimSpace(opts.Source); source != "" {
		clauses = append(clauses, "source = ?")
		args = append(args, source)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT yfinance_symbol, original_code, market, exchange, stock_name, currency,
		       source, is_active, sync_daily, created_at, updated_at
		FROM stock_pool %s
		ORDER BY yfinance_symbol
	`, where), args...)
	if err != nil {
		return nil, fmt.Errorf("list stock pool: %w", err)
	}
	defer rows.Close()

	var out []StockPoolEntry
	for rows.Next() {
		item, err := scanStockPoolEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// EnsureStockInPool 确保股票存在于数据底座池（用户手动添加时调用）。
func (s *Store) EnsureStockInPool(ctx context.Context, symbol string, market Market, stockName string) error {
	entry, err := ParsePoolSymbol(symbol, market)
	if err != nil {
		return err
	}
	if stockName != "" {
		entry.StockName = stockName
	}
	if _, ok, err := s.GetStockPoolEntry(ctx, entry.YfinanceSymbol); err != nil {
		return err
	} else if ok {
		return nil
	}
	entry.Source = PoolSourceAPIManual
	entry.IsActive = true
	entry.SyncDaily = true
	return s.UpsertStockPoolEntry(ctx, entry)
}

// ParsePoolSymbol 将用户输入解析为股票池条目字段。
func ParsePoolSymbol(symbol string, market Market) (StockPoolEntry, error) {
	raw := strings.TrimSpace(strings.ToUpper(symbol))
	if raw == "" {
		return StockPoolEntry{}, fmt.Errorf("symbol is required")
	}

	entry := StockPoolEntry{
		Currency: "USD",
	}
	if market == MarketA {
		entry.Currency = "CNY"
		entry.Market = MarketA
	} else if market != "" {
		entry.Market = market
	}

	if strings.Contains(raw, ".") {
		parts := strings.SplitN(raw, ".", 2)
		entry.OriginalCode = parts[0]
		suffix := parts[1]
		switch suffix {
		case "SS", "SH":
			entry.Exchange = "SH"
			entry.YfinanceSymbol = entry.OriginalCode + ".SS"
			entry.Market = MarketA
			entry.Currency = "CNY"
		case "SZ":
			entry.Exchange = "SZ"
			entry.YfinanceSymbol = entry.OriginalCode + ".SZ"
			entry.Market = MarketA
			entry.Currency = "CNY"
		default:
			entry.YfinanceSymbol = raw
			entry.OriginalCode = parts[0]
			if entry.Market == "" {
				entry.Market = MarketUS
			}
		}
	} else {
		entry.OriginalCode = raw
		entry.YfinanceSymbol = raw
		if entry.Market == "" {
			entry.Market = MarketUS
		}
	}

	if entry.StockName == "" {
		entry.StockName = entry.YfinanceSymbol
	}
	return entry, nil
}

// ToYfinanceSymbol 将常见内部写法转为 yfinance 符号。
func ToYfinanceSymbol(stockCode string) string {
	entry, err := ParsePoolSymbol(stockCode, "")
	if err != nil {
		return strings.TrimSpace(strings.ToUpper(stockCode))
	}
	return entry.YfinanceSymbol
}
