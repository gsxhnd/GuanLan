package data

import (
	"context"
	"testing"
)

func TestStockPoolUpsertAndList(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	if err := store.UpsertStockPoolEntry(ctx, StockPoolEntry{
		YfinanceSymbol: "600519.SS",
		OriginalCode:   "600519",
		Market:         MarketA,
		StockName:      "贵州茅台",
		Exchange:       "SH",
		Currency:       "CNY",
		Source:         PoolSourceAPIManual,
		IsActive:       true,
		SyncDaily:      true,
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	items, err := store.ListStockPool(ctx, StockPoolListOptions{DailySyncOnly: true})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || items[0].YfinanceSymbol != "600519.SS" {
		t.Fatalf("unexpected pool: %+v", items)
	}

	if err := store.EnsureStockInPool(ctx, "AAPL", MarketUS, "Apple"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	items, _ = store.ListStockPool(ctx, StockPoolListOptions{})
	if len(items) != 2 {
		t.Fatalf("expected 2 in pool, got %d", len(items))
	}
}

func TestParsePoolSymbol(t *testing.T) {
	entry, err := ParsePoolSymbol("000001.SZ", MarketA)
	if err != nil {
		t.Fatal(err)
	}
	if entry.YfinanceSymbol != "000001.SZ" || entry.OriginalCode != "000001" || entry.Exchange != "SZ" {
		t.Fatalf("unexpected: %+v", entry)
	}

	entry, err = ParsePoolSymbol("600519.SH", MarketA)
	if err != nil {
		t.Fatal(err)
	}
	if entry.YfinanceSymbol != "600519.SS" || entry.Exchange != "SH" {
		t.Fatalf("unexpected: %+v", entry)
	}
}

func TestStockPoolPreservesAPIManualSource(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	if err := store.UpsertStockPoolEntry(ctx, StockPoolEntry{
		YfinanceSymbol: "000001.SZ",
		OriginalCode:   "000001",
		Market:         MarketA,
		Exchange:       "SZ",
		StockName:      "平安银行",
		Currency:       "CNY",
		Source:         PoolSourceAPIManual,
		IsActive:       true,
		SyncDaily:      true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertStockPoolEntry(ctx, StockPoolEntry{
		YfinanceSymbol: "000001.SZ",
		OriginalCode:   "000001",
		Market:         MarketA,
		Exchange:       "SZ",
		StockName:      "平安银行",
		Currency:       "CNY",
		Source:         PoolSourceCSVImport,
		IsActive:       true,
		SyncDaily:      true,
	}); err != nil {
		t.Fatal(err)
	}
	item, ok, err := store.GetStockPoolEntry(ctx, "000001.SZ")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if item.Source != PoolSourceAPIManual {
		t.Fatalf("expected api_manual, got %s", item.Source)
	}
}
