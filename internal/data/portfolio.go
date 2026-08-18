package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", s, err)
	}
	return t, nil
}

func yearFilter(year string) (start, end *time.Time, err error) {
	year = strings.TrimSpace(year)
	if year == "" {
		return nil, nil, nil
	}
	t, err := time.Parse("2006", year)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid year %q", year)
	}
	s := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	e := time.Date(t.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	return &s, &e, nil
}

// CreateTrade 记录交易并写入关联现金流。
func (s *Store) CreateTrade(ctx context.Context, trade PortfolioTrade) (PortfolioTrade, error) {
	if trade.StockCode == "" {
		return PortfolioTrade{}, fmt.Errorf("stock_code is required")
	}
	if trade.Side != TradeSideBuy && trade.Side != TradeSideSell {
		return PortfolioTrade{}, fmt.Errorf("invalid side %q", trade.Side)
	}
	if trade.Quantity <= 0 || trade.Price < 0 {
		return PortfolioTrade{}, fmt.Errorf("invalid price or quantity")
	}

	trade.TradeID = uuid.NewString()
	trade.CreatedAt = time.Now().UTC()
	trade.StockCode = strings.ToUpper(strings.TrimSpace(trade.StockCode))

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO portfolio_trades (
			trade_id, trade_date, stock_code, stock_name, side,
			price, quantity, total_fee, note, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, trade.TradeID, trade.TradeDate, trade.StockCode, trade.StockName,
		trade.Side, trade.Price, trade.Quantity, trade.TotalFee, nullString(trade.Note), trade.CreatedAt)
	if err != nil {
		return PortfolioTrade{}, fmt.Errorf("create trade: %w", err)
	}

	amount := trade.Price * trade.Quantity
	if trade.Side == TradeSideBuy {
		amount = -(amount + trade.TotalFee)
	} else {
		amount = amount - trade.TotalFee
	}
	_, _ = s.CreateCashFlow(ctx, PortfolioCashFlow{
		FlowDate:  trade.TradeDate,
		Amount:    amount,
		FlowType:  CashFlowTrade,
		SourceRef: &trade.TradeID,
		Note:      fmt.Sprintf("trade %s %s", trade.Side, trade.StockCode),
	})

	return trade, nil
}

// ListTrades 返回交易记录。
func (s *Store) ListTrades(ctx context.Context, year string, limit int) ([]PortfolioTrade, error) {
	if limit <= 0 {
		limit = 200
	}
	start, end, err := yearFilter(year)
	if err != nil {
		return nil, err
	}

	var (
		args  []any
		where []string
	)
	if start != nil {
		where = append(where, "trade_date >= ?")
		args = append(args, *start)
	}
	if end != nil {
		where = append(where, "trade_date <= ?")
		args = append(args, *end)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT trade_id, trade_date, stock_code, stock_name, side,
		       price, quantity, total_fee, note, created_at
		FROM portfolio_trades %s
		ORDER BY trade_date DESC, created_at DESC
		LIMIT ?
	`, whereSQL), args...)
	if err != nil {
		return nil, fmt.Errorf("list trades: %w", err)
	}
	defer rows.Close()

	return scanTrades(rows)
}

func scanTrades(rows *sql.Rows) ([]PortfolioTrade, error) {
	var out []PortfolioTrade
	for rows.Next() {
		var t PortfolioTrade
		var note sql.NullString
		if err := rows.Scan(
			&t.TradeID, &t.TradeDate, &t.StockCode, &t.StockName, &t.Side,
			&t.Price, &t.Quantity, &t.TotalFee, &note, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		if note.Valid {
			t.Note = note.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CreateDividend 记录分红并写入关联现金流。
func (s *Store) CreateDividend(ctx context.Context, div PortfolioDividend) (PortfolioDividend, error) {
	if div.StockCode == "" {
		return PortfolioDividend{}, fmt.Errorf("stock_code is required")
	}
	div.DividendID = uuid.NewString()
	div.CreatedAt = time.Now().UTC()
	div.StockCode = strings.ToUpper(strings.TrimSpace(div.StockCode))

	if div.TotalDividend <= 0 && div.DividendPerShare != nil && *div.DividendPerShare > 0 {
		return PortfolioDividend{}, fmt.Errorf("total_dividend required when resolving per-share in Store; use biz.Portfolio.CreateDividend")
	}
	if div.TotalDividend <= 0 {
		return PortfolioDividend{}, fmt.Errorf("total_dividend is required")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO portfolio_dividends (
			dividend_id, dividend_date, stock_code, dividend_per_share,
			total_dividend, bonus_share_ratio, transfer_share_ratio, note, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, div.DividendID, div.DividendDate, div.StockCode, div.DividendPerShare,
		div.TotalDividend, div.BonusShareRatio, div.TransferShareRatio, nullString(div.Note), div.CreatedAt)
	if err != nil {
		return PortfolioDividend{}, fmt.Errorf("create dividend: %w", err)
	}

	ref := div.DividendID
	_, _ = s.CreateCashFlow(ctx, PortfolioCashFlow{
		FlowDate:  div.DividendDate,
		Amount:    div.TotalDividend,
		FlowType:  CashFlowDividend,
		SourceRef: &ref,
		Note:      fmt.Sprintf("dividend %s", div.StockCode),
	})

	return div, nil
}

// ListDividends 返回分红记录。
func (s *Store) ListDividends(ctx context.Context, year string, limit int) ([]PortfolioDividend, error) {
	if limit <= 0 {
		limit = 200
	}
	start, end, err := yearFilter(year)
	if err != nil {
		return nil, err
	}

	var (
		args  []any
		where []string
	)
	if start != nil {
		where = append(where, "dividend_date >= ?")
		args = append(args, *start)
	}
	if end != nil {
		where = append(where, "dividend_date <= ?")
		args = append(args, *end)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT dividend_id, dividend_date, stock_code, dividend_per_share,
		       total_dividend, bonus_share_ratio, transfer_share_ratio, note, created_at
		FROM portfolio_dividends %s
		ORDER BY dividend_date DESC, created_at DESC
		LIMIT ?
	`, whereSQL), args...)
	if err != nil {
		return nil, fmt.Errorf("list dividends: %w", err)
	}
	defer rows.Close()

	var out []PortfolioDividend
	for rows.Next() {
		var d PortfolioDividend
		var perShare, bonus, transfer sql.NullFloat64
		var note sql.NullString
		if err := rows.Scan(
			&d.DividendID, &d.DividendDate, &d.StockCode, &perShare,
			&d.TotalDividend, &bonus, &transfer, &note, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		if perShare.Valid {
			v := perShare.Float64
			d.DividendPerShare = &v
		}
		if note.Valid {
			d.Note = note.String
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// CreateCashFlow 记录出入金等现金流。
func (s *Store) CreateCashFlow(ctx context.Context, flow PortfolioCashFlow) (PortfolioCashFlow, error) {
	if flow.FlowType == "" {
		return PortfolioCashFlow{}, fmt.Errorf("flow_type is required")
	}
	flow.CashFlowID = uuid.NewString()
	flow.CreatedAt = time.Now().UTC()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO portfolio_cash_flows (
			cash_flow_id, flow_date, amount, flow_type, source_ref, note, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, flow.CashFlowID, flow.FlowDate, flow.Amount, flow.FlowType,
		flow.SourceRef, nullString(flow.Note), flow.CreatedAt)
	if err != nil {
		return PortfolioCashFlow{}, fmt.Errorf("create cash flow: %w", err)
	}
	return flow, nil
}

// ListCashFlows 返回现金流记录。
func (s *Store) ListCashFlows(ctx context.Context, year string, limit int) ([]PortfolioCashFlow, error) {
	if limit <= 0 {
		limit = 200
	}
	start, end, err := yearFilter(year)
	if err != nil {
		return nil, err
	}

	var (
		args  []any
		where []string
	)
	if start != nil {
		where = append(where, "flow_date >= ?")
		args = append(args, *start)
	}
	if end != nil {
		where = append(where, "flow_date <= ?")
		args = append(args, *end)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT cash_flow_id, flow_date, amount, flow_type, source_ref, note, created_at
		FROM portfolio_cash_flows %s
		ORDER BY flow_date DESC, created_at DESC
		LIMIT ?
	`, whereSQL), args...)
	if err != nil {
		return nil, fmt.Errorf("list cash flows: %w", err)
	}
	defer rows.Close()

	var out []PortfolioCashFlow
	for rows.Next() {
		var f PortfolioCashFlow
		var sourceRef, note sql.NullString
		if err := rows.Scan(
			&f.CashFlowID, &f.FlowDate, &f.Amount, &f.FlowType, &sourceRef, &note, &f.CreatedAt,
		); err != nil {
			return nil, err
		}
		if sourceRef.Valid {
			v := sourceRef.String
			f.SourceRef = &v
		}
		if note.Valid {
			f.Note = note.String
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// CreateValuation 记录估值快照。
func (s *Store) CreateValuation(ctx context.Context, v PortfolioValuation) (PortfolioValuation, error) {
	v.ValuationID = uuid.NewString()
	v.CreatedAt = time.Now().UTC()
	if v.Source == "" {
		v.Source = "manual"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO portfolio_valuations (
			valuation_id, valuation_date, stock_code, price,
			total_asset_override, source, note, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, v.ValuationID, v.ValuationDate, v.StockCode, v.Price,
		v.TotalAssetOverride, v.Source, nullString(v.Note), v.CreatedAt)
	if err != nil {
		return PortfolioValuation{}, fmt.Errorf("create valuation: %w", err)
	}
	return v, nil
}

// ListValuations 返回估值快照。
func (s *Store) ListValuations(ctx context.Context, year string) ([]PortfolioValuation, error) {
	start, end, err := yearFilter(year)
	if err != nil {
		return nil, err
	}

	var (
		args  []any
		where []string
	)
	if start != nil {
		where = append(where, "valuation_date >= ?")
		args = append(args, *start)
	}
	if end != nil {
		where = append(where, "valuation_date <= ?")
		args = append(args, *end)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT valuation_id, valuation_date, stock_code, price,
		       total_asset_override, source, note, created_at
		FROM portfolio_valuations %s
		ORDER BY valuation_date DESC
	`, whereSQL), args...)
	if err != nil {
		return nil, fmt.Errorf("list valuations: %w", err)
	}
	defer rows.Close()

	var out []PortfolioValuation
	for rows.Next() {
		var v PortfolioValuation
		var stockCode, note sql.NullString
		var price, totalOverride sql.NullFloat64
		if err := rows.Scan(
			&v.ValuationID, &v.ValuationDate, &stockCode, &price,
			&totalOverride, &v.Source, &note, &v.CreatedAt,
		); err != nil {
			return nil, err
		}
		if stockCode.Valid {
			s := stockCode.String
			v.StockCode = &s
		}
		if price.Valid {
			p := price.Float64
			v.Price = &p
		}
		if totalOverride.Valid {
			t := totalOverride.Float64
			v.TotalAssetOverride = &t
		}
		if note.Valid {
			v.Note = note.String
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ParseDate exported helper for handlers.
func ParseDate(s string) (time.Time, error) {
	return parseDate(s)
}
