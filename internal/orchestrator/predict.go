package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gsxhnd/guanlan/internal/biz"
	"github.com/gsxhnd/guanlan/internal/data"
	pb "github.com/gsxhnd/guanlan/internal/proto/quant/v1"
)

// PredictExecutor calls prediction gRPC and persists scores.
type PredictExecutor struct {
	Store  *data.Store
	Client pb.PredictionServiceClient
}

func (p *PredictExecutor) Run(ctx context.Context, task data.DataSyncTask) error {
	if p.Client == nil {
		return fmt.Errorf("prediction client not configured")
	}
	tradeDate, modelVersion, codes := biz.DecodeAnalysisTarget(task.TargetObject)
	stockCodes, err := p.resolveCodes(ctx, codes)
	if err != nil {
		return err
	}
	if len(stockCodes) == 0 {
		return fmt.Errorf("no stocks to predict")
	}

	resp, err := p.Client.PredictBatch(ctx, &pb.PredictBatchRequest{
		StockCodes:    stockCodes,
		TradeDate:     tradeDate,
		ModelVersion:  modelVersion,
	})
	if err != nil {
		return fmt.Errorf("predict batch: %w", err)
	}

	rows, err := toStoreRows(resp.GetPredictions(), tradeDate)
	if err != nil {
		return err
	}
	if err := p.Store.InsertPredictions(ctx, rows); err != nil {
		return err
	}

	version := modelVersion
	if len(rows) > 0 && rows[0].ModelVersion != "" {
		version = rows[0].ModelVersion
	}
	return p.Store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusSuccess, nil, &version)
}

// PredictOne runs on-demand inference and persists a single row.
func (p *PredictExecutor) PredictOne(ctx context.Context, stockCode, tradeDate, modelVersion string) (data.Prediction, error) {
	if p.Client == nil {
		return data.Prediction{}, fmt.Errorf("prediction client not configured")
	}
	stockCode = strings.ToUpper(strings.TrimSpace(stockCode))
	if stockCode == "" {
		return data.Prediction{}, fmt.Errorf("stock_code is required")
	}
	if tradeDate == "" {
		tradeDate = time.Now().UTC().Format("2006-01-02")
	}
	if modelVersion == "" {
		modelVersion = "latest"
	}

	resp, err := p.Client.Predict(ctx, &pb.PredictRequest{
		StockCode:     stockCode,
		TradeDate:     tradeDate,
		ModelVersion:  modelVersion,
	})
	if err != nil {
		return data.Prediction{}, fmt.Errorf("predict: %w", err)
	}
	rows, err := toStoreRows([]*pb.PredictResponse{resp}, tradeDate)
	if err != nil {
		return data.Prediction{}, err
	}
	if err := p.Store.InsertPredictions(ctx, rows); err != nil {
		return data.Prediction{}, err
	}
	return rows[0], nil
}

func (p *PredictExecutor) resolveCodes(ctx context.Context, codes []string) ([]string, error) {
	if len(codes) == 1 && (codes[0] == biz.AnalysisTargetWatchlist || codes[0] == "") {
		items, err := p.Store.ListWatchlistItems(ctx, true)
		if err != nil {
			return nil, err
		}
		out := make([]string, 0, len(items))
		for _, item := range items {
			out = append(out, item.StockCode)
		}
		if len(out) == 0 {
			pool, err := p.Store.ListStockPool(ctx, data.StockPoolListOptions{DailySyncOnly: true})
			if err != nil {
				return nil, err
			}
			for _, e := range pool {
				out = append(out, e.YfinanceSymbol)
			}
		}
		return out, nil
	}
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c != "" && c != biz.AnalysisTargetWatchlist {
			out = append(out, c)
		}
	}
	return out, nil
}

func toStoreRows(in []*pb.PredictResponse, fallbackDate string) ([]data.Prediction, error) {
	now := time.Now().UTC()
	out := make([]data.Prediction, 0, len(in))
	for _, r := range in {
		if r == nil || r.GetStockCode() == "" {
			continue
		}
		dateStr := r.GetTradeDate()
		if dateStr == "" {
			dateStr = fallbackDate
		}
		td, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid trade_date %q: %w", dateStr, err)
		}
		out = append(out, data.Prediction{
			StockCode:    r.GetStockCode(),
			TradeDate:    td,
			Score:        r.GetScore(),
			ModelVersion: r.GetModelVersion(),
			CreatedAt:    now,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("prediction service returned no scores")
	}
	return out, nil
}
