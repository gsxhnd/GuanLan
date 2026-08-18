package biz

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

// ComputePositions recomputes holdings and cash from ledger tables.
func (p *Portfolio) ComputePositions(ctx context.Context) ([]data.PortfolioPosition, float64, error) {
	trades, err := p.Store.ListTrades(ctx, "", 10000)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(trades, func(i, j int) bool {
		if trades[i].TradeDate.Equal(trades[j].TradeDate) {
			return trades[i].CreatedAt.Before(trades[j].CreatedAt)
		}
		return trades[i].TradeDate.Before(trades[j].TradeDate)
	})

	dividends, err := p.Store.ListDividends(ctx, "", 10000)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(dividends, func(i, j int) bool {
		return dividends[i].DividendDate.Before(dividends[j].DividendDate)
	})

	cashFlows, err := p.Store.ListCashFlows(ctx, "", 10000)
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
		if f.FlowType == data.CashFlowDeposit || f.FlowType == data.CashFlowWithdrawal {
			cash += f.Amount
		}
	}

	for _, t := range trades {
		ps := positions[t.StockCode]
		if ps == nil {
			ps = &posState{name: t.StockName}
			positions[t.StockCode] = ps
		}
		if ps.name == "" {
			ps.name = t.StockName
		}

		amt := t.Price * t.Quantity
		if t.Side == data.TradeSideBuy {
			cash -= amt + t.TotalFee
			ps.qty += t.Quantity
			ps.totalCost += amt + t.TotalFee
		} else {
			cash += amt - t.TotalFee
			if ps.qty > 0 {
				avg := ps.totalCost / ps.qty
				costPortion := avg * t.Quantity
				ps.realized += amt - t.TotalFee - costPortion
				ps.qty -= t.Quantity
				ps.totalCost -= costPortion
				if ps.qty <= 0 {
					ps.qty = 0
					ps.totalCost = 0
				}
			}
		}
	}

	for _, d := range dividends {
		cash += d.TotalDividend
		ps := positions[d.StockCode]
		if ps != nil && ps.qty > 0 {
			ps.totalCost -= d.TotalDividend
			if ps.totalCost < 0 {
				ps.totalCost = 0
			}
			ps.dividendIncome += d.TotalDividend
		}
	}

	valuations, _ := p.Store.ListValuations(ctx, "")
	latestPrice := map[string]float64{}
	for _, v := range valuations {
		if v.StockCode != nil && v.Price != nil {
			latestPrice[*v.StockCode] = *v.Price
		}
	}

	if closes, err := p.Store.LatestCloses(ctx); err == nil {
		for code, close := range closes {
			if _, ok := latestPrice[code]; !ok {
				latestPrice[code] = close
			}
		}
	}

	var out []data.PortfolioPosition
	for code, ps := range positions {
		if ps.qty <= 0 {
			continue
		}
		avg := 0.0
		if ps.qty > 0 {
			avg = ps.totalCost / ps.qty
		}
		pos := data.PortfolioPosition{
			StockCode:      code,
			StockName:      ps.name,
			Quantity:       ps.qty,
			TotalCost:      ps.totalCost,
			AverageCost:    avg,
			RealizedPnL:    ps.realized,
			DividendIncome: ps.dividendIncome,
		}
		if price, ok := latestPrice[code]; ok {
			pos.LatestPrice = &price
			mv := price * ps.qty
			pos.MarketValue = &mv
			u := mv - ps.totalCost
			pos.UnrealizedPnL = &u
		}
		out = append(out, pos)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StockCode < out[j].StockCode })
	return out, cash, nil
}

// GetAssets returns current assets and historical snapshots.
func (p *Portfolio) GetAssets(ctx context.Context, startDate, endDate *time.Time) (float64, float64, float64, []data.AssetSnapshot, error) {
	positions, cash, err := p.ComputePositions(ctx)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	var holdingMV float64
	for _, pos := range positions {
		if pos.MarketValue != nil {
			holdingMV += *pos.MarketValue
		}
	}
	total := cash + holdingMV

	today := time.Now().UTC().Truncate(24 * time.Hour)
	_ = p.Store.UpsertAssetSnapshot(ctx, data.AssetSnapshot{
		SnapshotDate:       today,
		CashBalance:        cash,
		HoldingMarketValue: holdingMV,
		TotalAsset:         total,
		Source:             "valuation",
	})

	snaps, err := p.Store.ListAssetSnapshots(ctx, startDate, endDate)
	if err != nil {
		return cash, holdingMV, total, nil, err
	}
	return cash, holdingMV, total, snaps, nil
}

// GetAnnualReview aggregates yearly review metrics.
func (p *Portfolio) GetAnnualReview(ctx context.Context, year int) (data.AnnualReview, error) {
	if year <= 0 {
		year = time.Now().Year()
	}
	yearStr := fmt.Sprintf("%d", year)

	dividends, err := p.Store.ListDividends(ctx, yearStr, 10000)
	if err != nil {
		return data.AnnualReview{}, err
	}
	cashFlows, err := p.Store.ListCashFlows(ctx, yearStr, 10000)
	if err != nil {
		return data.AnnualReview{}, err
	}

	positions, _, err := p.ComputePositions(ctx)
	if err != nil {
		return data.AnnualReview{}, err
	}
	realizedByStock := map[string]float64{}
	divByStock := map[string]float64{}
	for _, pos := range positions {
		realizedByStock[pos.StockCode] = pos.RealizedPnL
	}
	for _, d := range dividends {
		divByStock[d.StockCode] += d.TotalDividend
	}

	var netCashFlow float64
	for _, f := range cashFlows {
		if f.FlowType == data.CashFlowDeposit || f.FlowType == data.CashFlowWithdrawal {
			netCashFlow += f.Amount
		}
	}

	var dividendIncome float64
	for _, d := range dividends {
		dividendIncome += d.TotalDividend
	}

	yearRealized := 0.0
	for _, pos := range positions {
		yearRealized += pos.RealizedPnL
	}

	var byStock []data.StockContribution
	codes := map[string]struct{}{}
	for c := range realizedByStock {
		codes[c] = struct{}{}
	}
	for c := range divByStock {
		codes[c] = struct{}{}
	}
	for c := range codes {
		byStock = append(byStock, data.StockContribution{
			StockCode:      c,
			RealizedPnL:    realizedByStock[c],
			DividendIncome: divByStock[c],
		})
	}
	sort.Slice(byStock, func(i, j int) bool { return byStock[i].StockCode < byStock[j].StockCode })

	review := data.AnnualReview{
		Year:             year,
		RealizedPnL:      yearRealized,
		DividendIncome:   dividendIncome,
		NetCashFlow:      netCashFlow,
		ByStockBreakdown: byStock,
	}

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	_, _, total, snaps, err := p.GetAssets(ctx, &start, &end)
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

// CreateDividend writes a dividend, resolving total from per-share × position when needed.
func (p *Portfolio) CreateDividend(ctx context.Context, div data.PortfolioDividend) (data.PortfolioDividend, error) {
	if div.TotalDividend <= 0 && div.DividendPerShare != nil && *div.DividendPerShare > 0 {
		positions, _, err := p.ComputePositions(ctx)
		if err != nil {
			return data.PortfolioDividend{}, err
		}
		for _, pos := range positions {
			if pos.StockCode == div.StockCode {
				div.TotalDividend = *div.DividendPerShare * pos.Quantity
				break
			}
		}
	}
	return p.Store.CreateDividend(ctx, div)
}
