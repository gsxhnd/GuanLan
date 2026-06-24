package data

import (
	"context"
	"testing"
	"time"
)

func TestStoreMigrateAndDataPageQueries(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("open memory store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	prev := now.AddDate(0, 0, -1)

	if err := store.UpsertIndexDataset(ctx, IndexDataset{
		IndexCode:        "000905.SH",
		Market:           MarketA,
		IndexName:        "中证500",
		DataCompleteness: 98.6,
		LastSyncTime:     &now,
		SyncStatus:       "ready",
	}); err != nil {
		t.Fatalf("upsert index: %v", err)
	}

	if err := store.UpsertStockDataStatus(ctx, StockDataStatus{
		StockCode:    "600519.SH",
		StockName:    "贵州茅台",
		Market:       MarketA,
		Completeness: 99.2,
		LastUpdate:   &now,
		SyncStatus:   StockStatusReady,
	}); err != nil {
		t.Fatalf("upsert stock status: %v", err)
	}

	for _, bar := range []DailyBar{
		{StockCode: "600519.SH", Market: MarketA, TradeDate: prev, Open: 1680, High: 1690, Low: 1670, Close: 1688, Volume: 2_000_000, Source: "qlib", DataVersion: "v1"},
		{StockCode: "600519.SH", Market: MarketA, TradeDate: now, Open: 1688, High: 1712.5, Low: 1675, Close: 1705.2, Volume: 2_840_000, Source: "qlib", DataVersion: "v1"},
	} {
		if err := store.UpsertDailyBar(ctx, bar); err != nil {
			t.Fatalf("upsert daily bar: %v", err)
		}
	}

	indexes, err := store.ListIndexDatasets(ctx)
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	if len(indexes) != 1 || indexes[0].IndexCode != "000905.SH" {
		t.Fatalf("unexpected indexes: %+v", indexes)
	}

	market := MarketA
	stocks, err := store.ListStocks(ctx, ListStocksFilter{Market: &market, Sort: "code"})
	if err != nil {
		t.Fatalf("list stocks: %v", err)
	}
	if len(stocks) != 1 {
		t.Fatalf("expected 1 stock, got %d", len(stocks))
	}
	if stocks[0].Close != 1705.2 {
		t.Fatalf("unexpected close: %v", stocks[0].Close)
	}
	if stocks[0].Change <= 0 {
		t.Fatalf("expected positive change, got %v", stocks[0].Change)
	}

	bars, err := store.ListDailyBars(ctx, ListDailyBarsParams{StockCode: "600519.SH"})
	if err != nil {
		t.Fatalf("list daily bars: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}

	task, err := store.CreateStockSyncTask(ctx, "600519.SH", TriggerManual)
	if err != nil {
		t.Fatalf("create sync task: %v", err)
	}
	if task.Status != TaskStatusPending {
		t.Fatalf("unexpected task status: %s", task.Status)
	}

	tasks, err := store.ListDataSyncTasks(ctx, 10)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}
