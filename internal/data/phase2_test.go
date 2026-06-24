package data

import (
	"context"
	"testing"
	"time"
)

func TestPhase2TaskWatchlistPortfolio(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("open memory store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()

	item, err := store.AddWatchlistItem(ctx, WatchlistItem{
		StockCode: "600519.SH",
		Market:    MarketA,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("add watchlist: %v", err)
	}
	if item.StockCode != "600519.SH" {
		t.Fatalf("unexpected item: %+v", item)
	}

	taskRec, err := store.CreateStockSyncTask(ctx, "600519.SH", TriggerManual)
	if err != nil {
		t.Fatalf("create sync task: %v", err)
	}
	if taskRec.Status != TaskStatusPending {
		t.Fatalf("unexpected task status: %s", taskRec.Status)
	}
	if err := store.UpdateTaskStatus(ctx, taskRec.TaskID, TaskStatusSuccess, nil, strPtr("v20260624")); err != nil {
		t.Fatalf("update task: %v", err)
	}
	final, err := store.GetTask(ctx, taskRec.TaskID)
	if err != nil {
		t.Fatalf("get final task: %v", err)
	}
	if final.Status != TaskStatusSuccess {
		t.Fatalf("expected success, got %s", final.Status)
	}
	if final.DataVersion == nil || *final.DataVersion == "" {
		t.Fatal("expected data_version on success task")
	}

	tradeDate := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)
	if _, err := store.CreateTrade(ctx, PortfolioTrade{
		TradeDate: tradeDate,
		StockCode: "600519.SH",
		StockName: "贵州茅台",
		Side:      TradeSideBuy,
		Price:     1680,
		Quantity:  100,
		TotalFee:  50,
	}); err != nil {
		t.Fatalf("create trade: %v", err)
	}

	if _, err := store.CreateDividend(ctx, PortfolioDividend{
		DividendDate:  time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		StockCode:     "600519.SH",
		TotalDividend: 2767,
	}); err != nil {
		t.Fatalf("create dividend: %v", err)
	}

	if _, err := store.CreateCashFlow(ctx, PortfolioCashFlow{
		FlowDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Amount:   200000,
		FlowType: CashFlowDeposit,
	}); err != nil {
		t.Fatalf("create cash flow: %v", err)
	}

	positions, cash, err := store.ComputePositions(ctx)
	if err != nil {
		t.Fatalf("compute positions: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	if positions[0].Quantity != 100 {
		t.Fatalf("unexpected qty: %v", positions[0].Quantity)
	}
	expectedCash := 200000.0 - (1680*100 + 50) + 2767.0
	if cash != expectedCash {
		t.Fatalf("cash: got %v want %v", cash, expectedCash)
	}

	_, _, total, _, err := store.GetAssets(ctx, nil, nil)
	if err != nil {
		t.Fatalf("get assets: %v", err)
	}
	if total <= 0 {
		t.Fatalf("expected positive total asset, got %v", total)
	}

	disabled, err := store.RemoveWatchlistItem(ctx, "600519.SH", false)
	if err != nil {
		t.Fatalf("remove watchlist: %v", err)
	}
	if disabled.IsActive {
		t.Fatal("expected disabled watchlist item")
	}
}

func strPtr(s string) *string { return &s }

func TestRetryFailedTask(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	taskRec, err := store.CreateTask(ctx, TaskTypeDataSync, "FAILME", TriggerManual, 0)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	reason := "simulated failure"
	if err := store.UpdateTaskStatus(ctx, taskRec.TaskID, TaskStatusFailed, &reason, nil); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	retry, err := store.RetryTask(ctx, taskRec.TaskID)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if retry.TriggerMethod != TriggerRetry {
		t.Fatalf("expected retry trigger, got %s", retry.TriggerMethod)
	}
	if retry.RetryCount != 1 {
		t.Fatalf("expected retry_count=1, got %d", retry.RetryCount)
	}
}
