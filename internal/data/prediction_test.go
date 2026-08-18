package data

import (
	"context"
	"testing"
	"time"
)

func TestInsertAndListPredictions(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	day := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	if err := store.InsertPredictions(ctx, []Prediction{
		{StockCode: "AAPL", TradeDate: day, Score: 0.4, ModelVersion: "baseline"},
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := store.InsertPredictions(ctx, []Prediction{
		{StockCode: "AAPL", TradeDate: day, Score: 0.7, ModelVersion: "baseline"},
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	rows, err := store.ListPredictions(ctx, "AAPL", "2026-08-17", "baseline", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].Score != 0.7 {
		t.Fatalf("score=%v", rows[0].Score)
	}
}
