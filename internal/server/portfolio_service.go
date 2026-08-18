package server

import (
	"context"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
	pb "github.com/gsxhnd/guanlan/internal/proto/quant/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Services) CreateTrade(ctx context.Context, req *pb.CreateTradeRequest) (*pb.Trade, error) {
	tradeDate, err := data.ParseDate(req.GetTradeDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	trade, err := s.Store.CreateTrade(ctx, data.PortfolioTrade{
		TradeDate: tradeDate,
		StockCode: req.GetStockCode(),
		StockName: req.GetStockName(),
		Side:      data.TradeSide(req.GetSide()),
		Price:     req.GetPrice(),
		Quantity:  req.GetQuantity(),
		TotalFee:  req.GetTotalFee(),
		Note:      req.GetNote(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create trade: %v", err)
	}
	return &pb.Trade{
		TradeId:   trade.TradeID,
		TradeDate: dateStr(trade.TradeDate),
		StockCode: trade.StockCode,
		StockName: trade.StockName,
		Side:      string(trade.Side),
		Price:     trade.Price,
		Quantity:  trade.Quantity,
		TotalFee:  trade.TotalFee,
		Note:      trade.Note,
		CreatedAt: tsPtr(&trade.CreatedAt),
	}, nil
}

func (s *Services) ListTrades(ctx context.Context, req *pb.ListTradesRequest) (*pb.ListTradesResponse, error) {
	trades, err := s.Store.ListTrades(ctx, req.GetYear(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list trades: %v", err)
	}
	out := make([]*pb.Trade, 0, len(trades))
	for _, t := range trades {
		out = append(out, &pb.Trade{
			TradeId:   t.TradeID,
			TradeDate: dateStr(t.TradeDate),
			StockCode: t.StockCode,
			StockName: t.StockName,
			Side:      string(t.Side),
			Price:     t.Price,
			Quantity:  t.Quantity,
			TotalFee:  t.TotalFee,
			Note:      t.Note,
			CreatedAt: tsPtr(&t.CreatedAt),
		})
	}
	return &pb.ListTradesResponse{Trades: out}, nil
}

func (s *Services) CreateDividend(ctx context.Context, req *pb.CreateDividendRequest) (*pb.Dividend, error) {
	divDate, err := data.ParseDate(req.GetDividendDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	var perShare *float64
	if req.GetDividendPerShare() > 0 {
		v := req.GetDividendPerShare()
		perShare = &v
	}
	div, err := s.Biz.Portfolio.CreateDividend(ctx, data.PortfolioDividend{
		DividendDate:     divDate,
		StockCode:        req.GetStockCode(),
		DividendPerShare: perShare,
		TotalDividend:    req.GetTotalDividend(),
		Note:             req.GetNote(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create dividend: %v", err)
	}
	return &pb.Dividend{
		DividendId:       div.DividendID,
		DividendDate:     dateStr(div.DividendDate),
		StockCode:        div.StockCode,
		DividendPerShare: floatPtr(div.DividendPerShare),
		TotalDividend:    div.TotalDividend,
		Note:             div.Note,
		CreatedAt:        tsPtr(&div.CreatedAt),
	}, nil
}

func (s *Services) ListDividends(ctx context.Context, req *pb.ListDividendsRequest) (*pb.ListDividendsResponse, error) {
	dividends, err := s.Store.ListDividends(ctx, req.GetYear(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list dividends: %v", err)
	}
	out := make([]*pb.Dividend, 0, len(dividends))
	for _, d := range dividends {
		out = append(out, &pb.Dividend{
			DividendId:       d.DividendID,
			DividendDate:     dateStr(d.DividendDate),
			StockCode:        d.StockCode,
			DividendPerShare: floatPtr(d.DividendPerShare),
			TotalDividend:    d.TotalDividend,
			Note:             d.Note,
			CreatedAt:        tsPtr(&d.CreatedAt),
		})
	}
	return &pb.ListDividendsResponse{Dividends: out}, nil
}

func (s *Services) CreateCashFlow(ctx context.Context, req *pb.CreateCashFlowRequest) (*pb.CashFlow, error) {
	flowDate, err := data.ParseDate(req.GetFlowDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	flow, err := s.Store.CreateCashFlow(ctx, data.PortfolioCashFlow{
		FlowDate: flowDate,
		Amount:   req.GetAmount(),
		FlowType: data.CashFlowType(req.GetFlowType()),
		Note:     req.GetNote(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create cash flow: %v", err)
	}
	return &pb.CashFlow{
		CashFlowId: flow.CashFlowID,
		FlowDate:   dateStr(flow.FlowDate),
		Amount:     flow.Amount,
		FlowType:   string(flow.FlowType),
		SourceRef:  strPtr(flow.SourceRef),
		Note:       flow.Note,
		CreatedAt:  tsPtr(&flow.CreatedAt),
	}, nil
}

func (s *Services) ListCashFlows(ctx context.Context, req *pb.ListCashFlowsRequest) (*pb.ListCashFlowsResponse, error) {
	flows, err := s.Store.ListCashFlows(ctx, req.GetYear(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list cash flows: %v", err)
	}
	out := make([]*pb.CashFlow, 0, len(flows))
	for _, f := range flows {
		out = append(out, &pb.CashFlow{
			CashFlowId: f.CashFlowID,
			FlowDate:   dateStr(f.FlowDate),
			Amount:     f.Amount,
			FlowType:   string(f.FlowType),
			SourceRef:  strPtr(f.SourceRef),
			Note:       f.Note,
			CreatedAt:  tsPtr(&f.CreatedAt),
		})
	}
	return &pb.ListCashFlowsResponse{CashFlows: out}, nil
}

func (s *Services) ListPositions(ctx context.Context, _ *pb.Empty) (*pb.ListPositionsResponse, error) {
	positions, cash, err := s.Biz.Portfolio.ComputePositions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list positions: %v", err)
	}
	out := make([]*pb.Position, 0, len(positions))
	for _, p := range positions {
		out = append(out, &pb.Position{
			StockCode:      p.StockCode,
			StockName:      p.StockName,
			Quantity:       p.Quantity,
			TotalCost:      p.TotalCost,
			AverageCost:    p.AverageCost,
			RealizedPnl:    p.RealizedPnL,
			DividendIncome: p.DividendIncome,
			LatestPrice:    floatPtr(p.LatestPrice),
			MarketValue:    floatPtr(p.MarketValue),
			UnrealizedPnl:  floatPtr(p.UnrealizedPnL),
		})
	}
	return &pb.ListPositionsResponse{Positions: out, CashBalance: cash}, nil
}

func (s *Services) CreateValuation(ctx context.Context, req *pb.CreateValuationRequest) (*pb.Valuation, error) {
	valDate, err := data.ParseDate(req.GetValuationDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	var stockCode *string
	if req.GetStockCode() != "" {
		v := req.GetStockCode()
		stockCode = &v
	}
	var price, totalOverride *float64
	if req.GetPrice() > 0 {
		v := req.GetPrice()
		price = &v
	}
	if req.GetTotalAssetOverride() > 0 {
		v := req.GetTotalAssetOverride()
		totalOverride = &v
	}
	v, err := s.Store.CreateValuation(ctx, data.PortfolioValuation{
		ValuationDate:      valDate,
		StockCode:          stockCode,
		Price:              price,
		TotalAssetOverride: totalOverride,
		Source:             req.GetSource(),
		Note:               req.GetNote(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create valuation: %v", err)
	}
	return &pb.Valuation{
		ValuationId:        v.ValuationID,
		ValuationDate:      dateStr(v.ValuationDate),
		StockCode:          strPtr(v.StockCode),
		Price:              floatPtr(v.Price),
		TotalAssetOverride: floatPtr(v.TotalAssetOverride),
		Source:             v.Source,
		Note:               v.Note,
	}, nil
}

func (s *Services) ListValuations(ctx context.Context, req *pb.ListValuationsRequest) (*pb.ListValuationsResponse, error) {
	vals, err := s.Store.ListValuations(ctx, req.GetYear())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list valuations: %v", err)
	}
	out := make([]*pb.Valuation, 0, len(vals))
	for _, v := range vals {
		out = append(out, &pb.Valuation{
			ValuationId:        v.ValuationID,
			ValuationDate:      dateStr(v.ValuationDate),
			StockCode:          strPtr(v.StockCode),
			Price:              floatPtr(v.Price),
			TotalAssetOverride: floatPtr(v.TotalAssetOverride),
			Source:             v.Source,
			Note:               v.Note,
		})
	}
	return &pb.ListValuationsResponse{Valuations: out}, nil
}

func (s *Services) GetAssets(ctx context.Context, req *pb.GetAssetsRequest) (*pb.GetAssetsResponse, error) {
	var startDate, endDate *time.Time
	if req.GetStartDate() != "" {
		t, err := data.ParseDate(req.GetStartDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		startDate = &t
	}
	if req.GetEndDate() != "" {
		t, err := data.ParseDate(req.GetEndDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		endDate = &t
	}
	cash, holding, total, snaps, err := s.Biz.Portfolio.GetAssets(ctx, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get assets: %v", err)
	}
	out := make([]*pb.AssetSnapshot, 0, len(snaps))
	for _, snap := range snaps {
		out = append(out, &pb.AssetSnapshot{
			SnapshotDate:       dateStr(snap.SnapshotDate),
			CashBalance:        snap.CashBalance,
			HoldingMarketValue: snap.HoldingMarketValue,
			TotalAsset:         snap.TotalAsset,
			Source:             snap.Source,
		})
	}
	return &pb.GetAssetsResponse{
		CashBalance:        cash,
		HoldingMarketValue: holding,
		TotalAsset:         total,
		Snapshots:          out,
	}, nil
}

func (s *Services) GetAnnualReview(ctx context.Context, req *pb.GetAnnualReviewRequest) (*pb.AnnualReview, error) {
	year := int(req.GetYear())
	if year == 0 {
		year = 0
	}
	review, err := s.Biz.Portfolio.GetAnnualReview(ctx, year)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get annual review: %v", err)
	}
	byStock := make([]*pb.StockContribution, 0, len(review.ByStockBreakdown))
	for _, c := range review.ByStockBreakdown {
		byStock = append(byStock, &pb.StockContribution{
			StockCode:      c.StockCode,
			RealizedPnl:    c.RealizedPnL,
			DividendIncome: c.DividendIncome,
		})
	}
	return &pb.AnnualReview{
		Year:            int32(review.Year),
		RealizedPnl:     review.RealizedPnL,
		DividendIncome:  review.DividendIncome,
		NetCashFlow:     review.NetCashFlow,
		BeginTotalAsset: floatPtr(review.BeginTotalAsset),
		EndTotalAsset:   floatPtr(review.EndTotalAsset),
		ReturnRate:      floatPtr(review.ReturnRate),
		ByStock:         byStock,
	}, nil
}
