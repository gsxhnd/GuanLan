package data

import "context"

// Repository 数据页底层读写接口。
type Repository interface {
	ListStockPool(ctx context.Context, opts StockPoolListOptions) ([]StockPoolEntry, error)
	ListStocks(ctx context.Context, filter ListStocksFilter) ([]StockListItem, error)
	ListDailyBars(ctx context.Context, params ListDailyBarsParams) ([]DailyBar, error)
	CreateTask(ctx context.Context, taskType TaskType, target string, trigger TriggerMethod, retryCount int) (DataSyncTask, error)
	ListDataSyncTasks(ctx context.Context, limit int) ([]DataSyncTask, error)
}

// 编译期断言 Store 实现 Repository。
var _ Repository = (*Store)(nil)
