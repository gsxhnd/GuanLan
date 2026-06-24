package data

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
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
		positions, _, err := s.ComputePositions(ctx)
		if err != nil {
			return PortfolioDividend{}, err
		}
		for _, p := range positions {
			if p.StockCode == div.StockCode {
				total := *div.DividendPerShare * p.Quantity
				div.TotalDividend = total
				break
			}
		}
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

// ComputePositions 基于交易和分红重算持仓与现金余额。
func (s *Store) ComputePositions(ctx context.Context) ([]PortfolioPosition, float64, error) {
	trades, err := s.ListTrades(ctx, "", 10000)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(trades, func(i, j int) bool {
		if trades[i].TradeDate.Equal(trades[j].TradeDate) {
			return trades[i].CreatedAt.Before(trades[j].CreatedAt)
		}
		return trades[i].TradeDate.Before(trades[j].TradeDate)
	})

	dividends, err := s.ListDividends(ctx, "", 10000)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(dividends, func(i, j int) bool {
		return dividends[i].DividendDate.Before(dividends[j].DividendDate)
	})

	cashFlows, err := s.ListCashFlows(ctx, "", 10000)
	if err != nil {
		return nil, 0, err
	}

	type posState struct {
		name           string
		qty            float64
		totalCost      float64
		realized       float64
		dividendIncome float64
	}

	positions := map[string]*posState{}
	var cash float64

	for _, f := range cashFlows {
		if f.FlowType == CashFlowDeposit || f.FlowType == CashFlowWithdrawal {
			cash += f.Amount
		}
	}

	for _, t := range trades {
		p := positions[t.StockCode]
		if p == nil {
			p = &posState{name: t.StockName}
			positions[t.StockCode] = p
		}
		if p.name == "" {
			p.name = t.StockName
		}

		amt := t.Price * t.Quantity
		if t.Side == TradeSideBuy {
			cash -= amt + t.TotalFee
			p.qty += t.Quantity
			p.totalCost += amt + t.TotalFee
		} else {
			cash += amt - t.TotalFee
			if p.qty > 0 {
				avg := p.totalCost / p.qty
				costPortion := avg * t.Quantity
				p.realized += amt - t.TotalFee - costPortion
				p.qty -= t.Quantity
				p.totalCost -= costPortion
				if p.qty <= 0 {
					p.qty = 0
					p.totalCost = 0
				}
			}
		}
	}

	for _, d := range dividends {
		cash += d.TotalDividend
		p := positions[d.StockCode]
		if p != nil && p.qty > 0 {
			p.totalCost -= d.TotalDividend
			if p.totalCost < 0 {
				p.totalCost = 0
			}
			p.dividendIncome += d.TotalDividend
		}
	}

	valuations, _ := s.ListValuations(ctx, "")
	latestPrice := map[string]float64{}
	for _, v := range valuations {
		if v.StockCode != nil && v.Price != nil {
			latestPrice[*v.StockCode] = *v.Price
		}
	}

	// 回退到日频收盘价
	rows, err := s.db.QueryContext(ctx, `
		SELECT stock_code, close FROM daily_bars
		QUALIFY ROW_NUMBER() OVER (PARTITION BY stock_code ORDER BY trade_date DESC) = 1
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code string
			var close float64
			if err := rows.Scan(&code, &close); err == nil {
				if _, ok := latestPrice[code]; !ok {
					latestPrice[code] = close
				}
			}
		}
	}

	var out []PortfolioPosition
	for code, p := range positions {
		if p.qty <= 0 {
			continue
		}
		avg := 0.0
		if p.qty > 0 {
			avg = p.totalCost / p.qty
		}
		pos := PortfolioPosition{
			StockCode:      code,
			StockName:      p.name,
			Quantity:       p.qty,
			TotalCost:      p.totalCost,
			AverageCost:    avg,
			RealizedPnL:    p.realized,
			DividendIncome: p.dividendIncome,
		}
		if price, ok := latestPrice[code]; ok {
			pos.LatestPrice = &price
			mv := price * p.qty
			pos.MarketValue = &mv
			u := mv - p.totalCost
			pos.UnrealizedPnL = &u
		}
		out = append(out, pos)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StockCode < out[j].StockCode })
	return out, cash, nil
}

// GetAssets 返回当前资产概览与历史快照。
func (s *Store) GetAssets(ctx context.Context, startDate, endDate *time.Time) (float64, float64, float64, []AssetSnapshot, error) {
	positions, cash, err := s.ComputePositions(ctx)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	var holdingMV float64
	for _, p := range positions {
		if p.MarketValue != nil {
			holdingMV += *p.MarketValue
		}
	}
	total := cash + holdingMV

	// 写入/更新今日快照
	today := time.Now().UTC().Truncate(24 * time.Hour)
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO portfolio_asset_snapshots (
			snapshot_date, cash_balance, holding_market_value, total_asset, source
		) VALUES (?, ?, ?, ?, 'valuation')
		ON CONFLICT (snapshot_date) DO UPDATE SET
			cash_balance = excluded.cash_balance,
			holding_market_value = excluded.holding_market_value,
			total_asset = excluded.total_asset
	`, today, cash, holdingMV, total)

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
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT snapshot_date, cash_balance, holding_market_value, total_asset, source
		FROM portfolio_asset_snapshots %s
		ORDER BY snapshot_date ASC
	`, whereSQL), args...)
	if err != nil {
		return cash, holdingMV, total, nil, fmt.Errorf("list asset snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []AssetSnapshot
	for rows.Next() {
		var snap AssetSnapshot
		if err := rows.Scan(&snap.SnapshotDate, &snap.CashBalance, &snap.HoldingMarketValue, &snap.TotalAsset, &snap.Source); err != nil {
			return 0, 0, 0, nil, err
		}
		snapshots = append(snapshots, snap)
	}
	return cash, holdingMV, total, snapshots, rows.Err()
}

// GetAnnualReview 按自然年汇总复盘数据。
func (s *Store) GetAnnualReview(ctx context.Context, year int) (AnnualReview, error) {
	if year <= 0 {
		year = time.Now().Year()
	}
	yearStr := fmt.Sprintf("%d", year)

	trades, err := s.ListTrades(ctx, yearStr, 10000)
	if err != nil {
		return AnnualReview{}, err
	}
	_ = trades
	dividends, err := s.ListDividends(ctx, yearStr, 10000)
	if err != nil {
		return AnnualReview{}, err
	}
	cashFlows, err := s.ListCashFlows(ctx, yearStr, 10000)
	if err != nil {
		return AnnualReview{}, err
	}

	positions, _, err := s.ComputePositions(ctx)
	if err != nil {
		return AnnualReview{}, err
	}
	realizedByStock := map[string]float64{}
	divByStock := map[string]float64{}
	for _, p := range positions {
		realizedByStock[p.StockCode] = p.RealizedPnL
	}
	for _, d := range dividends {
		divByStock[d.StockCode] += d.TotalDividend
	}

	var netCashFlow float64
	for _, f := range cashFlows {
		if f.FlowType == CashFlowDeposit || f.FlowType == CashFlowWithdrawal {
			netCashFlow += f.Amount
		}
	}

	var dividendIncome float64
	for _, d := range dividends {
		dividendIncome += d.TotalDividend
	}

	// 已实现盈亏取当前累计值（Phase 2 简化口径）
	yearRealized := 0.0
	for _, p := range positions {
		yearRealized += p.RealizedPnL
	}

	var byStock []StockContribution
	codes := map[string]struct{}{}
	for c := range realizedByStock {
		codes[c] = struct{}{}
	}
	for c := range divByStock {
		codes[c] = struct{}{}
	}
	for c := range codes {
		byStock = append(byStock, StockContribution{
			StockCode:      c,
			RealizedPnL:    realizedByStock[c],
			DividendIncome: divByStock[c],
		})
	}
	sort.Slice(byStock, func(i, j int) bool { return byStock[i].StockCode < byStock[j].StockCode })

	review := AnnualReview{
		Year:             year,
		RealizedPnL:      yearRealized,
		DividendIncome:   dividendIncome,
		NetCashFlow:      netCashFlow,
		ByStockBreakdown: byStock,
	}

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	_, _, total, snaps, err := s.GetAssets(ctx, &start, &end)
	if err == nil && len(snaps) > 0 {
		begin := snaps[0].TotalAsset
		endAsset := snaps[len(snaps)-1].TotalAsset
		review.BeginTotalAsset = &begin
		review.EndTotalAsset = &endAsset
		if begin > 0 {
			rate := (endAsset - begin) / begin
			review.ReturnRate = &rate
		}
	} else {
		review.EndTotalAsset = &total
	}

	return review, nil
}

// ParseDate exported helper for handlers.
func ParseDate(s string) (time.Time, error) {
	return parseDate(s)
}
