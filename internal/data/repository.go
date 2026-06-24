package data

import "context"

// Repository 数据页底层读写接口，对应 docs/dev/02-architecture.md §5 数据相关端点。
type Repository interface {
	ListIndexDatasets(ctx context.Context) ([]IndexDataset, error)
	ListIndexConstituents(ctx context.Context, indexCode string) ([]IndexConstituent, error)
	ListStocks(ctx context.Context, filter ListStocksFilter) ([]StockListItem, error)
	ListDailyBars(ctx context.Context, params ListDailyBarsParams) ([]DailyBar, error)
	CreateStockSyncTask(ctx context.Context, stockCode string, trigger TriggerMethod) (DataSyncTask, error)
	ListDataSyncTasks(ctx context.Context, limit int) ([]DataSyncTask, error)
}

// 编译期断言 Store 实现 Repository。
var _ Repository = (*Store)(nil)
