package biz

import (
	"context"

	"github.com/gsxhnd/guanlan/internal/data"
)

// AddItem adds a watchlist entry, ensures pool membership, and may enqueue sync.
func (w *Watchlist) AddItem(ctx context.Context, item data.WatchlistItem) (data.WatchlistItem, error) {
	out, err := w.Store.AddWatchlistItem(ctx, item)
	if err != nil {
		return data.WatchlistItem{}, err
	}
	_ = w.Store.EnsureStockInPool(ctx, out.StockCode, out.Market, out.StockCode)
	if w.Tasks != nil {
		_, _, _ = w.Tasks.EnsureSyncIfNeeded(ctx, out.StockCode, data.TriggerManual)
	}
	return out, nil
}
