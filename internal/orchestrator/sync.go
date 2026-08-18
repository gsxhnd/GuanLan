package orchestrator

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
	pb "github.com/gsxhnd/guanlan/internal/proto/quant/v1"
)

// SyncExecutor fetches cleaned bars from crawler gRPC and ingests into DuckDB.
type SyncExecutor struct {
	Store         *data.Store
	Crawler       pb.CrawlerServiceClient
	LookbackDays  int
	FullHistory   bool
}

func (e *SyncExecutor) Run(ctx context.Context, task data.DataSyncTask) error {
	if task.TaskType != data.TaskTypeDataSync {
		return fmt.Errorf("unsupported task type: %s", task.TaskType)
	}
	if e.Crawler == nil {
		return fmt.Errorf("crawler client not configured")
	}

	lookback := e.LookbackDays
	if lookback <= 0 {
		lookback = 7
	}

	req := &pb.FetchDailyBarsRequest{
		StockCodes: []string{task.TargetObject},
	}
	if e.FullHistory {
		req.StartDate = ""
		req.EndDate = ""
	} else {
		start := time.Now().UTC().AddDate(0, 0, -(lookback + 5))
		req.StartDate = start.Format("2006-01-02")
		req.EndDate = time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	}

	stream, err := e.Crawler.FetchDailyBars(ctx, req)
	if err != nil {
		_ = e.Store.MarkStockSyncFailed(ctx, task.TargetObject)
		return fmt.Errorf("fetch daily bars: %w", err)
	}

	var lastVersion string
	got := false
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = e.Store.MarkStockSyncFailed(ctx, task.TargetObject)
			return fmt.Errorf("crawler stream: %w", err)
		}
		got = true
		if msg.GetError() != "" {
			_ = e.Store.MarkStockSyncFailed(ctx, task.TargetObject)
			return fmt.Errorf("crawler %s: %s", msg.GetStockCode(), msg.GetError())
		}

		bars := make([]data.DailyBar, 0, len(msg.GetBars()))
		for _, b := range msg.GetBars() {
			td, err := time.Parse("2006-01-02", b.GetTradeDate())
			if err != nil {
				continue
			}
			bars = append(bars, data.DailyBar{
				StockCode: b.GetStockCode(),
				Market:    data.Market(b.GetMarket()),
				TradeDate: td,
				Open:      b.GetOpen(),
				High:      b.GetHigh(),
				Low:       b.GetLow(),
				Close:     b.GetClose(),
				Volume:    b.GetVolume(),
				Source:    b.GetSource(),
			})
		}

		completeness := 0.0
		if len(bars) > 0 {
			completeness = min(100.0, float64(len(bars))/500.0*100)
		}
		version := fmt.Sprintf("v%s-%s", time.Now().UTC().Format("20060102T150405"), strings.ToLower(msg.GetStockCode()))
		res, err := e.Store.IngestDailyBars(ctx, msg.GetStockCode(), data.Market(firstMarket(bars)), msg.GetStockCode(), bars, nil, completeness, version)
		if err != nil {
			_ = e.Store.MarkStockSyncFailed(ctx, task.TargetObject)
			return err
		}
		lastVersion = res.DataVersion
	}

	if !got {
		_ = e.Store.MarkStockSyncFailed(ctx, task.TargetObject)
		return fmt.Errorf("crawler returned no messages for %s", task.TargetObject)
	}

	v := lastVersion
	return e.Store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusSuccess, nil, &v)
}

func firstMarket(bars []data.DailyBar) data.Market {
	if len(bars) == 0 {
		return ""
	}
	return bars[0].Market
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
